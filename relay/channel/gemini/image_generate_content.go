package gemini

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

// maxGeminiImageInputBytes caps reference images inlined into a generateContent
// request, matching the Veo image task limit.
const maxGeminiImageInputBytes = 20 * 1024 * 1024

// geminiSupportedAspectRatios lists the values accepted by
// generationConfig.imageConfig.aspectRatio. Unsupported client values are
// dropped instead of being relayed as a request the upstream rejects.
var geminiSupportedAspectRatios = map[string]struct{}{
	"1:1": {}, "1:4": {}, "1:8": {}, "2:3": {}, "3:2": {}, "3:4": {},
	"4:1": {}, "4:3": {}, "4:5": {}, "5:4": {}, "8:1": {}, "9:16": {},
	"16:9": {}, "21:9": {},
}

// geminiAspectRatioBySize maps OpenAI image sizes and the Studio ratio presets
// onto Gemini aspect ratios. Unknown sizes keep the model default.
var geminiAspectRatioBySize = map[string]string{
	"256x256": "1:1", "512x512": "1:1", "1024x1024": "1:1",
	"1216x832": "3:2", "1536x1024": "3:2",
	"1024x1536": "2:3",
	"864x1152":  "3:4", "1152x864": "4:3",
	"1024x1792": "9:16", "1792x1024": "16:9",
	"1792x768": "21:9",
}

// buildGeminiImageGenerateContentRequest converts an OpenAI image request into
// the generateContent request used by chat-style Gemini image models
// (gemini-*-image*, "Nano Banana"). Imagen models keep using :predict.
func buildGeminiImageGenerateContentRequest(c *gin.Context, request dto.ImageRequest) (*dto.GeminiChatRequest, error) {
	if request.Stream != nil && *request.Stream {
		return nil, errors.New("Gemini image models do not support streaming image generation")
	}

	parts, err := geminiImageInputParts(c, request)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.Prompt) != "" {
		parts = append(parts, dto.GeminiPart{Text: request.Prompt})
	}
	if len(parts) == 0 {
		return nil, errors.New("prompt or a reference image is required")
	}

	geminiRequest := &dto.GeminiChatRequest{
		Contents: []dto.GeminiChatContent{{Role: "user", Parts: parts}},
		GenerationConfig: dto.GeminiChatGenerationConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	}
	if imageConfig := geminiImageConfig(request); imageConfig != nil {
		geminiRequest.GenerationConfig.ImageConfig = imageConfig
	}
	return geminiRequest, nil
}

// geminiImageConfig maps size/quality onto generationConfig.imageConfig. The
// default 1K image size is omitted so image models that do not know imageSize
// keep working with the same request shape as before.
func geminiImageConfig(request dto.ImageRequest) json.RawMessage {
	config := make(map[string]string, 2)
	if ratio := geminiAspectRatio(request.Size); ratio != "" {
		config["aspectRatio"] = ratio
	}
	if size := geminiImageSize(request.Quality); size != "" {
		config["imageSize"] = size
	}
	if len(config) == 0 {
		return nil
	}
	data, err := common.Marshal(config)
	if err != nil {
		// Marshaling a map[string]string cannot fail; keep the request usable.
		return nil
	}
	return data
}

func geminiAspectRatio(size string) string {
	size = strings.ToLower(strings.TrimSpace(size))
	if ratio, ok := geminiAspectRatioBySize[size]; ok {
		return ratio
	}
	if _, ok := geminiSupportedAspectRatios[size]; ok {
		return size
	}
	return ""
}

func geminiImageSize(quality string) string {
	switch strings.ToLower(strings.TrimSpace(quality)) {
	case "hd", "high", "2k":
		return "2K"
	case "4k":
		return "4K"
	}
	return ""
}

// geminiImageInputParts returns the reference-image parts of an OpenAI image
// request. Reference images arrive either as multipart uploads (images/edits)
// or as data URIs in the JSON body; Gemini needs both inlined as parts.
func geminiImageInputParts(c *gin.Context, request dto.ImageRequest) ([]dto.GeminiPart, error) {
	parts := make([]dto.GeminiPart, 0, 2)

	if c != nil && c.Request != nil && c.Request.MultipartForm != nil {
		for _, field := range []string{"image", "image[]"} {
			for _, header := range c.Request.MultipartForm.File[field] {
				inline, err := geminiInlineDataFromUpload(header)
				if err != nil {
					return nil, err
				}
				parts = append(parts, dto.GeminiPart{InlineData: inline})
			}
		}
	}

	for _, raw := range []json.RawMessage{request.Image, request.Images} {
		values, err := geminiImageValues(raw)
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			inline, err := geminiInlineDataFromDataURI(value)
			if err != nil {
				return nil, err
			}
			parts = append(parts, dto.GeminiPart{InlineData: inline})
		}
	}
	return parts, nil
}

// geminiImageValues normalizes the image fields, which may hold a single data
// URI string or an array of them.
func geminiImageValues(raw json.RawMessage) ([]string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}

	var values []string
	if err := common.Unmarshal(trimmed, &values); err != nil {
		var single string
		if err := common.Unmarshal(trimmed, &single); err != nil {
			return nil, errors.New("image must be a data URI string or an array of data URI strings")
		}
		values = []string{single}
	}

	compacted := make([]string, 0, len(values))
	for _, value := range values {
		if trimmedValue := strings.TrimSpace(value); trimmedValue != "" {
			compacted = append(compacted, trimmedValue)
		}
	}
	return compacted, nil
}

