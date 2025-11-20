// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

/**
 * DM/GM Permission Utilities
 * 
 * This module provides permission checking utilities for Direct Messages (DM)
 * and Group Messages (GM). These functions can be overridden or extended in
 * enterprise builds to add additional permission constraints.
 */

import type {Channel} from '@mattermost/types/channels';
import type {GlobalState} from '@mattermost/types/store';

import {General, Permissions} from 'mattermost-redux/constants';
import {haveISystemPermission} from 'mattermost-redux/selectors/entities/roles';

/**
 * Check if the current user can interact with a DM/GM channel
 * (post, edit, delete messages, add reactions, etc.)
 * 
 * For non-DM/GM channels, this always returns true.
 * For DM channels, checks CREATE_DIRECT_CHANNEL permission.
 * For GM channels, checks CREATE_GROUP_CHANNEL permission.
 * 
 * @param state - Redux global state
 * @param channel - Channel to check (optional)
 * @returns true if user can interact with the channel
 */
export function canInteractWithDMGMChannel(state: GlobalState, channel?: Channel): boolean {
    if (!channel) {
        return true;
    }

    const isDM = channel.type === General.DM_CHANNEL;
    const isGM = channel.type === General.GM_CHANNEL;

    if (!isDM && !isGM) {
        return true;
    }

    if (!state.entities?.roles?.roles) {
        return false;
    }

    if (isDM) {
        return haveISystemPermission(state, {permission: Permissions.CREATE_DIRECT_CHANNEL});
    }

    if (isGM) {
        return haveISystemPermission(state, {permission: Permissions.CREATE_GROUP_CHANNEL});
    }

    return true;
}

/**
 * Check if the current user can post in a DM/GM channel
 * 
 * @param state - Redux global state
 * @param channel - Channel to check (optional)
 * @returns true if user can post in the channel
 */
export function canPostInDMGMChannel(state: GlobalState, channel?: Channel): boolean {
    return canInteractWithDMGMChannel(state, channel);
}

/**
 * Check if the current user can create any DM or GM channel
 * 
 * Returns true if user has either CREATE_DIRECT_CHANNEL or CREATE_GROUP_CHANNEL permission.
 * This is used for UI visibility (e.g., showing "Direct Messages" category in sidebar).
 * 
 * @param state - Redux global state
 * @returns true if user can create DM or GM channels
 */
export function canCreateDMGMChannel(state: GlobalState): boolean {
    if (!state.entities?.roles?.roles) {
        return false;
    }

    const canCreateDM = haveISystemPermission(state, {permission: Permissions.CREATE_DIRECT_CHANNEL});
    const canCreateGM = haveISystemPermission(state, {permission: Permissions.CREATE_GROUP_CHANNEL});
    return canCreateDM || canCreateGM;
}

/**
 * Determine if a DM/GM channel should be hidden in the sidebar
 * 
 * Channels are hidden if:
 * 1. They are not the current channel AND
 * 2. User lacks the specific permission for that channel type
 * 
 * @param hasPermissions - Whether user has DM/GM create permissions
 * @param channel - Channel to check
 * @param currentChannelId - ID of the currently active channel
 * @returns true if the channel should be hidden
 */
export function shouldHideDMGMChannel(
    hasPermissions: boolean,
    channel: Channel,
    currentChannelId: string,
): boolean {
    // Always show the current channel
    if (channel.id === currentChannelId) {
        return false;
    }

    // Hide DM/GM channels if user lacks permissions
    if (!hasPermissions && (channel.type === General.DM_CHANNEL || channel.type === General.GM_CHANNEL)) {
        return true;
    }

    return false;
}
