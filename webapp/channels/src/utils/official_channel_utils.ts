// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Channel} from '@mattermost/types/channels';

/**
 * Check if a channel is an official tunag channel based on its name pattern.
 * Official channels follow the pattern: tunag-{company_id}-{subdomain}-admin
 * Example: tunag-00002-stmn-admin
 * Pattern: tunag-digits-alphanumeric-admin
 *
 * @param {Channel | string} channel - Channel object or channel name string
 * @returns {boolean} - true if channel is an official tunag channel, false otherwise
 */
export function isOfficialTunagChannel(channel: Channel | string): boolean {
    const channelName = typeof channel === 'string' ? channel : channel.name;

    if (!channelName) {
        return false;
    }

    // Pattern: tunag-{digits}-{alphanumeric}-admin
    // Example: tunag-00002-stmn-admin
    const officialChannelPattern = /^tunag-\d+-[a-zA-Z0-9]+-admin$/;

    return officialChannelPattern.test(channelName);
}
