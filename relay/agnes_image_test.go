package relay

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAgnesImageHelperPreservesExtensionsAndMapping(t *testing.T) {
	service.InitHttpClient()
	var captured []byte
	var path string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		path = r.URL.Path
		// Stop before billing: this test exercises parsing, deep-copy, model
		// mapping, conversion and actual HTTP dispatch against a mock upstream.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write([]byte(`{"error":{"message":"fixture stop","type":"invalid_request_error"}}`))
	}))
	defer upstream.Close()
	body := `{"model":"public-image","prompt":"make it blue","size":"2K","ratio":"16:9","return_base64":false,"response_format":"b64_json","extra_body":{"image":["https://example.com/source.png"]}}`
	var request dto.ImageRequest
	require.NoError(t, common.Unmarshal([]byte(body), &request))
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/images/edits", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	common.SetContextKey(c, constant.ContextKeyChannelType, constant.ChannelTypeAgnesAI)
	common.SetContextKey(c, constant.ContextKeyChannelBaseUrl, upstream.URL+"/v1")
	common.SetContextKey(c, constant.ContextKeyChannelKey, "test-key")
	c.Set("model_mapping", `{"public-image":"agnes-image-2.5-flash"}`)
	info := &relaycommon.RelayInfo{OriginModelName: "public-image", Request: &request, RequestURLPath: "/v1/images/edits", RelayMode: relayconstant.RelayModeImagesEdits, RelayFormat: types.RelayFormatOpenAI}
	apiErr := ImageHelper(c, info)
	require.NotNil(t, apiErr)
	require.NotEmpty(t, captured, "conversion failed before upstream: %v", apiErr)
	require.Equal(t, "/v1/images/generations", path)
	var got map[string]any
	require.NoError(t, common.Unmarshal(captured, &got))
	require.Equal(t, "agnes-image-2.5-flash", got["model"])
	require.Equal(t, "16:9", got["ratio"])
	require.Equal(t, false, got["return_base64"])
	extra := got["extra_body"].(map[string]any)
	require.Equal(t, []any{"https://example.com/source.png"}, extra["image"])
	require.Equal(t, "b64_json", extra["response_format"])
}

func TestAgnesImageHelperRejectsInvalidInputLocally(t *testing.T) {
	for _, body := range []string{
		`{"model":"agnes-image-2.5-flash","prompt":"test"}`,
		`{"model":"agnes-image-2.5-flash","prompt":"test","image":"https://example.com/a.png","n":0}`,
		`{"model":"agnes-image-2.5-flash","prompt":"test","image":"https://example.com/a.png","n":2}`,
	} {
		c := agnesPipelineContext("/v1/images/edits", body, "http://127.0.0.1:1")
		var request dto.ImageRequest
		require.NoError(t, common.Unmarshal([]byte(body), &request))
		info := &relaycommon.RelayInfo{Request: &request, OriginModelName: request.Model, RequestURLPath: c.Request.URL.Path, RelayMode: relayconstant.RelayModeImagesEdits}
		apiErr := ImageHelper(c, info)
		require.NotNil(t, apiErr)
		require.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		require.Equal(t, types.ErrorCodeInvalidRequest, apiErr.GetErrorCode())
	}
}
