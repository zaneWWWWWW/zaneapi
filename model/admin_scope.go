package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AdminScopeKindChannel  = "channel"
	AdminScopeKindUser     = "user"
	AdminScopeModeAll      = "all"
	AdminScopeModeAssigned = "assigned"
)

type AdminScope struct {
	Id     uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId int    `json:"user_id" gorm:"uniqueIndex:uk_admin_scope_user_kind,priority:1;not null"`
	Kind   string `json:"kind" gorm:"size:32;uniqueIndex:uk_admin_scope_user_kind,priority:2;not null"`
	Mode   string `json:"mode" gorm:"size:32;not null"`
}

func (AdminScope) TableName() string {
	return "admin_scopes"
}

type AdminScopeItem struct {
	Id     uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId int    `json:"user_id" gorm:"uniqueIndex:uk_admin_scope_item,priority:1;index;not null"`
	Kind   string `json:"kind" gorm:"size:32;uniqueIndex:uk_admin_scope_item,priority:2;not null"`
	ItemId int    `json:"item_id" gorm:"uniqueIndex:uk_admin_scope_item,priority:3;not null"`
}

func (AdminScopeItem) TableName() string {
	return "admin_scope_items"
}

type DataScope struct {
	All bool
	IDs []int
}

func (s DataScope) Allows(id int) bool {
	if s.All {
		return true
	}
	for _, itemID := range s.IDs {
		if itemID == id {
			return true
		}
	}
	return false
}

func ApplyIDScope(query *gorm.DB, column string, scope DataScope) *gorm.DB {
	if scope.All {
		return query
	}
	if len(scope.IDs) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where(column+" IN ?", scope.IDs)
}

func LoadDataScope(userID int, kind string) DataScope {
	return loadDataScope(DB, userID, kind)
}

func loadDataScope(db *gorm.DB, userID int, kind string) DataScope {
	if db == nil || userID <= 0 {
		return DataScope{}
	}
	var scope AdminScope
	err := db.Where("user_id = ? AND kind = ?", userID, kind).First(&scope).Error
	if err != nil {
		return DataScope{}
	}
	if scope.Mode == AdminScopeModeAll {
		return DataScope{All: true}
	}
	var items []AdminScopeItem
	if err := db.Where("user_id = ? AND kind = ?", userID, kind).Find(&items).Error; err != nil {
		return DataScope{}
	}
	ids := make([]int, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ItemId)
	}
	return DataScope{IDs: ids}
}

type AdminScopePayload struct {
	Mode string `json:"mode"`
	IDs  []int  `json:"ids"`
}

type AdminScopesPayload struct {
	Channel *AdminScopePayload `json:"channel,omitempty"`
	User    *AdminScopePayload `json:"user,omitempty"`
}

func GetAdminScopesPayload(userID int) AdminScopesPayload {
	return AdminScopesPayload{
		Channel: scopeToPayload(LoadDataScope(userID, AdminScopeKindChannel)),
		User:    scopeToPayload(LoadDataScope(userID, AdminScopeKindUser)),
	}
}

func scopeToPayload(scope DataScope) *AdminScopePayload {
	if scope.All {
		return &AdminScopePayload{Mode: AdminScopeModeAll}
	}
	ids := scope.IDs
	if ids == nil {
		ids = []int{}
	}
	return &AdminScopePayload{Mode: AdminScopeModeAssigned, IDs: ids}
}

func SetAdminScopeInTx(tx *gorm.DB, userID int, kind string, payload AdminScopePayload) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	mode := payload.Mode
	if mode != AdminScopeModeAll {
		mode = AdminScopeModeAssigned
	}
	scope := AdminScope{UserId: userID, Kind: kind, Mode: mode}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "kind"}},
		DoUpdates: clause.AssignmentColumns([]string{"mode"}),
	}).Create(&scope).Error; err != nil {
		return err
	}
	if err := tx.Where("user_id = ? AND kind = ?", userID, kind).Delete(&AdminScopeItem{}).Error; err != nil {
		return err
	}
	if mode == AdminScopeModeAll {
		return nil
	}
	seen := make(map[int]struct{}, len(payload.IDs))
	items := make([]AdminScopeItem, 0, len(payload.IDs))
	for _, id := range payload.IDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		items = append(items, AdminScopeItem{UserId: userID, Kind: kind, ItemId: id})
	}
	if len(items) == 0 {
		return nil
	}
	return tx.Create(&items).Error
}

func AssignScopeItemInTx(tx *gorm.DB, userID int, kind string, itemID int) error {
	if tx == nil || userID <= 0 || itemID <= 0 {
		return nil
	}
	var scope AdminScope
	err := tx.Where("user_id = ? AND kind = ?", userID, kind).First(&scope).Error
	if err == nil && scope.Mode == AdminScopeModeAll {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := tx.Create(&AdminScope{
			UserId: userID,
			Kind:   kind,
			Mode:   AdminScopeModeAssigned,
		}).Error; err != nil {
			return err
		}
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&AdminScopeItem{
		UserId: userID,
		Kind:   kind,
		ItemId: itemID,
	}).Error
}

func ClearAdminScopesInTx(tx *gorm.DB, userID int) error {
	if tx == nil || userID <= 0 {
		return nil
	}
	if err := tx.Where("user_id = ?", userID).Delete(&AdminScopeItem{}).Error; err != nil {
		return err
	}
	return tx.Where("user_id = ?", userID).Delete(&AdminScope{}).Error
}

func EnsureAdminScopesForExistingAdmins() error {
	if DB == nil {
		return nil
	}
	var admins []User
	if err := DB.Select("id").Where("role >= ?", common.RoleAdminUser).Find(&admins).Error; err != nil {
		return err
	}
	for _, admin := range admins {
		if err := ensureDefaultAllScope(admin.Id, AdminScopeKindChannel); err != nil {
			return err
		}
		if err := ensureDefaultAllScope(admin.Id, AdminScopeKindUser); err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultAllScope(userID int, kind string) error {
	var count int64
	if err := DB.Model(&AdminScope{}).Where("user_id = ? AND kind = ?", userID, kind).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return DB.Create(&AdminScope{
		UserId: userID,
		Kind:   kind,
		Mode:   AdminScopeModeAll,
	}).Error
}
