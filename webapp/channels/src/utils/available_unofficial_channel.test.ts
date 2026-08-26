// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Permissions} from 'mattermost-redux/constants';
import {getChannel} from 'mattermost-redux/selectors/entities/channels';
import {haveIChannelPermission, haveICurrentTeamPermission} from 'mattermost-redux/selectors/entities/roles';

import {fetchChannelAccessible} from 'actions/engage_chat';
import store from 'stores/redux_store';

import {isAvailableUnofficialChannel, isAvailableDMOrGMChannel, isAvailableDMChannel} from './available_unofficial_channel';
import {isOfficialTunagChannel} from './official_channel_utils';

jest.mock('stores/redux_store', () => ({
    getState: jest.fn(),
    dispatch: jest.fn(),
}));

jest.mock('mattermost-redux/selectors/entities/channels', () => ({
    ...jest.requireActual('mattermost-redux/selectors/entities/channels'),
    getChannel: jest.fn(),
}));

jest.mock('mattermost-redux/selectors/entities/roles', () => ({
    ...jest.requireActual('mattermost-redux/selectors/entities/roles'),
    haveIChannelPermission: jest.fn(),
    haveICurrentTeamPermission: jest.fn(),
}));

jest.mock('actions/engage_chat', () => ({
    fetchChannelAccessible: jest.fn().mockReturnValue(() => Promise.resolve()),
}));

jest.mock('./official_channel_utils', () => ({
    ...jest.requireActual('./official_channel_utils'),
    isOfficialTunagChannel: jest.fn(),
}));

// Mock reducer_registry because it runs reducerRegistry.register() at module load time
jest.mock('mattermost-redux/store/reducer_registry', () => ({
    default: {register: jest.fn()},
}));

let mockChannel: any;
let mockPermissionResult: boolean;

