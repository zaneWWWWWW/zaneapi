package gemini

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeminiBillingProbeExtractsUsageAcrossChunkedWrites(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"inlineData":{"mimeType":"image/png","data":"` + strings.Repeat("A", 4096) + `"}}]}}],"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":22,"thoughtsTokenCount":3,"totalTokenCount":36}}`)
	probe := newGeminiBillingProbe()
	probe.w = io.Discard
	for _, b := range payload {
		_, err := probe.Write([]byte{b})
		require.NoError(t, err)
	}
	assert.Equal(t, 11, probe.usage.PromptTokenCount)
	assert.Equal(t, 22, probe.usage.CandidatesTokenCount)
	assert.Equal(t, 3, probe.usage.ThoughtsTokenCount)
	assert.Equal(t, 36, probe.usage.TotalTokenCount)
	assert.Equal(t, 1, probe.imageCount)
}

func TestServeGeminiNativeJSONStreamsLargeBodyWithoutHoldingWholeTree(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-test:generateContent", nil)
	info := &relaycommon.RelayInfo{
		OriginModelName: "gemini-test",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gemini-test",
		},
	}

	blob := strings.Repeat("B", geminiNativeBufferLimit+32)
	payload := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"inlineData":{"mimeType":"image/png","data":"` + blob + `"}}]}}],"promptFeedback":{"blockReason":null},"usageMetadata":{"promptTokenCount":9,"candidatesTokenCount":1400,"totalTokenCount":1409},"responseId":"keep-large"}`)
	require.Greater(t, len(payload), geminiNativeBufferLimit)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(payload)),
	}
	usage, apiErr := serveGeminiNativeJSON(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 9, usage.PromptTokens)
	assert.Equal(t, 1400, usage.CompletionTokens)
	assert.Equal(t, payload, recorder.Body.Bytes())
}

func TestServeGeminiNativeJSONDoesNotFailRequestAfterMalformedLargeBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-test:generateContent", nil)
	info := &relaycommon.RelayInfo{
		OriginModelName: "gemini-test",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gemini-test",
		},
	}

	payload := []byte(`{"candidates":[{"content":{"parts":[{"text":"` + strings.Repeat("x", geminiNativeBufferLimit) + `"}]}}]`)
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(payload)),
	}
	usage, apiErr := serveGeminiNativeJSON(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 0, usage.CompletionTokens)
	assert.Equal(t, payload, recorder.Body.Bytes())
}

func TestCopyGeminiNativeBodyDrainsUsageAfterClientWriteError(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"inlineData":{"mimeType":"image/png","data":"` + strings.Repeat("A", geminiNativeBufferLimit) + `"}}]}}],"usageMetadata":{"promptTokenCount":9,"candidatesTokenCount":1400,"totalTokenCount":1409}}`)
	require.Greater(t, len(payload), geminiNativeBufferLimit)

	probe := newGeminiBillingProbe()
	client := &failAfterBuffer{after: 128}
	clientErr, readErr := copyGeminiNativeBody(bytes.NewReader(payload), client, probe, nil)
	require.Error(t, clientErr)
	require.NoError(t, readErr)
	assert.Equal(t, 9, probe.usage.PromptTokenCount)
	assert.Equal(t, 1400, probe.usage.CandidatesTokenCount)
	assert.Equal(t, 1409, probe.usage.TotalTokenCount)
	assert.Equal(t, 1, probe.imageCount)
	assert.Less(t, client.buf.Len(), len(payload))
}

func TestServeGeminiNativeJSONSettlesUsageAfterClientBrokenPipe(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Writer = &failAfterGinWriter{ResponseWriter: c.Writer, after: 256}
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-test:generateContent", nil)
	info := &relaycommon.RelayInfo{
		OriginModelName: "gemini-test",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gemini-test",
		},
	}

	blob := strings.Repeat("B", geminiNativeBufferLimit+32)
	payload := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"inlineData":{"mimeType":"image/png","data":"` + blob + `"}}]}}],"usageMetadata":{"promptTokenCount":9,"candidatesTokenCount":1400,"totalTokenCount":1409}}`)
	require.Greater(t, len(payload), geminiNativeBufferLimit)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(payload)),
	}
	usage, apiErr := serveGeminiNativeJSON(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 9, usage.PromptTokens)
	assert.Equal(t, 1400, usage.CompletionTokens)
	assert.Less(t, recorder.Body.Len(), len(payload))
}

type failAfterBuffer struct {
	buf   bytes.Buffer
	after int
}

