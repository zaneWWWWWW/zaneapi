package controller

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hupijiaoRoundTripper func(*http.Request) (*http.Response, error)

func (f hupijiaoRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func signedHupijiaoResponse(t *testing.T, secret string, errCode int, fields map[string]string) string {
	t.Helper()

	signParams := make(map[string]string, len(fields)+2)
	payload := make(map[string]any, len(fields)+3)
	for key, value := range fields {
		signParams[key] = value
		payload[key] = value
	}
	signParams["errcode"] = strconv.Itoa(errCode)
	payload["errcode"] = errCode
	payload["hash"] = hupijiaoSign(signParams, secret)

	body, err := common.Marshal(payload)
	require.NoError(t, err)
	return string(body)
}

func configureHupijiaoForControllerTest(t *testing.T) {
	t.Helper()
	confirmPaymentComplianceForTest(t)

	originalEnabled := setting.HupijiaoEnabled
	originalEndpoint := setting.HupijiaoEndpoint
	originalDisplayName := setting.HupijiaoDisplayName
	originalIcon := setting.HupijiaoIcon
	originalAppID := setting.HupijiaoAppID
	originalAppSecret := setting.HupijiaoAppSecret
	originalCallbackAddress := operation_setting.CustomCallbackAddress
	originalServerAddress := system_setting.ServerAddress
	t.Cleanup(func() {
		setting.HupijiaoEnabled = originalEnabled
		setting.HupijiaoEndpoint = originalEndpoint
		setting.HupijiaoDisplayName = originalDisplayName
		setting.HupijiaoIcon = originalIcon
		setting.HupijiaoAppID = originalAppID
		setting.HupijiaoAppSecret = originalAppSecret
		operation_setting.CustomCallbackAddress = originalCallbackAddress
		system_setting.ServerAddress = originalServerAddress
	})

	setting.HupijiaoEnabled = true
	setting.HupijiaoEndpoint = "https://api.xunhupay.com/payment/do.html"
	setting.HupijiaoDisplayName = "支付宝扫码"
	setting.HupijiaoIcon = "SiAlipay"
	setting.HupijiaoAppID = "app-123"
	setting.HupijiaoAppSecret = "test-secret"
	operation_setting.CustomCallbackAddress = "https://merchant.example.com"
	system_setting.ServerAddress = "https://merchant.example.com"
}

func performHupijiaoNotify(t *testing.T, params map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/user/hupijiao/notify", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	HupijiaoNotify(ctx)
	return recorder
}

func TestHupijiaoSignSortsAndFiltersParameters(t *testing.T) {
	actual := hupijiaoSign(map[string]string{
		"total_fee": "9.90",
		"time":      "1700000000",
		"nonce_str": "abc",
		"appid":     "app",
		"empty":     "",
		"hash":      "ignored",
	}, "secret")

	assert.Equal(t, "6998ecf9a765fd777eae86a8a8d73f4d", actual)
}

func TestGenerateHupijiaoOrderIDConformsToProtocol(t *testing.T) {
	tradeNo, err := generateHupijiaoOrderID()
	require.NoError(t, err)
	assert.LessOrEqual(t, len(tradeNo), 32)
	assert.True(t, regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(tradeNo))
}

func TestCreateHupijiaoPaymentAcceptsSignedFlatResponse(t *testing.T) {
	const secret = "test-secret"
	originalClient := hupijiaoHTTPClient
	hupijiaoHTTPClient = &http.Client{Transport: hupijiaoRoundTripper(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, "https://api.xunhupay.com/payment/do.html", request.URL.String())
		require.Equal(t, "application/json;charset=UTF-8", request.Header.Get("Content-Type"))

		var requestParams map[string]string
		require.NoError(t, common.DecodeJson(request.Body, &requestParams))
		assert.Equal(t, hupijiaoAPIVersion, requestParams["version"])
		assert.Equal(t, "app-123", requestParams["appid"])
		assert.Equal(t, "order-123", requestParams["trade_order_id"])
		assert.Equal(t, "9.90", requestParams["total_fee"])
		assert.Equal(t, "32-char-nonce", requestParams["nonce_str"])
		assert.True(t, verifyHupijiaoSign(requestParams, secret))

		responseBody := signedHupijiaoResponse(t, secret, 0, map[string]string{
			"openid": "2019081202",
			"url":    "https://api.xunhupay.com/alipay/pay/index.html?id=123",
			"errmsg": "success!",
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     make(http.Header),
		}, nil
	})}
	t.Cleanup(func() {
		hupijiaoHTTPClient = originalClient
	})

	paymentURL, err := createHupijiaoPayment(
		context.Background(),
		"https://api.xunhupay.com/payment/do.html",
		secret,
		map[string]string{
			"version":        hupijiaoAPIVersion,
			"appid":          "app-123",
			"trade_order_id": "order-123",
			"total_fee":      "9.90",
			"title":          "Balance topup",
			"notify_url":     "https://merchant.example.com/api/user/hupijiao/notify",
			"time":           "1700000000",
			"nonce_str":      "32-char-nonce",
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "https://api.xunhupay.com/alipay/pay/index.html?id=123", paymentURL)
}

func TestCreateHupijiaoPaymentAcceptsSuccessfulResponseWithUnverifiableSignature(t *testing.T) {
	originalClient := hupijiaoHTTPClient
	hupijiaoHTTPClient = &http.Client{Transport: hupijiaoRoundTripper(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"url":"https://api.xunhupay.com/pay/123","errcode":0,"errmsg":"success","hash":"00000000000000000000000000000000"}`,
			)),
			Header: make(http.Header),
		}, nil
	})}
	t.Cleanup(func() {
		hupijiaoHTTPClient = originalClient
	})

	paymentURL, err := createHupijiaoPayment(
		context.Background(),
		"https://api.xunhupay.com/payment/do.html",
		"test-secret",
		map[string]string{"appid": "app-123"},
	)

	require.NoError(t, err)
	assert.Equal(t, "https://api.xunhupay.com/pay/123", paymentURL)
}

func TestCreateHupijiaoPaymentRejectsSignedGatewayBusinessError(t *testing.T) {
	const secret = "test-secret"
	originalClient := hupijiaoHTTPClient
	hupijiaoHTTPClient = &http.Client{Transport: hupijiaoRoundTripper(func(_ *http.Request) (*http.Response, error) {
		responseBody := signedHupijiaoResponse(t, secret, 500, map[string]string{
			"errmsg": "signature rejected",
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     make(http.Header),
		}, nil
	})}
	t.Cleanup(func() {
		hupijiaoHTTPClient = originalClient
	})

	_, err := createHupijiaoPayment(
		context.Background(),
		"https://api.xunhupay.com/payment/do.html",
		secret,
		map[string]string{"appid": "app-123"},
	)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errHupijiaoRequestRejected))
}

func TestIsPublicHupijiaoIPRejectsPrivateAndSharedAddresses(t *testing.T) {
	testCases := []struct {
		address string
		public  bool
	}{
		{address: "8.8.8.8", public: true},
		{address: "1.1.1.1", public: true},
		{address: "127.0.0.1", public: false},
		{address: "10.0.0.1", public: false},
		{address: "169.254.169.254", public: false},
		{address: "100.64.0.1", public: false},
		{address: "::1", public: false},
		{address: "fd00::1", public: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.address, func(t *testing.T) {
			assert.Equal(t, testCase.public, isPublicHupijiaoIP(net.ParseIP(testCase.address)))
		})
	}
}

func TestGetTopUpInfoExposesHupijiaoWithoutEpay(t *testing.T) {
	configureHupijiaoForControllerTest(t)

	originalPayAddress := operation_setting.PayAddress
	originalEpayID := operation_setting.EpayId
	originalEpayKey := operation_setting.EpayKey
	originalPayMethods := operation_setting.PayMethods
	t.Cleanup(func() {
		operation_setting.PayAddress = originalPayAddress
		operation_setting.EpayId = originalEpayID
		operation_setting.EpayKey = originalEpayKey
		operation_setting.PayMethods = originalPayMethods
	})

	operation_setting.PayAddress = ""
	operation_setting.EpayId = ""
	operation_setting.EpayKey = ""
	operation_setting.PayMethods = []map[string]string{
		{"name": "旧的虎皮椒配置", "type": model.PaymentMethodHupijiao},
		{"name": "支付宝", "type": "alipay"},
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/user/topup/info", nil)
	GetTopUpInfo(ctx)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			EnableOnlineTopUp   bool                `json:"enable_online_topup"`
			EnableHupijiaoTopUp bool                `json:"enable_hupijiao_topup"`
			PayMethods          []map[string]string `json:"pay_methods"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	assert.False(t, response.Data.EnableOnlineTopUp)
	assert.True(t, response.Data.EnableHupijiaoTopUp)
	require.Len(t, response.Data.PayMethods, 1)
	assert.Equal(t, model.PaymentMethodHupijiao, response.Data.PayMethods[0]["type"])
	assert.Equal(t, "支付宝扫码", response.Data.PayMethods[0]["name"])
	assert.Equal(t, "SiAlipay", response.Data.PayMethods[0]["icon"])
}

func TestHupijiaoNotifyCreditsOnceAcrossRepeatedCallbacks(t *testing.T) {
	configureHupijiaoForControllerTest(t)
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.TopUp{}, &model.Log{}))

	user := &model.User{
		Id:       701,
		Username: "hupijiao_notify_user",
		Status:   common.UserStatusEnabled,
	}
	require.NoError(t, db.Create(user).Error)
	topUp := &model.TopUp{
		UserId:          user.Id,
		Amount:          2,
		Money:           9.99,
		TradeNo:         "HPJNotifyOrder123",
		PaymentMethod:   model.PaymentMethodHupijiao,
		PaymentProvider: model.PaymentProviderHupijiao,
		CreateTime:      common.GetTimestamp(),
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, topUp.Insert())

	params := map[string]string{
		"trade_order_id": topUp.TradeNo,
		"total_fee":      "9.99",
		"transaction_id": "transaction-123",
		"open_order_id":  "open-order-123",
		"order_title":    "额度充值",
		"status":         hupijiaoPaidStatus,
		"appid":          setting.HupijiaoAppID,
		"time":           "1700000000",
		"nonce_str":      "callback-nonce",
	}
	params["hash"] = hupijiaoSign(params, setting.HupijiaoAppSecret)

	first := performHupijiaoNotify(t, params)
	assert.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, hupijiaoNotifySuccess, first.Body.String())
	second := performHupijiaoNotify(t, params)
	assert.Equal(t, http.StatusOK, second.Code)
	assert.Equal(t, hupijiaoNotifySuccess, second.Body.String())

	var updatedUser model.User
	require.NoError(t, db.Select("quota").Where("id = ?", user.Id).First(&updatedUser).Error)
	assert.Equal(t, int(2*common.QuotaPerUnit), updatedUser.Quota)
	updatedTopUp := model.GetTopUpByTradeNo(topUp.TradeNo)
	require.NotNil(t, updatedTopUp)
	assert.Equal(t, common.TopUpStatusSuccess, updatedTopUp.Status)
}

func TestHupijiaoNotifyAcknowledgesSignedNonPaidStatus(t *testing.T) {
	configureHupijiaoForControllerTest(t)
	params := map[string]string{
		"trade_order_id": "HPJRefundOrder123",
		"total_fee":      "9.99",
		"status":         "CD",
		"appid":          setting.HupijiaoAppID,
		"time":           "1700000000",
		"nonce_str":      "callback-nonce",
	}
	params["hash"] = hupijiaoSign(params, setting.HupijiaoAppSecret)

	response := performHupijiaoNotify(t, params)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, hupijiaoNotifySuccess, response.Body.String())
}

func TestHupijiaoNotifyRejectsInvalidSignature(t *testing.T) {
	configureHupijiaoForControllerTest(t)
	params := map[string]string{
		"trade_order_id": "HPJInvalidSignature123",
		"total_fee":      "9.99",
		"status":         hupijiaoPaidStatus,
		"appid":          setting.HupijiaoAppID,
		"time":           "1700000000",
		"nonce_str":      "callback-nonce",
		"hash":           "invalid",
	}

	response := performHupijiaoNotify(t, params)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "fail", response.Body.String())
}
