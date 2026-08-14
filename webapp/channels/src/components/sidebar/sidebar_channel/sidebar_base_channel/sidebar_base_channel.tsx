// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useCallback} from 'react';
import {useIntl} from 'react-intl';
import {useSelector} from 'react-redux';

import type {Channel} from '@mattermost/types/channels';

import LeaveChannelModal from 'components/leave_channel_modal';
import SidebarChannelLink from 'components/sidebar/sidebar_channel/sidebar_channel_link';
import BuildingIcon from 'components/widgets/icons/building_icon';

import Constants, {ModalIdentifiers} from 'utils/constants';
import {isOfficialTunagChannel} from 'utils/official_channel_utils';

import type {GlobalState} from 'types/store';

import SidebarBaseChannelIcon from './sidebar_base_channel_icon';

import type {PropsFromRedux} from './index';

export interface Props extends PropsFromRedux {
    channel: Channel;
    currentTeamName: string;
}

const SidebarBaseChannel = ({
    channel,
    currentTeamName,
    actions,
}: Props) => {
    const intl = useIntl();

    const handleLeavePublicChannel = useCallback((callback: () => void) => {
        actions.leaveChannel(channel.id);
        callback();
    }, [channel.id, actions.leaveChannel]);

    const handleLeavePrivateChannel = useCallback((callback: () => void) => {
        actions.openModal({modalId: ModalIdentifiers.LEAVE_PRIVATE_CHANNEL_MODAL, dialogType: LeaveChannelModal, dialogProps: {channel}});
        callback();
    }, [channel, actions.openModal]);

    let channelLeaveHandler = null;
    if (channel.type === Constants.OPEN_CHANNEL && channel.name !== Constants.DEFAULT_CHANNEL) {
        channelLeaveHandler = handleLeavePublicChannel;
    } else if (channel.type === Constants.PRIVATE_CHANNEL) {
        channelLeaveHandler = handleLeavePrivateChannel;
    }

    const isOfficial = useSelector((state: GlobalState) => {
        if (channel.type === Constants.DM_CHANNEL || channel.type === Constants.GM_CHANNEL) {
            return false;
        }
        return isOfficialTunagChannel(channel, state);
    });

    const channelIcon = isOfficial ? (
        <BuildingIcon/>
    ) : (
        <SidebarBaseChannelIcon
            channelType={channel.type}
        />
    );

    let ariaLabelPrefix;
    if (channel.type === Constants.OPEN_CHANNEL) {
        ariaLabelPrefix = intl.formatMessage({id: 'accessibility.sidebar.types.public', defaultMessage: 'public channel'});
    } else if (channel.type === Constants.PRIVATE_CHANNEL) {
        ariaLabelPrefix = intl.formatMessage({id: 'accessibility.sidebar.types.private', defaultMessage: 'private channel'});
    }

    return (
        <SidebarChannelLink
            channel={channel}
            link={`/${currentTeamName}/channels/${channel.name}`}
            label={channel.display_name}
            ariaLabelPrefix={ariaLabelPrefix}
            channelLeaveHandler={channelLeaveHandler!}
            icon={channelIcon}
            isSharedChannel={channel.shared}
        />
    );
};

export default SidebarBaseChannel;
