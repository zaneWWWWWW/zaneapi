package controller

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const (
	hupijiaoAPIVersion       = "1.1"
	hupijiaoPaymentPath      = "/payment/do.html"
	hupijiaoPaidStatus       = "OD"
	hupijiaoNotifySuccess    = "success"
	hupijiaoMaxResponseBytes = int64(1 << 20)
)

var (
	errHupijiaoRequestRejected = errors.New("hupijiao payment request rejected")
	hupijiaoHTTPClient         = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext:           dialHupijiaoPublicAddress,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   5,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

type HupijiaoPayRequest struct {
	Amount int64 `json:"amount"`
}

type hupijiaoPaymentResponse struct {
	ErrCode   int
	ErrMsg    string
	URL       string
	URLQrcode string
}

func isPublicHupijiaoIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsPrivate() {
		return false
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		// Shared address space (RFC 6598) is not public even though net.IP does
		// not classify it as private.
		return !(ipv4[0] == 100 && ipv4[1]&0xc0 == 64)
	}
	return true
}

func dialHupijiaoPublicAddress(ctx context.Context, network string, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("parse Hupijiao gateway address: %w", err)
	}

	resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve Hupijiao gateway: %w", err)
	}
	if len(resolved) == 0 {
		return nil, errors.New("Hupijiao gateway did not resolve to an address")
	}
	for _, candidate := range resolved {
		if !isPublicHupijiaoIP(candidate.IP) {
			return nil, errors.New("Hupijiao gateway resolved to a non-public address")
		}
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	var lastErr error
	for _, candidate := range resolved {
		connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
		if dialErr == nil {
			return connection, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("connect to Hupijiao gateway: %w", lastErr)
}

func hupijiaoSign(params map[string]string, secret string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if key != "hash" && value != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+params[key])
	}
	sum := md5.Sum([]byte(strings.Join(parts, "&") + secret))
	return hex.EncodeToString(sum[:])
}

func verifyHupijiaoSign(params map[string]string, secret string) bool {
	received := strings.ToLower(strings.TrimSpace(params["hash"]))
	if received == "" || strings.TrimSpace(secret) == "" {
		return false
	}
	expected := hupijiaoSign(params, secret)
	return subtle.ConstantTimeCompare([]byte(received), []byte(expected)) == 1
}

func generateHupijiaoOrderID() (string, error) {
	// The protocol limits trade_order_id to 32 characters. A 112-bit random
	// suffix keeps the identifier short while making collisions negligible.
	randomBytes := make([]byte, 14)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return "HPJ" + hex.EncodeToString(randomBytes), nil
}

func generateHupijiaoNonce() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

func normalizeHupijiaoTopUpAmount(amount int64) int64 {
	if operation_setting.GetQuotaDisplayType() != operation_setting.QuotaDisplayTypeTokens {
		return amount
	}

	normalized := decimal.NewFromInt(amount).
		Div(decimal.NewFromFloat(common.QuotaPerUnit)).
		IntPart()
	if normalized < 1 {
		return 1
	}
	return normalized
}

func getHupijiaoCallbackURL() (string, bool) {
	base := strings.TrimRight(strings.TrimSpace(service.GetCallbackAddress()), "/")
	if !isHupijiaoCallbackOriginValid(base) {
		return "", false
	}
	callbackURL := base + "/api/user/hupijiao/notify"
	return callbackURL, len(callbackURL) <= 128
}

func decodeHupijiaoPaymentResponse(body []byte) (hupijiaoPaymentResponse, error) {
	var rawFields map[string]json.RawMessage
	if err := common.Unmarshal(body, &rawFields); err != nil {
		return hupijiaoPaymentResponse{}, fmt.Errorf("decode Hupijiao payment response: %w", err)
	}
	if len(rawFields) == 0 {
		return hupijiaoPaymentResponse{}, errors.New("Hupijiao returned an empty response")
	}

	params := make(map[string]string, len(rawFields))
	for key, rawValue := range rawFields {
		valueType := common.GetJsonType(rawValue)
		if valueType == "object" || valueType == "array" {
			return hupijiaoPaymentResponse{}, fmt.Errorf("Hupijiao returned unsupported field %q", key)
		}
		params[key] = common.JsonRawMessageToString(rawValue)
	}
	responseHash, err := hex.DecodeString(strings.TrimSpace(params["hash"]))
	if err != nil || len(responseHash) != md5.Size {
		return hupijiaoPaymentResponse{}, errors.New("Hupijiao payment response has an invalid signature format")
	}
	errCode, err := strconv.Atoi(params["errcode"])
	if err != nil {
		return hupijiaoPaymentResponse{}, errors.New("Hupijiao payment response has an invalid error code")
	}
	return hupijiaoPaymentResponse{
		ErrCode:   errCode,
		ErrMsg:    params["errmsg"],
		URL:       params["url"],
		URLQrcode: params["url_qrcode"],
	}, nil
}

