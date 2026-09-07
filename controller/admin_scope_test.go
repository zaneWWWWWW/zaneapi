package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminChannelScopeHidesUnassignedChannels(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	visible := &model.Channel{Name: "visible-channel", Key: "k1"}
	hidden := &model.Channel{Name: "hidden-channel", Key: "k2"}
	require.NoError(t, db.Create(visible).Error)
	require.NoError(t, db.Create(hidden).Error)
	require.NoError(t, db.Create(&model.AdminScope{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		Mode:   model.AdminScopeModeAssigned,
	}).Error)
	require.NoError(t, db.Create(&model.AdminScopeItem{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		ItemId: visible.Id,
	}).Error)

	listRecorder := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listRecorder)
	listCtx.Set("id", 8)
	listCtx.Set("role", common.RoleAdminUser)
	listCtx.Request = httptest.NewRequest(http.MethodGet, "/api/channel/?page_size=50", nil)
	GetAllChannels(listCtx)

	var listPayload struct {
		Success bool `json:"success"`
		Data    struct {
			Items []struct {
				Id   int    `json:"id"`
				Name string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(listRecorder.Body.Bytes(), &listPayload), listRecorder.Body.String())
	require.True(t, listPayload.Success, listRecorder.Body.String())
	require.Len(t, listPayload.Data.Items, 1)
	assert.Equal(t, visible.Id, listPayload.Data.Items[0].Id)

	hiddenRecorder := httptest.NewRecorder()
	hiddenCtx, _ := gin.CreateTestContext(hiddenRecorder)
	hiddenCtx.Set("id", 8)
	hiddenCtx.Set("role", common.RoleAdminUser)
	hiddenCtx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", hidden.Id)}}
	hiddenCtx.Request = httptest.NewRequest(http.MethodGet, "/api/channel/"+fmt.Sprintf("%d", hidden.Id), nil)
	GetChannel(hiddenCtx)
	assert.Contains(t, hiddenRecorder.Body.String(), "auth.insufficient_privilege")

	visibleRecorder := httptest.NewRecorder()
	visibleCtx, _ := gin.CreateTestContext(visibleRecorder)
	visibleCtx.Set("id", 8)
	visibleCtx.Set("role", common.RoleAdminUser)
	visibleCtx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", visible.Id)}}
	visibleCtx.Request = httptest.NewRequest(http.MethodGet, "/api/channel/"+fmt.Sprintf("%d", visible.Id), nil)
	GetChannel(visibleCtx)
	assert.Contains(t, visibleRecorder.Body.String(), `"success":true`)
}

func TestAdminUserScopeHidesUnassignedUsers(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	visible := &model.User{Username: "scope-visible", Password: "password1", Role: common.RoleCommonUser, AffCode: "aff-v"}
	hidden := &model.User{Username: "scope-hidden", Password: "password1", Role: common.RoleCommonUser, AffCode: "aff-h"}
	require.NoError(t, db.Create(visible).Error)
	require.NoError(t, db.Create(hidden).Error)
	require.NoError(t, db.Create(&model.AdminScope{
		UserId: 8,
		Kind:   model.AdminScopeKindUser,
		Mode:   model.AdminScopeModeAssigned,
	}).Error)
	require.NoError(t, db.Create(&model.AdminScopeItem{
		UserId: 8,
		Kind:   model.AdminScopeKindUser,
		ItemId: visible.Id,
	}).Error)

	listRecorder := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listRecorder)
	listCtx.Set("id", 8)
	listCtx.Set("role", common.RoleAdminUser)
	listCtx.Request = httptest.NewRequest(http.MethodGet, "/api/user/?page_size=50", nil)
	GetAllUsers(listCtx)

	var listPayload struct {
		Success bool `json:"success"`
		Data    struct {
			Items []struct {
				Id       int    `json:"id"`
				Username string `json:"username"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(listRecorder.Body.Bytes(), &listPayload), listRecorder.Body.String())
	require.True(t, listPayload.Success, listRecorder.Body.String())
	require.Len(t, listPayload.Data.Items, 1)
	assert.Equal(t, visible.Username, listPayload.Data.Items[0].Username)

	hiddenRecorder := httptest.NewRecorder()
	hiddenCtx, _ := gin.CreateTestContext(hiddenRecorder)
	hiddenCtx.Set("id", 8)
	hiddenCtx.Set("role", common.RoleAdminUser)
	hiddenCtx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", hidden.Id)}}
	hiddenCtx.Request = httptest.NewRequest(http.MethodGet, "/api/user/"+fmt.Sprintf("%d", hidden.Id), nil)
	GetUser(hiddenCtx)
	assert.Contains(t, hiddenRecorder.Body.String(), "auth.insufficient_privilege")
}

func TestDisableTagChannelsOnlyTouchesAssignedChannels(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	tag := "scope-tag"
	visible := &model.Channel{Name: "tag-visible", Key: "k1", Tag: &tag, Status: common.ChannelStatusEnabled}
	hidden := &model.Channel{Name: "tag-hidden", Key: "k2", Tag: &tag, Status: common.ChannelStatusEnabled}
	require.NoError(t, db.Create(visible).Error)
	require.NoError(t, db.Create(hidden).Error)
	require.NoError(t, db.Create(&model.AdminScope{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		Mode:   model.AdminScopeModeAssigned,
	}).Error)
	require.NoError(t, db.Create(&model.AdminScopeItem{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		ItemId: visible.Id,
	}).Error)

	body, err := common.Marshal(ChannelTag{Tag: tag})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 8)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/channel/tag/disabled", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	DisableTagChannels(ctx)
	require.Contains(t, recorder.Body.String(), `"success":true`)

	var storedVisible, storedHidden model.Channel
	require.NoError(t, db.Select("status").First(&storedVisible, visible.Id).Error)
	require.NoError(t, db.Select("status").First(&storedHidden, hidden.Id).Error)
	assert.Equal(t, common.ChannelStatusManuallyDisabled, storedVisible.Status)
	assert.Equal(t, common.ChannelStatusEnabled, storedHidden.Status)
}

func TestDeleteDisabledChannelsOnlyDeletesAssigned(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	visible := &model.Channel{Name: "disabled-visible", Key: "k1", Status: common.ChannelStatusManuallyDisabled}
	hidden := &model.Channel{Name: "disabled-hidden", Key: "k2", Status: common.ChannelStatusManuallyDisabled}
	require.NoError(t, db.Create(visible).Error)
	require.NoError(t, db.Create(hidden).Error)
	require.NoError(t, db.Create(&model.AdminScope{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		Mode:   model.AdminScopeModeAssigned,
	}).Error)
	require.NoError(t, db.Create(&model.AdminScopeItem{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		ItemId: visible.Id,
	}).Error)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 8)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/api/channel/disabled", nil)
	DeleteDisabledChannel(ctx)
	require.Contains(t, recorder.Body.String(), `"success":true`)

	var count int64
	require.NoError(t, db.Model(&model.Channel{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
	var remaining model.Channel
	require.NoError(t, db.First(&remaining, hidden.Id).Error)
	assert.Equal(t, hidden.Name, remaining.Name)
}

func TestEnabledListModelsRespectsChannelScope(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	visible := &model.Channel{
		Name:   "models-visible",
		Key:    "k1",
		Models: "gpt-visible",
		Group:  "default",
		Status: common.ChannelStatusEnabled,
	}
	hidden := &model.Channel{
		Name:   "models-hidden",
		Key:    "k2",
		Models: "gpt-hidden",
		Group:  "default",
		Status: common.ChannelStatusEnabled,
	}
	require.NoError(t, db.Create(visible).Error)
	require.NoError(t, db.Create(hidden).Error)
	require.NoError(t, visible.AddAbilities(nil))
	require.NoError(t, hidden.AddAbilities(nil))
	require.NoError(t, db.Create(&model.AdminScope{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		Mode:   model.AdminScopeModeAssigned,
	}).Error)
	require.NoError(t, db.Create(&model.AdminScopeItem{
		UserId: 8,
		Kind:   model.AdminScopeKindChannel,
		ItemId: visible.Id,
	}).Error)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 8)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/channel/models_enabled", nil)
	EnabledListModels(ctx)

	var payload struct {
		Success bool     `json:"success"`
		Data    []string `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload), recorder.Body.String())
	require.True(t, payload.Success, recorder.Body.String())
	assert.Contains(t, payload.Data, "gpt-visible")
	assert.NotContains(t, payload.Data, "gpt-hidden")
}

func TestFixAbilityInScopeDoesNotRebuildUnassignedChannels(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	visible := &model.Channel{Name: "fix-visible", Key: "k1", Models: "gpt-visible", Group: "default"}
	hidden := &model.Channel{Name: "fix-hidden", Key: "k2", Models: "gpt-hidden", Group: "default"}
	require.NoError(t, db.Create(visible).Error)
	require.NoError(t, db.Create(hidden).Error)
	require.NoError(t, visible.AddAbilities(nil))
	require.NoError(t, hidden.AddAbilities(nil))
	require.NoError(t, db.Model(&model.Ability{}).Where("channel_id = ?", hidden.Id).Update("model", "gpt-stale").Error)

	success, fails, err := model.FixAbilityInScope(model.DataScope{IDs: []int{visible.Id}})
	require.NoError(t, err)
	assert.Equal(t, 1, success)
	assert.Equal(t, 0, fails)

	var hiddenAbility model.Ability
	require.NoError(t, db.Where("channel_id = ?", hidden.Id).First(&hiddenAbility).Error)
	assert.Equal(t, "gpt-stale", hiddenAbility.Model)
}

func TestFilterChannelsByScopeKeepsAssignedOnly(t *testing.T) {
	channels := []*model.Channel{{Id: 1}, {Id: 2}, {Id: 3}}
	filtered := filterChannelsByScope(channels, model.DataScope{IDs: []int{2}})
	require.Len(t, filtered, 1)
	assert.Equal(t, 2, filtered[0].Id)
	assert.Equal(t, 3, len(filterChannelsByScope(channels, model.DataScope{All: true})))
	assert.Empty(t, filterChannelsByScope(channels, model.DataScope{}))
}
