// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Channel} from '@mattermost/types/channels';
import type {UserProfile} from '@mattermost/types/users';

import * as OfficialChannelUtils from 'utils/official_channel_utils';

// Mock Redux store
jest.mock('stores/redux_store', () => ({
    getState: jest.fn(),
}));

jest.mock('mattermost-redux/selectors/entities/users', () => ({
    getCurrentUser: jest.fn(),
    getUser: jest.fn(),
}));

import store from 'stores/redux_store';
import {getUser} from 'mattermost-redux/selectors/entities/users';

const mockStore = store as jest.Mocked<typeof store>;
const mockGetUser = getUser as jest.MockedFunction<typeof getUser>;

describe('Official Channel Utils', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockStore.getState.mockReturnValue({} as any);
    });

    describe('isOfficialTunagChannel', () => {
        test('returns false for string inputs', () => {
            // String inputs cannot be validated for creator, so should return false
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stmn-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('any-channel-name')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('')).toBe(false);
        });

        test('returns false for null or undefined inputs', () => {
            expect(OfficialChannelUtils.isOfficialTunagChannel(null)).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel(undefined)).toBe(false);
        });

        test('returns false for channels without creator_id', () => {
            const channelWithoutCreator: Partial<Channel> = {
                name: 'test-channel',
                id: 'channel_id',
                display_name: 'Test Channel',
            };

            expect(OfficialChannelUtils.isOfficialTunagChannel(channelWithoutCreator as Channel)).toBe(false);
        });

        test('returns false when creator user is not found', () => {
            const channel: Partial<Channel> = {
                name: 'test-channel',
                id: 'channel_id',
                display_name: 'Test Channel',
                creator_id: 'user_id_1',
            };

            mockGetUser.mockReturnValue(null as any);

            expect(OfficialChannelUtils.isOfficialTunagChannel(channel as Channel)).toBe(false);
        });

        test('returns true when creator username matches official pattern', () => {
            const channel: Partial<Channel> = {
                name: 'test-channel',
                id: 'channel_id',
                display_name: 'Test Channel',
                creator_id: 'user_id_1',
            };

            // Test various valid integration admin usernames
            const validUsernames = [
                'tunag-00002-stmn-admin',
                'tunag-12345-abc-admin',
                'tunag-0-z-admin',
                'tunag-999-company123-admin',
                'tunag-1-a-admin',
                'tunag-123456789-subdomain123-admin'
            ];

            validUsernames.forEach((username) => {
                const integrationAdminUser: Partial<UserProfile> = {
                    id: 'user_id_1',
                    username: username,
                };

                mockGetUser.mockReturnValue(integrationAdminUser as UserProfile);
                expect(OfficialChannelUtils.isOfficialTunagChannel(channel as Channel)).toBe(true);
            });
        });

        test('returns false when creator username does not match official pattern', () => {
            const channel: Partial<Channel> = {
                name: 'test-channel',
                id: 'channel_id',
                display_name: 'Test Channel',
                creator_id: 'user_id_1',
            };

            // Test invalid usernames
            const invalidUsernames = [
                'regular_user',
                'tunag-admin',
                'tunag-00002-admin',
                'tunag-stmn-admin',
                'tunagg-00002-stmn-admin',
                'tuna-00002-stmn-admin',
                'tunag-00002-stmn-admins',
                'tunag-00002-stmn-user',
                'tunag-abc-stmn-admin',
                ' tunag-00002-stmn-admin',
                'tunag-00002-stmn-admin ',
                'tunag-00002-stmn-admin-extra',
                'TUNAG-00002-stmn-admin',
                'tunag-00002-stmn-ADMIN',
                'tunag-00002-stm-n-admin',
                'tunag-00002-stm_n-admin',
                'tunag-00002-stm.n-admin'
            ];

            invalidUsernames.forEach((username) => {
                const regularUser: Partial<UserProfile> = {
                    id: 'user_id_1',
                    username: username,
                };

                mockGetUser.mockReturnValue(regularUser as UserProfile);
                expect(OfficialChannelUtils.isOfficialTunagChannel(channel as Channel)).toBe(false);
            });
        });

        test('returns false when creator has no username', () => {
            const channel: Partial<Channel> = {
                name: 'test-channel',
                id: 'channel_id',
                display_name: 'Test Channel',
                creator_id: 'user_id_1',
            };

            const userWithoutUsername: Partial<UserProfile> = {
                id: 'user_id_1',
            };

            const userWithEmptyUsername: Partial<UserProfile> = {
                id: 'user_id_1',
                username: '',
            };

            mockGetUser.mockReturnValue(userWithoutUsername as UserProfile);
            expect(OfficialChannelUtils.isOfficialTunagChannel(channel as Channel)).toBe(false);

            mockGetUser.mockReturnValue(userWithEmptyUsername as UserProfile);
            expect(OfficialChannelUtils.isOfficialTunagChannel(channel as Channel)).toBe(false);
        });
    });
});
