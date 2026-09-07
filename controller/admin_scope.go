package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service/authz"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func currentChannelScope(c *gin.Context) model.DataScope {
	return authz.ChannelScope(c.GetInt("id"), c.GetInt("role"))
}

func currentUserScope(c *gin.Context) model.DataScope {
	return authz.UserScope(c.GetInt("id"), c.GetInt("role"))
}

func abortIfChannelOutOfScope(c *gin.Context, channelID int) bool {
	if currentChannelScope(c).Allows(channelID) {
		return false
	}
	common.ApiErrorI18n(c, i18n.MsgAuthInsufficientPrivilege)
	return true
}

func abortIfUserOutOfScope(c *gin.Context, userID int) bool {
	if currentUserScope(c).Allows(userID) {
		return false
	}
	common.ApiErrorI18n(c, i18n.MsgAuthInsufficientPrivilege)
	return true
}

func filterChannelsByScope(channels []*model.Channel, scope model.DataScope) []*model.Channel {
	if scope.All {
		return channels
	}
	filtered := make([]*model.Channel, 0, len(channels))
	for _, channel := range channels {
		if channel != nil && scope.Allows(channel.Id) {
			filtered = append(filtered, channel)
		}
	}
	return filtered
}

func abortIfAnyChannelOutOfScope(c *gin.Context, ids []int) bool {
	scope := currentChannelScope(c)
	for _, id := range ids {
		if !scope.Allows(id) {
			common.ApiErrorI18n(c, i18n.MsgAuthInsufficientPrivilege)
			return true
		}
	}
	return false
}

func assignCreatedChannels(c *gin.Context, channelIDs []int) error {
	if len(channelIDs) == 0 {
		return nil
	}
	actorID := c.GetInt("id")
	actorRole := c.GetInt("role")
	return model.DB.Transaction(func(tx *gorm.DB) error {
		for _, channelID := range channelIDs {
			if err := authz.AssignCreatedChannelInTx(tx, actorID, actorRole, channelID); err != nil {
				return err
			}
		}
		return nil
	})
}

func emptyAssignedScopes() model.AdminScopesPayload {
	empty := &model.AdminScopePayload{Mode: model.AdminScopeModeAssigned, IDs: []int{}}
	return model.AdminScopesPayload{Channel: empty, User: empty}
}

func updateAdminScopesForUserInTx(c *gin.Context, tx *gorm.DB, userID int, userRole int, payload *model.AdminScopesPayload, initializeIfMissing bool) error {
	if userRole < common.RoleAdminUser {
		return model.ClearAdminScopesInTx(tx, userID)
	}
	if c.GetInt("role") != common.RoleRootUser {
		return nil
	}
	if payload == nil {
		if !initializeIfMissing {
			return nil
		}
		empty := emptyAssignedScopes()
		payload = &empty
	}
	if payload.Channel != nil {
		if err := model.SetAdminScopeInTx(tx, userID, model.AdminScopeKindChannel, *payload.Channel); err != nil {
			return err
		}
	}
	if payload.User != nil {
		if err := model.SetAdminScopeInTx(tx, userID, model.AdminScopeKindUser, *payload.User); err != nil {
			return err
		}
	}
	return nil
}
