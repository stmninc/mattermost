// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {GlobalState} from '@mattermost/types/store';

import {General, Permissions} from 'mattermost-redux/constants';

import {getChannel} from './channels';
import {haveIChannelPermission, haveISystemPermission} from './roles';

export function canAddReactions(state: GlobalState, channelId: string) {
    const channel = getChannel(state, channelId);

    if (!channel || channel.delete_at > 0) {
        return false;
    }

    // DM/GMチャンネルの権限チェック
    const isDM = channel.type === General.DM_CHANNEL;
    const isGM = channel.type === General.GM_CHANNEL;

    if (isDM) {
        const canCreateDM = haveISystemPermission(state, {permission: Permissions.CREATE_DIRECT_CHANNEL});
        if (!canCreateDM) {
            return false;
        }
    }

    if (isGM) {
        const canCreateGM = haveISystemPermission(state, {permission: Permissions.CREATE_GROUP_CHANNEL});
        if (!canCreateGM) {
            return false;
        }
    }

    return haveIChannelPermission(state, channel.team_id, channelId, Permissions.ADD_REACTION);
}

export function canRemoveReactions(state: GlobalState, channelId: string) {
    const channel = getChannel(state, channelId);

    if (!channel || channel.delete_at > 0) {
        return false;
    }

    // DM/GMチャンネルの権限チェック
    const isDM = channel.type === General.DM_CHANNEL;
    const isGM = channel.type === General.GM_CHANNEL;

    if (isDM) {
        const canCreateDM = haveISystemPermission(state, {permission: Permissions.CREATE_DIRECT_CHANNEL});
        if (!canCreateDM) {
            return false;
        }
    }

    if (isGM) {
        const canCreateGM = haveISystemPermission(state, {permission: Permissions.CREATE_GROUP_CHANNEL});
        if (!canCreateGM) {
            return false;
        }
    }

    return haveIChannelPermission(state, channel.team_id, channelId, Permissions.REMOVE_REACTION);
}
