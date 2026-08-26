// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Client4} from 'mattermost-redux/client';

import {RECEIVED_CHANNEL_ACCESSIBLE} from 'reducers/engage_chat';

import type {ActionFuncAsync} from 'types/store';

// Unlike typical Mattermost Redux actions (which are dispatched from component lifecycle
// methods and therefore run at most once per mount), fetchChannelAccessible is dispatched
// from isAvailableUnofficialChannel — a plain function called synchronously during render.
// This means it can be called multiple times before the API response arrives and populates
// the Redux cache. The pendingRequests Set prevents redundant in-flight requests for the
// same channelId during that window.
const pendingRequests = new Set<string>();

export function fetchChannelAccessible(channelId: string): ActionFuncAsync<void> {
    return async (dispatch) => {
        if (!channelId) {
            return {data: undefined};
        }
        if (pendingRequests.has(channelId)) {
            return {data: undefined};
        }
        pendingRequests.add(channelId);
        try {
            const response = await fetch(
                `${Client4.getBaseRoute()}/engage_chat/channels/${channelId}/accessible`,
                Client4.getOptions({method: 'GET'}),
            );
            if (!response.ok) {
                dispatch({
                    type: RECEIVED_CHANNEL_ACCESSIBLE,
                    channelId,
                    accessible: false,
                });
                return {data: undefined};
            }
            const data: {is_accessible: boolean} = await response.json();
            dispatch({
                type: RECEIVED_CHANNEL_ACCESSIBLE,
                channelId,
                accessible: data.is_accessible,
            });
        } catch {
            // On API failure, cache false so the channel remains inaccessible
            // until the user reloads the page (no automatic retry).
            dispatch({
                type: RECEIVED_CHANNEL_ACCESSIBLE,
                channelId,
                accessible: false,
            });
        } finally {
            pendingRequests.delete(channelId);
        }
        return {data: undefined};
    };
}
