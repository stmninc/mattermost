// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"os"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// IsOfficialTunagChannel checks if a channel is official by comparing creator with integration admin user.
func (a *App) IsOfficialTunagChannel(c request.CTX, channel *model.Channel) (bool, *model.AppError) {
	if channel == nil {
		return false, nil
	}

	integrationAdminUsername := os.Getenv("INTEGRATION_ADMIN_USERNAME")
	if integrationAdminUsername == "" {
		return false, nil
	}

	creatorUser, err := a.GetUser(channel.CreatorId)
	if err != nil {
		return false, err
	}

	if creatorUser == nil {
		return false, nil
	}

	return creatorUser.Username == integrationAdminUsername, nil
}
