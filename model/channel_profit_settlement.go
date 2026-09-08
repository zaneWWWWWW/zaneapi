package model

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelProfitSettlement stores the latest committed funding totals awaiting
// projection into the profit ledger. Replaying it never touches user balances.
type ChannelProfitSettlement struct {
	EventKey     string `gorm:"primaryKey;size:128"`
	BaseQuota    float64
	RevenueQuota int64
	Pending      bool  `gorm:"index:idx_profit_pending,priority:1"`
	AttemptedAt  int64 `gorm:"index:idx_profit_pending,priority:2"`
}

func queueChannelProfitTotal(tx *gorm.DB, eventKey string, baseQuota float64, revenueQuota int64) error {
	if eventKey == "" {
		return nil
	}
	if baseQuota < 0 || math.IsNaN(baseQuota) || math.IsInf(baseQuota, 0) || revenueQuota < 0 || revenueQuota > common.MaxQuota {
		return fmt.Errorf("invalid profit settlement totals")
	}
	if len(eventKey) > 128 {
		hash := sha256.Sum256([]byte(eventKey))
		eventKey = hex.EncodeToString(hash[:])
	}
	job := ChannelProfitSettlement{EventKey: eventKey, BaseQuota: baseQuota, RevenueQuota: revenueQuota, Pending: true}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_key"}}, DoUpdates: clause.AssignmentColumns([]string{"base_quota", "revenue_quota", "pending", "attempted_at"})}).Create(&job).Error
}

// SetChannelProfitTotal durably records intent before attempting ledger writes.
func SetChannelProfitTotal(eventKey string, baseQuota float64, revenueQuota int64) error {
	if eventKey == "" {
		return nil
	}
	if err := queueChannelProfitTotal(DB, eventKey, baseQuota, revenueQuota); err != nil {
		return err
	}
	return ApplyChannelProfitSettlement(eventKey)
}

func ApplyChannelProfitSettlement(eventKey string) error {
	if eventKey == "" {
		return nil
	}
	if len(eventKey) > 128 {
		hash := sha256.Sum256([]byte(eventKey))
		eventKey = hex.EncodeToString(hash[:])
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		var job ChannelProfitSettlement
		if err := lockForUpdate(tx).Where("event_key = ?", eventKey).First(&job).Error; err != nil {
			return err
		}
		if !job.Pending {
			return nil
		}
		if err := setChannelProfitTotalInTx(tx, job.EventKey, job.BaseQuota, job.RevenueQuota); err != nil {
			return err
		}
		return tx.Model(&job).Update("pending", false).Error
	})
	if err != nil {
		// Retain Pending=true. Advancing this time lets a bounded sweep make
		// progress past a permanently broken event without discarding it.
		if updateErr := DB.Model(&ChannelProfitSettlement{}).Where("event_key = ? AND pending = ?", eventKey, true).Update("attempted_at", common.GetTimestamp()).Error; updateErr != nil {
			common.SysError("failed to update profit retry time: " + updateErr.Error())
		}
	}
	return err
}

func RetryPendingChannelProfits(ctx context.Context, limit int) error {
	if limit <= 0 {
		return nil
	}
	var jobs []ChannelProfitSettlement
	if err := DB.WithContext(ctx).Where("pending = ?", true).Order("attempted_at asc, event_key asc").Limit(limit).Find(&jobs).Error; err != nil {
		return err
	}
	var lastErr error
	for _, job := range jobs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := ApplyChannelProfitSettlement(job.EventKey); err != nil {
			common.SysError(fmt.Sprintf("profit retry failed: event=%s error=%s", job.EventKey, err))
			lastErr = err
		}
	}
	return lastErr
}

type ProfitFundingAdjustment struct {
	EventKey       string
	UserID         int
	SubscriptionID int
	TaskID         int64
	Delta          int
	BaseQuota      float64
	RevenueQuota   int64
}

// AdjustFundingWithProfit commits funding, task quota and profit intent in one
// transaction. Ledger errors can then be retried without transferring money.
func AdjustFundingWithProfit(adjustment ProfitFundingAdjustment) error {
	if adjustment.EventKey == "" {
		return fmt.Errorf("missing profit event key")
	}
	if adjustment.Delta > common.MaxQuota || adjustment.Delta < -common.MaxQuota {
		return fmt.Errorf("funding delta out of range")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		if adjustment.TaskID > 0 {
			var task Task
			if err := lockForUpdate(tx).Select("id").Where("id = ?", adjustment.TaskID).First(&task).Error; err != nil {
				return err
			}
		}
		if adjustment.Delta != 0 {
			if adjustment.SubscriptionID > 0 {
				var sub UserSubscription
				if err := lockForUpdate(tx).Where("id = ?", adjustment.SubscriptionID).First(&sub).Error; err != nil {
					return err
				}
				if sub.AmountUsed < 0 || (adjustment.Delta > 0 && sub.AmountUsed > math.MaxInt64-int64(adjustment.Delta)) {
					return fmt.Errorf("subscription amount out of range")
				}
				used := sub.AmountUsed + int64(adjustment.Delta)
				if used < 0 {
					used = 0
				}
				if sub.AmountTotal > 0 && used > sub.AmountTotal {
					return ErrInsufficientQuota
				}
				if err := tx.Model(&sub).Update("amount_used", used).Error; err != nil {
					return err
				}
			} else {
				query := tx.Model(&User{}).Where("id = ?", adjustment.UserID)
				if adjustment.Delta > 0 {
					query = query.Where("quota >= ?", adjustment.Delta)
				}
				result := query.Update("quota", gorm.Expr("quota - ?", adjustment.Delta))
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return ErrInsufficientQuota
				}
			}
		}
		if adjustment.TaskID > 0 {
			result := tx.Model(&Task{}).Where("id = ?", adjustment.TaskID).Updates(map[string]interface{}{"quota": adjustment.RevenueQuota, "updated_at": common.GetTimestamp()})
			if result.Error != nil {
				return result.Error
			}
		}
		return queueChannelProfitTotal(tx, adjustment.EventKey, adjustment.BaseQuota, adjustment.RevenueQuota)
	})
	if err == nil && adjustment.SubscriptionID == 0 && adjustment.Delta != 0 {
		if cacheErr := cacheIncrUserQuota(adjustment.UserID, -int64(adjustment.Delta)); cacheErr != nil {
			common.SysError("failed to update quota cache after profit settlement: " + cacheErr.Error())
		}
	}
	return err
}
