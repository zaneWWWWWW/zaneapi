package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

const playgroundImageFetchMaxBytes = 15 << 20

func setupPlaygroundUserToken(c *gin.Context, format types.RelayFormat) (*relaycommon.RelayInfo, *types.NewAPIError) {
	useAccessToken := c.GetBool("use_access_token")
	if useAccessToken {
		return nil, types.NewError(errors.New("暂不支持使用 access token"), types.ErrorCodeAccessDenied, types.ErrOptionWithSkipRetry())
	}

	relayInfo, err := relaycommon.GenRelayInfo(c, format, nil, nil)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeInvalidRequest, types.ErrOptionWithSkipRetry())
	}

	userId := c.GetInt("id")
	userCache, err := model.GetUserCache(userId)
	if err != nil {
		return nil, types.NewError(err, types.ErrorCodeQueryDataError, types.ErrOptionWithSkipRetry())
	}
	userCache.WriteContext(c)

	tempToken := &model.Token{
		UserId: userId,
		Name:   fmt.Sprintf("playground-%s", relayInfo.UsingGroup),
		Group:  relayInfo.UsingGroup,
	}
	_ = middleware.SetupContextForToken(c, tempToken)
	return relayInfo, nil
}

func PlaygroundImageFile(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("url"))
	parsed, err := url.Parse(raw)
	if err != nil || raw == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid url"})
		return
	}

	resp, err := service.DoDownloadRequest(raw, "playground studio reference")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "failed to fetch image"})
		return
	}
	defer service.CloseResponseBodyGracefully(resp)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "failed to fetch image"})
		return
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, playgroundImageFetchMaxBytes+1))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "failed to read image"})
		return
	}
	if len(data) > playgroundImageFetchMaxBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"success": false, "message": "image too large"})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		contentType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "url is not an image"})
		return
	}
	c.Data(http.StatusOK, contentType, data)
}

func Playground(c *gin.Context) {
	var newAPIError *types.NewAPIError

	defer func() {
		if newAPIError != nil {
			c.JSON(newAPIError.StatusCode, gin.H{
				"error": newAPIError.ToOpenAIError(),
			})
		}
	}()

	_, newAPIError = setupPlaygroundUserToken(c, types.RelayFormatOpenAI)
	if newAPIError != nil {
		return
	}

	Relay(c, types.RelayFormatOpenAI)
}

func PlaygroundImage(c *gin.Context) {
	var newAPIError *types.NewAPIError

	defer func() {
		if newAPIError != nil {
			c.JSON(newAPIError.StatusCode, gin.H{
				"error": newAPIError.ToOpenAIError(),
			})
		}
	}()

	_, newAPIError = setupPlaygroundUserToken(c, types.RelayFormatOpenAIImage)
	if newAPIError != nil {
		return
	}

	Relay(c, types.RelayFormatOpenAIImage)
}

func PlaygroundVideo(c *gin.Context) {
	_, newAPIError := setupPlaygroundUserToken(c, types.RelayFormatTask)
	if newAPIError != nil {
		c.JSON(newAPIError.StatusCode, gin.H{
			"error": newAPIError.ToOpenAIError(),
		})
		return
	}

	RelayTask(c)
}

func PlaygroundVideoFetch(c *gin.Context) {
	_, newAPIError := setupPlaygroundUserToken(c, types.RelayFormatTask)
	if newAPIError != nil {
		c.JSON(newAPIError.StatusCode, gin.H{
			"error": newAPIError.ToOpenAIError(),
		})
		return
	}

	RelayTaskFetch(c)
}
