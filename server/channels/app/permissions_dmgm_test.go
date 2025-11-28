// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestCheckDMGMChannelPermissions(t *testing.T) {
	mainHelper.Parallel(t)
	th := Setup(t)
	defer th.TearDown()

	// Create test users
	user1 := th.CreateUser()
	user2 := th.CreateUser()
	botUser := th.CreateUser()
	botUser.IsBot = true
	_, err := th.App.UpdateUser(th.Context, botUser, false)
	require.Nil(t, err)

	// Create a DM channel
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, user1.Id, user2.Id)
	require.Nil(t, appErr)
	require.NotNil(t, dmChannel)

	// Create a GM channel
	gmChannel, appErr := th.App.createGroupChannel(th.Context, []string{user1.Id, user2.Id, botUser.Id}, user1.Id)
	require.Nil(t, appErr)
	require.NotNil(t, gmChannel)

	t.Run("should allow DM access with CREATE_DIRECT_CHANNEL permission", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should allow GM access with CREATE_GROUP_CHANNEL permission", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, gmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should allow bots to access DM channels", func(t *testing.T) {
		// Create a context with bot's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: botUser.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, botUser.Id)
		assert.Nil(t, err)
	})

	t.Run("should skip check with nil channel", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, nil, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should skip check with nil session", func(t *testing.T) {
		// Create a context without a session (or with nil session)
		err := th.App.CheckDMGMChannelPermissions(th.Context, dmChannel, user1.Id)
		// Should skip check when session is nil
		assert.Nil(t, err)
	})

	t.Run("should allow system admin to access DM channels", func(t *testing.T) {
		// Create a context with admin session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemAdminRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should handle non-existent user gracefully", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		// Pass a non-existent user ID
		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, "nonexistent-user-id")
		// Should not error out, should continue to permission checks
		assert.Nil(t, err)
	})
}