func createHupijiaoPayment(ctx context.Context, endpoint string, secret string, params map[string]string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if !isHupijiaoPaymentURLValid(endpoint) {
		return "", errors.New("Hupijiao payment URL is invalid")
	}

	requestParams := make(map[string]string, len(params)+1)
	for key, value := range params {
		requestParams[key] = value
	}
	requestParams["hash"] = hupijiaoSign(requestParams, secret)
	body, err := common.Marshal(requestParams)
	if err != nil {
		return "", fmt.Errorf("encode Hupijiao payment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create Hupijiao payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := hupijiaoHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send Hupijiao payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Hupijiao payment request returned HTTP %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, hupijiaoMaxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("read Hupijiao payment response: %w", err)
	}
	if int64(len(responseBody)) > hupijiaoMaxResponseBytes {
		return "", errors.New("Hupijiao payment response is too large")
	}

	// Some live Hupijiao gateways return a well-formed response hash that does
	// not verify with their documented algorithm. The request is protected by
	// HTTPS and the callback remains signature-verified before credit is added.
	result, err := decodeHupijiaoPaymentResponse(responseBody)
	if err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("%w: %s", errHupijiaoRequestRejected, strings.TrimSpace(result.ErrMsg))
	}

	paymentURL, err := url.Parse(strings.TrimSpace(result.URL))
	if err != nil || paymentURL.Scheme != "https" || paymentURL.Host == "" || paymentURL.User != nil {
		return "", errors.New("Hupijiao returned an invalid payment URL")
	}
	return paymentURL.String(), nil
}

