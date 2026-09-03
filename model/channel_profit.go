package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelProfitRecord is an immutable accounting event. CostRatio is copied
// from the channel when the event is created so later configuration changes do
// not rewrite historical profit.
type ChannelProfitRecord struct {
	Id           int     `json:"id"`
	EventKey     string  `json:"event_key" gorm:"type:varchar(128);uniqueIndex"`
	ChannelId    int     `json:"channel_id" gorm:"index:idx_channel_profit_created,priority:1"`
	CreatedAt    int64   `json:"created_at" gorm:"bigint;index:idx_channel_profit_created,priority:2"`
	ModelName    string  `json:"model_name" gorm:"type:varchar(255)"`
	EventType    string  `json:"event_type" gorm:"type:varchar(16)"`
	RevenueQuota int64   `json:"revenue_quota"`
	CostQuota    int64   `json:"cost_quota"`
	ProfitQuota  int64   `json:"profit_quota"`
	CostRatio    float64 `json:"cost_ratio"`
}

type ChannelProfitSummary struct {
	RevenueQuota int64 `json:"revenue_quota"`
	CostQuota    int64 `json:"cost_quota"`
	ProfitQuota  int64 `json:"profit_quota"`
	RequestCount int64 `json:"request_count"`
}

type ChannelProfitByChannel struct {
	ChannelProfitSummary
	ChannelId   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
}

type ChannelProfitReport struct {
	ChannelProfitSummary
	Channels                 []ChannelProfitByChannel `json:"channels"`
	UnconfiguredChannelCount int64                    `json:"unconfigured_channel_count"`
}

// RecordChannelProfit records a settled revenue event. Refunds use a negative
// revenue amount and therefore reverse both revenue and cost.
func RecordChannelProfit(eventKey string, channelId int, modelName string, revenueQuota int64, createdAt int64) {
	if eventKey == "" || channelId <= 0 || revenueQuota == 0 {
		return
	}
	if len(eventKey) > 128 {
		sum := sha256.Sum256([]byte(eventKey))
		eventKey = hex.EncodeToString(sum[:])
	}
	if len(modelName) > 255 {
		modelName = modelName[:255]
	}
	var channel Channel
	if err := DB.Select("cost_ratio").First(&channel, channelId).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			common.SysError(fmt.Sprintf("failed to load channel profit configuration: channel_id=%d error=%s", channelId, err.Error()))
		}
		return
	}
	if channel.CostRatio == nil {
		return
	}
	ratio := *channel.CostRatio
	if ratio < 0 || ratio > 1 || math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		common.SysError(fmt.Sprintf("invalid channel profit cost ratio: channel_id=%d ratio=%g", channelId, ratio))
		return
	}
	costQuota := int64(common.QuotaFromDecimal(decimal.NewFromInt(revenueQuota).Mul(decimal.NewFromFloat(ratio))))
	record := ChannelProfitRecord{
		EventKey:     eventKey,
		ChannelId:    channelId,
		CreatedAt:    createdAt,
		ModelName:    modelName,
		EventType:    "consume",
		RevenueQuota: revenueQuota,
		CostQuota:    costQuota,
		ProfitQuota:  revenueQuota - costQuota,
		CostRatio:    ratio,
	}
	if revenueQuota < 0 {
		record.EventType = "refund"
	}
	if err := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error; err != nil {
		common.SysError("failed to record channel profit: " + err.Error())
	}
}

func GetChannelProfitReport(startTimestamp int64, endTimestamp int64) (ChannelProfitReport, error) {
	report := ChannelProfitReport{Channels: make([]ChannelProfitByChannel, 0)}
	applyRange := func(query *gorm.DB) *gorm.DB {
		if startTimestamp > 0 {
			query = query.Where("created_at >= ?", startTimestamp)
		}
		if endTimestamp > 0 {
			query = query.Where("created_at <= ?", endTimestamp)
		}
		return query
	}
	query := applyRange(DB.Model(&ChannelProfitRecord{}))
	if err := query.Select("COALESCE(SUM(revenue_quota), 0) AS revenue_quota, COALESCE(SUM(cost_quota), 0) AS cost_quota, COALESCE(SUM(profit_quota), 0) AS profit_quota, COALESCE(SUM(CASE WHEN event_type = 'consume' THEN 1 ELSE 0 END), 0) AS request_count").Scan(&report.ChannelProfitSummary).Error; err != nil {
		return report, err
	}
	channelQuery := applyRange(DB.Model(&ChannelProfitRecord{}))
	if err := channelQuery.Select("channel_profit_records.channel_id, COALESCE(channels.name, '') AS channel_name, COALESCE(SUM(channel_profit_records.revenue_quota), 0) AS revenue_quota, COALESCE(SUM(channel_profit_records.cost_quota), 0) AS cost_quota, COALESCE(SUM(channel_profit_records.profit_quota), 0) AS profit_quota, COALESCE(SUM(CASE WHEN channel_profit_records.event_type = 'consume' THEN 1 ELSE 0 END), 0) AS request_count").Joins("LEFT JOIN channels ON channels.id = channel_profit_records.channel_id").Group("channel_profit_records.channel_id, channels.name").Order("SUM(channel_profit_records.profit_quota) DESC").Scan(&report.Channels).Error; err != nil {
		return report, err
	}
	if err := DB.Model(&Channel{}).Where("cost_ratio IS NULL AND status = ?", common.ChannelStatusEnabled).Count(&report.UnconfiguredChannelCount).Error; err != nil {
		return report, err
	}
	return report, nil
}
