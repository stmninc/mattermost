// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// checkDMGMChannelPermissions checks if the user has permission to perform actions in DM/GM channels
func checkDMGMChannelPermissions(c *Context, channel *model.Channel) {
	switch channel.Type {
	case model.ChannelTypeDirect:
		if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionCreateDirectChannel) {
			c.SetPermissionError(model.PermissionCreateDirectChannel)
		}
	case model.ChannelTypeGroup:
		if !c.App.SessionHasPermissionTo(*c.AppContext.Session(), model.PermissionCreateGroupChannel) {
			c.SetPermissionError(model.PermissionCreateGroupChannel)
		}
	}
}
