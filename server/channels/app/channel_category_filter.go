// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

func (a *App) FilterSidebarCategories(rctx request.CTX, categories *model.OrderedSidebarCategories) {
	session := rctx.Session()

	if session.Roles == "" {
		return
	}

	canCreateDirectChannel := a.SessionHasPermissionTo(*session, model.PermissionCreateDirectChannel)
	canCreateGroupChannel := a.SessionHasPermissionTo(*session, model.PermissionCreateGroupChannel)

	if canCreateDirectChannel || canCreateGroupChannel {
		return
	}

	filteredCategories := make(model.SidebarCategoriesWithChannels, 0, len(categories.Categories))
	filteredOrder := make(model.SidebarCategoryOrder, 0, len(categories.Order))

	for _, category := range categories.Categories {
		// When the user lacks permissions to create DMGM and DM category is empty, remove DM category
		if category.Type == model.SidebarCategoryDirectMessages && len(category.Channels) == 0 {
			continue
		}
		filteredCategories = append(filteredCategories, category)
		filteredOrder = append(filteredOrder, category.Id)
	}

	categories.Categories = filteredCategories
	categories.Order = filteredOrder
}
