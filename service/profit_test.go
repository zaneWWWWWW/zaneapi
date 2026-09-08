package service

import (
	"context"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/billingexpr"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProfitTextBasisPreservesFreeAndDiscountedUsage(t *testing.T) {
	for _, group := range []float64{0, .9} {
		info := &relaycommon.RelayInfo{OriginModelName: "profit-text", StartTime: time.Now(), PriceData: types.PriceData{ModelRatio: 1, CompletionRatio: 2, CacheRatio: .1, GroupRatioInfo: types.GroupRatioInfo{GroupRatio: group}}}
		ctx, _ := gin.CreateTestContext(nil)
		usage := &dto.Usage{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 200}}
		summary := calculateTextQuotaSummary(ctx, info, usage)
		assert.Equal(t, 1020.0, summary.BaseQuota.InexactFloat64())
		assert.Equal(t, common.QuotaRound(1020*group), summary.Quota)
		info.PriceData.UsePrice = true
		info.PriceData.ModelPrice = .01
		info.PriceData.AddOtherRatio("n", 3)
		summary = calculateTextQuotaSummary(ctx, info, usage)
		assert.InDelta(t, .03*common.QuotaPerUnit, summary.BaseQuota.InexactFloat64(), .00001)
	}
}

func setupProfitServiceDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Channel{}, &model.ChannelProfitRecord{}, &model.ChannelProfitSettlement{}, &model.Task{}, &model.User{}, &model.Token{}, &model.Log{}, &model.UserSubscription{}))
	previousDB, previousLogDB := model.DB, model.LOG_DB
	model.DB, model.LOG_DB = db, db
	t.Cleanup(func() { model.DB, model.LOG_DB = previousDB, previousLogDB })
}

func TestProfitFreeTaskSettlementAndRefund(t *testing.T) {
	setupProfitServiceDB(t)
	upstream := .7
	model.RecordChannelProfit("task-free", 1, "model", 0, 100, types.ProfitBasis{BaseQuota: 1000, GroupRatio: 0, UpstreamRatio: &upstream})
	task := &model.Task{TaskID: "free", ChannelId: 1, Quota: 0, Group: "free", PrivateData: model.TaskPrivateData{BillingContext: &model.TaskBillingContext{ProfitEventKey: "task-free", ModelRatio: 1, GroupRatio: 0}}}
	RecalculateTaskQuotaByTokens(context.Background(), task, 2000)
	report, err := model.GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Equal(t, int64(1400), report.CostQuota)
	assert.Equal(t, int64(-1400), report.ProfitQuota)
	require.True(t, RefundTaskQuota(context.Background(), task, "upstream refunded"))
	report, err = model.GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Zero(t, report.CostQuota)
	assert.Zero(t, report.ProfitQuota)
}

func TestProfitTieredFreeUsageUsesActualStandardCost(t *testing.T) {
	setupProfitServiceDB(t)
	upstream := .7
	info := &relaycommon.RelayInfo{ChannelMeta: &relaycommon.ChannelMeta{ChannelId: 1, UpstreamRatio: &upstream}, OriginModelName: "tiered-free", StartTime: time.Now(), PriceData: types.PriceData{GroupRatioInfo: types.GroupRatioInfo{GroupRatio: 0}}, TieredBillingSnapshot: &billingexpr.BillingSnapshot{BillingMode: "tiered_expr", ExprString: "p * 2", GroupRatio: 0, QuotaPerUnit: common.QuotaPerUnit}}
	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set(common.RequestIdKey, "tiered-free")
	PostTextConsumeQuota(ctx, info, &dto.Usage{PromptTokens: 1000, TotalTokens: 1000}, nil)
	var record model.ChannelProfitRecord
	require.NoError(t, model.DB.First(&record).Error)
	assert.Equal(t, 1000.0, record.BaseQuota)
	assert.Zero(t, record.RevenueQuota)
	assert.Equal(t, int64(700), record.CostQuota)
}
