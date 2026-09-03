package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveQuotaDataCacheRecordsSuccessfulExportTime(t *testing.T) {
	CacheQuotaDataLock.Lock()
	previousCache := CacheQuotaData
	previousExportAt := lastQuotaDataExportAt.Load()
	CacheQuotaData = make(map[string]*QuotaData)
	lastQuotaDataExportAt.Store(0)
	CacheQuotaDataLock.Unlock()
	t.Cleanup(func() {
		CacheQuotaDataLock.Lock()
		CacheQuotaData = previousCache
		lastQuotaDataExportAt.Store(previousExportAt)
		CacheQuotaDataLock.Unlock()
	})

	LogQuotaData(QuotaDataLogParams{
		UserID:    98765,
		Username:  "quota-export-status",
		ModelName: "gpt-test",
		CreatedAt: time.Now().Unix(),
	})
	before := time.Now().Unix()
	require.NoError(t, SaveQuotaDataCache())
	after := time.Now().Unix()

	assert.GreaterOrEqual(t, GetLastQuotaDataExportAt(), before)
	assert.LessOrEqual(t, GetLastQuotaDataExportAt(), after)
}
