package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newAdminScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&User{}, &AdminScope{}, &AdminScopeItem{}))
	previous := DB
	DB = db
	t.Cleanup(func() {
		DB = previous
		_ = sqlDB.Close()
	})
	return db
}

func TestDataScopeAllows(t *testing.T) {
	t.Parallel()
	assert.True(t, DataScope{All: true}.Allows(9))
	assert.False(t, DataScope{}.Allows(1))
	assert.True(t, DataScope{IDs: []int{2, 4}}.Allows(4))
	assert.False(t, DataScope{IDs: []int{2, 4}}.Allows(3))
}

func TestSetAdminScopeAssignedAndAll(t *testing.T) {
	db := newAdminScopeTestDB(t)
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return SetAdminScopeInTx(tx, 7, AdminScopeKindChannel, AdminScopePayload{
			Mode: AdminScopeModeAssigned,
			IDs:  []int{1, 1, 3, 0},
		})
	}))
	scope := LoadDataScope(7, AdminScopeKindChannel)
	assert.False(t, scope.All)
	assert.ElementsMatch(t, []int{1, 3}, scope.IDs)

	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return SetAdminScopeInTx(tx, 7, AdminScopeKindChannel, AdminScopePayload{Mode: AdminScopeModeAll})
	}))
	scope = LoadDataScope(7, AdminScopeKindChannel)
	assert.True(t, scope.All)
	assert.Empty(t, scope.IDs)
}

func TestAssignScopeItemSkippedWhenModeAll(t *testing.T) {
	db := newAdminScopeTestDB(t)
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		if err := SetAdminScopeInTx(tx, 8, AdminScopeKindChannel, AdminScopePayload{Mode: AdminScopeModeAll}); err != nil {
			return err
		}
		return AssignScopeItemInTx(tx, 8, AdminScopeKindChannel, 12)
	}))
	var count int64
	require.NoError(t, db.Model(&AdminScopeItem{}).Where("user_id = ?", 8).Count(&count).Error)
	assert.Equal(t, int64(0), count)
	assert.True(t, LoadDataScope(8, AdminScopeKindChannel).All)
}

func TestAssignScopeItemCreatesAssignedRow(t *testing.T) {
	db := newAdminScopeTestDB(t)
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return AssignScopeItemInTx(tx, 9, AdminScopeKindUser, 44)
	}))
	scope := LoadDataScope(9, AdminScopeKindUser)
	assert.False(t, scope.All)
	assert.Equal(t, []int{44}, scope.IDs)
}

func TestEnsureAdminScopesForExistingAdmins(t *testing.T) {
	db := newAdminScopeTestDB(t)
	require.NoError(t, db.Create(&User{Id: 2, Username: "admin-a", Password: "password1", Role: common.RoleAdminUser, AffCode: "aff-admin"}).Error)
	require.NoError(t, db.Create(&User{Id: 3, Username: "user-a", Password: "password1", Role: common.RoleCommonUser, AffCode: "aff-user"}).Error)
	require.NoError(t, EnsureAdminScopesForExistingAdmins())

	assert.True(t, LoadDataScope(2, AdminScopeKindChannel).All)
	assert.True(t, LoadDataScope(2, AdminScopeKindUser).All)
	assert.False(t, LoadDataScope(3, AdminScopeKindChannel).All)
}

func TestMissingScopeIsEmptyAssigned(t *testing.T) {
	newAdminScopeTestDB(t)
	scope := LoadDataScope(99, AdminScopeKindChannel)
	assert.False(t, scope.All)
	assert.Empty(t, scope.IDs)
}
