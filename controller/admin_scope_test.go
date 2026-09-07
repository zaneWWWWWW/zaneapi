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

func TestAdminBindInvalidateDeleteSubscriptionRejectUnassignedUser(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.UserSubscription{}))
	confirmPaymentComplianceForTest(t)

	visible := &model.User{Username: "sub-visible", Password: "password1", Role: common.RoleCommonUser, AffCode: "sub-v"}
	hidden := &model.User{Username: "sub-hidden", Password: "password1", Role: common.RoleCommonUser, AffCode: "sub-h"}
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

	hiddenSub := &model.UserSubscription{UserId: hidden.Id, PlanId: 1, Status: "active"}
	require.NoError(t, db.Create(hiddenSub).Error)

	bindBody, err := common.Marshal(AdminBindSubscriptionRequest{UserId: hidden.Id, PlanId: 1})
	require.NoError(t, err)
	bindRecorder := httptest.NewRecorder()
	bindCtx, _ := gin.CreateTestContext(bindRecorder)
	bindCtx.Set("id", 8)
	bindCtx.Set("role", common.RoleAdminUser)
	bindCtx.Request = httptest.NewRequest(http.MethodPost, "/api/subscription/admin/bind", bytes.NewReader(bindBody))
	bindCtx.Request.Header.Set("Content-Type", "application/json")
	AdminBindSubscription(bindCtx)
	assert.Contains(t, bindRecorder.Body.String(), "auth.insufficient_privilege")

	invalidateRecorder := httptest.NewRecorder()
	invalidateCtx, _ := gin.CreateTestContext(invalidateRecorder)
	invalidateCtx.Set("id", 8)
	invalidateCtx.Set("role", common.RoleAdminUser)
	invalidateCtx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", hiddenSub.Id)}}
	invalidateCtx.Request = httptest.NewRequest(http.MethodPost, "/api/subscription/admin/invalidate/"+fmt.Sprintf("%d", hiddenSub.Id), nil)
	AdminInvalidateUserSubscription(invalidateCtx)
	assert.Contains(t, invalidateRecorder.Body.String(), "auth.insufficient_privilege")

	deleteRecorder := httptest.NewRecorder()
	deleteCtx, _ := gin.CreateTestContext(deleteRecorder)
	deleteCtx.Set("id", 8)
	deleteCtx.Set("role", common.RoleAdminUser)
	deleteCtx.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", hiddenSub.Id)}}
	deleteCtx.Request = httptest.NewRequest(http.MethodDelete, "/api/subscription/admin/"+fmt.Sprintf("%d", hiddenSub.Id), nil)
	AdminDeleteUserSubscription(deleteCtx)
	assert.Contains(t, deleteRecorder.Body.String(), "auth.insufficient_privilege")

	var stored model.UserSubscription
	require.NoError(t, db.First(&stored, hiddenSub.Id).Error)
	assert.Equal(t, "active", stored.Status)
}

func TestAdminSensitiveUserOpsRejectUnassignedUser(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	visible := &model.User{Username: "sens-visible", Password: "password1", Role: common.RoleCommonUser, AffCode: "sens-v"}
	hidden := &model.User{Username: "sens-hidden", Password: "password1", Role: common.RoleCommonUser, AffCode: "sens-h"}
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

	hiddenID := fmt.Sprintf("%d", hidden.Id)
	cases := []struct {
		name   string
		params gin.Params
		handle func(*gin.Context)
	}{
		{
			name:   "oauth bindings",
			params: gin.Params{{Key: "id", Value: hiddenID}},
			handle: GetUserOAuthBindingsByAdmin,
		},
		{
			name:   "oauth unbind",
			params: gin.Params{{Key: "id", Value: hiddenID}, {Key: "provider_id", Value: "1"}},
			handle: UnbindCustomOAuthByAdmin,
		},
		{
			name:   "clear binding",
			params: gin.Params{{Key: "id", Value: hiddenID}, {Key: "binding_type", Value: "github"}},
			handle: AdminClearUserBinding,
		},
		{
			name:   "reset passkey",
			params: gin.Params{{Key: "id", Value: hiddenID}},
			handle: AdminResetPasskey,
		},
		{
			name:   "disable 2fa",
			params: gin.Params{{Key: "id", Value: hiddenID}},
			handle: AdminDisable2FA,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Set("id", 8)
			ctx.Set("role", common.RoleAdminUser)
			ctx.Params = tc.params
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			tc.handle(ctx)
			assert.Contains(t, recorder.Body.String(), "auth.insufficient_privilege")
		})
	}
}

