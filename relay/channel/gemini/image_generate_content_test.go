package gemini

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGeminiImageGenerateContentRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		request         dto.ImageRequest
		wantErr         string
		wantText        string
		wantInlineMime  string
		wantInlineData  string
		wantImageConfig string
	}{
		{
			name:            "studio default size maps to square aspect ratio",
			request:         dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "a red apple", Size: "1024x1024", Quality: "standard"},
			wantText:        "a red apple",
			wantImageConfig: `{"aspectRatio":"1:1"}`,
		},
		{
			name:            "hd quality and landscape size map to image size and ratio",
			request:         dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "a red apple", Size: "1792x1024", Quality: "hd"},
			wantText:        "a red apple",
			wantImageConfig: `{"aspectRatio":"16:9","imageSize":"2K"}`,
		},
		{
			name:            "supported explicit aspect ratio is preserved",
			request:         dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "a red apple", Size: "21:9"},
			wantText:        "a red apple",
			wantImageConfig: `{"aspectRatio":"21:9"}`,
		},
		{
			name:     "unknown size and quality leave the model defaults",
			request:  dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "a red apple", Size: "1234x567", Quality: "standard"},
			wantText: "a red apple",
		},
		{
			name:           "data URI reference image becomes an inline part",
			request:        dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "make it blue", Image: mustGeminiRawMessage(t, "data:image/png;base64,"+testPNGBase64)},
			wantText:       "make it blue",
			wantInlineMime: "image/png",
			wantInlineData: testPNGBase64,
		},
		{
			name:    "http reference image is rejected",
			request: dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "make it blue", Image: mustGeminiRawMessage(t, "https://example.com/apple.png")},
			wantErr: "only accept base64 data URIs",
		},
		{
			name:    "invalid base64 reference image is rejected",
			request: dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "make it blue", Image: mustGeminiRawMessage(t, "data:image/png;base64,!!!not-base64!!!")},
			wantErr: "not valid base64",
		},
		{
			name:    "streaming image generation is rejected",
			request: dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "a red apple", Stream: common.GetPointer(true)},
			wantErr: "do not support streaming image generation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := buildGeminiImageGenerateContentRequest(nil, tt.request)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)

			require.Len(t, got.Contents, 1)
			assert.Equal(t, "user", got.Contents[0].Role)
			assert.Equal(t, []string{"TEXT", "IMAGE"}, got.GenerationConfig.ResponseModalities)
			assert.Equal(t, tt.wantImageConfig, string(got.GenerationConfig.ImageConfig))

			if tt.wantInlineMime == "" {
				require.Len(t, got.Contents[0].Parts, 1)
				assert.Equal(t, tt.wantText, got.Contents[0].Parts[0].Text)
				return
			}

			require.Len(t, got.Contents[0].Parts, 2)
			require.NotNil(t, got.Contents[0].Parts[0].InlineData)
			assert.Equal(t, tt.wantInlineMime, got.Contents[0].Parts[0].InlineData.MimeType)
			assert.Equal(t, tt.wantInlineData, got.Contents[0].Parts[0].InlineData.Data)
			assert.Equal(t, tt.wantText, got.Contents[0].Parts[1].Text)
		})
	}
}

func TestBuildGeminiImageGenerateContentRequestWithUploadedReference(t *testing.T) {
	t.Parallel()

	pngBytes, err := base64.StdEncoding.DecodeString(testPNGBase64)
	require.NoError(t, err)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "reference.png")
	require.NoError(t, err)
	_, err = part.Write(pngBytes)
	require.NoError(t, err)
	require.NoError(t, writer.WriteField("prompt", "make it blue"))
	require.NoError(t, writer.Close())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/edits", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, c.Request.ParseMultipartForm(8<<20))

	got, err := buildGeminiImageGenerateContentRequest(c, dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "make it blue"})
	require.NoError(t, err)

	require.Len(t, got.Contents[0].Parts, 2)
	require.NotNil(t, got.Contents[0].Parts[0].InlineData)
	assert.Equal(t, "image/png", got.Contents[0].Parts[0].InlineData.MimeType)
	assert.Equal(t, testPNGBase64, got.Contents[0].Parts[0].InlineData.Data)
	assert.Equal(t, "make it blue", got.Contents[0].Parts[1].Text)
}

