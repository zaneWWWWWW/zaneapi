package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type legacyProfitChannel struct {
	Channel   `gorm:"embedded"`
	CostRatio *float64 `gorm:"column:cost_ratio"`
}

type legacyProfitRecord struct {
	ChannelProfitRecord `gorm:"embedded"`
	CostRatio           float64 `gorm:"column:cost_ratio"`
}

func TestChannelProfitMigrationPreservesHistoryAndIsRepeatable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Table("channels").AutoMigrate(&legacyProfitChannel{}))
	require.NoError(t, db.Table("channel_profit_records").AutoMigrate(&legacyProfitRecord{}))
	ratio := .7
	channel := legacyProfitChannel{Channel: Channel{Name: "existing", Key: "secret", Models: "model"}, CostRatio: &ratio}
	require.NoError(t, db.Table("channels").Create(&channel).Error)
	record := legacyProfitRecord{ChannelProfitRecord: ChannelProfitRecord{EventKey: "legacy", ChannelId: channel.Id, RevenueQuota: 1000, CostQuota: 700, ProfitQuota: 300}, CostRatio: .7}
	require.NoError(t, db.Table("channel_profit_records").Create(&record).Error)
	// Startup runs AutoMigrate first: it must not reinterpret the old value.
	require.NoError(t, db.AutoMigrate(&Channel{}, &ChannelProfitRecord{}, &ChannelProfitSettlement{}))
	require.NoError(t, migrateChannelProfitSchema(db))
	require.NoError(t, migrateChannelProfitSchema(db))
	assert.False(t, db.Migrator().HasColumn(&Channel{}, "cost_ratio"))
	assert.False(t, db.Migrator().HasColumn(&ChannelProfitRecord{}, "cost_ratio"))
	var storedChannel Channel
	require.NoError(t, db.First(&storedChannel, channel.Id).Error)
	assert.Nil(t, storedChannel.UpstreamRatio)
	assert.Equal(t, "secret", storedChannel.Key)
	var storedRecord ChannelProfitRecord
	require.NoError(t, db.First(&storedRecord, record.Id).Error)
	assert.Equal(t, int64(1000), storedRecord.RevenueQuota)
	assert.Equal(t, int64(700), storedRecord.CostQuota)
	assert.Equal(t, int64(300), storedRecord.ProfitQuota)
	assert.Zero(t, storedRecord.AccountingVersion)
	// New writes and the original unique event constraint survive SQLite's
	// table rebuild, as well as a later startup migration.
	require.NoError(t, db.AutoMigrate(&Channel{}, &ChannelProfitRecord{}))
	upstream := 1.2
	require.NoError(t, db.Model(&storedChannel).Update("upstream_ratio", upstream).Error)
	require.NoError(t, db.First(&storedChannel, channel.Id).Error)
	require.NotNil(t, storedChannel.UpstreamRatio)
	assert.Equal(t, upstream, *storedChannel.UpstreamRatio)
	err = db.Create(&ChannelProfitRecord{EventKey: "legacy"}).Error
	require.Error(t, err)
}

func TestChannelProfitMigrationOnFreshDatabase(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Channel{}, &ChannelProfitRecord{}, &ChannelProfitSettlement{}))
	require.NoError(t, migrateChannelProfitSchema(db))
	require.NoError(t, migrateChannelProfitSchema(db))
	assert.True(t, db.Migrator().HasColumn(&Channel{}, "upstream_ratio"))
	assert.True(t, db.Migrator().HasTable(&ChannelProfitSettlement{}))
}
