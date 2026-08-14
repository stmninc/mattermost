// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestCheckChannelPermissions(t *testing.T) {
	th := Setup(t).InitBasic(t)

	// Save original environment variable
	originalValue := os.Getenv("INTEGRATION_ADMIN_USERNAME")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("INTEGRATION_ADMIN_USERNAME")
		} else {
			os.Setenv("INTEGRATION_ADMIN_USERNAME", originalValue)
		}
	}()

	// Create test users
	user1 := th.BasicUser
	user2 := th.BasicUser2
	adminUser := th.SystemAdminUser
	botUser := th.CreateUser(t)
	botUser.IsBot = true
	_, err := th.App.UpdateUser(th.Context, botUser, false)
	require.Nil(t, err)

	// Create sessions
	user1Session, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: user1.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, err)

	botSession, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: botUser.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, err)

	adminSession, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: adminUser.Id,
		Roles:  model.SystemAdminRoleId,
	})
	require.Nil(t, err)

	// Create a DM channel
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, user1.Id, user2.Id)
	require.Nil(t, appErr)
	require.NotNil(t, dmChannel)

	// Create a GM channel
	gmChannel, appErr := th.App.createGroupChannel(th.Context, []string{user1.Id, user2.Id, botUser.Id}, user1.Id)
	require.Nil(t, appErr)
	require.NotNil(t, gmChannel)

	// Create a official Channel
	channel1 := &model.Channel{
		DisplayName: "test channel",
		Name:        "test-channel-" + model.NewId(),
		Type:        model.ChannelTypePrivate,
		TeamId:      th.BasicTeam.Id,
		CreatorId:   adminUser.Id,
	}
	officialChannel, appErr := th.App.CreateChannel(th.Context, channel1, false)
	require.Nil(t, appErr)
	require.NotNil(t, officialChannel)

	// Create an unofficial channel
	channel2 := &model.Channel{
		DisplayName: "test channel",
		Name:        "test-channel-" + model.NewId(),
		Type:        model.ChannelTypePrivate,
		TeamId:      th.BasicTeam.Id,
		CreatorId:   user2.Id,
	}
	unofficialChannel, appErr := th.App.CreateChannel(th.Context, channel2, true)
	require.Nil(t, appErr)
	require.NotNil(t, unofficialChannel)

	t.Run("should allow DM access with CREATE_DIRECT_CHANNEL permission", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(user1Session)

		err := th.App.CheckChannelPermissions(ctxWithSession, dmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should allow GM access with CREATE_GROUP_CHANNEL permission", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(user1Session)

		err := th.App.CheckChannelPermissions(ctxWithSession, gmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should allow unofficial channel access with CREATE_PRIVATE_CHANNEL permission", func(t *testing.T) {
		resetIntegrationAdminUsernameForTesting()
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUser.Username)

		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(user1Session)

		err := th.App.CheckChannelPermissions(ctxWithSession, unofficialChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should not allow unofficial channel access without CREATE_PRIVATE_CHANNEL permission", func(t *testing.T) {
		resetIntegrationAdminUsernameForTesting()
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUser.Username)

		th.RemovePermissionFromRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)

		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(user1Session)

		err := th.App.CheckChannelPermissions(ctxWithSession, unofficialChannel, user1.Id)
		assert.NotNil(t, err)
	})

	t.Run("should allow official channel access", func(t *testing.T) {
		resetIntegrationAdminUsernameForTesting()
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUser.Username)

		th.RemovePermissionFromRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)

		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(user1Session)

		err := th.App.CheckChannelPermissions(ctxWithSession, officialChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should allow bots to access DM channels", func(t *testing.T) {
		// Create a context with bot's session
		ctxWithSession := th.Context.WithSession(botSession)

		err := th.App.CheckChannelPermissions(ctxWithSession, dmChannel, botUser.Id)
		assert.Nil(t, err)
	})

	t.Run("should skip check with nil channel", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(user1Session)

		err := th.App.CheckChannelPermissions(ctxWithSession, nil, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should skip check with nil session", func(t *testing.T) {
		// Create a context without a session (or with nil session)
		err := th.App.CheckChannelPermissions(th.Context, dmChannel, user1.Id)
		// Should skip check when session is nil
		assert.Nil(t, err)
	})

	t.Run("should allow system admin to access DM channels", func(t *testing.T) {
		// Create a context with admin session
		ctxWithSession := th.Context.WithSession(adminSession)

		err := th.App.CheckChannelPermissions(ctxWithSession, dmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should handle non-existent user gracefully", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(user1Session)

		// Pass a non-existent user ID
		err := th.App.CheckChannelPermissions(ctxWithSession, dmChannel, "nonexistent-user-id")
		// Should not error out, should continue to permission checks
		assert.Nil(t, err)
	})
}
