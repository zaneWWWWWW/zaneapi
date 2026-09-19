package agnes

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/stretchr/testify/require"
)

func TestVideo25ModesMappingAndFixedBilling(t *testing.T) {
	for _, fields := range []string{
		`"mode":"text","seconds":"4","size":"720P","seed":0`,
		`"mode":"keyframe","seconds":"12","size":"2K","first_frame":"https://example.com/first.png","last_frame":"https://example.com/last.png"`,
		`"mode":"reference","seconds":"8","size":"1080P","images":["https://example.com/1.png","https://example.com/2.png"],"audios":["https://example.com/a.mp3"],"videos":[{"url":"https://example.com/v.mp4","start_seconds":0,"require_audio":false}]`,
		`"mode":"text","extra_body":{"mode":"reference","seconds":"6","seed":0}`,
	} {
		t.Run(fields, func(t *testing.T) {
			body := `{"model":"alias","prompt":"test",` + fields + `}`
			c, _ := newAgnesVideoContext(body)
			info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: ModelVideo25}}
			a := &TaskAdaptor{}
			require.Nil(t, a.ValidateRequestAndSetAction(c, info))
			require.Nil(t, a.ValidateMappedRequest(c, info))
			reader, err := a.BuildRequestBody(c, info)
			require.NoError(t, err)
			data, err := io.ReadAll(reader)
			require.NoError(t, err)
			var got map[string]any
			require.NoError(t, common.Unmarshal(data, &got))
			require.Equal(t, ModelVideo25, got["model"])
			require.NotContains(t, got, "extra_body")
			if strings.Contains(body, `"seed":0`) {
				require.Equal(t, float64(0), got["seed"])
			}
			if videos, ok := got["videos"].([]any); ok {
				video := videos[0].(map[string]any)
				require.Equal(t, float64(0), video["start_seconds"])
				require.Equal(t, false, video["require_audio"])
			}
			require.Empty(t, a.EstimateBilling(c, info))
			require.Empty(t, a.AdjustBillingOnSubmit(info, []byte(`{"seconds":"12","size":"2K"}`)))
			require.Zero(t, a.AdjustBillingOnComplete(&model.Task{Quota: 62500}, &relaycommon.TaskInfo{TotalTokens: 99999}))
		})
	}
}

