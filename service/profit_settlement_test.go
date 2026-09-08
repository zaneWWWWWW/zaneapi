package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProfitRecordsRetainedPrepaymentWhenSettlementFails(t *testing.T) {
	setupProfitServiceDB(t)
	user := model.User{Username: "payer", Quota: 0}
	require.NoError(t, model.DB.Create(&user).Error)
	upstream := .7
	info := &relaycommon.RelayInfo{UserId: user.Id, TokenUnlimited: true, ChannelMeta: &relaycommon.ChannelMeta{ChannelId: 1, UpstreamRatio: &upstream}, OriginModelName: "model", StartTime: time.Now(), PriceData: types.PriceData{ModelRatio: 1, GroupRatioInfo: types.GroupRatioInfo{GroupRatio: .9}}}
	info.Billing = &BillingSession{relayInfo: info, funding: &WalletFunding{userId: user.Id, consumed: 10}, preConsumedQuota: 10}
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set(common.RequestIdKey, "underpaid")
	PostTextConsumeQuota(ctx, info, &dto.Usage{PromptTokens: 100, TotalTokens: 100}, nil)
	var record model.ChannelProfitRecord
	require.NoError(t, model.DB.Where("event_key = ?", "underpaid").First(&record).Error)
	assert.Equal(t, int64(10), record.RevenueQuota)
	assert.Equal(t, int64(70), record.CostQuota)
	assert.Equal(t, int64(-60), record.ProfitQuota)
}

func TestProfitFundingSnapshotSurvivesTokenRefundFailure(t *testing.T) {
	setupProfitServiceDB(t)
	user := model.User{Username: "payer", Quota: 10}
	require.NoError(t, model.DB.Create(&user).Error)
	info := &relaycommon.RelayInfo{UserId: user.Id, TokenId: 999}
	info.Billing = &BillingSession{relayInfo: info, funding: &WalletFunding{userId: user.Id, consumed: 90}, preConsumedQuota: 90}
	ctx, _ := gin.CreateTestContext(nil)
	require.Error(t, SettleBilling(ctx, info, 10))
	require.NotNil(t, info.ProfitRevenueQuota)
	assert.Equal(t, 10, *info.ProfitRevenueQuota)
	require.NoError(t, model.DB.First(&user, user.Id).Error)
	assert.Equal(t, 90, user.Quota)
}

