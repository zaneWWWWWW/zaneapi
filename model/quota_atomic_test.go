package model

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecreaseUserQuotaIsAtomicAndNeverNegative(t *testing.T) {
	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = false
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })
	user := User{Username: "atomic-user", Password: "password", Quota: 10, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(&user).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&user) })

	require.NoError(t, DecreaseUserQuota(user.Id, 7, false))
	err := DecreaseUserQuota(user.Id, 7, false)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInsufficientQuota))

	var got User
	require.NoError(t, DB.Select("quota").First(&got, user.Id).Error)
	assert.Equal(t, 3, got.Quota)
}

func TestConcurrentUserQuotaDecrementsOnlyAllowAvailableBalance(t *testing.T) {
	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = false
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })
	user := User{Username: "atomic-concurrent-user", Password: "password", Quota: 10, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(&user).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&user) })

	start := make(chan struct{})
	results := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			results <- DecreaseUserQuota(user.Id, 7, false)
		}()
	}
	close(start)

	succeeded := 0
	insufficient := 0
	for range 2 {
		err := <-results
		if err == nil {
			succeeded++
		} else if errors.Is(err, ErrInsufficientQuota) {
			insufficient++
		} else {
			require.NoError(t, err)
		}
	}
	assert.Equal(t, 1, succeeded)
	assert.Equal(t, 1, insufficient)

	var got User
	require.NoError(t, DB.Select("quota").First(&got, user.Id).Error)
	assert.Equal(t, 3, got.Quota)
}

func TestDecreaseTokenQuotaIsAtomicAndNeverNegative(t *testing.T) {
	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = false
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })
	token := Token{UserId: 1, Key: "atomic-token", RemainQuota: 10}
	require.NoError(t, DB.Create(&token).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&token) })

	require.NoError(t, DecreaseTokenQuota(token.Id, token.Key, 7))
	err := DecreaseTokenQuota(token.Id, token.Key, 7)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInsufficientQuota))

	var got Token
	require.NoError(t, DB.Select("remain_quota").First(&got, token.Id).Error)
	assert.Equal(t, 3, got.RemainQuota)
}

func TestQuotaDecrementsStayImmediateWhenBatchUpdatesEnabled(t *testing.T) {
	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = true
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })

	user := User{Username: "atomic-batch-user", Password: "password", Quota: 10, Status: common.UserStatusEnabled}
	require.NoError(t, DB.Create(&user).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&user) })
	token := Token{UserId: user.Id, Key: "atomic-batch-token", RemainQuota: 10}
	require.NoError(t, DB.Create(&token).Error)
	t.Cleanup(func() { DB.Unscoped().Delete(&token) })

	require.NoError(t, DecreaseUserQuota(user.Id, 7, false))
	require.NoError(t, DecreaseTokenQuota(token.Id, token.Key, 7))

	var gotUser User
	require.NoError(t, DB.Select("quota").First(&gotUser, user.Id).Error)
	assert.Equal(t, 3, gotUser.Quota)
	var gotToken Token
	require.NoError(t, DB.Select("remain_quota").First(&gotToken, token.Id).Error)
	assert.Equal(t, 3, gotToken.RemainQuota)
}

func TestZeroQuotaDecrementDoesNotRequireExistingRecord(t *testing.T) {
	require.NoError(t, DecreaseUserQuota(-1, 0, false))
	require.NoError(t, DecreaseTokenQuota(-1, "missing-token", 0))
}
