// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// Validates for the TUNAG setting "Allow DM, GM, and creation of unofficial channel".
// If the permission checked by this function has been removed from the role,
// the user is considered not to be allowed the corresponding TUNAG setting.
//
//   - unofficial channel:    channel not created via TUNAG. It can be checked by isOfficialChannel().
func (a *App) CheckChannelPermissions(rctx request.CTX, channel *model.Channel, userID string) *model.AppError {
	session := rctx.Session()
	if session == nil || session.Id == "" {
		return nil // No session means no permission check needed
	}

	// If channel is nil, skip permission check
	if channel == nil {
		return nil
	}

	// Bot is always assumed to have permissions
	if session.IsBotUser() {
		return nil
	}
	if userID != "" {
		user, err := a.GetUser(userID)
		if err != nil {
			if err.StatusCode != http.StatusNotFound {
				return err
			}
			// User not found, so not a bot. Continue to permission checks.
		} else if user.IsBot {
			return nil
		}
	}

	// If user is not a member of any team, skip permission check for api-test (TestCreatePostAll)
	if len(session.TeamMembers) == 0 {
		return nil
	}

	// System admins always assumed to have permission to DM/GM/Channels
	if a.SessionHasPermissionTo(*session, model.PermissionManageSystem) {
		return nil
	}

	var requiredPermission *model.Permission
	var hasPermission bool

	switch channel.Type {
	case model.ChannelTypeDirect:
		requiredPermission = model.PermissionCreateDirectChannel
		hasPermission = a.SessionHasPermissionTo(*session, requiredPermission)

	case model.ChannelTypeGroup:
		requiredPermission = model.PermissionCreateGroupChannel
		hasPermission = a.SessionHasPermissionTo(*session, requiredPermission)

	case model.ChannelTypePrivate:
		// If channel is official channel, user always has permission to access.
		isOfficial, err := a.IsOfficialChannel(rctx, channel)
		if err != nil {
			return err
		}
		if isOfficial {
			return nil
		}

		requiredPermission = model.PermissionCreatePrivateChannel
		hasPermission = a.SessionHasPermissionToTeam(*session, channel.TeamId, requiredPermission)

	default:
		return nil
	}

	if requiredPermission != nil && !hasPermission {
		return model.NewAppError(
			"CheckChannelPermissions",
			"api.context.permissions.app_error",
			nil,
			"",
			http.StatusForbidden,
		)
	}

	return nil
}
