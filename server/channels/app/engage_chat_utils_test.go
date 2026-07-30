// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

func TestIsChannelAccessible(t *testing.T) {
	th := Setup(t).InitBasic(t)

	// 1. --- Roles and Test Users Setup ---
	_, err := th.App.EnableCustomRoles(th.Context, []string{model.SystemEngageAdmin, model.TeamEngageAdmin})
	require.Nil(t, err)

	restrictedUser := th.CreateUser(t)
	th.LinkUserToTeam(t, restrictedUser, th.BasicTeam)

	botRec := th.CreateBot(t)
	botUser, err := th.App.GetUser(botRec.UserId)
	require.Nil(t, err)

	// User with SystemEngageAdmin role
	systemExceptionUser := th.CreateUser(t)
	th.LinkUserToTeam(t, systemExceptionUser, th.BasicTeam)
	_, err = th.App.UpdateUserRoles(th.Context, systemExceptionUser.Id, model.SystemUserRoleId+" "+model.SystemEngageAdmin, false)
	require.Nil(t, err)

	// User with no special roles
	regularUser := th.CreateUser(t)
	th.LinkUserToTeam(t, regularUser, th.BasicTeam)

	session, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: restrictedUser.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, err)
	ctxWithSession := th.Context.WithSession(session)

	botSession, err := th.App.CreateSession(th.Context, &model.Session{
		UserId: botUser.Id,
		Roles:  model.SystemUserRoleId,
	})
	require.Nil(t, err)
	ctxWithBotSession := th.Context.WithSession(botSession)

	// 2. --- Channels Setup ---
	publicChannel, err := th.App.CreateChannel(th.Context, &model.Channel{
		TeamId: th.BasicTeam.Id, DisplayName: "Public Channel", Name: "public-channel-" + model.NewId(), Type: model.ChannelTypeOpen,
	}, false)
	require.Nil(t, err)
	th.AddUserToChannel(t, restrictedUser, publicChannel)

	// --- Channels for permission tests ---
	dmChannel, _ := th.App.GetOrCreateDirectChannel(ctxWithSession, restrictedUser.Id, regularUser.Id)
	dmWithException, _ := th.App.GetOrCreateDirectChannel(ctxWithSession, restrictedUser.Id, systemExceptionUser.Id)
	gmChannel := th.CreateGroupChannel(t, restrictedUser, regularUser)
	gmWithException := th.CreateGroupChannel(t, restrictedUser, systemExceptionUser)
	privateChannel := th.CreatePrivateChannel(t, th.BasicTeam)
	th.AddUserToChannel(t, restrictedUser, privateChannel)

	// 3. --- Run Test Cases ---
	t.Run("General Access Scenarios", func(t *testing.T) {
		testCases := []struct {
			name          string
			channelID     string
			expectedError bool
		}{
			{"Invalid channel ID should return an error", "invalid-channel-id", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				accessible, err := th.App.IsChannelAccessible(ctxWithSession, tc.channelID)
				if tc.expectedError {
					require.NotNil(t, err)
				} else {
					require.Nil(t, err)
				}
				require.False(t, accessible)
			})
		}
	})

	t.Run("Situation no restriction", func(t *testing.T) {
		testCases := []struct {
			name     string
			channel  *model.Channel
			ctx      request.CTX
			expected bool
		}{
			{"DM Channel", dmChannel, ctxWithSession, true},
			{"Group Channel", gmChannel, ctxWithSession, true},
			{"Private Channel", privateChannel, ctxWithSession, true},
			{"Bot user should always be accessible", publicChannel, ctxWithBotSession, true},
			{"Public channel should always be accessible", publicChannel, ctxWithSession, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				accessible, err := th.App.IsChannelAccessible(tc.ctx, tc.channel.Id)
				require.Nil(t, err)
				require.Equal(t, tc.expected, accessible)
			})
		}
	})

	t.Run("Situation under restriction", func(t *testing.T) {
		// Remove permissions for the restricted user (setup once for all sub-scenarios)
		th.RemovePermissionFromRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		th.RemovePermissionFromRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
		th.RemovePermissionFromRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
		defer func() {
			th.AddPermissionToRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
			th.AddPermissionToRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
			th.AddPermissionToRole(t, model.PermissionCreatePrivateChannel.Id, model.TeamUserRoleId)
		}()

		t.Run("without exception", func(t *testing.T) {
			testCases := []struct {
				name     string
				channel  *model.Channel
				ctx      request.CTX
				expected bool
			}{
				{"DM without exception should NOT be accessible", dmChannel, ctxWithSession, false},
				{"Group without exception should NOT be accessible", gmChannel, ctxWithSession, false},
				{"Private channel should always be accessible", privateChannel, ctxWithSession, true},
				{"Public channel should always be accessible", publicChannel, ctxWithSession, true},
				{"Bot user should always be accessible", publicChannel, ctxWithBotSession, true},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					accessible, err := th.App.IsChannelAccessible(tc.ctx, tc.channel.Id)
					require.Nil(t, err)
					require.Equal(t, tc.expected, accessible)
				})
			}
		})

		t.Run("with exception", func(t *testing.T) {
			testCases := []struct {
				name     string
				channel  *model.Channel
				expected bool
			}{
				{"DM with exception member should be accessible", dmWithException, true},
				{"Group with exception member should be accessible", gmWithException, true},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					accessible, err := th.App.IsChannelAccessible(ctxWithSession, tc.channel.Id)
					require.Nil(t, err)
					require.Equal(t, tc.expected, accessible)
				})
			}
		})
	})

	t.Run("Read-only Town Square", func(t *testing.T) {
		townSquare, err := th.App.GetChannelByName(th.Context, model.DefaultChannelName, th.BasicTeam.Id, false)
		require.Nil(t, err)

		adminSession, err := th.App.CreateSession(th.Context, &model.Session{
			UserId: th.SystemAdminUser.Id,
			Roles:  model.SystemUserRoleId + " " + model.SystemAdminRoleId,
		})
		require.Nil(t, err)
		ctxWithAdminSession := th.Context.WithSession(adminSession)

		t.Run("disabled: everyone can access town-square", func(t *testing.T) {
			t.Setenv(model.TownSquareReadOnlyEnvVar, "")
			model.ResetTownSquareReadOnlyCache()
			accessible, err := th.App.IsChannelAccessible(ctxWithSession, townSquare.Id)
			require.Nil(t, err)
			require.True(t, accessible)
		})

		t.Run("enabled", func(t *testing.T) {
			t.Setenv(model.TownSquareReadOnlyEnvVar, "true")
			model.ResetTownSquareReadOnlyCache()
			defer model.ResetTownSquareReadOnlyCache()

			testCases := []struct {
				name     string
				ctx      request.CTX
				expected bool
			}{
				{"regular user cannot post to town-square", ctxWithSession, false},
				{"system admin can post to town-square", ctxWithAdminSession, true},
				{"bot can post to town-square", ctxWithBotSession, true},
			}
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					accessible, err := th.App.IsChannelAccessible(tc.ctx, townSquare.Id)
					require.Nil(t, err)
					require.Equal(t, tc.expected, accessible)
				})
			}

			t.Run("other public channel remains accessible", func(t *testing.T) {
				accessible, err := th.App.IsChannelAccessible(ctxWithSession, publicChannel.Id)
				require.Nil(t, err)
				require.True(t, accessible)
			})
		})
	})
}
