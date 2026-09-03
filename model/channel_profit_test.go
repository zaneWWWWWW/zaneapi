package model

import (
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupChannelProfitDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Channel{}, &ChannelProfitRecord{}))

	originalDB, originalLogDB := DB, LOG_DB
	DB, LOG_DB = db, db
	t.Cleanup(func() {
		DB, LOG_DB = originalDB, originalLogDB
	})
}

func TestChannelProfitRecordsSnapshotCostAndReverseRefunds(t *testing.T) {
	setupChannelProfitDB(t)

	ratio := 0.75
	channel := &Channel{Name: "profit-channel", CostRatio: &ratio}
	require.NoError(t, DB.Create(channel).Error)

	RecordChannelProfit("consume-1", channel.Id, "gpt-test", 1000, 100)
	RecordChannelProfit("consume-1", channel.Id, "gpt-test", 1000, 100)
	RecordChannelProfit("refund-1", channel.Id, "gpt-test", -200, 101)

	var consumed ChannelProfitRecord
	require.NoError(t, DB.Where("event_key = ?", "consume-1").First(&consumed).Error)
	assert.Equal(t, 0.75, consumed.CostRatio)
	assert.Equal(t, int64(750), consumed.CostQuota)

	report, err := GetChannelProfitReport(1, 200)
	require.NoError(t, err)
	require.Len(t, report.Channels, 1)
	assert.Equal(t, int64(800), report.RevenueQuota)
	assert.Equal(t, int64(600), report.CostQuota)
	assert.Equal(t, int64(200), report.ProfitQuota)
	assert.Equal(t, int64(1), report.RequestCount)
	assert.Equal(t, int64(1), report.Channels[0].RequestCount)
	assert.Equal(t, "profit-channel", report.Channels[0].ChannelName)
}

func TestChannelProfitSkipsChannelsWithoutCostRatio(t *testing.T) {
	setupChannelProfitDB(t)

	channel := &Channel{Name: "unconfigured-channel", Status: common.ChannelStatusEnabled}
	require.NoError(t, DB.Create(channel).Error)
	disabled := &Channel{Name: "disabled-unconfigured", Status: common.ChannelStatusManuallyDisabled}
	require.NoError(t, DB.Create(disabled).Error)

	RecordChannelProfit("consume-1", channel.Id, "gpt-test", 1000, 100)
	report, err := GetChannelProfitReport(1, 200)
	require.NoError(t, err)
	assert.Empty(t, report.Channels)
	assert.Equal(t, int64(1), report.UnconfiguredChannelCount)
}

func TestChannelProfitRecordsConsumptionWhenUsageLogsAreDisabled(t *testing.T) {
	setupChannelProfitDB(t)
	originalLogConsumeEnabled := common.LogConsumeEnabled
	common.LogConsumeEnabled = false
	t.Cleanup(func() { common.LogConsumeEnabled = originalLogConsumeEnabled })

	ratio := 0.4
	channel := &Channel{Name: "no-log-channel", CostRatio: &ratio}
	require.NoError(t, DB.Create(channel).Error)

	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set(common.RequestIdKey, "no-log-consume")
	RecordConsumeLog(ctx, 1, RecordConsumeLogParams{
		ChannelId: channel.Id,
		ModelName: "gpt-test",
		Quota:     1000,
	})

	report, err := GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), report.RevenueQuota)
	assert.Equal(t, int64(400), report.CostQuota)
}

func TestChannelProfitSkipsChannelTests(t *testing.T) {
	setupChannelProfitDB(t)

	ratio := 0.5
	channel := &Channel{Name: "test-channel", CostRatio: &ratio}
	require.NoError(t, DB.Create(channel).Error)

	ctx, _ := gin.CreateTestContext(nil)
	ctx.Set(common.RequestIdKey, "channel-test")
	RecordConsumeLog(ctx, 1, RecordConsumeLogParams{
		ChannelId:  channel.Id,
		ModelName:  "gpt-test",
		Quota:      1000,
		SkipProfit: true,
	})

	report, err := GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), report.RevenueQuota)
	assert.Empty(t, report.Channels)
}

