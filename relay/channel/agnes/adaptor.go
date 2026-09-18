package agnes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/common"
	channelconstant "github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/claude"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type Adaptor struct {
	openai.Adaptor
}

type imageRequest struct {
	Model        string         `json:"model"`
	Prompt       string         `json:"prompt"`
	Size         string         `json:"size"`
	Ratio        *string        `json:"ratio,omitempty"`
	ReturnBase64 *bool          `json:"return_base64,omitempty"`
	ExtraBody    map[string]any `json:"extra_body,omitempty"`
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	relayMode := relayconstant.RelayModeImagesGenerations
	if info != nil {
		relayMode = info.RelayMode
	}
	switch relayMode {
	case relayconstant.RelayModeImagesGenerations:
		return convertImageRequest(info, request, false)
	case relayconstant.RelayModeImagesEdits:
		return convertImageRequest(info, request, true)
	default:
		return a.Adaptor.ConvertImageRequest(c, info, request)
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if info == nil || info.ChannelMeta == nil {
		return "", errors.New("relay info is nil")
	}
	path := info.RequestURLPath
	if info.RelayFormat == types.RelayFormatClaude {
		path = "/v1/messages"
	} else if info.RelayMode == relayconstant.RelayModeImagesEdits {
		path = "/v1/images/generations"
	}
	return NormalizeBaseURL(info.ChannelBaseUrl) + path, nil
}

// NormalizeBaseURL accepts the API root or the /v1 base used by official SDKs.
func NormalizeBaseURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return channelconstant.ChannelBaseURLs[channelconstant.ChannelTypeAgnesAI]
	}
	return strings.TrimSuffix(base, "/v1")
}

