// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

func (a *App) CheckDMGMChannelPermissions(c request.CTX, channel *model.Channel, userID string) *model.AppError {
	session := c.Session()
	if session == nil || session.Id == "" {
		return nil // No session means no DM/GM permission check needed
	}

	// If channel is nil, skip DM/GM permission check
	if channel == nil {
		return nil
	}

	// Check if the user is a bot - bots should be allowed to post in DM/GM channels
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

	// System admins always have permission to DM/GM channels
	if a.SessionHasPermissionTo(*session, model.PermissionManageSystem) {
		return nil
	}

	var requiredPermission *model.Permission
	switch channel.Type {
	case model.ChannelTypeDirect:
		requiredPermission = model.PermissionCreateDirectChannel
	case model.ChannelTypeGroup:
		requiredPermission = model.PermissionCreateGroupChannel
	}

	if requiredPermission != nil && !a.SessionHasPermissionTo(*session, requiredPermission) {
		return model.NewAppError(
			"CheckDMGMChannelPermissions",
			"api.context.permissions.app_error",
			nil,
			"",
			http.StatusForbidden,
		)
	}

	return nil
}
