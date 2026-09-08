package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelProfitRecord is an immutable accounting event. UpstreamRatio is copied
// from the channel when the event is created so later configuration changes do
// not rewrite historical profit.
type ChannelProfitRecord struct {
	Id                int     `json:"id"`
	EventKey          string  `json:"event_key" gorm:"type:varchar(128);uniqueIndex"`
	ChannelId         int     `json:"channel_id" gorm:"index:idx_channel_profit_created,priority:1"`
	CreatedAt         int64   `json:"created_at" gorm:"bigint;index:idx_channel_profit_created,priority:2"`
	ModelName         string  `json:"model_name" gorm:"type:varchar(255)"`
	EventType         string  `json:"event_type" gorm:"type:varchar(16)"`
	RevenueQuota      int64   `json:"revenue_quota"`
	CostQuota         int64   `json:"cost_quota"`
	ProfitQuota       int64   `json:"profit_quota"`
	UpstreamRatio     float64 `json:"upstream_ratio"`
	BaseQuota         float64 `json:"base_quota"`
	UserGroupRatio    float64 `json:"user_group_ratio"`
	AccountingVersion int     `json:"accounting_version"`
	CostSaturated     bool    `json:"cost_saturated"`
	RootEventKey      string  `json:"root_event_key" gorm:"size:128;index"`
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

// RecordChannelProfit records the initial settled charge, including free usage.
func RecordChannelProfit(eventKey string, channelId int, modelName string, revenueQuota int64, createdAt int64, basis types.ProfitBasis) {
	if eventKey == "" || channelId <= 0 || basis.UpstreamRatio == nil {
		return
	}
	if len(eventKey) > 128 {
		sum := sha256.Sum256([]byte(eventKey))
		eventKey = hex.EncodeToString(sum[:])
	}
	if len([]rune(modelName)) > 255 {
		modelName = string([]rune(modelName)[:255])
	}
	ratio := *basis.UpstreamRatio
	if ratio < 0 || math.IsNaN(ratio) || math.IsInf(ratio, 0) || basis.BaseQuota < 0 || math.IsNaN(basis.BaseQuota) || math.IsInf(basis.BaseQuota, 0) || revenueQuota < 0 || basis.GroupRatio < 0 || math.IsNaN(basis.GroupRatio) || math.IsInf(basis.GroupRatio, 0) {
		common.SysError(fmt.Sprintf("invalid channel profit basis: channel_id=%d", channelId))
		return
	}
	costQuota, clamp := common.QuotaFromDecimalChecked(decimal.NewFromFloat(basis.BaseQuota).Mul(decimal.NewFromFloat(ratio)))
	record := ChannelProfitRecord{
		EventKey:          eventKey,
		ChannelId:         channelId,
		CreatedAt:         createdAt,
		ModelName:         modelName,
		EventType:         "consume",
		RevenueQuota:      revenueQuota,
		CostQuota:         int64(costQuota),
		ProfitQuota:       revenueQuota - int64(costQuota),
		UpstreamRatio:     ratio,
		BaseQuota:         basis.BaseQuota,
		UserGroupRatio:    basis.GroupRatio,
		AccountingVersion: 2,
		CostSaturated:     clamp != nil,
		RootEventKey:      eventKey,
	}
	if clamp != nil {
		common.SysError(fmt.Sprintf("channel profit cost saturated: event=%s %s", eventKey, clamp.Error()))
	}
	if err := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error; err != nil {
		common.SysError("failed to record channel profit: " + err.Error())
	}
}

// SetChannelProfitTotal records the difference from the last settled totals.
// Locking the original event serializes replays and preserves its cost snapshot,
// even after a channel is edited or deleted. Zero totals fully reverse the cost.
func setChannelProfitTotalInTx(tx *gorm.DB, eventKey string, baseQuota float64, revenueQuota int64) error {
	if eventKey == "" {
		return nil
	}
	if len(eventKey) > 128 {
		sum := sha256.Sum256([]byte(eventKey))
		eventKey = hex.EncodeToString(sum[:])
	}
	if baseQuota < 0 || math.IsNaN(baseQuota) || math.IsInf(baseQuota, 0) || revenueQuota < 0 {
		return fmt.Errorf("invalid profit settlement totals")
	}
	var original ChannelProfitRecord
	if err := lockForUpdate(tx).Where("event_key = ? AND accounting_version = ?", eventKey, 2).First(&original).Error; err != nil {
		return err
	}
	var totals struct {
		RevenueQuota int64
		CostQuota    int64
		BaseQuota    float64
		Events       int64
	}
	if err := tx.Model(&ChannelProfitRecord{}).Where("root_event_key = ?", eventKey).Select("SUM(revenue_quota) AS revenue_quota, SUM(cost_quota) AS cost_quota, SUM(base_quota) AS base_quota, COUNT(*) AS events").Scan(&totals).Error; err != nil {
		return err
	}
	cost, clamp := common.QuotaFromDecimalChecked(decimal.NewFromFloat(baseQuota).Mul(decimal.NewFromFloat(original.UpstreamRatio)))
	if clamp != nil {
		common.SysError(fmt.Sprintf("channel profit cost saturated: event=%s %s", eventKey, clamp.Error()))
	}
	if totals.RevenueQuota == revenueQuota && totals.CostQuota == int64(cost) && totals.BaseQuota == baseQuota {
		return nil
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:adjustment:%d", eventKey, totals.Events)))
	record := ChannelProfitRecord{EventKey: hex.EncodeToString(sum[:]), RootEventKey: eventKey, ChannelId: original.ChannelId, ModelName: original.ModelName, CreatedAt: common.GetTimestamp(), EventType: "adjustment", AccountingVersion: 2, UpstreamRatio: original.UpstreamRatio, UserGroupRatio: original.UserGroupRatio, BaseQuota: baseQuota - totals.BaseQuota, RevenueQuota: revenueQuota - totals.RevenueQuota, CostQuota: int64(cost) - totals.CostQuota}
	record.ProfitQuota = record.RevenueQuota - record.CostQuota
	record.CostSaturated = clamp != nil
	if record.RevenueQuota < 0 || record.CostQuota < 0 {
		record.EventType = "refund"
	}
	return tx.Create(&record).Error
}

func GetChannelProfitReport(startTimestamp int64, endTimestamp int64) (ChannelProfitReport, error) {
	report := ChannelProfitReport{Channels: make([]ChannelProfitByChannel, 0)}
	applyRange := func(query *gorm.DB) *gorm.DB {
		query = query.Where("accounting_version = ?", 2)
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
	if err := DB.Model(&Channel{}).Where("upstream_ratio IS NULL AND status = ?", common.ChannelStatusEnabled).Count(&report.UnconfiguredChannelCount).Error; err != nil {
		return report, err
	}
	return report, nil
}