describe('available_unofficial_channel utils', () => {
    const mockGetState = store.getState as jest.Mock;
    const mockDispatch = store.dispatch as jest.Mock;
    const mockGetChannel = getChannel as jest.Mock;
    const mockHaveIChannelPermission = haveIChannelPermission as jest.Mock;
    const mockHaveICurrentTeamPermission = haveICurrentTeamPermission as jest.Mock;
    const mockFetchChannelAccessible = fetchChannelAccessible as jest.Mock;
    const mockIsOfficialTunagChannel = isOfficialTunagChannel as jest.Mock;

    beforeEach(() => {
        jest.resetAllMocks();

        mockChannel = {id: 'channel_id', team_id: 'team_id', type: 'O'};
        mockPermissionResult = true; // Default to having permission

        // Setup mock behavior (no engageChat cache by default)
        mockGetState.mockReturnValue({});
        mockGetChannel.mockImplementation(() => mockChannel);
        mockFetchChannelAccessible.mockReturnValue(() => Promise.resolve());
        mockIsOfficialTunagChannel.mockReturnValue(false);

        // Simplify permission check to return mockPermissionResult
        // (Set mockPermissionResult to false in individual test cases to simulate denial)
        mockHaveIChannelPermission.mockImplementation(() => mockPermissionResult);
        mockHaveICurrentTeamPermission.mockImplementation(() => mockPermissionResult);
    });

    describe('isAvailableUnofficialChannel', () => {
        test('returns false immediately for empty channelId without dispatching', () => {
            expect(isAvailableUnofficialChannel('')).toBe(false);
            expect(mockDispatch).not.toHaveBeenCalled();
        });

        describe('when API cache (engageChat) is populated', () => {
            test('returns true from cache without performing a permission check', () => {
                mockGetState.mockReturnValue({
                    engageChat: {channelAccessible: {channel_id: true}},
                });

                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockHaveIChannelPermission).not.toHaveBeenCalled();
                expect(mockDispatch).not.toHaveBeenCalled();
            });

            test('returns false from cache without performing a permission check', () => {
                mockGetState.mockReturnValue({
                    engageChat: {channelAccessible: {channel_id: false}},
                });

                expect(isAvailableUnofficialChannel('channel_id')).toBe(false);
                expect(mockHaveIChannelPermission).not.toHaveBeenCalled();
                expect(mockDispatch).not.toHaveBeenCalled();
            });
        });

        describe('when API cache is not populated', () => {
            test('returns true for official Tunag channel without permission check or API call', () => {
                mockChannel.type = 'D';
                mockPermissionResult = false;
                mockIsOfficialTunagChannel.mockReturnValue(true);

                expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                expect(mockHaveIChannelPermission).not.toHaveBeenCalled();
                expect(mockDispatch).not.toHaveBeenCalled();
            });

            describe('when local permission check passes (fast path)', () => {
                test('returns true for open channel (Type: O) without dispatching an API fetch', () => {
                    mockChannel.type = 'O';
                    expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                    expect(mockDispatch).not.toHaveBeenCalled();
                });

                test('returns true for private channel (Type: P) without dispatching an API fetch', () => {
                    mockChannel.type = 'P';
                    expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                    expect(mockDispatch).not.toHaveBeenCalled();
                });

                test('returns true for unknown channel type without dispatching an API fetch', () => {
                    mockChannel.type = 'Unknown';
                    expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                    expect(mockDispatch).not.toHaveBeenCalled();
                });

                test('returns true for direct message (Type: D) when CREATE_DIRECT_CHANNEL is granted', () => {
                    mockChannel.type = 'D';
                    mockPermissionResult = true;

                    expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                    expect(mockHaveIChannelPermission).toHaveBeenCalledWith(
                        expect.anything(),
                        'team_id',
                        'channel_id',
                        Permissions.CREATE_DIRECT_CHANNEL,
                    );
                    expect(mockDispatch).not.toHaveBeenCalled();
                });

                test('returns true for group message (Type: G) when CREATE_GROUP_CHANNEL is granted', () => {
                    mockChannel.type = 'G';
                    mockPermissionResult = true;

                    expect(isAvailableUnofficialChannel('channel_id')).toBe(true);
                    expect(mockHaveIChannelPermission).toHaveBeenCalledWith(
                        expect.anything(),
                        'team_id',
                        'channel_id',
                        Permissions.CREATE_GROUP_CHANNEL,
                    );
                    expect(mockDispatch).not.toHaveBeenCalled();
                });
            });

            describe('when local permission check fails — falls back to API', () => {
                test('dispatches an API fetch and returns false while waiting', () => {
                    mockChannel.type = 'D';
                    mockPermissionResult = false;

                    expect(isAvailableUnofficialChannel('channel_id')).toBe(false);
                    expect(mockDispatch).toHaveBeenCalledTimes(1);
                });

                test('dispatches an API fetch and returns false when channel is not found locally', () => {
                    mockGetChannel.mockReturnValue(null);

                    expect(isAvailableUnofficialChannel('missing_channel')).toBe(false);
                    expect(mockDispatch).toHaveBeenCalledTimes(1);
                });
            });
        });
    });

    describe('isAvailableDMOrGMChannel', () => {
        test('should return true when both DM and GM creation permissions are granted', () => {
            mockHaveICurrentTeamPermission.mockReturnValue(true);
            expect(isAvailableDMOrGMChannel()).toBe(true);

            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledTimes(1);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_DIRECT_CHANNEL);
            expect(mockHaveICurrentTeamPermission).not.toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_GROUP_CHANNEL);
        });

        test('should return true when DM creation permission is denied but GM is granted', () => {
            // Change return value for each call: 1st(DM) is False, 2nd(GM) is True
            mockHaveICurrentTeamPermission.
                mockReturnValueOnce(false).
                mockReturnValueOnce(true);

            expect(isAvailableDMOrGMChannel()).toBe(true);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledTimes(2);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_DIRECT_CHANNEL);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_GROUP_CHANNEL);
        });

        test('should return true when GM creation permission is denied but DM is granted', () => {
            // Short-circuits: 1st(DM) is True, so GM check is never performed.
            mockHaveICurrentTeamPermission.mockReturnValueOnce(true);

            expect(isAvailableDMOrGMChannel()).toBe(true);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledTimes(1);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_DIRECT_CHANNEL);
            expect(mockHaveICurrentTeamPermission).not.toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_GROUP_CHANNEL);
        });

        test('should return false when both DM and GM creation permissions are denied', () => {
            mockHaveICurrentTeamPermission.mockReturnValue(false);
            expect(isAvailableDMOrGMChannel()).toBe(false);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledTimes(2);
        });
    });

    describe('isAvailableDMChannel', () => {
        test('should return true when DM creation permission is granted', () => {
            mockPermissionResult = true;
            expect(isAvailableDMChannel()).toBe(true);
            expect(mockHaveICurrentTeamPermission).toHaveBeenCalledWith(expect.anything(), Permissions.CREATE_DIRECT_CHANNEL);
        });

        test('should return false when DM creation permission is denied', () => {
            mockPermissionResult = false;
            expect(isAvailableDMChannel()).toBe(false);
        });
    });
});
