// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"github.com/mattermost/mattermost/server/public/model"
)

// checkOfficialChannelPermission checks if the user can perform actions on an official channel.
// Returns false if the channel is official and the user is not the creator. Sets c.Err on errors.
func checkOfficialChannelPermission(c *Context, channel *model.Channel) bool {
	isOfficial, appErr := c.App.IsOfficialChannel(c.AppContext, channel)
	if appErr != nil {
		c.Err = appErr
		return false
	}

	if isOfficial {
		// For official channels, only the creator can perform actions
		if channel.CreatorId != c.AppContext.Session().UserId {
			return false
		}
	}

	// For non-official channels, return true (caller handles permissions)
	return true
}
