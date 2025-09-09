// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/v8/channels/store/storetest/mocks"
	"github.com/stretchr/testify/require"
)

func TestSendNotificationCallEnd(t *testing.T) {
	t.Run("should return early if post type is not custom_calls", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		post := &model.Post{
			Id:        model.NewId(),
			ChannelId: th.BasicChannel.Id,
			Type:      "regular",
			Props:     model.StringInterface{"end_at": 1234567890},
		}

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)
	})

	t.Run("should return early if end_at is not present in props", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		post := &model.Post{
			Id:        model.NewId(),
			ChannelId: th.BasicChannel.Id,
			Type:      "custom_calls",
			Props:     model.StringInterface{},
		}

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)
	})

	t.Run("should return early if end_at is nil in props", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		post := &model.Post{
			Id:        model.NewId(),
			ChannelId: th.BasicChannel.Id,
			Type:      "custom_calls",
			Props:     model.StringInterface{"end_at": nil},
		}

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)
	})

	t.Run("should  early if channel is not direct or group", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		post := &model.Post{
			Id:        model.NewId(),
			ChannelId: th.BasicChannel.Id,
			Type:      "custom_calls",
			Props:     model.StringInterface{"end_at": 1234567890},
		}

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)
	})

	t.Run("should handle error when getting channel", func(t *testing.T) {
		th := SetupWithStoreMock(t)
		defer th.TearDown()

		post := &model.Post{
			Id:        model.NewId(),
			ChannelId: "invalid_channel_id",
			Type:      "custom_calls",
			Props:     model.StringInterface{"end_at": 1234567890},
		}

		mockStore := th.App.Srv().Store().(*mocks.Store)
		mockChannelStore := mocks.ChannelStore{}
		mockChannelStore.On("Get", "invalid_channel_id", true).Return(nil, model.NewAppError("test", "channel.not.found", nil, "", http.StatusNotFound))
		mockStore.On("Channel").Return(&mockChannelStore)

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.NotNil(t, err)
	})

	t.Run("should handle error when getting channel members", func(t *testing.T) {
		th := SetupWithStoreMock(t)
		defer th.TearDown()

		directChannel := &model.Channel{
			Id:   model.NewId(),
			Type: model.ChannelTypeDirect,
			Name: "direct_channel",
		}

		post := &model.Post{
			Id:        model.NewId(),
			ChannelId: directChannel.Id,
			Type:      "custom_calls",
			Props:     model.StringInterface{"end_at": 1234567890},
		}

		mockStore := th.App.Srv().Store().(*mocks.Store)
		mockChannelStore := mocks.ChannelStore{}
		mockChannelStore.On("Get", directChannel.Id, true).Return(directChannel, nil)
		mockChannelStore.On("GetMembers", model.ChannelMembersGetOptions{
			ChannelID:    directChannel.Id,
			Offset:       0,
			Limit:        model.ChannelGroupMaxUsers,
			UpdatedAfter: 0,
		}).Return(nil, model.NewAppError("test", "channel.members.not.found", nil, "", http.StatusNotFound))
		mockStore.On("Channel").Return(&mockChannelStore)

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.NotNil(t, err)
	})

	t.Run("should send notification to channel members excluding participants", func(t *testing.T) {
		th := SetupWithStoreMock(t)
		defer th.TearDown()

		teamId := model.NewId()
		directChannel := &model.Channel{
			Id:          model.NewId(),
			Type:        model.ChannelTypeDirect,
			Name:        "direct_channel",
			DisplayName: "Direct Channel",
			TeamId:      teamId,
		}

		senderId := model.NewId()
		member1Id := model.NewId()
		member2Id := model.NewId()
		member3Id := model.NewId()

		post := &model.Post{
			Id:        model.NewId(),
			UserId:    senderId,
			ChannelId: directChannel.Id,
			Type:      "custom_calls",
			Props: model.StringInterface{
				"end_at":      1234567890,
				"participants": []interface{}{senderId, member1Id, member2Id},
			},
		}

		channelMembers := model.ChannelMembers{
			{ChannelId: directChannel.Id, UserId: senderId},
			{ChannelId: directChannel.Id, UserId: member1Id},
			{ChannelId: directChannel.Id, UserId: member2Id},
			{ChannelId: directChannel.Id, UserId: member3Id},
		}

		mockStore := th.App.Srv().Store().(*mocks.Store)
		mockChannelStore := mocks.ChannelStore{}
		mockChannelStore.On("Get", directChannel.Id, true).Return(directChannel, nil)
		mockChannelStore.On("GetMembers", model.ChannelMembersGetOptions{
			ChannelID:    directChannel.Id,
			Offset:       0,
			Limit:        model.ChannelGroupMaxUsers,
			UpdatedAfter: 0,
		}).Return(channelMembers, nil)
		mockStore.On("Channel").Return(&mockChannelStore)

		mockSessionStore := mocks.SessionStore{}
		mockSessionStore.On("GetSessionsWithActiveDeviceIds", member3Id).Return([]*model.Session{}, nil)
		mockStore.On("Session").Return(&mockSessionStore)

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)

		mockChannelStore.AssertExpectations(t)
		mockSessionStore.AssertExpectations(t)
	})
}