func TestVideo25RejectsInvalidRequestsAfterMapping(t *testing.T) {
	for _, tc := range []struct{ name, model, fields string }{
		{"missing mode", ModelVideo25, ``},
		{"duration type", ModelVideo25, `"mode":"text","seconds":5`},
		{"duration low", ModelVideo25, `"mode":"text","seconds":"3"`},
		{"duration high", ModelVideo25, `"mode":"text","seconds":"13"`},
		{"pixel size", ModelVideo25, `"mode":"text","size":"1280x720"`},
		{"ratio", ModelVideo25, `"mode":"text","aspect_ratio":"auto"`},
		{"batch", ModelVideo25, `"mode":"text","n":2`},
		{"zero batch", ModelVideo25, `"mode":"text","n":0`},
		{"string batch", ModelVideo25, `"mode":"text","n":"1"`},
		{"string seed", ModelVideo25, `"mode":"text","seed":"0"`},
		{"frames", ModelVideo25, `"mode":"text","num_frames":121`},
		{"legacy reference", ModelVideo25, `"mode":"reference","video_url":"https://example.com/a.mp4"`},
		{"text media", ModelVideo25, `"mode":"text","images":["https://example.com/a.png"]`},
		{"empty keyframe", ModelVideo25, `"mode":"keyframe"`},
		{"keyframe references", ModelVideo25, `"mode":"keyframe","first_frame":"https://example.com/a.png","audios":["https://example.com/a.mp3"]`},
		{"empty reference", ModelVideo25, `"mode":"reference"`},
		{"reference frames", ModelVideo25, `"mode":"reference","first_frame":"https://example.com/a.png","images":["https://example.com/a.png"]`},
		{"bad URL", ModelVideo25, `"mode":"keyframe","first_frame":"file:///tmp/a.png"`},
		{"video start", ModelVideo25, `"mode":"reference","videos":[{"url":"https://example.com/v.mp4","start_seconds":-1}]`},
		{"video audio", ModelVideo25, `"mode":"reference","videos":[{"url":"https://example.com/v.mp4","require_audio":"false"}]`},
		{"extra shape", ModelVideo25, `"mode":"text","extra_body":[]`},
		{"flash size", ModelVideo25Flash, `"mode":"text","size":"1080P"`},
		{"flash video", ModelVideo25Flash, `"mode":"reference","videos":[{"url":"https://example.com/v.mp4"}]`},
		{"legacy duration", ModelVideoV20, `"seconds":"5"`},
		{"legacy frames", ModelVideoV20, `"num_frames":120`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"model":"alias","prompt":"test"`
			if tc.fields != "" {
				body += "," + tc.fields
			}
			body += "}"
			c, _ := newAgnesVideoContext(body)
			info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: tc.model}}
			a := &TaskAdaptor{}
			require.Nil(t, a.ValidateRequestAndSetAction(c, info))
			taskErr := a.ValidateMappedRequest(c, info)
			require.NotNil(t, taskErr)
			require.Equal(t, http.StatusBadRequest, taskErr.StatusCode)
		})
	}
	for _, tc := range []struct {
		model, field string
		limit        int
	}{
		{ModelVideo25, "images", 8}, {ModelVideo25Flash, "images", 5}, {ModelVideo25, "audios", 3}, {ModelVideo25Flash, "audios", 3}, {ModelVideo25, "videos", 1},
	} {
		for _, count := range []int{tc.limit, tc.limit + 1} {
			items := make([]any, count)
			for i := range items {
				items[i] = "https://example.com/media"
				if tc.field == "videos" {
					items[i] = map[string]any{"url": "https://example.com/v.mp4"}
				}
			}
			req := map[string]any{"model": tc.model, "prompt": "test", "mode": "reference", tc.field: items}
			_, err := prepareRequest(req, nil)
			if count > tc.limit {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		}
	}
}

func TestFetchTaskNewAndLegacyURLs(t *testing.T) {
	service.InitHttpClient()
	var gotPath, gotVideo, gotModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotVideo = r.URL.Query().Get("video_id")
		gotModel = r.URL.Query().Get("model_name")
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.Write([]byte(`{"status":"queued"}`))
	}))
	defer server.Close()
	for _, modelName := range ModelList {
		resp, err := (&TaskAdaptor{}).FetchTask(server.URL+"/v1/", "test-key", map[string]any{"task_id": "upstream", "video_id": "video/a+b?c=&", "model_name": modelName}, "")
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, "/agnesapi", gotPath)
		require.Equal(t, "video/a+b?c=&", gotVideo)
		require.Equal(t, modelName, gotModel)
	}
	resp, err := (&TaskAdaptor{}).FetchTask(server.URL, "test-key", map[string]any{"task_id": "legacy", "model_name": ModelVideoV20}, "")
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, "/v1/videos/legacy", gotPath)
	_, err = (&TaskAdaptor{}).FetchTask(server.URL, "test-key", map[string]any{"task_id": "new", "model_name": ModelVideo25}, "")
	require.ErrorContains(t, err, "missing video_id")
}

func TestVideoResponseFormsAndPrivateRetrievalID(t *testing.T) {
	for _, wrapper := range []string{`%s`, `{"body":%s}`, `{"status":"completed","response":{"body":%s}}`, `{"final_response":{"body":%s}}`} {
		body := fmt.Sprintf(wrapper, `{"id":"upstream","video_id":"retrieval","status":"completed","url":"https://example.com/video.mp4","seconds":"1.0","size":"1088x832","size_mapping":{"width":1088,"height":832},"completed_at":1234}`)
		a := &TaskAdaptor{}
		result, err := a.ParseTaskResult([]byte(body))
		require.NoError(t, err)
		require.Equal(t, "https://example.com/video.mp4", result.Url)
		c, w := newAgnesVideoContext(`{}`)
		info := &relaycommon.RelayInfo{OriginModelName: "alias", TaskRelayInfo: &relaycommon.TaskRelayInfo{PublicTaskID: "public"}}
		id, data, taskErr := a.DoResponse(c, &http.Response{Body: io.NopCloser(strings.NewReader(body))}, info)
		require.Nil(t, taskErr)
		require.Equal(t, "upstream", id)
		require.Equal(t, "retrieval", info.UpstreamVideoID)
		require.NotContains(t, w.Body.String(), "retrieval")
		task := &model.Task{TaskID: "public", Status: model.TaskStatusSuccess, Data: data, Properties: model.Properties{OriginModelName: "alias"}, PrivateData: model.TaskPrivateData{UpstreamVideoID: info.UpstreamVideoID}}
		public, err := a.ConvertToOpenAIVideo(task)
		require.NoError(t, err)
		require.Contains(t, string(public), `"size_mapping"`)
		require.Contains(t, string(public), `"size":"1088x832"`)
		require.Contains(t, string(public), `"seconds":"1.0"`)
		require.NotContains(t, string(public), "retrieval")
		task.Status = model.TaskStatusInProgress
		public, err = a.ConvertToOpenAIVideo(task)
		require.NoError(t, err)
		require.NotContains(t, string(public), `"completed_at":1234`)
		require.NotContains(t, string(public), "https://example.com/video.mp4")
	}
}

func TestVideoErrorsAndSanitizedHistory(t *testing.T) {
	a := &TaskAdaptor{}
	for _, body := range []string{
		`{"body":{"error":{"message":"bad media","code":"invalid_media"}}}`,
		`{"response":{"body":{"status":"queued"}},"final_response":{"body":{"status":"failed","error":{"message":"bad media"}}}}`,
	} {
		result, err := a.ParseTaskResult([]byte(body))
		require.NoError(t, err)
		require.Equal(t, "FAILURE", result.Status)
		require.Equal(t, "bad media", result.Reason)
	}
	for _, code := range []int{429, 500, 503} {
		_, err := a.ParseTaskResult([]byte(fmt.Sprintf(`{"error":{"message":"retry later","code":%d}}`, code)))
		require.Error(t, err)
	}
	data, err := a.ConvertToOpenAIVideo(&model.Task{TaskID: "public", Status: model.TaskStatusFailure, FailReason: "task_failed", FinishTime: 1234})
	require.NoError(t, err)
	require.Contains(t, string(data), `"status":"failed"`)
	require.Contains(t, string(data), `"message":"task_failed"`)
}
