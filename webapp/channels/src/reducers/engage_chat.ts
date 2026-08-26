// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {AnyAction} from 'redux';

import reducerRegistry from 'mattermost-redux/store/reducer_registry';

export const RECEIVED_CHANNEL_ACCESSIBLE = 'RECEIVED_CHANNEL_ACCESSIBLE';

export type EngageChatState = {
    channelAccessible: Record<string, boolean>;
};

const initialState: EngageChatState = {
    channelAccessible: {},
};

function engageChatReducer(state = initialState, action: AnyAction): EngageChatState {
    switch (action.type) {
    case RECEIVED_CHANNEL_ACCESSIBLE:
        return {
            ...state,
            channelAccessible: {
                ...state.channelAccessible,
                [action.channelId]: action.accessible,
            },
        };
    default:
        return state;
    }
}

// Dynamically inject the reducer into the store without modifying existing files
reducerRegistry.register('engageChat', engageChatReducer);

export default engageChatReducer;