func TestProfitLegacyFundingSnapshotTracksRollback(t *testing.T) {
	setupProfitServiceDB(t)
	user := model.User{Username: "payer", Quota: 100}
	require.NoError(t, model.DB.Create(&user).Error)
	info := &relaycommon.RelayInfo{UserId: user.Id, TokenId: 999, FinalPreConsumedQuota: 10}
	ctx, _ := gin.CreateTestContext(nil)
	require.Error(t, SettleBilling(ctx, info, 90))
	require.NotNil(t, info.ProfitRevenueQuota)
	assert.Equal(t, 10, *info.ProfitRevenueQuota)
	require.NoError(t, model.DB.First(&user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
}

func TestProfitTaskSubmissionUsesActualFunding(t *testing.T) {
	setupProfitServiceDB(t)
	user := model.User{Username: "payer", Quota: 0}
	require.NoError(t, model.DB.Create(&user).Error)
	upstream := .7
	info := &relaycommon.RelayInfo{UserId: user.Id, TokenUnlimited: true, ChannelMeta: &relaycommon.ChannelMeta{ChannelId: 1, UpstreamRatio: &upstream}, OriginModelName: "model", PriceData: types.PriceData{Quota: 90, BaseQuota: 100, GroupRatioInfo: types.GroupRatioInfo{GroupRatio: .9}}}
	info.TaskRelayInfo = &relaycommon.TaskRelayInfo{PublicTaskID: "submit"}
	info.Billing = &BillingSession{relayInfo: info, funding: &WalletFunding{userId: user.Id, consumed: 10}, preConsumedQuota: 10}
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Request = httptest.NewRequest("POST", "/v1/videos", nil)
	require.Error(t, SettleBilling(ctx, info, 90))
	LogTaskConsumption(ctx, info)
	var record model.ChannelProfitRecord
	require.NoError(t, model.DB.Where("event_key = ?", "task:1:submit").First(&record).Error)
	assert.Equal(t, int64(10), record.RevenueQuota)
	assert.Equal(t, int64(70), record.CostQuota)
}

func TestProfitRefundRetriesLedgerWithoutRefundingMoneyAgain(t *testing.T) {
	for _, subscription := range []bool{false, true} {
		t.Run(map[bool]string{false: "wallet", true: "subscription"}[subscription], func(t *testing.T) {
			setupProfitServiceDB(t)
			user := model.User{Username: "payer", Quota: 100}
			require.NoError(t, model.DB.Create(&user).Error)
			upstream := .7
			model.RecordChannelProfit("refund", 1, "model", 900, 100, types.ProfitBasis{BaseQuota: 1000, GroupRatio: .9, UpstreamRatio: &upstream})
			task := model.Task{TaskID: "refund-task", UserId: user.Id, ChannelId: 1, Status: model.TaskStatusFailure, Quota: 900, PrivateData: model.TaskPrivateData{BillingContext: &model.TaskBillingContext{ProfitEventKey: "refund", GroupRatio: .9}}}
			var sub model.UserSubscription
			if subscription {
				sub = model.UserSubscription{UserId: user.Id, AmountTotal: 1000, AmountUsed: 900}
				require.NoError(t, model.DB.Create(&sub).Error)
				task.PrivateData.BillingSource = BillingSourceSubscription
				task.PrivateData.SubscriptionId = sub.Id
			}
			require.NoError(t, model.DB.Create(&task).Error)
			require.NoError(t, model.DB.Callback().Create().Before("gorm:create").Register("fail_profit_projection", func(tx *gorm.DB) {
				if tx.Statement.Table == "channel_profit_records" {
					tx.AddError(errors.New("ledger unavailable"))
				}
			}))
			t.Cleanup(func() { _ = model.DB.Callback().Create().Remove("fail_profit_projection") })
			require.True(t, RefundTaskQuota(context.Background(), &task, "upstream refunded"))
			var stored model.Task
			require.NoError(t, model.DB.First(&stored, task.ID).Error)
			assert.Zero(t, stored.Quota)
			var job model.ChannelProfitSettlement
			require.NoError(t, model.DB.First(&job, "event_key = ?", "refund").Error)
			assert.True(t, job.Pending)
			assert.Zero(t, job.RevenueQuota)
			// Drop the caller's state, as on restart. The durable queue drives retries.
			task = model.Task{}
			require.NoError(t, model.DB.Callback().Create().Remove("fail_profit_projection"))
			require.NoError(t, model.RetryPendingChannelProfits(context.Background(), 100))
			require.NoError(t, model.RetryPendingChannelProfits(context.Background(), 100))
			report, err := model.GetChannelProfitReport(1, common.GetTimestamp()+1)
			require.NoError(t, err)
			assert.Zero(t, report.RevenueQuota)
			assert.Zero(t, report.CostQuota)
			require.NoError(t, model.DB.First(&user, user.Id).Error)
			if subscription {
				require.NoError(t, model.DB.First(&sub, sub.Id).Error)
				assert.Zero(t, sub.AmountUsed)
				assert.Equal(t, 100, user.Quota)
			} else {
				assert.Equal(t, 1000, user.Quota)
			}
			require.NoError(t, model.DB.First(&job, "event_key = ?", "refund").Error)
			assert.False(t, job.Pending)
		})
	}
}

func TestProfitQueueFailureRollsBackFundingAndTaskQuota(t *testing.T) {
	setupProfitServiceDB(t)
	user := model.User{Username: "payer", Quota: 100}
	require.NoError(t, model.DB.Create(&user).Error)
	task := model.Task{TaskID: "queue-failure", UserId: user.Id, Quota: 900, PrivateData: model.TaskPrivateData{BillingContext: &model.TaskBillingContext{ProfitEventKey: "refund"}}}
	require.NoError(t, model.DB.Create(&task).Error)
	require.NoError(t, model.DB.Callback().Create().Before("gorm:create").Register("fail_profit_queue", func(tx *gorm.DB) {
		if tx.Statement.Table == "channel_profit_settlements" {
			tx.AddError(errors.New("queue unavailable"))
		}
	}))
	t.Cleanup(func() { _ = model.DB.Callback().Create().Remove("fail_profit_queue") })
	require.False(t, RefundTaskQuota(context.Background(), &task, "refund"))
	require.NoError(t, model.DB.First(&user, user.Id).Error)
	assert.Equal(t, 100, user.Quota)
	require.NoError(t, model.DB.First(&task, task.ID).Error)
	assert.Equal(t, 900, task.Quota)
}
