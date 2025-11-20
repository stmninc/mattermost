// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {shallow} from 'enzyme';
import React from 'react';
import * as redux from 'react-redux';

import {TestHelper} from 'utils/test_helper';

import SidebarCategorySortingMenu from './sidebar_category_sorting_menu';

const initialState = {
    entities: {
        users: {
            currentUserId: 'user_id',
            profiles: {
                user_id: {
                    id: 'user_id',
                    roles: 'system_user',
                },
            },
        },
        preferences: {
            myPreferences: {
                'sidebar_settings--limit_visible_dms_gms': {
                    value: '10',
                },
            },
        },
        roles: {
            roles: {
                system_user: {
                    permissions: ['create_direct_channel', 'create_group_channel'],
                },
            },
        },
    },
};

jest.spyOn(redux, 'useSelector').mockImplementation((cb) => cb(initialState));
jest.spyOn(redux, 'useDispatch').mockReturnValue((t: unknown) => t);

describe('components/sidebar/sidebar_category/sidebar_category_sorting_menu', () => {
    const baseProps = {
        category: TestHelper.getCategoryMock(),
        handleOpenDirectMessagesModal: jest.fn(),
    };

    test('should match snapshot', () => {
        const wrapper = shallow(
            <SidebarCategorySortingMenu {...baseProps}/>,
        );

        expect(wrapper).toMatchSnapshot();
    });
});
