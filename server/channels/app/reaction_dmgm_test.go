// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestDMGMPermissionChecksForReactions(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	user1 := th.BasicUser
	user2 := th.BasicUser2

	// Create DM channel (with permissions)
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, user1.Id, user2.Id)
	require.Nil(t, appErr)
	require.NotNil(t, dmChannel)

	// Create GM channel (with permissions)
	user3 := th.CreateUser()
	gmChannel, appErr := th.App.createGroupChannel(th.Context, []string{user1.Id, user2.Id, user3.Id}, user1.Id)
	require.Nil(t, appErr)
	require.NotNil(t, gmChannel)

	// Create posts in DM and GM for reaction tests
	dmPost := th.CreatePost(dmChannel)
	gmPost := th.CreatePost(gmChannel)

	// Now remove permissions for testing
	th.RemovePermissionFromRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

	th.RemovePermissionFromRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)

	t.Run("SaveReactionForPost denied in DM without permission", func(t *testing.T) {
		reaction := &model.Reaction{
			UserId:    user1.Id,
			PostId:    dmPost.Id,
			EmojiName: "smile",
		}

		session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		_, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.NotNil(t, appErr)
		assert.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})

	t.Run("SaveReactionForPost denied in GM without permission", func(t *testing.T) {
		reaction := &model.Reaction{
			UserId:    user1.Id,
			PostId:    gmPost.Id,
			EmojiName: "smile",
		}

		session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		_, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.NotNil(t, appErr)
		assert.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})

	t.Run("Bot can add reaction in DM without permission", func(t *testing.T) {
		bot := th.CreateBot()
		botUser, appErr := th.App.GetUser(bot.UserId)
		require.Nil(t, appErr)

		reaction := &model.Reaction{
			UserId:    botUser.Id,
			PostId:    dmPost.Id,
			EmojiName: "robot",
		}

		session := &model.Session{Id: model.NewId(), UserId: botUser.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		createdReaction, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.Nil(t, appErr)
		require.NotNil(t, createdReaction)
	})

	t.Run("System admin can add reaction in DM without permission", func(t *testing.T) {
		adminUser := th.SystemAdminUser

		reaction := &model.Reaction{
			UserId:    adminUser.Id,
			PostId:    dmPost.Id,
			EmojiName: "star",
		}

		session := &model.Session{Id: model.NewId(), UserId: adminUser.Id, Roles: model.SystemAdminRoleId}
		ctxWithSession := th.Context.WithSession(session)
		createdReaction, appErr := th.App.SaveReactionForPost(ctxWithSession, reaction)
		require.Nil(t, appErr)
		require.NotNil(t, createdReaction)
	})
}
