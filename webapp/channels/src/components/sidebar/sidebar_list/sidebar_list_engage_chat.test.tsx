// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import {CategorySorting} from '@mattermost/types/channel_categories';
import type {ChannelType} from '@mattermost/types/channels';
import type {TeamType} from '@mattermost/types/teams';

import {CategoryTypes} from 'mattermost-redux/constants/channel_categories';

import {renderWithContext, screen} from 'tests/react_testing_utils';
import {TestHelper} from 'utils/test_helper';

import {SidebarList as SidebarListComponent} from './sidebar_list';

jest.mock('components/async_load', () => ({
    makeAsyncComponent: (displayName: string) => {
        const Component = (props: {children?: React.ReactNode}) => (
            <div data-testid={displayName}>{props.children}</div>
        );
        Component.displayName = displayName;
        return Component;
    },
}));

jest.mock('components/common/scrollbars', () => {
    const React = require('react');

    return React.forwardRef(({children, onScroll}: {children?: React.ReactNode; onScroll?: () => void}, ref: any) => {
        const setRef = (node: HTMLDivElement | null) => {
            if (!node) {
                return;
            }
            if (!node.scrollTo) {
                node.scrollTo = jest.fn();
            }
            if (typeof ref === 'function') {
                ref(node);
            } else if (ref) {
                ref.current = node;
            }
        };

        return (
            <div
                data-testid='scrollbars'
                ref={setRef}
                onScroll={onScroll}
            >
                {children}
            </div>
        );
    });
});

jest.mock('components/sidebar/sidebar_category', () => () => <div data-testid='sidebar-category'/>);

describe('SidebarList - when component is not rendered', () => {
    const intl = {
        formatMessage: ({defaultMessage}: {defaultMessage: string}) => defaultMessage,
    } as any;

    const currentChannel = TestHelper.getChannelMock({
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
    });

    const baseProps = {
        currentTeam: TestHelper.getTeamMock({
            id: 'kemjcpu9bi877yegqjs18ndp4r',
            invite_id: 'ojsnudhqzbfzpk6e4n6ip1hwae',
            name: 'test',
            create_at: 123,
            update_at: 123,
            delete_at: 123,
            display_name: 'test',
            description: 'test',
            email: 'test',
            type: 'O' as TeamType,
            company_name: 'test',
            allowed_domains: 'test',
            allow_open_invite: false,
            scheme_id: 'test',
            group_constrained: false,
        }),
        currentChannelId: currentChannel.id,
        categories: [
            {
                id: 'category1',
                team_id: 'team1',
                user_id: '',
                type: CategoryTypes.CUSTOM,
                display_name: 'custom_category_1',
                sorting: CategorySorting.Alphabetical,
                channel_ids: ['channel_id', 'channel_id_2'],
                muted: false,
                collapsed: false,
            },
        ],
        unreadChannelIds: ['channel_id_2'],
        displayedChannels: [currentChannel],
        newCategoryIds: [],
        multiSelectedChannelIds: [],
        isUnreadFilterEnabled: false,
        draggingState: {},
        categoryCollapsedState: {},
        handleOpenMoreDirectChannelsModal: jest.fn(),
        onDragStart: jest.fn(),
        onDragEnd: jest.fn(),
        showUnreadsCategory: false,
        collapsedThreads: true,
        hasUnreadThreads: false,
        currentStaticPageId: '',
        staticPages: [],
        actions: {
            switchToChannelById: jest.fn(),
            switchToLhsStaticPage: jest.fn(),
            close: jest.fn(),
            moveChannelsInSidebar: jest.fn(),
            moveCategory: jest.fn(),
            removeFromCategory: jest.fn(),
            setDraggingState: jest.fn(),
            stopDragging: jest.fn(),
            clearChannelSelection: jest.fn(),
            multiSelectChannelAdd: jest.fn(),
        },
    };

    test('should not throw error on team change in componentDidUpdate when scrollbar is null', () => {
        const sidebarListRef = React.createRef<SidebarListComponent>();
        const component = renderWithContext(
            <SidebarListComponent
                {...baseProps}
                intl={intl}
                ref={sidebarListRef}
            />);

        const instance = sidebarListRef.current!;
        instance.scrollbar = {current: null};

        const newCurrentTeam = {
            ...baseProps.currentTeam,
            id: 'new_team',
        };

        expect(() => {
            component.rerender(
                <SidebarListComponent
                    {...baseProps}
                    intl={intl}
                    ref={sidebarListRef}
                    currentTeam={newCurrentTeam}
                />,
            );
        }).not.toThrow();
    });

    test('should not throw error on scroll animation update when scrollbar is null', () => {
        const sidebarListRef = React.createRef<SidebarListComponent>();
        renderWithContext(
            <SidebarListComponent
                {...baseProps}
                intl={intl}
                ref={sidebarListRef}
            />,
        );

        const instance = sidebarListRef.current!;
        instance.scrollbar = {current: null};

        const mockSpring = {
            getCurrentValue: jest.fn(() => 100),
        } as any;

        expect(() => {
            instance.handleScrollAnimationUpdate(mockSpring);
        }).not.toThrow();
    });

    test('should not throw error on scrolling to first unread channel when scrollbar is null', () => {
        const sidebarListRef = React.createRef<SidebarListComponent>();
        renderWithContext(
            <SidebarListComponent
                {...baseProps}
                intl={intl}
                ref={sidebarListRef}
            />,
        );

        const instance = sidebarListRef.current!;
        instance.scrollbar = {current: null};

        expect(() => {
            instance.scrollToFirstUnreadChannel();
        }).not.toThrow();
    });

    test('should not throw error on scrolling to position when scrollbar is null', () => {
        const sidebarListRef = React.createRef<SidebarListComponent>();
        renderWithContext(
            <SidebarListComponent
                {...baseProps}
                intl={intl}
                ref={sidebarListRef}
            />,
        );

        const instance = sidebarListRef.current!;
        instance.scrollbar = {current: null};

        expect(() => {
            instance.scrollToPosition(100);
        }).not.toThrow();
    });

    test('should not throw error on updating unread indicators when scrollbar is null', () => {
        const sidebarListRef = React.createRef<SidebarListComponent>();
        renderWithContext(
            <SidebarListComponent
                {...baseProps}
                intl={intl}
                ref={sidebarListRef}
            />,
        );

        const instance = sidebarListRef.current!;
        instance.scrollbar = {current: null};

        expect(() => {
            instance.updateUnreadIndicators();
        }).not.toThrow();
    });

    test('should render Scrollbars by default (when shouldRender is not provided)', () => {
        renderWithContext(
            <SidebarListComponent
                {...baseProps}
                intl={intl}
            />,
        );

        expect(screen.getByTestId('scrollbars')).toBeInTheDocument();
    });

    test('should not render Scrollbars when shouldRender is false', () => {
        renderWithContext(
            <SidebarListComponent
                {...baseProps}
                intl={intl}
                shouldRender={false}
            />,
        );

        expect(screen.queryByTestId('scrollbars')).not.toBeInTheDocument();
    });
});