func TestBuildGeminiImageGenerateContentRequestRejectsNonImageUpload(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "notes.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("this is not an image"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/edits", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, c.Request.ParseMultipartForm(8<<20))

	_, err = buildGeminiImageGenerateContentRequest(c, dto.ImageRequest{Model: "gemini-3.1-flash-image-preview", Prompt: "make it blue"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not an image")
}

func TestGeminiGenerateContentImageHandlerReturnsOpenAIImageResponse(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
		"candidates":[{"content":{"role":"model","parts":[
			{"text":"Here is the image"},
			{"inlineData":{"mimeType":"image/jpeg","data":"` + testPNGBase64 + `"}}
		]},"finishReason":"STOP"}],
		"usageMetadata":{"promptTokenCount":261,"candidatesTokenCount":1120,"totalTokenCount":1381}
	}`)

	c, recorder := newGeminiImageTestContext(t)
	info := newGeminiImageTestRelayInfo()
	usage, newAPIError := GeminiGenerateContentImageHandler(c, info, geminiImageResponse(payload))
	require.Nil(t, newAPIError)
	require.NotNil(t, usage)
	assert.Equal(t, 261, usage.PromptTokens)
	assert.Equal(t, 1120, usage.CompletionTokens)
	assert.Equal(t, 1381, usage.TotalTokens)

	// Fixed-price image models settle on the images actually returned.
	assert.Equal(t, float64(1), info.PriceData.OtherRatios()["n"])

	var response dto.ImageResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, testPNGBase64, response.Data[0].B64Json)
	assert.Greater(t, response.Created, int64(0))
}

func TestGeminiGenerateContentImageHandlerAcceptsMarkdownImages(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
		"candidates":[{"content":{"role":"model","parts":[
			{"text":"Here you go ![image](data:image/png;base64,` + testPNGBase64 + `)"}
		]},"finishReason":"STOP"}],
		"usageMetadata":{"promptTokenCount":12,"candidatesTokenCount":1300,"totalTokenCount":1312}
	}`)

	c, recorder := newGeminiImageTestContext(t)
	_, newAPIError := GeminiGenerateContentImageHandler(c, newGeminiImageTestRelayInfo(), geminiImageResponse(payload))
	require.Nil(t, newAPIError)

	var response dto.ImageResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, testPNGBase64, response.Data[0].B64Json)
}

func TestGeminiGenerateContentImageHandlerErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		payload    string
		wantCode   types.ErrorCode
		wantStatus int
	}{
		{
			name:       "empty candidates report an empty response",
			payload:    `{"candidates":[],"usageMetadata":{"promptTokenCount":11,"totalTokenCount":11}}`,
			wantCode:   types.ErrorCodeEmptyResponse,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "text-only candidates report a missing image",
			payload:    `{"candidates":[{"content":{"role":"model","parts":[{"text":"I cannot do that"}]}}],"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":7,"totalTokenCount":18}}`,
			wantCode:   types.ErrorCodeEmptyResponse,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "blocked prompt is reported as a prompt block",
			payload:    `{"candidates":[],"promptFeedback":{"blockReason":"SAFETY"},"usageMetadata":{"promptTokenCount":11,"totalTokenCount":11}}`,
			wantCode:   types.ErrorCodePromptBlocked,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "inline image with invalid base64 reports a missing image",
			payload:    `{"candidates":[{"content":{"role":"model","parts":[{"inlineData":{"mimeType":"image/png","data":"!!!not-base64!!!"}}]}}],"usageMetadata":{"promptTokenCount":11,"candidatesTokenCount":7,"totalTokenCount":18}}`,
			wantCode:   types.ErrorCodeEmptyResponse,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := newGeminiImageTestContext(t)
			_, newAPIError := GeminiGenerateContentImageHandler(c, newGeminiImageTestRelayInfo(), geminiImageResponse([]byte(tt.payload)))
			require.NotNil(t, newAPIError)
			assert.Equal(t, tt.wantCode, newAPIError.GetErrorCode())
			assert.Equal(t, tt.wantStatus, newAPIError.StatusCode)
		})
	}
}

func TestSetupRequestHeaderForcesJSONForImageRelays(t *testing.T) {
	t.Parallel()

	c, _ := newGeminiImageTestContext(t)
	c.Request.Header.Set("Content-Type", "multipart/form-data; boundary=test-boundary")

	info := newGeminiImageTestRelayInfo()
	info.RelayMode = relayconstant.RelayModeImagesEdits
	imageHeader := http.Header{}
	require.NoError(t, (&Adaptor{}).SetupRequestHeader(c, &imageHeader, info))
	assert.Equal(t, gin.MIMEJSON, imageHeader.Get("Content-Type"))

	info.RelayMode = relayconstant.RelayModeChatCompletions
	chatHeader := http.Header{}
	require.NoError(t, (&Adaptor{}).SetupRequestHeader(c, &chatHeader, info))
	assert.Equal(t, "multipart/form-data; boundary=test-boundary", chatHeader.Get("Content-Type"))
}

func newGeminiImageTestContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/pg/images/generations", nil)
	return c, recorder
}

func newGeminiImageTestRelayInfo() *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		RelayFormat:     types.RelayFormatOpenAIImage,
		OriginModelName: "gemini-3.1-flash-image-preview",
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gemini-3.1-flash-image-preview",
		},
		PriceData: types.PriceData{UsePrice: true},
	}
}

func geminiImageResponse(payload []byte) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(payload)),
	}
}