func (a *Adaptor) ConvertOpenAIRequest(_ *gin.Context, _ *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) ConvertClaudeRequest(_ *gin.Context, _ *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(_ *gin.Context, _ *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return &request, nil
}

// Dispatch through this adaptor so image and Messages URL/header overrides apply.
func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, body io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, body)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (any, *types.NewAPIError) {
	if info.RelayFormat == types.RelayFormatClaude {
		return (&claude.Adaptor{}).DoResponse(c, resp, info)
	}
	return a.Adaptor.DoResponse(c, resp, info)
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, header *http.Header, info *relaycommon.RelayInfo) error {
	if info == nil || info.ChannelMeta == nil {
		return errors.New("relay info is nil")
	}
	if info.RelayFormat == types.RelayFormatClaude {
		return (&claude.Adaptor{}).SetupRequestHeader(c, header, info)
	}
	if err := a.Adaptor.SetupRequestHeader(c, header, info); err != nil {
		return err
	}
	switch info.RelayMode {
	case relayconstant.RelayModeImagesGenerations, relayconstant.RelayModeImagesEdits:
		header.Set("Content-Type", gin.MIMEJSON)
	}
	return nil
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

func convertImageRequest(info *relaycommon.RelayInfo, request dto.ImageRequest, requireImage bool) (any, error) {
	if request.N != nil && *request.N != 1 {
		return nil, errors.New("agnes image API only supports n = 1")
	}

	modelName := strings.TrimSpace(request.Model)
	if info != nil && info.ChannelMeta != nil && strings.TrimSpace(info.UpstreamModelName) != "" {
		modelName = strings.TrimSpace(info.UpstreamModelName)
	}
	if modelName == "" {
		modelName = ModelImage21Flash
	}

	extraBody, images, err := buildImageFields(request)
	if err != nil {
		return nil, err
	}
	if requireImage && len(images) == 0 {
		return nil, errors.New("agnes image edits require an image URL or Data URI in image or extra_body.image; file uploads are not supported")
	}
	if len(images) > 0 {
		if extraBody == nil {
			extraBody = make(map[string]any)
		}
		extraBody["image"] = images
	}
	var returnBase64 *bool
	if raw := getRawExtra(request, "return_base64"); raw != nil {
		if err := common.Unmarshal(raw, &returnBase64); err != nil {
			return nil, errors.New("return_base64 must be a boolean")
		}
	}
	var ratio *string
	if raw := getRawExtra(request, "ratio"); raw != nil {
		if err := common.Unmarshal(raw, &ratio); err != nil {
			return nil, errors.New("ratio must be a string")
		}
		if ratio != nil && !validImageRatio(*ratio) {
			return nil, errors.New("unsupported image ratio")
		}
	}
	size := strings.TrimSpace(request.Size)
	if size == "" {
		size = "1K"
		if modelName == ModelImage20Flash {
			size = "1024x1024"
		}
	}

	converted := imageRequest{
		Model:        modelName,
		Prompt:       request.Prompt,
		Size:         size,
		Ratio:        ratio,
		ReturnBase64: returnBase64,
		ExtraBody:    extraBody,
	}
	if info != nil {
		// A retried request may carry ratios from another provider. Agnes image
		// calls always use one fixed charge, independent of media parameters.
		info.PriceData.ReplaceOtherRatios(map[string]float64{"n": 1})
	}

	return converted, nil
}

func buildImageFields(request dto.ImageRequest) (map[string]any, []string, error) {
	extraBody := make(map[string]any)
	var images []string

	if request.Extra != nil {
		if raw, ok := request.Extra["extra_body"]; ok && len(bytes.TrimSpace(raw)) > 0 {
			var parsed map[string]json.RawMessage
			if err := common.Unmarshal(raw, &parsed); err != nil {
				return nil, nil, fmt.Errorf("invalid extra_body: %w", err)
			}
			for key, value := range parsed {
				if len(bytes.TrimSpace(value)) == 0 {
					continue
				}
				if key == "image" {
					if len(bytes.TrimSpace(request.Image)) > 0 {
						continue // The explicit top-level input takes precedence.
					}
					normalized, ok, err := normalizeImageValue(value)
					if err != nil {
						return nil, nil, err
					}
					if ok {
						images = normalized
					}
				} else {
					extraBody[key] = value
				}
			}
		}
	}

	if len(bytes.TrimSpace(request.Image)) > 0 {
		normalized, _, err := normalizeImageValue(request.Image)
		if err != nil {
			return nil, nil, err
		}
		images = normalized
	}

	if _, ok := extraBody["response_format"]; !ok && strings.TrimSpace(request.ResponseFormat) != "" {
		extraBody["response_format"] = strings.TrimSpace(request.ResponseFormat)
	}

	if len(extraBody) == 0 {
		return nil, images, nil
	}
	return extraBody, images, nil
}

func getRawExtra(request dto.ImageRequest, key string) json.RawMessage {
	if request.Extra == nil {
		return nil
	}
	raw := request.Extra[key]
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	return raw
}

func normalizeImageValue(raw json.RawMessage) ([]string, bool, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, false, nil
	}

	var images []string
	if err := common.Unmarshal(trimmed, &images); err == nil {
		images = compactStrings(images)
		for _, input := range images {
			if !validImageInput(input) {
				return nil, false, errors.New("image must contain public HTTP URLs or image Data URIs")
			}
		}
		return images, len(images) > 0, nil
	}

	var image string
	if err := common.Unmarshal(trimmed, &image); err == nil {
		image = strings.TrimSpace(image)
		if image == "" {
			return nil, false, nil
		}
		if !validImageInput(image) {
			return nil, false, errors.New("image must be an HTTP URL or image Data URI")
		}
		return []string{image}, true, nil
	}

	return nil, false, errors.New("agnes image input must be a URL string or an array of URL strings")
}

func validImageInput(input string) bool {
	if strings.HasPrefix(input, "data:image/") {
		return strings.Contains(input, ";base64,")
	}
	u, err := url.Parse(input)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func validImageRatio(ratio string) bool {
	switch ratio {
	case "1:1", "3:4", "4:3", "16:9", "9:16", "2:3", "3:2", "21:9":
		return true
	}
	return false
}

func compactStrings(values []string) []string {
	compact := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			compact = append(compact, value)
		}
	}
	return compact
}
