// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {useSelector} from 'react-redux';

import type {UserProfile} from '@mattermost/types/users';

import {createSelector} from 'mattermost-redux/selectors/create_selector';
import {getMyGroupMentionKeysForChannel, getMyGroupMentionKeys} from 'mattermost-redux/selectors/entities/groups';
import {getTeammateNameDisplaySetting} from 'mattermost-redux/selectors/entities/preferences';
import {
    getCurrentUserMentionKeys,
    getUsersByUsername,
} from 'mattermost-redux/selectors/entities/users';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

import type {MentionKey} from 'utils/text_formatting';

import type {GlobalState} from 'types/store';

// Common logic for generating fullname mention keys
const generateFullnameMentionKeys = (
    mentionKeysWithoutGroups: MentionKey[],
    groupMentionKeys: MentionKey[],
    users: Record<string, UserProfile>,
    nameDisplaySetting: string,
): MentionKey[] => {
    const baseMentionKeys = mentionKeysWithoutGroups.concat(groupMentionKeys);

    // Generate fullname mention keys for all users (including current user)
    const fullnameMentionKeys: MentionKey[] = [];

    for (const user of Object.values(users)) {
        const displayName = displayUsername(user, nameDisplaySetting, false);
        if (displayName !== user.username) {
            fullnameMentionKeys.push({
                key: `@${displayName}`,
                caseSensitive: false,
            });
        }
    }

    // Return combined mention keys
    return baseMentionKeys.concat(fullnameMentionKeys);
};

// Memoized selector for fullname mention keys without channel-specific groups
const getFullnameMentionKeysBase = createSelector(
    'getFullnameMentionKeysBase',
    getCurrentUserMentionKeys,
    (state: GlobalState) => getMyGroupMentionKeys(state, false),
    getUsersByUsername,
    getTeammateNameDisplaySetting,
    generateFullnameMentionKeys,
);

// Memoized selector factory for channel-specific mention keys
const makeGetFullnameMentionKeysForChannel = () => createSelector(
    'getFullnameMentionKeysForChannel',
    getCurrentUserMentionKeys,
    (state: GlobalState, teamId: string, channelId: string) => getMyGroupMentionKeysForChannel(state, teamId, channelId),
    getUsersByUsername,
    getTeammateNameDisplaySetting,
    generateFullnameMentionKeys,
);

/**
 * Custom hook to generate mention keys including fullname format
 * @param channelId - Optional channelId to get group mentions specific to a channel
 * @param teamId - Optional teamId required when channelId is provided
 */
export const useFullnameMentionKeys = (channelId?: string, teamId?: string): MentionKey[] => {
    return useSelector((state: GlobalState) => {
        if (channelId && teamId) {
            // Use memoized selector for channel-specific case
            const selector = makeGetFullnameMentionKeysForChannel();
            return selector(state, teamId, channelId);
        }

        // Use memoized selector for base case
        return getFullnameMentionKeysBase(state);
    });
};
