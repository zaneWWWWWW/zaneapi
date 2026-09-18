package agnes

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestAgnesDispatchUsesNativeProtocol(t *testing.T) {
	service.InitHttpClient()
	for _, tc := range []struct {
		name, inputPath, wantPath string
		mode                      int
		format                    types.RelayFormat
	}{
		{"chat", "/v1/chat/completions", "/v1/chat/completions", relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI},
		{"responses", "/v1/responses", "/v1/responses", relayconstant.RelayModeResponses, types.RelayFormatOpenAI},
		{"messages", "/v1/messages", "/v1/messages", relayconstant.RelayModeChatCompletions, types.RelayFormatClaude},
		{"edits", "/v1/images/edits", "/v1/images/generations", relayconstant.RelayModeImagesEdits, types.RelayFormatOpenAI},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			var gotHeader http.Header
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotHeader = r.Header.Clone()
				w.Write([]byte(`{}`))
			}))
			defer server.Close()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", tc.inputPath, bytes.NewBufferString(`{}`))
			c.Request.Header.Set("Content-Type", "application/json")
			info := &relaycommon.RelayInfo{RelayMode: tc.mode, RelayFormat: tc.format, RequestURLPath: tc.inputPath, ChannelMeta: &relaycommon.ChannelMeta{ChannelType: constant.ChannelTypeAgnesAI, ChannelBaseUrl: server.URL + "/v1/", ApiKey: "test-key"}}
			response, err := (&Adaptor{}).DoRequest(c, info, bytes.NewBufferString(`{}`))
			require.NoError(t, err)
			response.(*http.Response).Body.Close()
			require.Equal(t, tc.wantPath, gotPath)
			require.Equal(t, "application/json", gotHeader.Get("Content-Type"))
			if tc.format == types.RelayFormatClaude {
				require.Equal(t, "test-key", gotHeader.Get("x-api-key"))
				require.Equal(t, "2023-06-01", gotHeader.Get("anthropic-version"))
			} else {
				require.Equal(t, "Bearer test-key", gotHeader.Get("Authorization"))
			}
		})
	}
}

func TestAgnesTextPreservesExplicitValues(t *testing.T) {
	var request dto.GeneralOpenAIRequest
	require.NoError(t, common.Unmarshal([]byte(`{"model":"o-custom-alias","temperature":0,"top_p":0,"stream":false,"chat_template_kwargs":{"enable_thinking":false},"tools":[{"type":"function","function":{"name":"test","parameters":{"type":"object"}}} ]}`), &request))
	converted, err := (&Adaptor{}).ConvertOpenAIRequest(nil, nil, &request)
	require.NoError(t, err)
	data, err := common.Marshal(converted)
	require.NoError(t, err)
	require.Contains(t, string(data), `"temperature":0`)
	require.Contains(t, string(data), `"top_p":0`)
	require.Contains(t, string(data), `"stream":false`)
	require.Contains(t, string(data), `"enable_thinking":false`)
	require.Contains(t, string(data), `"tools"`)
	var claudeRequest dto.ClaudeRequest
	require.NoError(t, common.Unmarshal([]byte(`{"model":"agnes-2.5-flash","max_tokens":2048,"temperature":0,"thinking":{"type":"enabled","budget_tokens":1024},"messages":[{"role":"user","content":"hi"}]}`), &claudeRequest))
	converted, err = (&Adaptor{}).ConvertClaudeRequest(nil, nil, &claudeRequest)
	require.NoError(t, err)
	data, err = common.Marshal(converted)
	require.NoError(t, err)
	require.Contains(t, string(data), `"thinking"`)
	require.Contains(t, string(data), `"temperature":0`)
}

func TestAgnesImageExtensionsAndValidation(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		invalid    bool
	}{
		{"zero n", `{"n":0}`, true},
		{"batch", `{"n":2}`, true},
		{"invalid ratio", `{"ratio":"auto"}`, true},
		{"invalid boolean", `{"return_base64":"false"}`, true},
		{"invalid input", `{"image":42}`, true},
		{"invalid URL", `{"image":"file:///tmp/input.png"}`, true},
		{"valid", `{"n":1,"ratio":"16:9","return_base64":false,"image":"data:image/png;base64,aGVsbG8=","extra_body":{"image":["https://example.com/ignored.png"],"response_format":"b64_json"}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var request dto.ImageRequest
			require.NoError(t, common.Unmarshal([]byte(tc.body), &request))
			info := &relaycommon.RelayInfo{RelayMode: relayconstant.RelayModeImagesGenerations}
			info.PriceData.AddOtherRatio("size", 4)
			converted, err := (&Adaptor{}).ConvertImageRequest(nil, info, request)
			if tc.invalid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			data, err := common.Marshal(converted)
			require.NoError(t, err)
			var got map[string]any
			require.NoError(t, common.Unmarshal(data, &got))
			require.Equal(t, "16:9", got["ratio"])
			require.Equal(t, false, got["return_base64"])
			require.Equal(t, "1K", got["size"])
			require.NotContains(t, got, "image")
			require.Equal(t, []any{"data:image/png;base64,aGVsbG8="}, got["extra_body"].(map[string]any)["image"])
			require.Equal(t, map[string]float64{"n": 1}, info.PriceData.OtherRatios())
		})
	}
}

func TestAgnesMessagesResponseUsesClaudeUsage(t *testing.T) {
	c, w := gin.CreateTestContext(httptest.NewRecorder())
	_ = w
	c.Request = httptest.NewRequest("POST", "/v1/messages", nil)
	info := &relaycommon.RelayInfo{RelayFormat: types.RelayFormatClaude, ChannelMeta: &relaycommon.ChannelMeta{}}
	response := &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewBufferString(`{"id":"msg_test","type":"message","role":"assistant","model":"agnes-2.5-flash","content":[{"type":"text","text":"OK"}],"stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":2,"cache_read_input_tokens":3}}`))}
	usage, apiErr := (&Adaptor{}).DoResponse(c, response, info)
	require.Nil(t, apiErr)
	require.Equal(t, types.RelayFormat(types.RelayFormatClaude), info.FinalRequestRelayFormat)
	require.Equal(t, 2, usage.(*dto.Usage).CompletionTokens)
}