func RequestHupijiaoPay(c *gin.Context) {
	if !isHupijiaoTopUpEnabled() {
		common.ApiErrorMsg(c, "虎皮椒支付未启用或配置不完整")
		return
	}

	var req HupijiaoPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if req.Amount < getMinTopup() {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", getMinTopup()))
		return
	}

	userID := c.GetInt("id")
	group, err := model.GetUserGroup(userID, true)
	if err != nil {
		common.ApiErrorMsg(c, "获取用户分组失败")
		return
	}

	payMoney := decimal.NewFromFloat(getPayMoney(req.Amount, group)).Round(2)
	if payMoney.LessThan(decimal.NewFromFloat(0.01)) {
		common.ApiErrorMsg(c, "充值金额过低")
		return
	}

	amount := normalizeHupijiaoTopUpAmount(req.Amount)
	quotaToAdd, clamp := common.QuotaFromDecimalChecked(decimal.NewFromInt(amount).Mul(decimal.NewFromFloat(common.QuotaPerUnit)))
	if clamp != nil || quotaToAdd <= 0 {
		common.ApiErrorMsg(c, "充值数量超出允许范围")
		return
	}

	notifyURL, ok := getHupijiaoCallbackURL()
	if !ok {
		common.ApiErrorMsg(c, "虎皮椒回调地址必须是有效的 HTTPS 站点根地址")
		return
	}

	tradeNo, err := generateHupijiaoOrderID()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒生成订单号失败 user_id=%d error=%q", userID, err.Error()))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}
	nonce, err := generateHupijiaoNonce()
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒生成随机串失败 user_id=%d error=%q", userID, err.Error()))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	topUp := &model.TopUp{
		UserId:          userID,
		Amount:          amount,
		Money:           payMoney.InexactFloat64(),
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodHupijiao,
		PaymentProvider: model.PaymentProviderHupijiao,
		CreateTime:      common.GetTimestamp(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒创建充值订单失败 user_id=%d trade_no=%s error=%q", userID, tradeNo, err.Error()))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	paymentParams := map[string]string{
		"version":        hupijiaoAPIVersion,
		"appid":          strings.TrimSpace(setting.HupijiaoAppID),
		"trade_order_id": tradeNo,
		"total_fee":      payMoney.StringFixed(2),
		"title":          "额度充值",
		"notify_url":     notifyURL,
		"time":           strconv.FormatInt(common.GetTimestamp(), 10),
		"nonce_str":      nonce,
	}
	if isHupijiaoCallbackOriginValid(system_setting.ServerAddress) {
		returnURL := paymentReturnPath("/wallet?show_history=true")
		callbackURL := paymentReturnPath("/wallet")
		if len(returnURL) <= 128 && len(callbackURL) <= 128 {
			paymentParams["return_url"] = returnURL
			paymentParams["callback_url"] = callbackURL
		}
	}

	paymentURL, err := createHupijiaoPayment(
		c.Request.Context(),
		setting.HupijiaoEndpoint,
		setting.HupijiaoAppSecret,
		paymentParams,
	)
	if err != nil {
		if errors.Is(err, errHupijiaoRequestRejected) {
			_ = model.UpdatePendingTopUpStatus(tradeNo, model.PaymentProviderHupijiao, common.TopUpStatusFailed)
		}
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒拉起支付失败 user_id=%d trade_no=%s error=%q", userID, tradeNo, err.Error()))
		common.ApiErrorMsg(c, "拉起虎皮椒支付失败")
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒充值订单创建成功 user_id=%d trade_no=%s amount=%d money=%s", userID, tradeNo, req.Amount, payMoney.StringFixed(2)))
	common.ApiSuccess(c, gin.H{"url": paymentURL})
}

func HupijiaoNotify(c *gin.Context) {
	if c.Request.Method != http.MethodPost || !isHupijiaoWebhookEnabled() {
		c.String(http.StatusServiceUnavailable, "fail")
		return
	}
	if err := c.Request.ParseForm(); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒回调表单解析失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	params := make(map[string]string, len(c.Request.PostForm))
	for key, values := range c.Request.PostForm {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	if !verifyHupijiaoSign(params, setting.HupijiaoAppSecret) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒回调验签失败 client_ip=%s", c.ClientIP()))
		c.String(http.StatusBadRequest, "fail")
		return
	}
	if params["appid"] != strings.TrimSpace(setting.HupijiaoAppID) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒回调 AppID 不匹配 trade_no=%s client_ip=%s", params["trade_order_id"], c.ClientIP()))
		c.String(http.StatusBadRequest, "fail")
		return
	}
	if params["status"] != hupijiaoPaidStatus {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒回调忽略非支付状态 trade_no=%s status=%s", params["trade_order_id"], params["status"]))
		c.String(http.StatusOK, hupijiaoNotifySuccess)
		return
	}

	tradeNo := strings.TrimSpace(params["trade_order_id"])
	paidMoney, err := decimal.NewFromString(params["total_fee"])
	if tradeNo == "" || len(tradeNo) > 32 || err != nil || paidMoney.LessThanOrEqual(decimal.Zero) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("虎皮椒回调参数无效 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	if err := model.RechargeHupijiao(tradeNo, paidMoney, c.ClientIP()); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("虎皮椒充值处理失败 trade_no=%s client_ip=%s error=%q", tradeNo, c.ClientIP(), err.Error()))
		statusCode := http.StatusInternalServerError
		if errors.Is(err, model.ErrTopUpNotFound) ||
			errors.Is(err, model.ErrPaymentMethodMismatch) ||
			errors.Is(err, model.ErrPaymentAmountMismatch) ||
			errors.Is(err, model.ErrTopUpStatusInvalid) {
			statusCode = http.StatusBadRequest
		}
		c.String(statusCode, "fail")
		return
	}

	logger.LogInfo(c.Request.Context(), fmt.Sprintf("虎皮椒充值成功 trade_no=%s client_ip=%s", tradeNo, c.ClientIP()))
	c.String(http.StatusOK, hupijiaoNotifySuccess)
}
