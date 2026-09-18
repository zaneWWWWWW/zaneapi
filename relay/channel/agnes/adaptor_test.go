package agnes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestConvertImageRequestBuildsAgnesExtraBody(t *testing.T) {
	var request dto.ImageRequest
	err := common.Unmarshal([]byte(`{
		"model": "agnes-image-2.1-flash",
		"prompt": "turn it into a rainy cyberpunk night",
		"size": "1024x768",
		"response_format": "url",
		"extra_body": {
			"image": "https://example.com/input.png"
		}
	}`), &request)
	require.NoError(t, err)

	converted, err := (&Adaptor{}).ConvertImageRequest(nil, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesGenerations,
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: ModelImage21Flash,
		},
	}, request)
	require.NoError(t, err)

	data, err := common.Marshal(converted)
	require.NoError(t, err)

	var payload struct {
		Model          string   `json:"model"`
		Prompt         string   `json:"prompt"`
		Size           string   `json:"size"`
		ResponseFormat string   `json:"response_format"`
		Image          []string `json:"image"`
		ExtraBody      struct {
			Image          []string `json:"image"`
			ResponseFormat string   `json:"response_format"`
		} `json:"extra_body"`
	}
	require.NoError(t, common.Unmarshal(data, &payload))

	require.Equal(t, ModelImage21Flash, payload.Model)
	require.Equal(t, "turn it into a rainy cyberpunk night", payload.Prompt)
	require.Equal(t, "1024x768", payload.Size)
	require.Empty(t, payload.ResponseFormat)
	require.Empty(t, payload.Image)
	require.Len(t, payload.ExtraBody.Image, 1)
	require.Equal(t, "https://example.com/input.png", payload.ExtraBody.Image[0])
	require.Equal(t, "url", payload.ExtraBody.ResponseFormat)
}

func TestConvertImageRequestRejectsMultipleImagesN(t *testing.T) {
	n := uint(2)
	request := dto.ImageRequest{
		Model:  ModelImage21Flash,
		Prompt: "a cute cat",
		N:      &n,
	}

	_, err := (&Adaptor{}).ConvertImageRequest(nil, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesGenerations,
	}, request)
	require.Error(t, err)
}

func TestConvertImageEditsRequestMapsTopLevelImage(t *testing.T) {
	var request dto.ImageRequest
	err := common.Unmarshal([]byte(`{
		"model": "agnes-image-2.1-flash",
		"prompt": "make the cube blue",
		"image": "https://example.com/edit-source.png"
	}`), &request)
	require.NoError(t, err)

	converted, err := (&Adaptor{}).ConvertImageRequest(nil, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesEdits,
	}, request)
	require.NoError(t, err)

	data, err := common.Marshal(converted)
	require.NoError(t, err)

	var payload struct {
		ExtraBody struct {
			Image []string `json:"image"`
		} `json:"extra_body"`
	}
	require.NoError(t, common.Unmarshal(data, &payload))
	require.Len(t, payload.ExtraBody.Image, 1)
	require.Equal(t, "https://example.com/edit-source.png", payload.ExtraBody.Image[0])
}

func TestConvertImageRequestForwardsReturnBase64(t *testing.T) {
	var request dto.ImageRequest
	err := common.Unmarshal([]byte(`{
		"model": "agnes-image-2.1-flash",
		"prompt": "a watercolor city map",
		"size": "1024x1024",
		"return_base64": false
	}`), &request)
	require.NoError(t, err)

	converted, err := (&Adaptor{}).ConvertImageRequest(nil, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesGenerations,
	}, request)
	require.NoError(t, err)

	data, err := common.Marshal(converted)
	require.NoError(t, err)

	var payload struct {
		ReturnBase64 *bool `json:"return_base64"`
	}
	require.NoError(t, common.Unmarshal(data, &payload))
	require.NotNil(t, payload.ReturnBase64)
	require.False(t, *payload.ReturnBase64)
}

func TestConvertImageEditsRequestRequiresImageURL(t *testing.T) {
	request := dto.ImageRequest{
		Model:  ModelImage21Flash,
		Prompt: "make the cube blue",
	}

	_, err := (&Adaptor{}).ConvertImageRequest(nil, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesEdits,
	}, request)
	require.Error(t, err)
}

func TestGetRequestURLEditsUsesGenerationsEndpoint(t *testing.T) {
	got, err := (&Adaptor{}).GetRequestURL(&relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesEdits,
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelBaseUrl: "https://apihub.agnes-ai.com",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "https://apihub.agnes-ai.com/v1/images/generations", got)
}

func TestSetupRequestHeaderForcesJSONForImageEdits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", nil)
	c.Request.Header.Set("Content-Type", "multipart/form-data; boundary=test")

	header := http.Header{}
	err := (&Adaptor{}).SetupRequestHeader(c, &header, &relaycommon.RelayInfo{
		RelayMode: relayconstant.RelayModeImagesEdits,
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiKey: "test-key",
		},
	})
	require.NoError(t, err)
	require.Equal(t, gin.MIMEJSON, header.Get("Content-Type"))
	require.Equal(t, "Bearer test-key", header.Get("Authorization"))
}

func TestGetModelListIncludesCurrentAgnesModels(t *testing.T) {
	models := (&Adaptor{}).GetModelList()
	seen := make(map[string]bool, len(models))
	for _, model := range models {
		seen[model] = true
	}
	for _, model := range []string{
		ModelText15Flash,
		ModelText20Flash,
		ModelText25Flash,
		ModelText25ProBeta,
		ModelText25Pro,
		ModelText30Flash,
		ModelImage20Flash,
		ModelImage21Flash,
		ModelImage25Flash,
		ModelVideoV20,
		ModelVideo25,
		ModelVideo25Flash,
	} {
		require.True(t, seen[model], "model list missing %s", model)
	}
}
