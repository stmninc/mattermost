// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestDMGMPermissionChecksForPosts(t *testing.T) {
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

	// Now remove permissions for testing
	th.RemovePermissionFromRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

	th.RemovePermissionFromRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
	defer th.AddPermissionToRole(model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)

	t.Run("CreatePost denied in DM without permission", func(t *testing.T) {
		post := &model.Post{
			UserId:    user1.Id,
			ChannelId: dmChannel.Id,
			Message:   "test message",
		}

		session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		_, err := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	t.Run("CreatePost denied in GM without permission", func(t *testing.T) {
		post := &model.Post{
			UserId:    user1.Id,
			ChannelId: gmChannel.Id,
			Message:   "test message",
		}

		session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		_, err := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
		require.NotNil(t, err)
		require.Equal(t, http.StatusForbidden, err.StatusCode)
	})

	// Create a post in DM for update/patch/delete tests (temporarily restore permissions)
	th.AddPermissionToRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
	post := &model.Post{
		UserId:    user1.Id,
		ChannelId: dmChannel.Id,
		Message:   "original message",
	}
	session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
	ctxWithSession := th.Context.WithSession(session)
	createdPost, appErr := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
	require.Nil(t, appErr)
	require.NotNil(t, createdPost)
	th.RemovePermissionFromRole(model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

	t.Run("UpdatePost denied in DM without permission", func(t *testing.T) {
		updatedPost := createdPost.Clone()
		updatedPost.Message = "updated message"

		session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		_, appErr := th.App.UpdatePost(ctxWithSession, updatedPost, nil)
		require.NotNil(t, appErr)
		require.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})

	t.Run("PatchPost denied in DM without permission", func(t *testing.T) {
		patchedMessage := "patched message"
		patch := &model.PostPatch{
			Message: &patchedMessage,
		}

		session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		_, appErr := th.App.PatchPost(ctxWithSession, createdPost.Id, patch, nil)
		require.NotNil(t, appErr)
		require.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})

	t.Run("DeletePost denied in DM without permission", func(t *testing.T) {
		session := &model.Session{Id: model.NewId(), UserId: user1.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		_, appErr := th.App.DeletePost(ctxWithSession, createdPost.Id, user1.Id)
		require.NotNil(t, appErr)
		require.Equal(t, http.StatusForbidden, appErr.StatusCode)
	})

	t.Run("Bot can create post in DM without permission", func(t *testing.T) {
		bot := th.CreateBot()
		botUser, appErr := th.App.GetUser(bot.UserId)
		require.Nil(t, appErr)

		post := &model.Post{
			UserId:    botUser.Id,
			ChannelId: dmChannel.Id,
			Message:   "bot message",
		}

		session := &model.Session{Id: model.NewId(), UserId: botUser.Id, Roles: model.SystemUserRoleId}
		ctxWithSession := th.Context.WithSession(session)
		createdPost, appErr := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
		require.Nil(t, appErr)
		require.NotNil(t, createdPost)
	})

	t.Run("System admin can create post in DM without permission", func(t *testing.T) {
		adminUser := th.SystemAdminUser

		post := &model.Post{
			UserId:    adminUser.Id,
			ChannelId: dmChannel.Id,
			Message:   "admin message",
		}

		session := &model.Session{Id: model.NewId(), UserId: adminUser.Id, Roles: model.SystemAdminRoleId}
		ctxWithSession := th.Context.WithSession(session)
		createdPost, appErr := th.App.CreatePostAsUser(ctxWithSession, post, session.Id, true)
		require.Nil(t, appErr)
		require.NotNil(t, createdPost)
	})
}
