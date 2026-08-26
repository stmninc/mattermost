// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/require"
)

func TestEnableCustomRoles(t *testing.T) {
	th := Setup(t).InitBasic(t)

	assertRolesMatch := func(t *testing.T, actualRoles []*model.Role, expectedRoleTemplates map[string]model.Role) {
		t.Helper()

		require.Len(t, actualRoles, len(expectedRoleTemplates))

		actualRolesMap := make(map[string]*model.Role)
		for _, role := range actualRoles {
			actualRolesMap[role.Name] = role
		}

		for name, expectedRole := range expectedRoleTemplates {
			actualRole, ok := actualRolesMap[name]
			require.True(t, ok)
			require.Equal(t, int64(0), actualRole.DeleteAt)
			require.Equal(t, expectedRole.Name, actualRole.Name)
			require.Equal(t, expectedRole.DisplayName, actualRole.DisplayName)
			require.ElementsMatch(t, expectedRole.Permissions, actualRole.Permissions)
		}
	}

	roleNames := model.AllCustomRoleNames()
	expectedRoles := model.MakeAllCustomRoleTemplates()

	t.Run("create roles for the first time", func(t *testing.T) {
		roles, err := th.App.EnableCustomRoles(th.Context, roleNames)
		require.Nil(t, err)
		require.Len(t, roles, len(expectedRoles))

		// Verify they exist in the database
		dbRoles, err := th.App.GetRolesByNames(roleNames)
		require.Nil(t, err)
		require.Len(t, dbRoles, len(roles))
		assertRolesMatch(t, roles, expectedRoles)
	})

	t.Run("roles already exist and are active", func(t *testing.T) {
		_, err := th.App.EnableCustomRoles(th.Context, roleNames)
		require.Nil(t, err)

		// Call a second time
		roles, err := th.App.EnableCustomRoles(th.Context, roleNames)
		require.Nil(t, err)

		require.Len(t, roles, len(expectedRoles))
		assertRolesMatch(t, roles, expectedRoles)
	})

	t.Run("restore soft-deleted roles", func(t *testing.T) {
		// Enable roles
		roles, err := th.App.EnableCustomRoles(th.Context, roleNames)
		require.Nil(t, err)

		for _, role := range roles {
			// Soft-delete them
			_, deleteErr := th.App.DeleteRole(role.Id)
			require.Nil(t, deleteErr)

			// Verify it's deleted
			deletedRole, getErr := th.App.GetRole(role.Id)
			require.Nil(t, getErr)
			require.NotEqual(t, int64(0), deletedRole.DeleteAt)
		}

		// Call EnableCustomRoles again to trigger restore
		_, err = th.App.EnableCustomRoles(th.Context, roleNames)
		require.Nil(t, err)

		// Verify it's restored
		for _, role := range roles {
			restoredRole, restoreErr := th.App.GetRole(role.Id)
			require.Nil(t, restoreErr)
			require.Equal(t, int64(0), restoredRole.DeleteAt)
		}
	})
}
