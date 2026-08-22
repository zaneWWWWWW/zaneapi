package service

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnlimitedTokenBillingOnlyDecreasesWallet(t *testing.T) {
	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = false
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

	user := model.User{Username: "unlimited-token-billing", Password: "password", Quota: 100, Status: common.UserStatusEnabled}
	require.NoError(t, model.DB.Create(&user).Error)
	t.Cleanup(func() { model.DB.Unscoped().Delete(&user) })

	token := model.Token{UserId: user.Id, Key: "unlimited-token-billing", RemainQuota: 0, UnlimitedQuota: true}
	require.NoError(t, model.DB.Create(&token).Error)
	t.Cleanup(func() { model.DB.Unscoped().Delete(&token) })

	info := &relaycommon.RelayInfo{
		UserId:         user.Id,
		TokenId:        token.Id,
		TokenKey:       token.Key,
		TokenUnlimited: true,
	}
	session := &BillingSession{relayInfo: info, funding: &WalletFunding{userId: user.Id}}
	require.Nil(t, session.preConsume(ctx, 40))
	assert.True(t, session.NeedsRefund())
	require.NoError(t, session.Settle(40))

	var gotUser model.User
	require.NoError(t, model.DB.Select("quota").First(&gotUser, user.Id).Error)
	assert.Equal(t, 60, gotUser.Quota)
	var gotToken model.Token
	require.NoError(t, model.DB.Select("remain_quota").First(&gotToken, token.Id).Error)
	assert.Equal(t, 0, gotToken.RemainQuota)
}

func TestPositiveSettlementKeepsWalletAndTokenConsistentOnFailure(t *testing.T) {
	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = false
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })

	t.Run("token insufficient does not charge wallet", func(t *testing.T) {
		user := model.User{Username: "settle-token-insufficient", Password: "password", Quota: 100, Status: common.UserStatusEnabled}
		require.NoError(t, model.DB.Create(&user).Error)
		t.Cleanup(func() { model.DB.Unscoped().Delete(&user) })
		token := model.Token{UserId: user.Id, Key: "settle-token-insufficient", RemainQuota: 10}
		require.NoError(t, model.DB.Create(&token).Error)
		t.Cleanup(func() { model.DB.Unscoped().Delete(&token) })

		info := &relaycommon.RelayInfo{UserId: user.Id, TokenId: token.Id, TokenKey: token.Key}
		session := &BillingSession{relayInfo: info, funding: &WalletFunding{userId: user.Id}}
		require.Error(t, session.Settle(20))

		var gotUser model.User
		require.NoError(t, model.DB.Select("quota").First(&gotUser, user.Id).Error)
		assert.Equal(t, 100, gotUser.Quota)
		var gotToken model.Token
		require.NoError(t, model.DB.Select("remain_quota").First(&gotToken, token.Id).Error)
		assert.Equal(t, 10, gotToken.RemainQuota)
	})

	t.Run("wallet insufficient rolls token reservation back", func(t *testing.T) {
		user := model.User{Username: "settle-wallet-insufficient", Password: "password", Quota: 10, Status: common.UserStatusEnabled}
		require.NoError(t, model.DB.Create(&user).Error)
		t.Cleanup(func() { model.DB.Unscoped().Delete(&user) })
		token := model.Token{UserId: user.Id, Key: "settle-wallet-insufficient", RemainQuota: 100}
		require.NoError(t, model.DB.Create(&token).Error)
		t.Cleanup(func() { model.DB.Unscoped().Delete(&token) })

		info := &relaycommon.RelayInfo{UserId: user.Id, TokenId: token.Id, TokenKey: token.Key}
		session := &BillingSession{relayInfo: info, funding: &WalletFunding{userId: user.Id}}
		require.Error(t, session.Settle(20))

		var gotUser model.User
		require.NoError(t, model.DB.Select("quota").First(&gotUser, user.Id).Error)
		assert.Equal(t, 10, gotUser.Quota)
		var gotToken model.Token
		require.NoError(t, model.DB.Select("remain_quota").First(&gotToken, token.Id).Error)
		assert.Equal(t, 100, gotToken.RemainQuota)
	})
}

func TestPostConsumeQuotaRollsWalletBackWhenTokenIsInsufficient(t *testing.T) {
	oldBatch := common.BatchUpdateEnabled
	common.BatchUpdateEnabled = false
	t.Cleanup(func() { common.BatchUpdateEnabled = oldBatch })

	user := model.User{Username: "post-consume-token-insufficient", Password: "password", Quota: 100, Status: common.UserStatusEnabled}
	require.NoError(t, model.DB.Create(&user).Error)
	t.Cleanup(func() { model.DB.Unscoped().Delete(&user) })
	token := model.Token{UserId: user.Id, Key: "post-consume-token-insufficient", RemainQuota: 10}
	require.NoError(t, model.DB.Create(&token).Error)
	t.Cleanup(func() { model.DB.Unscoped().Delete(&token) })

	info := &relaycommon.RelayInfo{UserId: user.Id, TokenId: token.Id, TokenKey: token.Key}
	require.Error(t, PostConsumeQuota(info, 20, 0, false))

	var gotUser model.User
	require.NoError(t, model.DB.Select("quota").First(&gotUser, user.Id).Error)
	assert.Equal(t, 100, gotUser.Quota)
	var gotToken model.Token
	require.NoError(t, model.DB.Select("remain_quota").First(&gotToken, token.Id).Error)
	assert.Equal(t, 10, gotToken.RemainQuota)
}