func TestAdminTaskMidjourneyTopupListsHideUnassignedUsers(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Task{}, &model.Midjourney{}, &model.TopUp{}))

	visible := &model.User{Username: "list-visible", Password: "password1", Role: common.RoleCommonUser, AffCode: "list-v"}
	hidden := &model.User{Username: "list-hidden", Password: "password1", Role: common.RoleCommonUser, AffCode: "list-h"}
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

	require.NoError(t, db.Create(&model.Task{UserId: visible.Id, TaskID: "visible-task"}).Error)
	require.NoError(t, db.Create(&model.Task{UserId: hidden.Id, TaskID: "hidden-task"}).Error)
	require.NoError(t, db.Create(&model.Midjourney{UserId: visible.Id, MjId: "visible-mj"}).Error)
	require.NoError(t, db.Create(&model.Midjourney{UserId: hidden.Id, MjId: "hidden-mj"}).Error)
	require.NoError(t, db.Create(&model.TopUp{UserId: visible.Id, TradeNo: "visible-trade", Status: common.TopUpStatusPending}).Error)
	require.NoError(t, db.Create(&model.TopUp{UserId: hidden.Id, TradeNo: "hidden-trade", Status: common.TopUpStatusPending}).Error)

	type listPayload struct {
		Success bool `json:"success"`
		Data    struct {
			Total int `json:"total"`
			Items []struct {
				UserId  int    `json:"user_id"`
				TaskID  string `json:"task_id"`
				MjId    string `json:"mj_id"`
				TradeNo string `json:"trade_no"`
			} `json:"items"`
		} `json:"data"`
	}

	taskRecorder := httptest.NewRecorder()
	taskCtx, _ := gin.CreateTestContext(taskRecorder)
	taskCtx.Set("id", 8)
	taskCtx.Set("role", common.RoleAdminUser)
	taskCtx.Request = httptest.NewRequest(http.MethodGet, "/api/task/?page_size=50", nil)
	GetAllTask(taskCtx)
	var taskPayload listPayload
	require.NoError(t, common.Unmarshal(taskRecorder.Body.Bytes(), &taskPayload), taskRecorder.Body.String())
	require.True(t, taskPayload.Success, taskRecorder.Body.String())
	require.Equal(t, 1, taskPayload.Data.Total)
	require.Len(t, taskPayload.Data.Items, 1)
	assert.Equal(t, visible.Id, taskPayload.Data.Items[0].UserId)

	mjRecorder := httptest.NewRecorder()
	mjCtx, _ := gin.CreateTestContext(mjRecorder)
	mjCtx.Set("id", 8)
	mjCtx.Set("role", common.RoleAdminUser)
	mjCtx.Request = httptest.NewRequest(http.MethodGet, "/api/mj/?page_size=50", nil)
	GetAllMidjourney(mjCtx)
	var mjPayload listPayload
	require.NoError(t, common.Unmarshal(mjRecorder.Body.Bytes(), &mjPayload), mjRecorder.Body.String())
	require.True(t, mjPayload.Success, mjRecorder.Body.String())
	require.Equal(t, 1, mjPayload.Data.Total)
	require.Len(t, mjPayload.Data.Items, 1)
	assert.Equal(t, visible.Id, mjPayload.Data.Items[0].UserId)

	topupRecorder := httptest.NewRecorder()
	topupCtx, _ := gin.CreateTestContext(topupRecorder)
	topupCtx.Set("id", 8)
	topupCtx.Set("role", common.RoleAdminUser)
	topupCtx.Request = httptest.NewRequest(http.MethodGet, "/api/topup/?page_size=50", nil)
	GetAllTopUps(topupCtx)
	var topupPayload listPayload
	require.NoError(t, common.Unmarshal(topupRecorder.Body.Bytes(), &topupPayload), topupRecorder.Body.String())
	require.True(t, topupPayload.Success, topupRecorder.Body.String())
	require.Equal(t, 1, topupPayload.Data.Total)
	require.Len(t, topupPayload.Data.Items, 1)
	assert.Equal(t, visible.Id, topupPayload.Data.Items[0].UserId)

	searchRecorder := httptest.NewRecorder()
	searchCtx, _ := gin.CreateTestContext(searchRecorder)
	searchCtx.Set("id", 8)
	searchCtx.Set("role", common.RoleAdminUser)
	searchCtx.Request = httptest.NewRequest(http.MethodGet, "/api/topup/?page_size=50&keyword=hidden-trade", nil)
	GetAllTopUps(searchCtx)
	var searchPayload listPayload
	require.NoError(t, common.Unmarshal(searchRecorder.Body.Bytes(), &searchPayload), searchRecorder.Body.String())
	require.True(t, searchPayload.Success, searchRecorder.Body.String())
	assert.Equal(t, 0, searchPayload.Data.Total)
	assert.Empty(t, searchPayload.Data.Items)
}

func TestAdminCompleteTopUpRejectsUnassignedUser(t *testing.T) {
	db := setupModelListControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.TopUp{}))

	visible := &model.User{Username: "topup-visible", Password: "password1", Role: common.RoleCommonUser, AffCode: "top-v"}
	hidden := &model.User{Username: "topup-hidden", Password: "password1", Role: common.RoleCommonUser, AffCode: "top-h"}
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

	hiddenTopUp := &model.TopUp{UserId: hidden.Id, TradeNo: "hidden-complete", Status: common.TopUpStatusPending, Amount: 1}
	require.NoError(t, db.Create(hiddenTopUp).Error)

	body, err := common.Marshal(AdminCompleteTopupRequest{TradeNo: hiddenTopUp.TradeNo})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("id", 8)
	ctx.Set("role", common.RoleAdminUser)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/topup/complete", bytes.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	AdminCompleteTopUp(ctx)
	assert.Contains(t, recorder.Body.String(), "auth.insufficient_privilege")

	var stored model.TopUp
	require.NoError(t, db.Where("trade_no = ?", hiddenTopUp.TradeNo).First(&stored).Error)
	assert.Equal(t, common.TopUpStatusPending, stored.Status)
}
