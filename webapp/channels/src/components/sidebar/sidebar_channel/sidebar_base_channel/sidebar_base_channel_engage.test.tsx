// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import * as reactRedux from 'react-redux';

import type {ChannelType} from '@mattermost/types/channels';

import SidebarBaseChannel from 'components/sidebar/sidebar_channel/sidebar_base_channel/sidebar_base_channel';
import SidebarChannelLink from 'components/sidebar/sidebar_channel/sidebar_channel_link';
import BuildingIcon from 'components/widgets/icons/building_icon';

import {renderWithContext} from 'tests/react_testing_utils';

jest.mock('components/tours/onboarding_tour', () => ({
    ChannelsAndDirectMessagesTour: () => null,
}));

jest.mock('components/sidebar/sidebar_channel/sidebar_channel_link', () => {
    return jest.fn(() => null);
});

describe('components/sidebar/sidebar_channel/sidebar_base_channel', () => {
    let useSelectorMock: jest.SpyInstance;

    beforeEach(() => {
        jest.clearAllMocks();
        useSelectorMock = jest.spyOn(reactRedux, 'useSelector').mockReturnValue(false);
    });

    afterEach(() => {
        jest.restoreAllMocks();
    });

    const baseProps = {
        channel: {
            id: 'channel_id',
            display_name: 'channel_display_name',
            create_at: 0,
            update_at: 0,
            delete_at: 0,
            team_id: '',
            type: 'O' as ChannelType,
            name: '',
            header: '',
            purpose: '',
            last_post_at: 0,
            last_root_post_at: 0,
            creator_id: '',
            scheme_id: '',
            group_constrained: false,
        },
        currentTeamName: 'team_name',
        actions: {
            leaveChannel: jest.fn(),
            openModal: jest.fn(),
        },
        isOfficial: false,
    };

    test('should change icon when channel is official', async () => {
        useSelectorMock.mockReturnValue(true);

        const props = {
            ...baseProps,
            isOfficial: true,
        };

        renderWithContext(<SidebarBaseChannel {...props}/>);

        expect(SidebarChannelLink).toHaveBeenCalledTimes(1);

        const passedProps = jest.mocked(SidebarChannelLink).mock.calls[0][0];

        expect(passedProps.icon).toBeDefined();
        expect(passedProps.icon?.type).toBe(BuildingIcon);
    });
});
