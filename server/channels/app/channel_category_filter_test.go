// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
)

func TestFilterSidebarCategories(t *testing.T) {
	th := Setup(t).InitBasic(t)

	t.Run("user with DM permissions should not filter", func(t *testing.T) {
		session := &model.Session{
			Roles: model.SystemUserRoleId,
		}
		ctx := th.Context.WithSession(session)

		categories := &model.OrderedSidebarCategories{
			Categories: model.SidebarCategoriesWithChannels{
				{
					SidebarCategory: model.SidebarCategory{
						Type: model.SidebarCategoryDirectMessages,
					},
					Channels: []string{}, // empty DM category
				},
			},
		}

		th.App.FilterSidebarCategories(ctx, categories)
		require.Len(t, categories.Categories, 1)
	})

	t.Run("user without DM permissions and empty DM category should filter", func(t *testing.T) {
		th.RemovePermissionFromRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		th.RemovePermissionFromRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)

		session := &model.Session{
			Roles: model.SystemUserRoleId,
		}
		ctx := th.Context.WithSession(session)

		categories := &model.OrderedSidebarCategories{
			Categories: model.SidebarCategoriesWithChannels{
				{
					SidebarCategory: model.SidebarCategory{
						Type: model.SidebarCategoryDirectMessages,
					},
					Channels: []string{}, // empty DM category
				},
			},
		}

		th.App.FilterSidebarCategories(ctx, categories)
		require.Len(t, categories.Categories, 0)
	})

	t.Run("user without DM permissions and have DM category should not filter", func(t *testing.T) {
		th.RemovePermissionFromRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		th.RemovePermissionFromRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreateDirectChannel.Id, model.SystemUserRoleId)
		defer th.AddPermissionToRole(t, model.PermissionCreateGroupChannel.Id, model.SystemUserRoleId)

		session := &model.Session{
			Roles: model.SystemUserRoleId,
		}
		ctx := th.Context.WithSession(session)

		categories := &model.OrderedSidebarCategories{
			Categories: model.SidebarCategoriesWithChannels{
				{
					SidebarCategory: model.SidebarCategory{
						Type: model.SidebarCategoryDirectMessages,
					},
					Channels: []string{th.BasicChannel.Id},
				},
			},
		}

		th.App.FilterSidebarCategories(ctx, categories)
		require.Len(t, categories.Categories, 1)
	})

	t.Run("test user with empty roles should not filter", func(t *testing.T) {
		session := &model.Session{
			UserId: th.BasicUser.Id,
			Roles:  "",
		}
		ctx := th.Context.WithSession(session)

		categories := &model.OrderedSidebarCategories{
			Categories: model.SidebarCategoriesWithChannels{
				{
					SidebarCategory: model.SidebarCategory{
						Type: model.SidebarCategoryDirectMessages,
					},
					Channels: []string{}, // empty DM category
				},
			},
		}

		th.App.FilterSidebarCategories(ctx, categories)
		require.Len(t, categories.Categories, 1)
	})
}
