package authz

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestChannelScopeRootIsUnrestricted(t *testing.T) {
	t.Parallel()
	assert.True(t, ChannelScope(1, common.RoleRootUser).All)
}

func TestChannelScopeCommonUserSeesNone(t *testing.T) {
	t.Parallel()
	assert.False(t, ChannelScope(4, common.RoleCommonUser).All)
	assert.Empty(t, ChannelScope(4, common.RoleCommonUser).IDs)
}

func TestAssignCreatedChannelInTx(t *testing.T) {
	db := newAuthzTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.AdminScope{}, &model.AdminScopeItem{}))
	previous := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = previous })

	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return AssignCreatedChannelInTx(tx, 5, common.RoleAdminUser, 21)
	}))
	scope := ChannelScope(5, common.RoleAdminUser)
	assert.False(t, scope.All)
	assert.Equal(t, []int{21}, scope.IDs)

	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return AssignCreatedChannelInTx(tx, 1, common.RoleRootUser, 22)
	}))
	var count int64
	require.NoError(t, db.Model(&model.AdminScopeItem{}).Where("user_id = ?", 1).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
