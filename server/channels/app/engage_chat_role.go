// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// EnableCustomRoles ensures that the given custom roles are active.
// It creates them if they don't exist, or restores them if they were soft-deleted.
func (a *App) EnableCustomRoles(rctx request.CTX, roleNames []string) ([]*model.Role, *model.AppError) {
	if len(roleNames) == 0 {
		return []*model.Role{}, nil
	}
	// Initialize custom role templates
	customRoleTemplates := model.MakeAllCustomRoleTemplates()

	// Get existing roles from the DB in a single batch.
	existingRoles, err := a.GetRolesByNames(roleNames)
	if err != nil {
		return nil, err
	}

	roleMap := make(map[string]*model.Role)
	for _, role := range existingRoles {
		roleMap[role.Name] = role
	}

	enabledRoles := make([]*model.Role, 0, len(roleNames))

	for _, rolename := range roleNames {
		role, exists := roleMap[rolename]

		if exists && role.DeleteAt == 0 {
			enabledRoles = append(enabledRoles, role)
			continue
		} else if exists && role.DeleteAt > 0 {
			restoredRole, err := a.restoreCustomRole(rctx, role)
			if err != nil {
				return nil, err
			}
			enabledRoles = append(enabledRoles, restoredRole)
		} else if !exists {
			customRole, ok := customRoleTemplates[rolename]
			if !ok {
				rctx.Logger().Warn("Custom role definition not found, skipping creation.", mlog.String("rolename", rolename))
				continue
			}
			createdRole, err := a.createCustomRole(rctx, &customRole)
			if err != nil {
				return nil, err
			}
			enabledRoles = append(enabledRoles, createdRole)
		}
	}

	return enabledRoles, nil
}

func (a *App) createCustomRole(rctx request.CTX, role *model.Role) (*model.Role, *model.AppError) {
	role, err := a.CreateRole(role)
	if err != nil {
		return nil, err
	}

	if appErr := a.sendUpdatedRoleEvent(role); appErr != nil {
		return nil, appErr
	}

	rctx.Logger().Info("Created custom role for engage-chat",
		mlog.String("role_id", role.Id),
		mlog.String("rolename", role.Name),
		mlog.String("display_name", role.DisplayName),
		mlog.String("description", role.Description),
		mlog.Array("permissions", role.Permissions),
		mlog.Bool("scheme_managed", role.SchemeManaged),
		mlog.Bool("built_in", role.BuiltIn),
	)
	return role, nil
}

func (a *App) restoreCustomRole(rctx request.CTX, role *model.Role) (*model.Role, *model.AppError) {
	role.DeleteAt = 0
	role, err := a.UpdateRole(role)
	if err != nil {
		return nil, err
	}

	if appErr := a.sendUpdatedRoleEvent(role); appErr != nil {
		return nil, appErr
	}

	rctx.Logger().Info("Restored custom role for engage-chat",
		mlog.String("role_id", role.Id),
		mlog.String("rolename", role.Name),
		mlog.String("display_name", role.DisplayName),
		mlog.String("description", role.Description),
		mlog.Array("permissions", role.Permissions),
		mlog.Bool("scheme_managed", role.SchemeManaged),
		mlog.Bool("built_in", role.BuiltIn),
	)
	return role, nil
}
