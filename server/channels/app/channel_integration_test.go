// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

// Resets the cached integration admin username from environment variable for tests.
func resetIntegrationAdminUsernameForTesting() {
	integrationAdminUsername = ""
	integrationAdminOnce = sync.Once{}
}

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
		// Reset cache for this test
		resetIntegrationAdminUsernameForTesting()

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
		// Official channels are created as private channels
		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Official Channel",
			Name:        "official-channel",
			Type:        model.ChannelTypePrivate,
			CreatorId:   adminUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.True(t, isOfficial)
	})

	t.Run("returns false when channel creator is not integration admin", func(t *testing.T) {
		// Reset cache for this test
		resetIntegrationAdminUsernameForTesting()

		// Set up environment variable with unique username
		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		// Use basic user (not admin) as creator
		// Non-official channels can be either public or private, using private here
		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Non-Official Channel",
			Name:        "non-official-channel",
			Type:        model.ChannelTypePrivate,
			CreatorId:   th.BasicUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns false when environment variable is not set", func(t *testing.T) {
		// Reset cache for this test
		resetIntegrationAdminUsernameForTesting()

		// Clear environment variable
		os.Unsetenv("INTEGRATION_ADMIN_USERNAME")

		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Test Channel",
			Name:        "test-channel",
			Type:        model.ChannelTypePrivate,
			CreatorId:   th.BasicUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns false when channel is nil", func(t *testing.T) {
		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		isOfficial, appErr := th.App.IsOfficialChannel(th.Context, nil)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns error when user does not exist", func(t *testing.T) {
		// Reset cache for this test
		resetIntegrationAdminUsernameForTesting()

		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Test Channel",
			Name:        "test-channel",
			Type:        model.ChannelTypePrivate,
			CreatorId:   "non-existent-user-id",
		}

		isOfficial, appErr := th.App.IsOfficialChannel(th.Context, channel)
		require.NotNil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns false when environment variable is empty string", func(t *testing.T) {
		// Reset cache for this test
		resetIntegrationAdminUsernameForTesting()

		// Set environment variable to empty string
		os.Setenv("INTEGRATION_ADMIN_USERNAME", "")

		channel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Test Channel",
			Name:        "test-channel",
			Type:        model.ChannelTypePrivate,
			CreatorId:   th.BasicUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialChannel(th.Context, channel)
		require.Nil(t, appErr)
		assert.False(t, isOfficial)
	})

	t.Run("returns true for public channel created by integration admin (edge case)", func(t *testing.T) {
		// Reset cache for this test
		resetIntegrationAdminUsernameForTesting()

		adminUsername := fmt.Sprintf("integration-admin-%d", time.Now().UnixNano())
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUsername)

		// Create admin user
		adminUser, err := th.App.CreateUser(th.Context, &model.User{
			Email:    fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()),
			Username: adminUsername,
			Password: "password123",
		})
		require.Nil(t, err)

		// Edge case: Test that the function works for public channels created by integration admin
		// (though official channels should normally be private)
		publicChannel := &model.Channel{
			TeamId:      th.BasicTeam.Id,
			DisplayName: "Public Channel by Admin",
			Name:        "public-channel-by-admin",
			Type:        model.ChannelTypeOpen,
			CreatorId:   adminUser.Id,
		}

		isOfficial, appErr := th.App.IsOfficialChannel(th.Context, publicChannel)
		require.Nil(t, appErr)
		assert.True(t, isOfficial)
	})
}
