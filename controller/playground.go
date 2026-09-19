package controller

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

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
