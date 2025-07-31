package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestGetMobileAppSessions(t *testing.T) {
	t.Run("should return sessions that have a VoIP DeviceId value.", func(t *testing.T) {
		th := Setup(t).InitBasic()
		defer th.TearDown()

		_, err := th.App.CreateSession(th.Context, &model.Session{
			UserId:    th.BasicUser.Id,
			DeviceId:  "DeviceId",
			VoipDeviceId: "VoipDeviceId",
			ExpiresAt: model.GetMillis() + 100000,
	  })
	  require.Nil(t, err)

		sessions, err := th.App.getMobileAppSessions(th.BasicUser.Id)
		require.Nil(t, err)
		require.Len(t, sessions, 1)
		require.Equal(t, "DeviceId", sessions[0].DeviceId)
		require.Equal(t, "VoipDeviceId", sessions[0].VoipDeviceId)
	})
}
