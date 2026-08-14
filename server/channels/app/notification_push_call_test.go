// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/v8/channels/store/storetest/mocks"
)

func TestSendPushNotificationsForCall(t *testing.T) {
	mainHelper.Parallel(t)
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

	// Start mock push proxy server & set proxy server url
	handler := &testPushNotificationHandler{
		t:        t,
		behavior: "simple",
	}
	mockPushServer := httptest.NewServer(http.HandlerFunc(handler.handleReq))
	defer mockPushServer.Close()

	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.EmailSettings.PushNotificationServer = mockPushServer.URL
	})

	userID := model.NewId()
	sessions := []*model.Session{
		{
			Id:           "sess1",
			UserId:       userID,
			DeviceId:     "apple:device1",
			VoipDeviceId: "voip:123",
			ExpiresAt:    model.GetMillis() + 100000000, // Active
		},
		{
			Id:           "sess2",
			UserId:       userID,
			DeviceId:     "apple:device2",
			VoipDeviceId: "voip:123",                    // Same as session1, should be skipped
			ExpiresAt:    model.GetMillis() + 100000000, // Active
		},
		{
			Id:           "sess3",
			UserId:       userID,
			DeviceId:     "apple:device3",
			VoipDeviceId: "voip:123",                    // Same as session1, should be skipped
			ExpiresAt:    model.GetMillis() + 100000000, // Active
		},
		{
			Id:           "sess4",
			UserId:       userID,
			DeviceId:     "apple:device4",
			VoipDeviceId: "voip:456",
			ExpiresAt:    model.GetMillis() + 100000000, // Active
		},
	}

	notification := &model.PushNotification{
		Type:    model.PushTypeMessage,
		SubType: model.PushSubTypeCalls,
	}

	mockSessionStore.On("GetSessionsWithActiveDeviceIds", userID).Return(sessions, nil)
	mockStore.On("Session").Return(&mockSessionStore)

	err := th.App.sendPushNotificationToAllSessions(th.Context, notification, userID, "")
	require.Nil(t, err)

	// for mockPushServer
	time.Sleep(1 * time.Second)
	// handler.numReqs() returns the total counts of http requests received by mockPushServer
	assert.Equal(t, 2, handler.numReqs(), "The number of sessions that should send push notification does not match the actual number of requests")

	mockSessionStore.AssertExpectations(t)
}
