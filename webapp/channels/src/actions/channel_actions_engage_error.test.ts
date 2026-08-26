// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import * as ChannelActions from 'mattermost-redux/actions/channels';
import {LogErrorBarMode} from 'mattermost-redux/actions/errors';

import {
    openDirectChannelToUserId,
    openGroupChannelToUserIds,
} from 'actions/channel_actions';

import mockStore from 'tests/test_store';
import {TestHelper} from 'utils/test_helper';

const initialState = {
    entities: {
        channels: {
            currentChannelId: 'current_channel_id',
            myMembers: {
                current_channel_id: {
                    channel_id: 'current_channel_id',
                    user_id: 'current_user_id',
                    roles: 'channel_role',
                    mention_count: 1,
                    msg_count: 9,
                },
            },
            channels: {
                current_channel_id: TestHelper.getChannelMock({
                    id: 'current_channel_id',
                    name: 'default-name',
                    display_name: 'Default',
                    delete_at: 0,
                    type: 'O',
                    team_id: 'team_id',
                }),
            },
            channelsInTeam: {},
            messageCounts: {},
        },
        teams: {
            currentTeamId: 'team-id',
            teams: {
                'team-id': {
                    id: 'team_id',
                    name: 'team-1',
                    display_name: 'Team 1',
                },
            },
            myMembers: {
                'team-id': {roles: 'team_role'},
            },
        },
        users: {
            currentUserId: 'current_user_id',
            profiles: {
                current_user_id: {roles: 'system_role', id: 'current_user_id'},
            },
        },
        preferences: {
            myPreferences: {},
        },
        roles: {
            roles: {},
        },
        general: {
            license: {IsLicensed: 'false'},
            config: {},
        },
    },
};

jest.mock('mattermost-redux/actions/channels', () => ({
    createDirectChannel: jest.fn().mockImplementation((...args: any) => (dispatch: any) => {
        const action = {type: 'MOCK_CREATE_DIRECT_CHANNEL', args, data: {id: 'new_channel_id'}};
        dispatch(action);
        return action;
    }),
    createGroupChannel: jest.fn().mockImplementation((...args: any) => (dispatch: any) => {
        const action = {type: 'MOCK_CREATE_GROUP_CHANNEL', args, data: {id: 'new_group_id'}};
        dispatch(action);
        return action;
    }),
}));

jest.mock('mattermost-redux/actions/errors', () => ({
    logError: jest.fn().mockImplementation((error: any, displayOptions: any) => ({type: 'MOCK_LOG_ERROR', error, displayOptions})),
    LogErrorBarMode: {
        Always: 'always',
    },
}));

describe('Actions.Channel (Error cases)', () => {
    test('openDirectChannelToUserId failure should dispatch logError with Always mode', async () => {
        const testStore = await mockStore(initialState);
        const error = {message: 'error'};
        (ChannelActions.createDirectChannel as jest.Mock).mockReturnValueOnce(() => ({error}));

        const userId = 'testid';

        await testStore.dispatch(openDirectChannelToUserId(userId));
        const actions = testStore.getActions();

        expect(actions).toContainEqual({
            type: 'MOCK_LOG_ERROR',
            error,
            displayOptions: {errorBarMode: LogErrorBarMode.Always},
        });
    });

    test('openGroupChannelToUserIds failure should dispatch logError with Always mode', async () => {
        const testStore = await mockStore(initialState);
        const error = {message: 'error'};
        (ChannelActions.createGroupChannel as jest.Mock).mockReturnValueOnce(() => ({error}));

        const userIds = ['testuserid1', 'testuserid2'];

        await testStore.dispatch(openGroupChannelToUserIds(userIds));
        const actions = testStore.getActions();

        expect(actions).toContainEqual({
            type: 'MOCK_LOG_ERROR',
            error,
            displayOptions: {errorBarMode: LogErrorBarMode.Always},
        });
    });
});
