// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetChannelAccessible checks if the current user can access the given channel.
func (c *Client4) GetChannelAccessible(ctx context.Context, channelID string) (bool, *Response, error) {
	r, err := c.DoAPIGet(ctx, fmt.Sprintf("/engage_chat/channels/%s/accessible", channelID), "")
	if err != nil {
		return false, BuildResponse(r), err
	}
	defer closeBody(r)

	var result map[string]bool
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		return false, nil, NewAppError("GetChannelAccessible", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return result["is_accessible"], BuildResponse(r), nil
}

// EnableCustomRoles enables and returns a list of custom roles specified by name.
func (c *Client4) EnableCustomRoles(ctx context.Context, roleNames []string) ([]*Role, *Response, error) {
	jsonData, err := json.Marshal(roleNames)
	if err != nil {
		return nil, nil, NewAppError("EnableCustomRoles", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	r, err := c.DoAPIPost(ctx, "/engage_chat/roles", string(jsonData))
	if err != nil {
		return nil, BuildResponse(r), err
	}
	defer closeBody(r)

	var roles []*Role
	if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
		return nil, nil, NewAppError("EnableCustomRoles", "api.unmarshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	return roles, BuildResponse(r), nil
}
