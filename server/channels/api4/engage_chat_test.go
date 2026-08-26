// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestEnableCustomRoles(t *testing.T) {
	th := Setup(t)

	roleNames := []string{model.SystemEngageAdmin, model.TeamEngageAdmin}

	t.Run("as regular user", func(t *testing.T) {
		_, resp, err := th.Client.EnableCustomRoles(context.Background(), roleNames)
		require.Error(t, err)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("websocket event on role creation", func(t *testing.T) {
		webSocketClient := th.CreateConnectedWebSocketClient(t)

		// Create role
		_, resp, err := th.SystemAdminClient.EnableCustomRoles(context.Background(), []string{model.SystemEngageAdmin})
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		assertExpectedWebsocketEvent(t, webSocketClient, model.WebsocketEventRoleUpdated, func(event *model.WebSocketEvent) {
			roleStr, ok := event.GetData()["role"].(string)
			require.True(t, ok, "expected role string")
			assert.Contains(t, roleStr, model.SystemEngageAdmin)
		})
	})

	t.Run("websocket event on role restoration", func(t *testing.T) {
		webSocketClient := th.CreateConnectedWebSocketClient(t)

		// Ensure the role exists first
		_, resp, err := th.SystemAdminClient.EnableCustomRoles(context.Background(), []string{model.TeamEngageAdmin})
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		// Flush the creation event so it doesn't interfere
		assertExpectedWebsocketEvent(t, webSocketClient, model.WebsocketEventRoleUpdated, func(event *model.WebSocketEvent) {})

		// Soft-delete the role
		role, err2 := th.App.GetRoleByName(th.Context, model.TeamEngageAdmin)
		require.Nil(t, err2)
		_, err2 = th.App.DeleteRole(role.Id)
		require.Nil(t, err2)

		// Restore the role
		_, resp, err = th.SystemAdminClient.EnableCustomRoles(context.Background(), []string{model.TeamEngageAdmin})
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		// Verify restoration event
		assertExpectedWebsocketEvent(t, webSocketClient, model.WebsocketEventRoleUpdated, func(event *model.WebSocketEvent) {
			roleStr, ok := event.GetData()["role"].(string)
			require.True(t, ok, "expected role string")
			assert.Contains(t, roleStr, model.TeamEngageAdmin)
		})
	})

	t.Run("as admin user", func(t *testing.T) {
		returnedRoles, resp, err := th.SystemAdminClient.EnableCustomRoles(context.Background(), roleNames)
		require.NoError(t, err)
		CheckOKStatus(t, resp)

		expectedRolesMap := model.MakeAllCustomRoleTemplates()
		require.Len(t, returnedRoles, len(expectedRolesMap))

		for _, returnedRole := range returnedRoles {
			expectedRole, ok := expectedRolesMap[returnedRole.Name]
			require.True(t, ok)

			assert.Equal(t, expectedRole.DisplayName, returnedRole.DisplayName)
			assert.ElementsMatch(t, expectedRole.Permissions, returnedRole.Permissions)
		}
	})
}

func TestCreatePublicChannelWithTeamEngageAdmin(t *testing.T) {
	th := Setup(t).InitBasic(t)

	// Enable custom roles and create the TeamEngageAdmin role.
	_, appErr := th.App.EnableCustomRoles(th.Context, []string{model.TeamEngageAdmin})
	require.Nil(t, appErr)

	// Save and restore default role permissions around the test.
	defer th.AddPermissionToRole(t, model.PermissionCreatePublicChannel.Id, model.TeamUserRoleId)

	// Remove PermissionCreatePublicChannel from team_user so regular users cannot create public channels.
	th.RemovePermissionFromRole(t, model.PermissionCreatePublicChannel.Id, model.TeamUserRoleId)

	// Create a user without special roles — should be blocked.
	regularUser := th.CreateUser(t)
	th.LinkUserToTeam(t, regularUser, th.BasicTeam)

	// Create a user with TeamEngageAdmin as a team member role — should be allowed.
	engageUser := th.CreateUser(t)
	th.LinkUserToTeam(t, engageUser, th.BasicTeam)
	resp, err := th.SystemAdminClient.UpdateTeamMemberRoles(context.Background(), th.BasicTeam.Id, engageUser.Id, model.TeamUserRoleId+" "+model.TeamEngageAdmin)
	require.NoError(t, err)
	CheckOKStatus(t, resp)

	t.Run("regular user cannot create public channel when permission is removed", func(t *testing.T) {
		client := th.CreateClient()
		_, _, loginErr := client.Login(context.Background(), regularUser.Email, regularUser.Password)
		require.NoError(t, loginErr)

		channel := &model.Channel{
			DisplayName: "No Permission Channel",
			Name:        GenerateTestChannelName(),
			Type:        model.ChannelTypeOpen,
			TeamId:      th.BasicTeam.Id,
		}
		_, resp, createErr := client.CreateChannel(context.Background(), channel)
		require.Error(t, createErr)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("user with TeamEngageAdmin team role can create public channel", func(t *testing.T) {
		client := th.CreateClient()
		_, _, loginErr := client.Login(context.Background(), engageUser.Email, engageUser.Password)
		require.NoError(t, loginErr)

		channel := &model.Channel{
			DisplayName: "TeamEngageAdmin Channel",
			Name:        GenerateTestChannelName(),
			Type:        model.ChannelTypeOpen,
			TeamId:      th.BasicTeam.Id,
		}
		created, _, createErr := client.CreateChannel(context.Background(), channel)
		require.NoError(t, createErr)
		assert.Equal(t, model.ChannelTypeOpen, created.Type)
	})
}

func TestGetChannelAccessible(t *testing.T) {
	th := Setup(t).InitBasic(t)

	// Enable custom roles
	_, appErr := th.App.EnableCustomRoles(th.Context, []string{model.SystemEngageAdmin, model.TeamEngageAdmin})
	require.Nil(t, appErr)

	// Setup exception user with SystemEngageAdmin
	exceptionUser := th.CreateUser(t)
	th.LinkUserToTeam(t, exceptionUser, th.BasicTeam)
	_, appErr = th.App.UpdateUserRoles(th.Context, exceptionUser.Id, model.SystemUserRoleId+" "+model.SystemEngageAdmin, false)
	require.Nil(t, appErr)

	// DM without exception member
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, th.BasicUser.Id, th.BasicUser2.Id)
	require.Nil(t, appErr)

	// DM with exception member
	dmWithException, appErr := th.App.GetOrCreateDirectChannel(th.Context, th.BasicUser.Id, exceptionUser.Id)
	require.Nil(t, appErr)

	t.Run("unauthenticated request returns 401", func(t *testing.T) {
		client := th.CreateClient()
		_, resp, err := client.GetChannelAccessible(context.Background(), th.BasicChannel.Id)
		require.Error(t, err)
		CheckUnauthorizedStatus(t, resp)
	})

	t.Run("invalid channel ID returns 400", func(t *testing.T) {
		_, resp, err := th.Client.GetChannelAccessible(context.Background(), "invalidid")
		require.Error(t, err)
		CheckBadRequestStatus(t, resp)
	})

	t.Run("non-existent channel ID returns error", func(t *testing.T) {
		_, resp, err := th.Client.GetChannelAccessible(context.Background(), model.NewId())
		require.Error(t, err)
		CheckNotFoundStatus(t, resp)
	})

	t.Run("public channel is always accessible", func(t *testing.T) {
		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), th.BasicChannel.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.True(t, accessible)
	})

	t.Run("DM channel is accessible when no restriction", func(t *testing.T) {
		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), dmChannel.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.True(t, accessible)
	})

	t.Run("under restriction: DM without exception is not accessible", func(t *testing.T) {
		th.RemovePermissionFromRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), dmChannel.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.False(t, accessible)
	})

	t.Run("under restriction: DM with exception member is accessible", func(t *testing.T) {
		th.RemovePermissionFromRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)

		accessible, resp, err := th.Client.GetChannelAccessible(context.Background(), dmWithException.Id)
		require.NoError(t, err)
		CheckOKStatus(t, resp)
		assert.True(t, accessible)
	})
}
