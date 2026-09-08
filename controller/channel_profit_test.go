package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupChannelProfitControllerTest(t *testing.T) *model.Channel {
	t.Helper()
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Log{}, &model.ChannelProfitRecord{}))

	ratio := 0.72
	channel := &model.Channel{
		Name:          "profit-sec-channel",
		Type:          1,
		Key:           "test-key",
		Models:        "gpt-test",
		Group:         "default",
		UpstreamRatio: &ratio,
	}
	require.NoError(t, db.Create(channel).Error)
	require.NoError(t, db.Create(&model.AdminScope{
		UserId: 2,
		Kind:   model.AdminScopeKindChannel,
		Mode:   model.AdminScopeModeAll,
	}).Error)
	return channel
}

func TestGetChannelProfitRejectsInvalidTimestamps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []string{
		"/api/channel/profit",
		"/api/channel/profit?start_timestamp=abc&end_timestamp=2",
		"/api/channel/profit?start_timestamp=1;drop&end_timestamp=2",
		"/api/channel/profit?start_timestamp=0&end_timestamp=2",
		"/api/channel/profit?start_timestamp=10&end_timestamp=9",
		"/api/channel/profit?start_timestamp=1&end_timestamp=-1",
	}
	for _, path := range cases {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, path, nil)
		GetChannelProfit(ctx)
		assert.False(t, recorder.Code == http.StatusOK && bytes.Contains(recorder.Body.Bytes(), []byte(`"success":true`)), path)
	}
}

func TestGetChannelProfitReturnsConfiguredReport(t *testing.T) {
	channel := setupChannelProfitControllerTest(t)
	model.RecordChannelProfit("sec-consume", channel.Id, "gpt-test", 1000, 100, types.ProfitBasis{BaseQuota: 1000, GroupRatio: 1, UpstreamRatio: channel.UpstreamRatio})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/channel/profit?start_timestamp=1&end_timestamp=200", nil)
	GetChannelProfit(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			RevenueQuota             int64 `json:"revenue_quota"`
			CostQuota                int64 `json:"cost_quota"`
			ProfitQuota              int64 `json:"profit_quota"`
			UnconfiguredChannelCount int64 `json:"unconfigured_channel_count"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	assert.Equal(t, int64(1000), payload.Data.RevenueQuota)
	assert.Equal(t, int64(720), payload.Data.CostQuota)
	assert.Equal(t, int64(280), payload.Data.ProfitQuota)
}

func TestGetChannelHidesUpstreamRatioFromNonRoot(t *testing.T) {
	channel := setupChannelProfitControllerTest(t)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 2)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", channel.Id)}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/channel/"+fmt.Sprintf("%d", channel.Id), nil)
	GetChannel(ctx)

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			UpstreamRatio *float64 `json:"upstream_ratio"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
	require.True(t, payload.Success)
	assert.Nil(t, payload.Data.UpstreamRatio)
}

func TestUpdateChannelDoesNotLeakOrOverwriteUpstreamRatioForAdmin(t *testing.T) {
	channel := setupChannelProfitControllerTest(t)
	body := fmt.Sprintf(`{"id":%d,"name":"profit-sec-channel-renamed"}`, channel.Id)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 2)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/channel/", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	UpdateChannel(ctx)

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Name          string   `json:"name"`
			UpstreamRatio *float64 `json:"upstream_ratio"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload), recorder.Body.String())
	require.True(t, payload.Success, recorder.Body.String())
	assert.Equal(t, "profit-sec-channel-renamed", payload.Data.Name)
	assert.Nil(t, payload.Data.UpstreamRatio)

	var stored model.Channel
	require.NoError(t, model.DB.Select("upstream_ratio", "name").First(&stored, channel.Id).Error)
	require.NotNil(t, stored.UpstreamRatio)
	assert.Equal(t, 0.72, *stored.UpstreamRatio)
}

func TestUpdateChannelRejectsNonRootUpstreamRatioWrite(t *testing.T) {
	channel := setupChannelProfitControllerTest(t)
	body := fmt.Sprintf(`{"id":%d,"name":"profit-sec-channel","upstream_ratio":0.1}`, channel.Id)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 2)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/channel/", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	UpdateChannel(ctx)

	assert.Contains(t, recorder.Body.String(), "auth.insufficient_privilege")
	var stored model.Channel
	require.NoError(t, model.DB.Select("upstream_ratio").First(&stored, channel.Id).Error)
	require.NotNil(t, stored.UpstreamRatio)
	assert.Equal(t, 0.72, *stored.UpstreamRatio)
}

func TestUpdateChannelRootCanClearUpstreamRatio(t *testing.T) {
	channel := setupChannelProfitControllerTest(t)
	body := fmt.Sprintf(`{"id":%d,"name":"profit-sec-channel","upstream_ratio":null}`, channel.Id)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleRootUser)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/api/channel/", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	UpdateChannel(ctx)

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			UpstreamRatio *float64 `json:"upstream_ratio"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload), recorder.Body.String())
	require.True(t, payload.Success, recorder.Body.String())
	assert.Nil(t, payload.Data.UpstreamRatio)

	var stored model.Channel
	require.NoError(t, model.DB.Select("upstream_ratio").First(&stored, channel.Id).Error)
	assert.Nil(t, stored.UpstreamRatio)
}

func TestAddChannelNilChannelDoesNotPanic(t *testing.T) {
	setupChannelProfitControllerTest(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleRootUser)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/channel/", bytes.NewBufferString(`{"mode":"single"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	AddChannel(ctx)
	assert.NotContains(t, recorder.Body.String(), "new_api_panic")
	assert.Contains(t, recorder.Body.String(), "channel cannot be empty")
}

func TestAddChannelRejectsNonRootUpstreamRatioBeforeValidation(t *testing.T) {
	setupChannelProfitControllerTest(t)
	body := `{"mode":"single","channel":{"name":"admin-create","type":1,"key":"k","models":"gpt-test","group":"default","upstream_ratio":9}}`

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 2)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/channel/", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	AddChannel(ctx)

	assert.Contains(t, recorder.Body.String(), "auth.insufficient_privilege")
	assert.NotContains(t, recorder.Body.String(), "cost ratio must be between")
}

func TestCopyChannelOmitsUpstreamRatioForNonRoot(t *testing.T) {
	channel := setupChannelProfitControllerTest(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 2)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", channel.Id)}}
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/channel/copy/"+fmt.Sprintf("%d", channel.Id), nil)
	CopyChannel(ctx)

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Id int `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload), recorder.Body.String())
	require.True(t, payload.Success, recorder.Body.String())

	var clone model.Channel
	require.NoError(t, model.DB.Select("upstream_ratio").First(&clone, payload.Data.Id).Error)
	assert.Nil(t, clone.UpstreamRatio)
}

func TestCopyChannelKeepsUpstreamRatioForRoot(t *testing.T) {
	channel := setupChannelProfitControllerTest(t)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 1)
	ctx.Set("role", common.RoleRootUser)
	ctx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", channel.Id)}}
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/channel/copy/"+fmt.Sprintf("%d", channel.Id), nil)
	CopyChannel(ctx)

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Id int `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload), recorder.Body.String())
	require.True(t, payload.Success, recorder.Body.String())

	var clone model.Channel
	require.NoError(t, model.DB.Select("upstream_ratio").First(&clone, payload.Data.Id).Error)
	require.NotNil(t, clone.UpstreamRatio)
	assert.Equal(t, 0.72, *clone.UpstreamRatio)
}
