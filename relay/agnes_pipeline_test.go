package relay

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type agnesBillingRecorder struct{ calls, quota int }

func (b *agnesBillingRecorder) Settle(quota int) error { b.calls++; b.quota = quota; return nil }
func (*agnesBillingRecorder) Refund(*gin.Context)      {}
func (*agnesBillingRecorder) NeedsRefund() bool        { return false }
func (*agnesBillingRecorder) GetPreConsumedQuota() int { return 0 }
func (*agnesBillingRecorder) Reserve(int) error        { return nil }

func agnesPipelineContext(path, body, baseURL string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	common.SetContextKey(c, constant.ContextKeyChannelType, constant.ChannelTypeAgnesAI)
	common.SetContextKey(c, constant.ContextKeyChannelBaseUrl, baseURL)
	common.SetContextKey(c, constant.ContextKeyChannelKey, "test-key")
	c.Set("model_mapping", `{"alias":"agnes-video-2.5"}`)
	return c
}

func TestAgnesTaskSubmitMappedValidationAndFixedQuota(t *testing.T) {
	service.InitHttpClient()
	previous := ratio_setting.ModelPrice2JSONString()
	t.Cleanup(func() { require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(previous)) })
	require.NoError(t, ratio_setting.UpdateModelPriceByJSONString(`{"alias":0.125}`))
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/videos", r.URL.Path)
		data, _ := io.ReadAll(r.Body)
		require.NoError(t, common.Unmarshal(data, &received))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"upstream-task","video_id":"retrieval-id","status":"queued","seconds":"12"}`))
	}))
	defer server.Close()
	for _, extra := range []string{
		`"mode":"text","seconds":"4","size":"720P"`,
		`"mode":"keyframe","seconds":"12","size":"2K","first_frame":"https://example.com/a.png"`,
		`"mode":"reference","seconds":"8","size":"1080P","images":["https://example.com/a.png","https://example.com/b.png"],"videos":[{"url":"https://example.com/v.mp4","start_seconds":0,"require_audio":false}]`,
	} {
		c := agnesPipelineContext("/v1/videos", `{"model":"alias","prompt":"test",`+extra+`}`, server.URL)
		info := &relaycommon.RelayInfo{TaskRelayInfo: &relaycommon.TaskRelayInfo{}, Billing: &agnesBillingRecorder{}, UsingGroup: "default"}
		result, taskErr := RelayTaskSubmit(c, info)
		require.Nil(t, taskErr)
		require.Equal(t, int(0.125*common.QuotaPerUnit), result.Quota)
		require.Equal(t, "agnes-video-2.5", received["model"])
		require.Equal(t, "retrieval-id", info.UpstreamVideoID)
		require.NotEqual(t, result.UpstreamTaskID, info.PublicTaskID)
		require.Empty(t, info.PriceData.OtherRatios())
	}
	// No billing session: validation must stop before price lookup, preconsume,
	// public ID allocation or upstream creation.
	c := agnesPipelineContext("/v1/videos", `{"model":"alias","prompt":"test","mode":"text","num_frames":121}`, server.URL)
	info := &relaycommon.RelayInfo{TaskRelayInfo: &relaycommon.TaskRelayInfo{}}
	result, taskErr := RelayTaskSubmit(c, info)
	require.Nil(t, result)
	require.Equal(t, http.StatusBadRequest, taskErr.StatusCode)
	require.Nil(t, info.Billing)
	require.Empty(t, info.PublicTaskID)
}

func TestAgnesImagePipelineSettlesOneCall(t *testing.T) {
	service.InitHttpClient()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMainDBType := common.MainDatabaseType()
	previousLogDBType := common.LogDatabaseType()
	previousBatch, previousLog, previousRedis := common.BatchUpdateEnabled, common.LogConsumeEnabled, common.RedisEnabled
	model.DB = db
	model.LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.BatchUpdateEnabled = false
	common.LogConsumeEnabled = true
	common.RedisEnabled = false
	t.Cleanup(func() {
		model.DB = previousDB
		model.LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainDBType, previousLogDBType)
		common.BatchUpdateEnabled = previousBatch
		common.LogConsumeEnabled = previousLog
		common.RedisEnabled = previousRedis
		sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Channel{}, &model.Log{}))
	require.NoError(t, db.Create(&model.User{Id: 1, Username: "agnes-image-test"}).Error)
	require.NoError(t, db.Create(&model.Channel{Id: 1, Name: "agnes-image-test"}).Error)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"created":1234,"data":[{"url":"https://example.com/image.png"}],"usage":{"prompt_tokens":9999,"completion_tokens":9999,"total_tokens":19998}}`))
	}))
	defer server.Close()
	for _, size := range []string{"1K", "2K", "4K"} {
		body := fmt.Sprintf(`{"model":"agnes-image-2.5-flash","prompt":"test","size":%q,"n":1}`, size)
		c := agnesPipelineContext("/v1/images/generations", body, server.URL)
		common.SetContextKey(c, constant.ContextKeyChannelId, 1)
		var request dto.ImageRequest
		require.NoError(t, common.Unmarshal([]byte(body), &request))
		billing := &agnesBillingRecorder{}
		info := &relaycommon.RelayInfo{Request: &request, OriginModelName: request.Model, RequestURLPath: c.Request.URL.Path, RelayMode: relayconstant.RelayModeImagesGenerations, RelayFormat: types.RelayFormatOpenAIImage, StartTime: time.Now(), UserId: 1, UserQuota: 1000000000, Billing: billing, PriceData: types.PriceData{UsePrice: true, ModelPrice: 0.01, GroupRatioInfo: types.GroupRatioInfo{GroupRatio: 1}}}
		require.Nil(t, ImageHelper(c, info))
		require.Equal(t, 1, billing.calls)
		require.Equal(t, int(0.01*common.QuotaPerUnit), billing.quota)
	}
	var logs int64
	require.NoError(t, db.Model(&model.Log{}).Count(&logs).Error)
	require.Equal(t, int64(3), logs)
}

