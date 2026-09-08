package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"math"
	"strings"
	"testing"
)

func setupChannelProfitDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Channel{}, &ChannelProfitRecord{}, &ChannelProfitSettlement{}))
	originalDB, originalLogDB := DB, LOG_DB
	DB, LOG_DB = db, db
	t.Cleanup(func() { DB, LOG_DB = originalDB, originalLogDB })
}

func TestChannelProfitUsesStandardCost(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		upstream, group       float64
		revenue, cost, profit int64
	}{
		{"discounted", .7, .9, 900, 700, 200},
		{"free upstream", 0, .9, 900, 0, 900},
		{"free user", .7, 0, 0, 700, -700},
		{"both free", 0, 0, 0, 0, 0},
		{"above one", 1.2, 1.5, 1500, 1200, 300},
		{"loss", 1.2, .9, 900, 1200, -300},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setupChannelProfitDB(t)
			basis := types.ProfitBasis{BaseQuota: 1000, GroupRatio: tc.group, UpstreamRatio: &tc.upstream}
			RecordChannelProfit("request", 1, "model", tc.revenue, 100, basis)
			RecordChannelProfit("request", 1, "model", tc.revenue, 100, basis)
			report, err := GetChannelProfitReport(1, 200)
			require.NoError(t, err)
			assert.Equal(t, tc.revenue, report.RevenueQuota)
			assert.Equal(t, tc.cost, report.CostQuota)
			assert.Equal(t, tc.profit, report.ProfitQuota)
			assert.Equal(t, int64(1), report.RequestCount)
		})
	}
}

func TestChannelProfitRefundUsesOriginalSnapshot(t *testing.T) {
	setupChannelProfitDB(t)
	ratio := .7
	channel := Channel{Name: "original", UpstreamRatio: &ratio}
	require.NoError(t, DB.Create(&channel).Error)
	RecordChannelProfit("task", channel.Id, "model", 900, 100, types.ProfitBasis{BaseQuota: 1000, GroupRatio: .9, UpstreamRatio: &ratio})
	require.NoError(t, DB.Model(&channel).Update("upstream_ratio", .2).Error)
	require.NoError(t, SetChannelProfitTotal("task", 2000, 1800))
	require.NoError(t, SetChannelProfitTotal("task", 2000, 1800))
	report, err := GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Equal(t, int64(1400), report.CostQuota)
	assert.Equal(t, int64(1), report.RequestCount)
	require.NoError(t, DB.Delete(&channel).Error)
	require.NoError(t, SetChannelProfitTotal("task", 0, 0))
	require.NoError(t, SetChannelProfitTotal("task", 0, 0))
	report, err = GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Zero(t, report.RevenueQuota)
	assert.Zero(t, report.CostQuota)
	assert.Zero(t, report.ProfitQuota)
	var count int64
	require.NoError(t, DB.Model(&ChannelProfitRecord{}).Count(&count).Error)
	assert.Equal(t, int64(3), count)
}

func TestChannelProfitRoundingAndRefund(t *testing.T) {
	setupChannelProfitDB(t)
	ratio := .7
	RecordChannelProfit("tiny", 1, "model", 1, 100, types.ProfitBasis{BaseQuota: 1.4, GroupRatio: .9, UpstreamRatio: &ratio})
	require.NoError(t, SetChannelProfitTotal("tiny", .6, 1))
	require.NoError(t, SetChannelProfitTotal("tiny", 0, 0))
	report, err := GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Zero(t, report.CostQuota)
	assert.Zero(t, report.ProfitQuota)
}

func TestChannelProfitIndependentOfUsageLogAvailability(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "log database unavailable", false: "logs disabled"}[enabled], func(t *testing.T) {
			setupChannelProfitDB(t)
			previous := common.LogConsumeEnabled
			common.LogConsumeEnabled = enabled
			t.Cleanup(func() { common.LogConsumeEnabled = previous })
			ratio := .7
			ctx, _ := gin.CreateTestContext(nil)
			ctx.Set(common.RequestIdKey, "consume")
			params := RecordConsumeLogParams{ChannelId: 1, ModelName: "model", Quota: 900, ProfitBasis: &types.ProfitBasis{BaseQuota: 1000, GroupRatio: .9, UpstreamRatio: &ratio}}
			RecordConsumeLog(ctx, 1, params)
			params.SkipProfit = true
			params.ProfitEventKey = "channel-test"
			RecordConsumeLog(ctx, 1, params)
			report, err := GetChannelProfitReport(1, common.GetTimestamp()+1)
			require.NoError(t, err)
			assert.Equal(t, int64(900), report.RevenueQuota)
			assert.Equal(t, int64(700), report.CostQuota)
			assert.Equal(t, int64(1), report.RequestCount)
		})
	}
}

func TestChannelProfitExcludesLegacyAndUnconfiguredRecords(t *testing.T) {
	setupChannelProfitDB(t)
	require.NoError(t, DB.Create(&Channel{Name: "unconfigured", Status: common.ChannelStatusEnabled}).Error)
	require.NoError(t, DB.Create(&Channel{Name: "disabled", Status: common.ChannelStatusManuallyDisabled}).Error)
	require.NoError(t, DB.Create(&ChannelProfitRecord{EventKey: "legacy", RevenueQuota: 1000, CostQuota: 750, ProfitQuota: 250, CreatedAt: 100}).Error)
	RecordChannelProfit("unconfigured", 1, "model", 900, 100, types.ProfitBasis{BaseQuota: 1000, GroupRatio: .9})
	report, err := GetChannelProfitReport(1, 200)
	require.NoError(t, err)
	assert.Empty(t, report.Channels)
	assert.Zero(t, report.RevenueQuota)
	assert.Equal(t, int64(1), report.UnconfiguredChannelCount)
}

func TestValidateProfitSettings(t *testing.T) {
	for _, ratio := range []float64{0, .7, 1, 1.2} {
		require.NoError(t, (&Channel{UpstreamRatio: &ratio}).ValidateProfitSettings())
	}
	for _, ratio := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		require.Error(t, (&Channel{UpstreamRatio: &ratio}).ValidateProfitSettings())
	}
}

func TestChannelProfitRejectsInvalidBasis(t *testing.T) {
	setupChannelProfitDB(t)
	for _, ratio := range []float64{-1, math.NaN(), math.Inf(1)} {
		RecordChannelProfit("invalid", 1, "model", 900, 100, types.ProfitBasis{BaseQuota: 1000, GroupRatio: .9, UpstreamRatio: &ratio})
	}
	report, err := GetChannelProfitReport(1, 200)
	require.NoError(t, err)
	assert.Empty(t, report.Channels)
}

func TestChannelProfitTimeRangeAndOversizedKeys(t *testing.T) {
	setupChannelProfitDB(t)
	ratio := .7
	basis := types.ProfitBasis{BaseQuota: 1000, GroupRatio: .9, UpstreamRatio: &ratio}
	key := strings.Repeat("k", 200)
	RecordChannelProfit(key, 1, strings.Repeat("模", 300), 900, 100, basis)
	RecordChannelProfit(key, 1, "model", 900, 100, basis)
	RecordChannelProfit("outside", 1, "model", 900, 300, basis)
	report, err := GetChannelProfitReport(100, 200)
	require.NoError(t, err)
	assert.Equal(t, int64(900), report.RevenueQuota)
	require.NoError(t, SetChannelProfitTotal(key, 0, 0))
}
