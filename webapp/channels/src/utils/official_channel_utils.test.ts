// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Channel} from '@mattermost/types/channels';

import * as OfficialChannelUtils from 'utils/official_channel_utils';

describe('Official Channel Utils', () => {
    describe('isOfficialTunagChannel', () => {
        test('returns true for valid official tunag channel names', () => {
            // Test with channel name string
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stmn-admin')).toBe(true);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-12345-abc-admin')).toBe(true);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-0-z-admin')).toBe(true);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-999-company123-admin')).toBe(true);
        });

        test('returns true for valid official tunag channel objects', () => {
            // Test with channel object
            const validChannel: Partial<Channel> = {
                name: 'tunag-00002-stmn-admin',
                id: 'channel_id',
                display_name: 'Official Channel',
            };

            expect(OfficialChannelUtils.isOfficialTunagChannel(validChannel as Channel)).toBe(true);
        });

        test('returns false for invalid official tunag channel names', () => {
            // Missing components
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-stmn-admin')).toBe(false);

            // Wrong prefix
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunagg-00002-stmn-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tuna-00002-stmn-admin')).toBe(false);

            // Wrong suffix
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stmn-admins')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stmn-user')).toBe(false);

            // Non-numeric company id
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-abc-stmn-admin')).toBe(false);

            // Extra characters or spacing
            expect(OfficialChannelUtils.isOfficialTunagChannel(' tunag-00002-stmn-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stmn-admin ')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stmn-admin-extra')).toBe(false);

            // Empty or undefined values
            expect(OfficialChannelUtils.isOfficialTunagChannel('')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('regular-channel')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('random-channel-name')).toBe(false);
        });

        test('returns false for channels with empty or undefined names', () => {
            const channelWithoutName: Partial<Channel> = {
                id: 'channel_id',
                display_name: 'Some Channel',
            };

            const channelWithEmptyName: Partial<Channel> = {
                name: '',
                id: 'channel_id',
                display_name: 'Some Channel',
            };

            expect(OfficialChannelUtils.isOfficialTunagChannel(channelWithoutName as Channel)).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel(channelWithEmptyName as Channel)).toBe(false);
        });

        test('handles edge cases correctly', () => {
            // Test various valid patterns
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-1-a-admin')).toBe(true);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-123456789-subdomain123-admin')).toBe(true);

            // Test invalid patterns with special characters
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stm-n-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stm_n-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stm.n-admin')).toBe(false);

            // Test case sensitivity
            expect(OfficialChannelUtils.isOfficialTunagChannel('TUNAG-00002-stmn-admin')).toBe(false);
            expect(OfficialChannelUtils.isOfficialTunagChannel('tunag-00002-stmn-ADMIN')).toBe(false);
        });
    });
});
