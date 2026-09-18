package sora

import (
	"testing"

	"github.com/QuantumNous/new-api/relay/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRequestURLUsesXaiGenerationsEndpoint(t *testing.T) {
	adaptor := &TaskAdaptor{baseURL: "https://api.x.ai"}
	info := &common.RelayInfo{OriginModelName: "grok-imagine-video"}

	url, err := adaptor.BuildRequestURL(info)

	require.NoError(t, err)
	assert.Equal(t, "https://api.x.ai/v1/videos/generations", url)
}

func TestNormalizeXaiVideoRequestBodyConvertsImageString(t *testing.T) {
	req := map[string]interface{}{
		"model":  "grok-imagine-video-1.5-preview",
		"prompt": "animate this",
		"image":  "https://example.com/image.png",
	}

	normalizeXaiVideoRequestBody(req)

	image, ok := req["image"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://example.com/image.png", image["url"])
}

func TestNormalizeXaiVideoRequestBodyConvertsImagesArray(t *testing.T) {
	req := map[string]interface{}{
		"model":  "grok-imagine-video-1.5-preview",
		"prompt": "animate this",
		"images": []interface{}{"https://example.com/image.png"},
	}

	normalizeXaiVideoRequestBody(req)

	image, ok := req["image"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://example.com/image.png", image["url"])
	assert.NotContains(t, req, "images")
}

func TestBuildRequestURL_MinimaxVariations(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		modelName   string
		requestPath string
		expectedURL string
	}{
		{
			name:        "base url without trailing v1",
			baseURL:     "http://10.149.9.30:28082",
			modelName:   "minimax-h3",
			expectedURL: "http://10.149.9.30:28082/v1/video/generations",
		},
		{
			name:        "base url with /v1",
			baseURL:     "http://10.149.9.30:28082/v1",
			modelName:   "minimax-h3-turbo",
			expectedURL: "http://10.149.9.30:28082/v1/video/generations",
		},
		{
			name:        "base url with full /v1/video/generations",
			baseURL:     "http://10.149.9.30:28082/v1/video/generations",
			modelName:   "minimax-video",
			expectedURL: "http://10.149.9.30:28082/v1/video/generations",
		},
		{
			name:        "model minimax/video-01",
			baseURL:     "http://10.149.9.30:28082",
			modelName:   "minimax/video-01",
			expectedURL: "http://10.149.9.30:28082/v1/video/generations",
		},
		{
			name:        "model h3",
			baseURL:     "http://10.149.9.30:28082",
			modelName:   "h3",
			expectedURL: "http://10.149.9.30:28082/v1/video/generations",
		},
		{
			name:        "request path explicitly specifies /video/generations",
			baseURL:     "http://10.149.9.30:28082",
			modelName:   "custom-other-model",
			requestPath: "/v1/video/generations",
			expectedURL: "http://10.149.9.30:28082/v1/video/generations",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adaptor := &TaskAdaptor{baseURL: tc.baseURL}
			info := &common.RelayInfo{
				OriginModelName: tc.modelName,
				RequestURLPath:  tc.requestPath,
			}
			url, err := adaptor.BuildRequestURL(info)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedURL, url)
		})
	}
}

func TestParseTaskResult_SynchronousCompletedVideo(t *testing.T) {
	adaptor := &TaskAdaptor{}
	syncResp := []byte(`{
		"id": "vid_sample_123",
		"created": 1789200000,
		"model": "minimax-h3-turbo",
		"prompt": "FPV drone cinematic shot",
		"duration_seconds": 5,
		"resolution": "1344x768",
		"data": [
			{
				"url": "http://10.149.9.30:28082/videos/result_123.mp4",
				"filename": "result_123.mp4",
				"content_type": "video/mp4"
			}
		]
	}`)

	taskInfo, err := adaptor.ParseTaskResult(syncResp)
	require.NoError(t, err)
	require.NotNil(t, taskInfo)
	assert.Equal(t, "SUCCESS", string(taskInfo.Status))
	assert.Equal(t, "http://10.149.9.30:28082/videos/result_123.mp4", taskInfo.Url)
}