func TestAgnesNativeProtocolPipelinesPreserveExplicitZero(t *testing.T) {
	service.InitHttpClient()
	for _, tc := range []struct {
		path, body string
		mode       int
		format     types.RelayFormat
		request    dto.Request
		helper     func(*gin.Context, *relaycommon.RelayInfo) *types.NewAPIError
		fields     []string
	}{
		{"/v1/chat/completions", `{"model":"alias","messages":[{"role":"user","content":[{"type":"text","text":"hello"},{"type":"image_url","image_url":{"url":"https://example.com/a.png"}}]}],"temperature":0,"top_p":0,"chat_template_kwargs":{"enable_thinking":false}}`, relayconstant.RelayModeChatCompletions, types.RelayFormatOpenAI, &dto.GeneralOpenAIRequest{}, TextHelper, []string{`"temperature":0`, `"top_p":0`, `"enable_thinking":false`, `"image_url"`}},
		{"/v1/responses", `{"model":"alias","input":"hello","temperature":0,"max_output_tokens":0}`, relayconstant.RelayModeResponses, types.RelayFormatOpenAIResponses, &dto.OpenAIResponsesRequest{}, ResponsesHelper, []string{`"temperature":0`, `"max_output_tokens":0`}},
		{"/v1/messages", `{"model":"alias","messages":[{"role":"user","content":"hello"}],"temperature":0,"max_tokens":0,"thinking":{"type":"enabled","budget_tokens":0}}`, relayconstant.RelayModeChatCompletions, types.RelayFormatClaude, &dto.ClaudeRequest{}, ClaudeHelper, []string{`"temperature":0`, `"max_tokens":0`, `"budget_tokens":0`}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			var captured string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tc.path, r.URL.Path)
				data, _ := io.ReadAll(r.Body)
				captured = string(data)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(400)
				w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"fixture stop"}}`))
			}))
			defer server.Close()
			require.NoError(t, common.Unmarshal([]byte(tc.body), tc.request))
			c := agnesPipelineContext(tc.path, tc.body, server.URL)
			c.Set("model_mapping", `{"alias":"agnes-2.5-flash"}`)
			info := &relaycommon.RelayInfo{Request: tc.request, OriginModelName: "alias", RequestURLPath: tc.path, RelayMode: tc.mode, RelayFormat: tc.format}
			require.NotNil(t, tc.helper(c, info))
			require.Contains(t, captured, `"model":"agnes-2.5-flash"`)
			for _, field := range tc.fields {
				require.Contains(t, captured, field)
			}
		})
	}
}
