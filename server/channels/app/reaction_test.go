// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/v8/channels/testlib"
)

func TestSaveReactionForPost(t *testing.T) {
	mainHelper.Parallel(t)
	th := Setup(t).InitBasic()

	post := th.CreatePost(th.BasicChannel)
	reaction1, err := th.App.SaveReactionForPost(th.Context, &model.Reaction{
		UserId:    th.BasicUser.Id,
		PostId:    post.Id,
		EmojiName: "cry",
	})
	require.NotNil(t, reaction1)
	require.Nil(t, err)
	reaction2, err := th.App.SaveReactionForPost(th.Context, &model.Reaction{
		UserId:    th.BasicUser.Id,
		PostId:    post.Id,
		EmojiName: "smile",
	})
	require.NotNil(t, reaction2)
	require.Nil(t, err)
	reaction3, err := th.App.SaveReactionForPost(th.Context, &model.Reaction{
		UserId:    th.BasicUser.Id,
		PostId:    post.Id,
		EmojiName: "rofl",
	})
	require.NotNil(t, reaction3)
	require.Nil(t, err)

	t.Run("should not add reaction if it does not exist on the system", func(t *testing.T) {
		reaction := &model.Reaction{
			UserId:    th.BasicUser.Id,
			PostId:    th.BasicPost.Id,
			EmojiName: "definitely-not-a-real-emoji",
		}

		result, err := th.App.SaveReactionForPost(th.Context, reaction)
		require.NotNil(t, err)
		require.Nil(t, result)
	})

	t.Run("should not add reaction if we are over the limit", func(t *testing.T) {
		var originalLimit *int
		th.UpdateConfig(func(cfg *model.Config) {
			originalLimit = cfg.ServiceSettings.UniqueEmojiReactionLimitPerPost
			*cfg.ServiceSettings.UniqueEmojiReactionLimitPerPost = 3
		})
		defer th.UpdateConfig(func(cfg *model.Config) {
			cfg.ServiceSettings.UniqueEmojiReactionLimitPerPost = originalLimit
		})

		reaction := &model.Reaction{
			UserId:    th.BasicUser.Id,
			PostId:    post.Id,
			EmojiName: "joy",
		}

		result, err := th.App.SaveReactionForPost(th.Context, reaction)
		require.NotNil(t, err)
		require.Nil(t, result)
	})

	t.Run("should always add reaction if we are over the limit but the reaction is not unique", func(t *testing.T) {
		user := th.CreateUser()

		var originalLimit *int
		th.UpdateConfig(func(cfg *model.Config) {
			originalLimit = cfg.ServiceSettings.UniqueEmojiReactionLimitPerPost
			*cfg.ServiceSettings.UniqueEmojiReactionLimitPerPost = 3
		})
		defer th.UpdateConfig(func(cfg *model.Config) {
			cfg.ServiceSettings.UniqueEmojiReactionLimitPerPost = originalLimit
		})

		reaction := &model.Reaction{
			UserId:    user.Id,
			PostId:    post.Id,
			EmojiName: "cry",
		}

		result, err := th.App.SaveReactionForPost(th.Context, reaction)
		require.Nil(t, err)
		require.NotNil(t, result)
	})
}

func TestSharedChannelSyncForReactionActions(t *testing.T) {
	mainHelper.Parallel(t)
	t.Run("adding a reaction in a shared channel performs a content sync when sync service is running on that node", func(t *testing.T) {
		th := setupSharedChannels(t).InitBasic()

		sharedChannelService := NewMockSharedChannelService(th.Server.GetSharedChannelSyncService())
		th.Server.SetSharedChannelSyncService(sharedChannelService)
		testCluster := &testlib.FakeClusterInterface{}
		th.Server.Platform().SetCluster(testCluster)

		user := th.BasicUser

		channel := th.CreateChannel(th.Context, th.BasicTeam, WithShared(true))

		post, err := th.App.CreatePost(th.Context, &model.Post{
			UserId:    user.Id,
			ChannelId: channel.Id,
			Message:   "Hello folks",
		}, channel, model.CreatePostFlags{SetOnline: true})
		require.Nil(t, err, "Creating a post should not error")

		reaction := &model.Reaction{
			UserId:    user.Id,
			PostId:    post.Id,
			EmojiName: "+1",
		}

		_, err = th.App.SaveReactionForPost(th.Context, reaction)
		require.Nil(t, err, "Adding a reaction should not error")

		th.TearDown() // We need to enforce teardown because reaction instrumentation happens in a goroutine

		assert.Len(t, sharedChannelService.channelNotifications, 2)
		assert.Equal(t, channel.Id, sharedChannelService.channelNotifications[0])
		assert.Equal(t, channel.Id, sharedChannelService.channelNotifications[1])
	})

	t.Run("removing a reaction in a shared channel performs a content sync when sync service is running on that node", func(t *testing.T) {
		th := setupSharedChannels(t).InitBasic()

		sharedChannelService := NewMockSharedChannelService(th.Server.GetSharedChannelSyncService())
		th.Server.SetSharedChannelSyncService(sharedChannelService)
		testCluster := &testlib.FakeClusterInterface{}
		th.Server.Platform().SetCluster(testCluster)

		user := th.BasicUser

		channel := th.CreateChannel(th.Context, th.BasicTeam, WithShared(true))

		post, err := th.App.CreatePost(th.Context, &model.Post{
			UserId:    user.Id,
			ChannelId: channel.Id,
			Message:   "Hello folks",
		}, channel, model.CreatePostFlags{SetOnline: true})
		require.Nil(t, err, "Creating a post should not error")

		reaction := &model.Reaction{
			UserId:    user.Id,
			PostId:    post.Id,
			EmojiName: "+1",
		}

		err = th.App.DeleteReactionForPost(th.Context, reaction)
		require.Nil(t, err, "Adding a reaction should not error")

		th.TearDown() // We need to enforce teardown because reaction instrumentation happens in a goroutine

		assert.Len(t, sharedChannelService.channelNotifications, 2)
		assert.Equal(t, channel.Id, sharedChannelService.channelNotifications[0])
		assert.Equal(t, channel.Id, sharedChannelService.channelNotifications[1])
	})
}

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

func (th *TestHelper) UpdateConfig(f func(*model.Config)) {
	if th.ConfigStore.IsReadOnly() {
		return
	}
	old := th.ConfigStore.Get()
	updated := old.Clone()
	f(updated)
	if _, _, err := th.ConfigStore.Set(updated); err != nil {
		panic(err)
	}
}
