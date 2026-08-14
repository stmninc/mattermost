// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/v8/channels/store/storetest/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSendNotificationCallEnd(t *testing.T) {
	t.Run("should return early if post type is not custom_calls", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

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
		th := Setup(t).InitBasic(t)

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
		th := Setup(t).InitBasic(t)

		post := &model.Post{
			Id:        model.NewId(),
			ChannelId: th.BasicChannel.Id,
			Type:      "custom_calls",
			Props:     model.StringInterface{"end_at": nil},
		}

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)
	})

	t.Run("should return early if channel is not direct or group", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

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

	t.Run("should handle user with multiple sessions, some expired, some lacks VoIP ID for iOS", func(t *testing.T) {
		th := SetupWithStoreMock(t)

		mockStore := th.App.Srv().Store().(*mocks.Store)
		mockUserStore := mocks.UserStore{}
		mockUserStore.On("Count", mock.Anything).Return(int64(10), nil)
		mockUserStore.On("GetUnreadCount", mock.AnythingOfType("string"), mock.AnythingOfType("bool")).Return(int64(1), nil)
		mockPostStore := mocks.PostStore{}
		mockPostStore.On("GetMaxPostSize").Return(65535, nil)
		mockSystemStore := mocks.SystemStore{}
		mockSystemStore.On("GetByName", "UpgradedFromTE").Return(&model.System{Name: "UpgradedFromTE", Value: "false"}, nil)
		mockSystemStore.On("GetByName", "InstallationDate").Return(&model.System{Name: "InstallationDate", Value: "10"}, nil)
		mockSystemStore.On("GetByName", "FirstServerRunTimestamp").Return(&model.System{Name: "FirstServerRunTimestamp", Value: "10"}, nil)
		mockSystemStore.On("GetByNameWithContext", mock.Anything, "DiagnosticId").Return(&model.System{Name: "DiagnosticId", Value: "diagnostic-id"}, nil)

		mockSessionStore := mocks.SessionStore{}
		mockPreferenceStore := mocks.PreferenceStore{}
		mockPreferenceStore.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&model.Preference{Value: "test"}, nil)
		mockStore.On("User").Return(&mockUserStore)
		mockStore.On("Post").Return(&mockPostStore)
		mockStore.On("System").Return(&mockSystemStore)
		mockStore.On("Session").Return(&mockSessionStore)
		mockStore.On("Preference").Return(&mockPreferenceStore)
		mockStore.On("GetDBSchemaVersion").Return(1, nil)

		directChannel := &model.Channel{
			Id:   model.NewId(),
			Type: model.ChannelTypeDirect,
			Name: "direct_channel",
		}

		senderId := model.NewId()
		userId := model.NewId()

		post := &model.Post{
			Id:        model.NewId(),
			UserId:    senderId,
			ChannelId: directChannel.Id,
			Type:      "custom_calls",
			Props:     model.StringInterface{"end_at": 1234567890},
		}

		channelMembers := model.ChannelMembers{
			{ChannelId: directChannel.Id, UserId: senderId},
			{ChannelId: directChannel.Id, UserId: userId},
		}

		// Create sessions: one active, one expired
		activeSession := &model.Session{
			Id:        model.NewId(),
			UserId:    userId,
			DeviceId:  "active_device",
			ExpiresAt: model.GetMillis() + 100000, // Active
			Props:     model.StringMap{model.SessionPropOs: "Android"},
		}

		expiredSession := &model.Session{
			Id:        model.NewId(),
			UserId:    userId,
			DeviceId:  "expired_device",
			ExpiresAt: model.GetMillis() - 100000, // Expired
			Props:     model.StringMap{model.SessionPropOs: "Android"},
		}

		// Create iOS session: one with VoipID, one without VoipID
		withVoipSession := &model.Session{
			Id:           model.NewId(),
			UserId:       userId,
			DeviceId:     "ios_device",
			ExpiresAt:    model.GetMillis() + 100000, // Active
			VoipDeviceId: "ios_voip_device",
			Props:        model.StringMap{model.SessionPropOs: "iOS"},
		}

		withoutVoipSession := &model.Session{
			Id:           model.NewId(),
			UserId:       userId,
			DeviceId:     "ios_device",
			ExpiresAt:    model.GetMillis() + 100000, // Active
			VoipDeviceId: "",
			Props:        model.StringMap{model.SessionPropOs: "iOS"},
		}

		mockChannelStore := mocks.ChannelStore{}
		mockChannelStore.On("Get", directChannel.Id, true).Return(directChannel, nil)
		mockChannelStore.On("GetMembers", model.ChannelMembersGetOptions{
			ChannelID:    directChannel.Id,
			Offset:       0,
			Limit:        model.ChannelGroupMaxUsers,
			UpdatedAfter: 0,
		}).Return(channelMembers, nil)
		mockStore.On("Channel").Return(&mockChannelStore)

		mockSessionStore.On("GetSessionsWithActiveDeviceIds", userId).Return([]*model.Session{
			activeSession,      // shoud send
			expiredSession,     // should not send
			withVoipSession,    // should send
			withoutVoipSession, // should not send
		}, nil)
		mockStore.On("Session").Return(&mockSessionStore)

		// Start mock push proxy & set proxy server url
		handler := &testPushNotificationHandler{
			t:        t,
			behavior: "simple",
		}
		mockProxyServer := httptest.NewServer(
			http.HandlerFunc(handler.handleReq),
		)
		defer mockProxyServer.Close()

		th.App.UpdateConfig(func(cfg *model.Config) {
			*cfg.EmailSettings.PushNotificationContents = model.GenericNotification
			*cfg.EmailSettings.PushNotificationServer = mockProxyServer.URL
		})

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)

		time.Sleep(1 * time.Second)
		assert.Equal(t, 2, handler.numReqs(), "The number of sessions that should send push notification does not match the actual number of requests")

		mockChannelStore.AssertExpectations(t)
		mockSessionStore.AssertExpectations(t)
	})

	t.Run("should send notification to all channel members except post user", func(t *testing.T) {
		th := SetupWithStoreMock(t)

		teamId := model.NewId()
		directChannel := &model.Channel{
			Id:          model.NewId(),
			Type:        model.ChannelTypeDirect,
			Name:        "direct_channel",
			DisplayName: "Direct Channel",
			TeamId:      teamId,
		}

		senderId := model.NewId()
		sessionUserId := model.NewId()
		member1Id := model.NewId()
		member2Id := model.NewId()

		th.Context.Session().UserId = sessionUserId

		post := &model.Post{
			Id:        model.NewId(),
			UserId:    senderId,
			ChannelId: directChannel.Id,
			Type:      "custom_calls",
			Props: model.StringInterface{
				"end_at": 1234567890,
			},
		}

		channelMembers := model.ChannelMembers{
			{ChannelId: directChannel.Id, UserId: senderId},
			{ChannelId: directChannel.Id, UserId: sessionUserId},
			{ChannelId: directChannel.Id, UserId: member1Id},
			{ChannelId: directChannel.Id, UserId: member2Id},
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
		// getMobileAppSessions should be called for every channel member except the post author (senderId)
		mockSessionStore.On("GetSessionsWithActiveDeviceIds", sessionUserId).Return([]*model.Session{}, nil)
		mockSessionStore.On("GetSessionsWithActiveDeviceIds", member1Id).Return([]*model.Session{}, nil)
		mockSessionStore.On("GetSessionsWithActiveDeviceIds", member2Id).Return([]*model.Session{}, nil)
		mockStore.On("Session").Return(&mockSessionStore)

		err := th.App.SendNotificationCallEnd(th.Context, post)
		require.Nil(t, err)

		mockChannelStore.AssertExpectations(t)
		mockSessionStore.AssertExpectations(t)
	})
}
