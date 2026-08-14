// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestChannelPermissionChecksForPosts(t *testing.T) {
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

	user1 := th.BasicUser
	user2 := th.BasicUser2
	adminUser := th.SystemAdminUser

	// Create sessions
	session, appErr := th.App.CreateSession(th.Context, &model.Session{
		UserId: user1.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, appErr)

	adminSession, appErr := th.App.CreateSession(th.Context, &model.Session{
		UserId: adminUser.Id,
		Roles:  model.SystemAdminRoleId,
	})
	require.Nil(t, appErr)

	// Create DM channel (with permissions)
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, user1.Id, user2.Id)
	require.Nil(t, appErr)
	require.NotNil(t, dmChannel)

	// Create GM channel (with permissions)
	user3 := th.CreateUser(t)
	gmChannel, appErr := th.App.createGroupChannel(th.Context, []string{user1.Id, user2.Id, user3.Id}, user1.Id)
	require.Nil(t, appErr)
	require.NotNil(t, gmChannel)

	// Create official channel
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

	// Create unofficial channel
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

	ctxWithSession := th.Context.WithSession(session)

	// Create a post in DM for update/patch/delete tests
	post := &model.Post{
		UserId:    user1.Id,
		ChannelId: dmChannel.Id,
		Message:   "original message",
	}
	createdPostInDM, _, appErr := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
	require.Nil(t, appErr)
	require.NotNil(t, createdPostInDM)

	// Create a post in unofficial channel for update/patch/delete tests
	post2 := &model.Post{
		UserId:    user1.Id,
		ChannelId: unofficialChannel.Id,
		Message:   "original message",
	}
	createdPostInUnofficial, _, appErr := th.App.CreatePostAsUser(ctxWithSession, post2, session.Id, true)
	require.Nil(t, appErr)
	require.NotNil(t, createdPostInUnofficial)

	// Now remove permissions for testing
	th.RemovePermissionFromRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

	th.RemovePermissionFromRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)

	th.RemovePermissionFromRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
	defer th.AddPermissionToRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)

	t.Run("CreatePost denied in DM without permission", func(t *testing.T) {
		post := &model.Post{
			UserId:    user1.Id,
			ChannelId: dmChannel.Id,
			Message:   "test message",
		}

		_, _, err := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("CreatePost denied in GM without permission", func(t *testing.T) {
		post := &model.Post{
			UserId:    user1.Id,
			ChannelId: gmChannel.Id,
			Message:   "test message",
		}

		_, _, err := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("CreatePost denied in unofficial channel without permission", func(t *testing.T) {
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUser.Username)
		resetIntegrationAdminUsernameForTesting()

		post := &model.Post{
			UserId:    user1.Id,
			ChannelId: unofficialChannel.Id,
			Message:   "test message",
		}

		_, _, err := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("UpdatePost denied in DM without permission", func(t *testing.T) {
		updatedPost := createdPostInDM.Clone()
		updatedPost.Message = "updated message"

		_, _, err := th.App.UpdatePost(ctxWithSession, updatedPost, nil)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("PatchPost denied in DM without permission", func(t *testing.T) {
		patchedMessage := "patched message"
		patch := &model.PostPatch{
			Message: &patchedMessage,
		}

		_, _, err := th.App.PatchPost(ctxWithSession, createdPostInDM.Id, patch, nil)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("DeletePost denied in DM without permission", func(t *testing.T) {
		_, err := th.App.DeletePost(ctxWithSession, createdPostInDM.Id, user1.Id)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("Bot can create post in DM without permission", func(t *testing.T) {
		bot := th.CreateBot(t)
		botUser, err := th.App.GetUser(bot.UserId)
		require.Nil(t, err)

		post := &model.Post{
			UserId:    botUser.Id,
			ChannelId: dmChannel.Id,
			Message:   "bot message",
		}
		botSession, err := th.App.CreateSession(th.Context, &model.Session{
			UserId: botUser.Id,
			Roles:  model.SystemUserRoleId,
		})
		require.Nil(t, err)

		botctxWithSession := th.Context.WithSession(botSession)
		createdPostByBot, _, err := th.App.CreatePostAsUser(botctxWithSession, post, botSession.Id, true)
		require.Nil(t, err)
		require.NotNil(t, createdPostByBot)
	})

	t.Run("System admin can create post in DM without permission", func(t *testing.T) {
		post := &model.Post{
			UserId:    adminUser.Id,
			ChannelId: dmChannel.Id,
			Message:   "admin message",
		}

		adminctxWithSession := th.Context.WithSession(adminSession)
		createdPostByAdmin, _, err := th.App.CreatePostAsUser(adminctxWithSession, post, adminSession.Id, true)
		require.Nil(t, err)
		require.NotNil(t, createdPostByAdmin)
	})

	t.Run("System admin can create post in unofficial channel without permission", func(t *testing.T) {
		os.Setenv("INTEGRATION_ADMIN_USERNAME", adminUser.Username)
		resetIntegrationAdminUsernameForTesting()

		post := &model.Post{
			UserId:    adminUser.Id,
			ChannelId: unofficialChannel.Id,
			Message:   "admin message",
		}

		adminctxWithSession := th.Context.WithSession(adminSession)
		createdPostByAdmin, _, err := th.App.CreatePostAsUser(adminctxWithSession, post, adminSession.Id, true)
		require.Nil(t, err)
		require.NotNil(t, createdPostByAdmin)
	})

	t.Run("UpdatePost denied in unofficial channel without permission", func(t *testing.T) {
		updatedPost := createdPostInUnofficial.Clone()
		updatedPost.Message = "updated message"

		_, _, err := th.App.UpdatePost(ctxWithSession, updatedPost, nil)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("PatchPost denied in unofficial channel without permission", func(t *testing.T) {
		patchedMessage := "patched message"
		patch := &model.PostPatch{
			Message: &patchedMessage,
		}

		_, _, err := th.App.PatchPost(ctxWithSession, createdPostInUnofficial.Id, patch, nil)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("DeletePost denied in unofficial channel without permission", func(t *testing.T) {
		_, err := th.App.DeletePost(ctxWithSession, createdPostInUnofficial.Id, user1.Id)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})
}