func geminiInlineDataFromUpload(header *multipart.FileHeader) (*dto.GeminiInlineData, error) {
	if header.Size > maxGeminiImageInputBytes {
		return nil, fmt.Errorf("reference image %s exceeds %d bytes", header.Filename, maxGeminiImageInputBytes)
	}
	file, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("open reference image failed: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxGeminiImageInputBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read reference image failed: %w", err)
	}
	if len(data) > maxGeminiImageInputBytes {
		return nil, fmt.Errorf("reference image %s exceeds %d bytes", header.Filename, maxGeminiImageInputBytes)
	}

	mimeType := strings.ToLower(strings.TrimSpace(header.Header.Get("Content-Type")))
	if !strings.HasPrefix(mimeType, "image/") {
		mimeType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("reference image %s is not an image (%s)", header.Filename, mimeType)
	}

	return &dto.GeminiInlineData{
		MimeType: mimeType,
		Data:     base64.StdEncoding.EncodeToString(data),
	}, nil
}

// geminiInlineDataFromDataURI inlines a JSON reference image. Only data URIs
// are supported: Gemini needs inline bytes and this relay does not fetch
// arbitrary URLs on behalf of clients.
func geminiInlineDataFromDataURI(value string) (*dto.GeminiInlineData, error) {
	if !strings.HasPrefix(value, "data:") {
		return nil, errors.New("Gemini image models only accept base64 data URIs or uploaded reference images")
	}
	meta, data, found := strings.Cut(value[len("data:"):], ",")
	if !found || data == "" {
		return nil, errors.New("invalid image data URI")
	}
	mediaType, encoding, _ := strings.Cut(meta, ";")
	mimeType := strings.ToLower(strings.TrimSpace(mediaType))
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("unsupported image data URI type: %s", mediaType)
	}
	if !strings.EqualFold(strings.TrimSpace(encoding), "base64") {
		return nil, errors.New("image data URI must be base64 encoded")
	}
	if !validGeminiImageBase64(data) {
		return nil, errors.New("image data URI is not valid base64")
	}
	if len(data) > maxGeminiImageInputBytes*4/3 {
		return nil, fmt.Errorf("reference image exceeds %d bytes", maxGeminiImageInputBytes)
	}
	return &dto.GeminiInlineData{MimeType: mimeType, Data: data}, nil
}

// GeminiGenerateContentImageHandler converts a generateContent response that
// carries inline images into the OpenAI image response consumed by the image
// relay (data[].b64_json), mirroring GeminiImageHandler for imagen.
func GeminiGenerateContentImageHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.Usage, *types.NewAPIError) {
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, types.NewOpenAIError(readErr, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}
	service.CloseResponseBodyGracefully(resp)

	// Gemini-compatible reverse endpoints may return generated images as
	// Markdown data URIs inside text parts; normalize them to inlineData so a
	// single extraction path handles both response shapes.
	responseBody, _, err := normalizeGeminiMarkdownImages(responseBody)
	if err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	var geminiResponse dto.GeminiChatResponse
	if err := common.Unmarshal(responseBody, &geminiResponse); err != nil {
		return nil, types.NewOpenAIError(err, types.ErrorCodeBadResponseBody, http.StatusInternalServerError)
	}

	if len(geminiResponse.Candidates) == 0 {
		usage := buildUsageFromGeminiResponse(c, info, &geminiResponse)
		if geminiResponse.PromptFeedback != nil && geminiResponse.PromptFeedback.BlockReason != nil {
			common.SetContextKey(c, constant.ContextKeyAdminRejectReason, fmt.Sprintf("gemini_block_reason=%s", *geminiResponse.PromptFeedback.BlockReason))
			return &usage, types.NewOpenAIError(
				errors.New("request blocked by Gemini API: "+*geminiResponse.PromptFeedback.BlockReason),
				types.ErrorCodePromptBlocked,
				http.StatusBadRequest,
			)
		}
		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "gemini_empty_candidates")
		return &usage, types.NewOpenAIError(
			errors.New("empty response from Gemini API"),
			types.ErrorCodeEmptyResponse,
			http.StatusInternalServerError,
		)
	}

	images := geminiInlineImages(&geminiResponse)
	if len(images) == 0 {
		common.SetContextKey(c, constant.ContextKeyAdminRejectReason, "gemini_image_missing")
		return nil, types.NewOpenAIError(
			errors.New("no image returned by Gemini image model"),
			types.ErrorCodeEmptyResponse,
			http.StatusInternalServerError,
		)
	}

	usage := buildUsageFromGeminiResponse(c, info, &geminiResponse)
	relaycommon.ApplyFixedPriceImageCount(info, int64(len(images)))

	openAIResponse := dto.ImageResponse{
		Created: common.GetTimestamp(),
		Data:    images,
	}
	jsonResponse, jsonErr := common.Marshal(openAIResponse)
	if jsonErr != nil {
		return nil, types.NewError(jsonErr, types.ErrorCodeBadResponseBody)
	}

	service.IOCopyBytesGracefully(c, resp, jsonResponse)
	return &usage, nil
}

// geminiInlineImages extracts inline image parts from a Gemini response.
// Non-image inline payloads (for example audio) are skipped.
func geminiInlineImages(response *dto.GeminiChatResponse) []dto.ImageData {
	if response == nil {
		return nil
	}
	images := make([]dto.ImageData, 0, len(response.Candidates))
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData == nil || part.InlineData.Data == "" {
				continue
			}
			if !strings.HasPrefix(strings.ToLower(part.InlineData.MimeType), "image/") {
				continue
			}
			// 上游可能返回损坏的图片数据：非 base64 的 payload 直接透传会让客户端
			// 拿到无法解码的 b64_json，这里与 markdown 路径保持一致地丢弃。
			if !validGeminiImageBase64(part.InlineData.Data) {
				continue
			}
			images = append(images, dto.ImageData{B64Json: part.InlineData.Data})
		}
	}
	return images
}
