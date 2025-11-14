// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

type testWriter struct {
	write func(p []byte) (int, error)
}

func (tw testWriter) Write(p []byte) (int, error) {
	return tw.write(p)
}

func TestExportPermissions(t *testing.T) {
	mainHelper.Parallel(t)
	th := Setup(t)
	defer th.TearDown()

	var scheme *model.Scheme
	var roles []*model.Role
	withMigrationMarkedComplete(t, th, func() {
		scheme, roles = th.CreateScheme()
	})

	results := [][]byte{}

	tw := testWriter{
		write: func(p []byte) (int, error) {
			results = append(results, p)
			return len(p), nil
		},
	}

	err := th.App.ExportPermissions(tw)
	if err != nil {
		t.Error(err)
	}

	if len(results) == 0 {
		t.Error("Expected export to have returned something.")
	}

	firstResult := results[0]

	var row map[string]any
	err = json.Unmarshal(firstResult, &row)
	if err != nil {
		t.Error(err)
	}

	getRoleByName := func(name string) string {
		for _, role := range roles {
			if role.Name == name {
				return role.Name
			}
		}
		return ""
	}

	expectations := map[string]func(str string) string{
		scheme.DisplayName:             func(_ string) string { return row["display_name"].(string) },
		scheme.Name:                    func(_ string) string { return row["name"].(string) },
		scheme.Description:             func(_ string) string { return row["description"].(string) },
		scheme.Scope:                   func(_ string) string { return row["scope"].(string) },
		scheme.DefaultTeamAdminRole:    func(str string) string { return getRoleByName(str) },
		scheme.DefaultTeamUserRole:     func(str string) string { return getRoleByName(str) },
		scheme.DefaultTeamGuestRole:    func(str string) string { return getRoleByName(str) },
		scheme.DefaultChannelAdminRole: func(str string) string { return getRoleByName(str) },
		scheme.DefaultChannelUserRole:  func(str string) string { return getRoleByName(str) },
		scheme.DefaultChannelGuestRole: func(str string) string { return getRoleByName(str) },
	}

	for key, valF := range expectations {
		expected := key
		actual := valF(key)
		if actual != expected {
			t.Errorf("Expected %v but got %v.", expected, actual)
		}
	}
}

func TestMigration(t *testing.T) {
	mainHelper.Parallel(t)
	th := Setup(t)
	defer th.TearDown()

	role, err := th.App.GetRoleByName(context.Background(), model.SystemAdminRoleId)
	require.Nil(t, err)
	assert.Contains(t, role.Permissions, model.PermissionCreateEmojis.Id)
	assert.Contains(t, role.Permissions, model.PermissionDeleteEmojis.Id)
	assert.Contains(t, role.Permissions, model.PermissionDeleteOthersEmojis.Id)
	assert.Contains(t, role.Permissions, model.PermissionUseGroupMentions.Id)

	appErr := th.App.ResetPermissionsSystem()
	require.Nil(t, appErr)

	role, err = th.App.GetRoleByName(context.Background(), model.SystemAdminRoleId)
	require.Nil(t, err)
	assert.Contains(t, role.Permissions, model.PermissionCreateEmojis.Id)
	assert.Contains(t, role.Permissions, model.PermissionDeleteEmojis.Id)
	assert.Contains(t, role.Permissions, model.PermissionDeleteOthersEmojis.Id)
	assert.Contains(t, role.Permissions, model.PermissionUseGroupMentions.Id)
}

func withMigrationMarkedComplete(t *testing.T, th *TestHelper, f func()) {
	// Mark the migration as done.
	_, err := th.App.Srv().Store().System().PermanentDeleteByName(model.MigrationKeyAdvancedPermissionsPhase2)
	require.NoError(t, err)
	err = th.App.Srv().Store().System().Save(&model.System{Name: model.MigrationKeyAdvancedPermissionsPhase2, Value: "true"})
	require.NoError(t, err)
	// Un-mark the migration at the end of the test.
	defer func() {
		_, err := th.App.Srv().Store().System().PermanentDeleteByName(model.MigrationKeyAdvancedPermissionsPhase2)
		require.NoError(t, err)
	}()
	f()
}

func TestCheckDMGMChannelPermissions(t *testing.T) {
	mainHelper.Parallel(t)
	th := Setup(t)
	defer th.TearDown()

	// Create test users
	user1 := th.CreateUser()
	user2 := th.CreateUser()
	botUser := th.CreateUser()
	botUser.IsBot = true
	_, err := th.App.UpdateUser(th.Context, botUser, false)
	require.Nil(t, err)

	// Create a DM channel
	dmChannel, appErr := th.App.GetOrCreateDirectChannel(th.Context, user1.Id, user2.Id)
	require.Nil(t, appErr)
	require.NotNil(t, dmChannel)

	// Create a GM channel
	gmChannel, appErr := th.App.createGroupChannel(th.Context, []string{user1.Id, user2.Id, botUser.Id}, user1.Id)
	require.Nil(t, appErr)
	require.NotNil(t, gmChannel)

	t.Run("should allow DM access with CREATE_DIRECT_CHANNEL permission", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should allow GM access with CREATE_GROUP_CHANNEL permission", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, gmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should allow bots to access DM channels", func(t *testing.T) {
		// Create a context with bot's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: botUser.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, botUser.Id)
		assert.Nil(t, err)
	})

	t.Run("should skip check with nil channel", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, nil, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should skip check with nil session", func(t *testing.T) {
		// Create a context without a session (or with nil session)
		err := th.App.CheckDMGMChannelPermissions(th.Context, dmChannel, user1.Id)
		// Should skip check when session is nil
		assert.Nil(t, err)
	})

	t.Run("should allow system admin to access DM channels", func(t *testing.T) {
		// Create a context with admin session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemAdminRoleId})

		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, user1.Id)
		assert.Nil(t, err)
	})

	t.Run("should handle non-existent user gracefully", func(t *testing.T) {
		// Create a context with user1's session
		ctxWithSession := th.Context.WithSession(&model.Session{UserId: user1.Id, Roles: model.SystemUserRoleId})

		// Pass a non-existent user ID
		err := th.App.CheckDMGMChannelPermissions(ctxWithSession, dmChannel, "nonexistent-user-id")
		// Should not error out, should continue to permission checks
		assert.Nil(t, err)
	})
}
