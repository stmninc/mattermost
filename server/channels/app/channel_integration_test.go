// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestIsOfficialChannel(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	// Save original environment variable
	originalValue := os.Getenv("INTEGRATION_ADMIN_USERNAME")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("INTEGRATION_ADMIN_USERNAME")
		} else {
			os.Setenv("INTEGRATION_ADMIN_USERNAME", originalValue)
		}
	}()

	t.Run("returns true when channel creator is integration admin", func(t *testing.T) {
		// Set up environment variable with unique username
		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		// Create admin user
		adminUser, err := th.App.CreateUser(th.Context, &model.User{
			Email:    fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()),
			Username: adminUsername,
			Password: "password123",
		})
		require.Nil(t, err)

		// Create channel with admin user as creator
		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Official Channel",
			Name:        "official-channel",
			Type:        model.ChannelTypeOpen,
			CreatorId:   adminUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialTunagChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.True(t, isOfficial)
	})

	t.Run("returns false when channel creator is not integration admin", func(t *testing.T) {
		// Set up environment variable with unique username
		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		// Use basic user (not admin) as creator
		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Non-Official Channel",
			Name:        "non-official-channel",
			Type:        model.ChannelTypeOpen,
			CreatorId:   th.BasicUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialTunagChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns false when environment variable is not set", func(t *testing.T) {
		// Clear environment variable
		os.Unsetenv("INTEGRATION_ADMIN_USERNAME")

		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Test Channel",
			Name:        "test-channel",
			Type:        model.ChannelTypeOpen,
			CreatorId:   th.BasicUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialTunagChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns false when channel is nil", func(t *testing.T) {
		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		isOfficial, appErr := th.App.IsOfficialTunagChannel(th.Context, nil)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns error when user does not exist", func(t *testing.T) {
		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Test Channel",
			Name:        "test-channel",
			Type:        model.ChannelTypeOpen,
			CreatorId:   "non-existent-user-id",
		}

		isOfficial, appErr := th.App.IsOfficialTunagChannel(th.Context, channel)
		require.NotNil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns false when environment variable is empty string", func(t *testing.T) {
		// Set environment variable to empty string
		os.Setenv("INTEGRATION_ADMIN_USERNAME", "")

		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Test Channel",
			Name:        "test-channel",
			Type:        model.ChannelTypeOpen,
			CreatorId:   th.BasicUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialTunagChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("works with different channel types", func(t *testing.T) {
		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		// Create admin user
		adminUser, err := th.App.CreateUser(th.Context, &model.User{
			Email:    fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()),
			Username: adminUsername,
			Password: "password123",
		})
		require.Nil(t, err)

		// Test with private channel
		privateChannel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Private Official Channel",
			Name:        "private-official-channel",
			Type:        model.ChannelTypePrivate,
			CreatorId:   adminUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialTunagChannel(th.Context, privateChannel)
		require.Nil(t, appErr)
		assert.True(t, isOfficial)
	})
}
