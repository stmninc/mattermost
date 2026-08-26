// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// IsChannelAccessible checks if a user has permissions to post and react in a given channel
// based on the full validation logic including DM/GM restrictions and exceptions.
// The user is derived from the session in the provided context.
// It returns true if accessible, false if not.
// Open and Private channels are always accessible.
// DM and Group channels are subject to permission and exception checks.
func (a *App) IsChannelAccessible(rctx request.CTX, channelID string) (bool, *model.AppError) {
	session := rctx.Session()
	if session == nil {
		return false, model.NewAppError("IsChannelAccessible", "api.context.session_expired.app_error", nil, "", http.StatusUnauthorized)
	}

	channel, err := a.GetChannel(rctx, channelID)
	if err != nil {
		return false, err
	}

	user, err := a.GetUser(session.UserId)
	if err != nil {
		return false, err
	}
	if user.IsBot {
		return true, nil
	}

	// System admins always have permission
	if a.SessionHasPermissionTo(*session, model.PermissionManageSystem) {
		return true, nil
	}

	// 1. Check if the situation is under restriction on DM/GM channel (by checking the user themselves has the required permission)
	var hasPermission bool
	switch channel.Type {
	case model.ChannelTypeDirect:
		hasPermission = a.SessionHasPermissionTo(*session, model.PermissionCreateDirectChannel)
		if !hasPermission {
			// Allow access if the DM channel has a bot member (e.g. system-generated bot messages)
			hasBotMember, botErr := a.Srv().Store().EngageChat().HasDMChannelBotMember(channelID)
			if botErr != nil {
				return false, model.NewAppError("IsChannelAccessible", "app.channel.has_bot_member.app_error", nil, "", http.StatusInternalServerError).Wrap(botErr)
			}
			if hasBotMember {
				return true, nil
			}
		}
	case model.ChannelTypeGroup:
		hasPermission = a.SessionHasPermissionTo(*session, model.PermissionCreateGroupChannel)
	default:
		// Open and Private channels are always accessible.
		return true, nil
	}

	if hasPermission {
		return true, nil
	}

	// 2. Check for the exception in restriction (by checking any member in the channel has one of the exception roles)
	hasMemberWithRole, hasRoleErr := a.Srv().Store().EngageChat().HasDMGMChannelMemberWithEngageAdmin(channelID)
	if hasRoleErr != nil {
		return false, model.NewAppError("IsChannelAccessible", "app.channel.has_engage_admin_role.app_error", nil, "", http.StatusInternalServerError).Wrap(hasRoleErr)
	}
	if hasMemberWithRole {
		return true, nil // Exception found, channel is accessible.
	}

	// No permissions and no exceptions found. The channel is not accessible.
	return false, nil
}
