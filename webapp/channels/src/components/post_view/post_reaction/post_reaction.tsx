// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import classNames from 'classnames';
import React, {useCallback} from 'react';
import {useIntl} from 'react-intl';
import {useSelector} from 'react-redux';

import type {Emoji} from '@mattermost/types/emojis';

import {getEmojiName} from 'mattermost-redux/utils/emoji_utils';
import {canAddReactions} from 'mattermost-redux/selectors/entities/reactions';

import useEmojiPicker from 'components/emoji_picker/use_emoji_picker';
import EmojiIcon from 'components/widgets/icons/emoji_icon';
import WithTooltip from 'components/with_tooltip';
import Gate from 'components/permissions_gates/gate';

import {Locations} from 'utils/constants';

import type {GlobalState} from 'types/store';

export type Props = {
    channelId?: string;
    postId: string;
    teamId: string;
    location?: keyof typeof Locations;
    setShowEmojiPicker: (showEmojiPicker: boolean) => void;
    showEmojiPicker: boolean;
    actions: {
        toggleReaction: (postId: string, emojiName: string) => void;
    };
}

export default function PostReaction({
    channelId,
    location = Locations.CENTER,
    postId,
    teamId,
    showEmojiPicker,
    setShowEmojiPicker,
    actions: {
        toggleReaction,
    },
}: Props) {
    const intl = useIntl();

    const canAdd = useSelector((state: GlobalState) => {
        if (!channelId) {
            return false;
        }
        return canAddReactions(state, channelId);
    });

    const handleEmojiClick = useCallback((emoji: Emoji) => {
        const emojiName = getEmojiName(emoji);
        toggleReaction(postId, emojiName);

        setShowEmojiPicker(false);
    }, [postId, setShowEmojiPicker, toggleReaction]);

    const {
        emojiPicker,
        getReferenceProps,
        setReference,
    } = useEmojiPicker({
        showEmojiPicker,
        setShowEmojiPicker,

        onEmojiClick: handleEmojiClick,
    });

    const ariaLabel = intl.formatMessage({id: 'post_info.tooltip.add_reactions', defaultMessage: 'Add Reaction'});

    return (
        <Gate hasPermission={canAdd}>
            <WithTooltip title={ariaLabel}>
                <button
                    ref={setReference}
                    data-testid='post-reaction-emoji-icon'
                    id={`${location}_reaction_${postId}`}
                    aria-label={ariaLabel}
                    className={classNames('post-menu__item', 'post-menu__item--reactions', {
                        'post-menu__item--active': showEmojiPicker,
                    })}
                    {...getReferenceProps()}
                >
                    <EmojiIcon className='icon icon--small'/>
                </button>
            </WithTooltip>
            {emojiPicker}
        </Gate>
    );
}
