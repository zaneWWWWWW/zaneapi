package authz

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

func ChannelScope(userID int, role int) model.DataScope {
	return dataScope(userID, role, model.AdminScopeKindChannel)
}

func UserScope(userID int, role int) model.DataScope {
	return dataScope(userID, role, model.AdminScopeKindUser)
}

func dataScope(userID int, role int, kind string) model.DataScope {
	if role >= common.RoleRootUser {
		return model.DataScope{All: true}
	}
	if role < common.RoleAdminUser {
		return model.DataScope{}
	}
	return model.LoadDataScope(userID, kind)
}

func AssignCreatedChannelInTx(tx *gorm.DB, actorID int, actorRole int, channelID int) error {
	if actorRole >= common.RoleRootUser || actorRole < common.RoleAdminUser {
		return nil
	}
	return model.AssignScopeItemInTx(tx, actorID, model.AdminScopeKindChannel, channelID)
}

func AssignCreatedUserInTx(tx *gorm.DB, actorID int, actorRole int, createdUserID int) error {
	if actorRole >= common.RoleRootUser || actorRole < common.RoleAdminUser {
		return nil
	}
	return model.AssignScopeItemInTx(tx, actorID, model.AdminScopeKindUser, createdUserID)
}
