// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

// checkOfficialChannelPermission checks if the current user has permission to perform actions on an official channel
// Returns true if the action is permitted, false otherwise. Sets c.Err if there's an error.
func checkOfficialChannelPermission(c *Context, channelId string) bool {
	channel, appErr := c.App.GetChannel(c.AppContext, channelId)
	if appErr != nil {
		c.Err = appErr
		return false
	}

	isOfficial, appErr := c.App.IsOfficialChannel(c.AppContext, channel)
	if appErr != nil {
		c.Err = appErr
		return false
	}

	if isOfficial && channel.CreatorId != c.AppContext.Session().UserId {
		return false
	}

	return true
}