func (w *failAfterBuffer) Write(p []byte) (int, error) {
	if w.buf.Len() >= w.after {
		return 0, errors.New("write: broken pipe")
	}
	remain := w.after - w.buf.Len()
	if len(p) > remain {
		n, err := w.buf.Write(p[:remain])
		if err != nil {
			return n, err
		}
		return n, errors.New("write: broken pipe")
	}
	return w.buf.Write(p)
}

type failAfterGinWriter struct {
	gin.ResponseWriter
	after int
	wrote int
}

func (w *failAfterGinWriter) Write(p []byte) (int, error) {
	if w.wrote >= w.after {
		return 0, errors.New("write: broken pipe")
	}
	remain := w.after - w.wrote
	if len(p) > remain {
		n, err := w.ResponseWriter.Write(p[:remain])
		w.wrote += n
		if err != nil {
			return n, err
		}
		return n, errors.New("write: broken pipe")
	}
	n, err := w.ResponseWriter.Write(p)
	w.wrote += n
	return n, err
}

func (w *failAfterGinWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func TestGeminiBillingProbeDoesNotRetainHugeTextStrings(t *testing.T) {
	t.Parallel()

	huge := strings.Repeat("C", 2<<20)
	payload := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"` + huge + `"}]}}],"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":7,"totalTokenCount":12}}`)
	probe := newGeminiBillingProbe()
	var out bytes.Buffer
	probe.w = &out
	n, err := probe.Write(payload)
	require.NoError(t, err)
	require.Equal(t, len(payload), n)
	assert.Equal(t, 5, probe.usage.PromptTokenCount)
	assert.Equal(t, 7, probe.usage.CandidatesTokenCount)
	assert.Equal(t, 12, probe.usage.TotalTokenCount)
	assert.Equal(t, 64<<10, probe.text.Len())
	assert.LessOrEqual(t, cap(probe.keyBuf), 64)
	assert.Nil(t, probe.capture)
	assert.Equal(t, payload, out.Bytes())
}

func TestGeminiBillingProbeIgnoresUsageMetadataInsideText(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"candidates":[{"content":{"parts":[{"text":"see usageMetadata: {\"promptTokenCount\":99}"}]}}],"usageMetadata":{"promptTokenCount":4,"candidatesTokenCount":1,"totalTokenCount":5}}`)
	probe := newGeminiBillingProbe()
	probe.w = io.Discard
	_, err := probe.Write(payload)
	require.NoError(t, err)
	assert.Equal(t, 4, probe.usage.PromptTokenCount)
	assert.Equal(t, 1, probe.usage.CandidatesTokenCount)
	assert.Equal(t, 5, probe.usage.TotalTokenCount)
}

func TestGeminiBillingProbeReadsNestedUsageDetails(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"usageMetadata":{"promptTokenCount":8,"promptTokensDetails":[{"modality":"TEXT","tokenCount":8}],"candidatesTokenCount":2,"totalTokenCount":10}}`)
	probe := newGeminiBillingProbe()
	probe.w = io.Discard
	_, err := probe.Write(payload)
	require.NoError(t, err)
	assert.Equal(t, 8, probe.usage.PromptTokenCount)
	assert.Equal(t, 2, probe.usage.CandidatesTokenCount)
	require.Len(t, probe.usage.PromptTokensDetails, 1)
	assert.Equal(t, "TEXT", probe.usage.PromptTokensDetails[0].Modality)
	assert.Equal(t, 8, probe.usage.PromptTokensDetails[0].TokenCount)
}

func TestGeminiBillingProbeCapturesBlockReason(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"promptFeedback":{"blockReason":"SAFETY","safetyRatings":[]},"usageMetadata":{"promptTokenCount":3,"totalTokenCount":3}}`)
	probe := newGeminiBillingProbe()
	probe.w = io.Discard
	_, err := probe.Write(payload)
	require.NoError(t, err)
	assert.Equal(t, "SAFETY", probe.blockReason)
	assert.Equal(t, 3, probe.usage.PromptTokenCount)
}

func TestGeminiChatHandlerGeminiFormatUsesNativeCopy(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	info := &relaycommon.RelayInfo{
		RelayFormat:     types.RelayFormatGemini,
		OriginModelName: "gemini-3-flash-preview",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gemini-3-flash-preview",
		},
	}
	body := []byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"ok"}]}}],"usageMetadata":{"promptTokenCount":4,"candidatesTokenCount":2,"totalTokenCount":6}}`)
	resp := &http.Response{Body: io.NopCloser(bytes.NewReader(body))}
	usage, apiErr := GeminiChatHandler(c, info, resp)
	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	assert.Equal(t, 4, usage.PromptTokens)
	assert.Equal(t, 2, usage.CompletionTokens)
	assert.JSONEq(t, string(body), recorder.Body.String())
}