func TestChannelProfitTaskBillingUsesStableEventKey(t *testing.T) {
	setupChannelProfitDB(t)
	originalLogConsumeEnabled := common.LogConsumeEnabled
	common.LogConsumeEnabled = false
	t.Cleanup(func() { common.LogConsumeEnabled = originalLogConsumeEnabled })

	ratio := 0.5
	channel := &Channel{Name: "task-channel", CostRatio: &ratio}
	require.NoError(t, DB.Create(channel).Error)

	params := RecordTaskBillingLogParams{
		LogType:   LogTypeConsume,
		ChannelId: channel.Id,
		ModelName: "gpt-test",
		Quota:     800,
		Other:     map[string]interface{}{"task_id": "task-1"},
	}
	RecordTaskBillingLog(params)
	RecordTaskBillingLog(params)

	report, err := GetChannelProfitReport(1, common.GetTimestamp()+1)
	require.NoError(t, err)
	assert.Equal(t, int64(800), report.RevenueQuota)
	assert.Equal(t, int64(400), report.CostQuota)
	assert.Equal(t, int64(1), report.RequestCount)
}

func TestValidateProfitSettingsRejectsRatioAboveOne(t *testing.T) {
	ratio := 1.5
	channel := &Channel{CostRatio: &ratio}
	require.EqualError(t, channel.ValidateProfitSettings(), "cost ratio must be between 0 and 1")

	valid := 1.0
	channel.CostRatio = &valid
	require.NoError(t, channel.ValidateProfitSettings())
}

func TestChannelProfitRecordsZeroAndFullCostRatio(t *testing.T) {
	setupChannelProfitDB(t)

	zero := 0.0
	full := 1.0
	free := &Channel{Name: "free-channel", CostRatio: &zero}
	costly := &Channel{Name: "costly-channel", CostRatio: &full}
	require.NoError(t, DB.Create(free).Error)
	require.NoError(t, DB.Create(costly).Error)

	RecordChannelProfit("free-1", free.Id, "gpt-test", 1000, 100)
	RecordChannelProfit("cost-1", costly.Id, "gpt-test", 1000, 100)

	report, err := GetChannelProfitReport(1, 200)
	require.NoError(t, err)
	assert.Equal(t, int64(2000), report.RevenueQuota)
	assert.Equal(t, int64(1000), report.CostQuota)
	assert.Equal(t, int64(1000), report.ProfitQuota)
}

func TestChannelProfitSkipsInvalidRatioAtRecordTime(t *testing.T) {
	setupChannelProfitDB(t)

	ratio := 1.25
	channel := &Channel{Name: "invalid-ratio", CostRatio: &ratio}
	require.NoError(t, DB.Create(channel).Error)

	RecordChannelProfit("invalid-1", channel.Id, "gpt-test", 1000, 100)
	report, err := GetChannelProfitReport(1, 200)
	require.NoError(t, err)
	assert.Equal(t, int64(0), report.RevenueQuota)
	assert.Empty(t, report.Channels)
}

func TestChannelProfitHashesOversizedEventKeys(t *testing.T) {
	setupChannelProfitDB(t)

	ratio := 0.5
	channel := &Channel{Name: "long-key-channel", CostRatio: &ratio}
	require.NoError(t, DB.Create(channel).Error)

	eventKey := strings.Repeat("k", 200)
	RecordChannelProfit(eventKey, channel.Id, strings.Repeat("m", 300), 500, 100)
	RecordChannelProfit(eventKey, channel.Id, "ignored", 500, 100)

	var count int64
	require.NoError(t, DB.Model(&ChannelProfitRecord{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	var record ChannelProfitRecord
	require.NoError(t, DB.First(&record).Error)
	assert.LessOrEqual(t, len(record.EventKey), 128)
	assert.LessOrEqual(t, len(record.ModelName), 255)
}

func TestChannelProfitReportRespectsTimeRange(t *testing.T) {
	setupChannelProfitDB(t)

	ratio := 0.5
	channel := &Channel{Name: "ranged-channel", CostRatio: &ratio}
	require.NoError(t, DB.Create(channel).Error)
	RecordChannelProfit("old", channel.Id, "gpt-test", 1000, 50)
	RecordChannelProfit("in-range", channel.Id, "gpt-test", 400, 150)
	RecordChannelProfit("new", channel.Id, "gpt-test", 800, 300)

	report, err := GetChannelProfitReport(100, 200)
	require.NoError(t, err)
	assert.Equal(t, int64(400), report.RevenueQuota)
	assert.Equal(t, int64(200), report.CostQuota)
	assert.Equal(t, int64(1), report.RequestCount)
}
