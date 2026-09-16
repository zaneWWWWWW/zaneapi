package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayInvalidRequestReturnsBadRequestBeforeUpstream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for name, body := range map[string]string{
		"negative max_tokens":      `{"model":"gpt-5.6-sol","messages":[{"role":"user","content":"hi"}],"max_tokens":-1}`,
		"temperature out of range": `{"model":"gpt-5.6-sol","messages":[{"role":"user","content":"hi"}],"temperature":10}`,
	} {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set(common.RequestIdKey, "test-request-id")

			Relay(c, types.RelayFormatOpenAI)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
			assert.Contains(t, recorder.Body.String(), `"code":"invalid_request"`)
			assert.Contains(t, recorder.Body.String(), `"message":"invalid request (request id: test-request-id)"`)
			assert.NotContains(t, recorder.Body.String(), "GeneralOpenAIRequest")
		})
	}
}
