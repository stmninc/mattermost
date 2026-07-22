// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {renderWithContext, screen} from 'tests/react_testing_utils';
import {Locations} from 'utils/constants';
import {TestHelper} from 'utils/test_helper';

import PostOptions from './post_options';

jest.mock('components/post_view/post_recent_reactions', () => {
    return jest.fn(({size}) => (
        <div
            data-testid='post-recent-reactions'
            data-size={size}
        />
    ));
});

describe('PostOptions', () => {
    const channel = TestHelper.getChannelMock({team_id: 'team_id'});
    const post = TestHelper.getPostMock({channel_id: channel.id, type: ''});

    const baseProps: React.ComponentProps<typeof PostOptions> = {
        post,
        teamId: 'team_id',
        isFlagged: false,
        removePost: jest.fn(),
        enableEmojiPicker: true,
        isReadOnly: false,
        channelIsArchived: false,
        oneClickReactionsEnabled: true,
        recentEmojis: [],
        hover: true, // directly enable hover
        isMobileView: false,
        location: Locations.CENTER,
        pluginActions: [],
        isChannelAutotranslated: false,
        handleDropdownOpened: jest.fn(),
        actions: {
            emitShortcutReactToLastPostFrom: jest.fn(),
        },
    };

    describe('should show 3 reaction suggestions in thread', () => {
        const testCase = [
            {location: Locations.CENTER, expectedSize: '3'},
            {location: Locations.RHS_ROOT, expectedSize: '3'},
            {location: Locations.RHS_COMMENT, expectedSize: '3'},
            {location: Locations.SEARCH, expectedSize: null},
        ];

        testCase.forEach(({location, expectedSize}) => {
            test(`should show ${expectedSize ? expectedSize + ' reactions' : 'no reactions'} when location is ${location}`, () => {
                const props = {
                    ...baseProps,
                    location,
                };

                renderWithContext(<PostOptions {...props}/>);

                const recentReactions = screen.queryByTestId('post-recent-reactions');
                if (expectedSize) {
                    expect(recentReactions).toBeInTheDocument();
                    expect(recentReactions!.getAttribute('data-size')).toBe(expectedSize);
                } else {
                    expect(recentReactions).not.toBeInTheDocument();
                }
            });
        });
    });
});
