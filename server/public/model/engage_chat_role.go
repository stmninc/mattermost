// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

// Custom roles
const (
	SystemEngageAdmin = "system_engage_admin"
	TeamEngageAdmin   = "team_engage_admin"
)

// allEngageCustomRoleNames is the single source of truth for all defined Engage custom role names.
// It is unexported to prevent modification from other packages.
var allCustomRoleNames = []string{
	SystemEngageAdmin,
	TeamEngageAdmin,
}

// MakeAllCustomRoleTemplates returns a map of all defined custom role templates.
func MakeAllCustomRoleTemplates() map[string]Role {
	return map[string]Role{
		SystemEngageAdmin: {
			Name:        SystemEngageAdmin,
			DisplayName: SystemEngageAdmin,
			Description: "Grants permissions for Engage Chat features.",
			Permissions: []string{
				PermissionCreateDirectChannel.Id,
				PermissionCreateGroupChannel.Id,
			},
		},
		TeamEngageAdmin: {
			Name:        TeamEngageAdmin,
			DisplayName: TeamEngageAdmin,
			Description: "Grants permissions for Engage Chat features.",
			Permissions: []string{
				PermissionCreatePrivateChannel.Id,
				PermissionCreatePublicChannel.Id,
			},
		},
	}
}

// AllCustomRoleNames returns a slice of all defined custom role names.
func AllCustomRoleNames() []string {
	return allCustomRoleNames
}
