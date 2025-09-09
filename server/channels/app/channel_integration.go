// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"os"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

var (
	integrationAdminUsername string
	integrationAdminOnce     sync.Once
)

// getIntegrationAdminUsername returns the integration admin username from environment variable
// using sync.Once to ensure it's only read once for performance.
func getIntegrationAdminUsername() string {
	integrationAdminOnce.Do(func() {
		integrationAdminUsername = os.Getenv("INTEGRATION_ADMIN_USERNAME")
	})
	return integrationAdminUsername
}

// IsOfficialChannel checks if a channel is official by comparing creator with integration admin user.
func (a *App) IsOfficialChannel(c request.CTX, channel *model.Channel) (bool, *model.AppError) {
	if channel == nil {
		return false, nil
	}

	// Get cached integration admin username
	adminUsername := getIntegrationAdminUsername()
	if adminUsername == "" {
		return false, nil
	}

	creatorUser, err := a.GetUser(channel.CreatorId)
	if err != nil {
		return false, err
	}

	return creatorUser.Username == adminUsername, nil
}
