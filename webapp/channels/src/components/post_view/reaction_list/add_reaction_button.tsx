// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import classNames from 'classnames';
import React, {useCallback, useState} from 'react';
import {useIntl} from 'react-intl';
import {useSelector} from 'react-redux';

import type {Emoji} from '@mattermost/types/emojis';
import type {Post} from '@mattermost/types/posts';

import {canAddReactions} from 'mattermost-redux/selectors/entities/reactions';

import useEmojiPicker from 'components/emoji_picker/use_emoji_picker';
import Gate from 'components/permissions_gates/gate';
import AddReactionIcon from 'components/widgets/icons/add_reaction_icon';
import WithTooltip from 'components/with_tooltip';

import type {GlobalState} from 'types/store';

type Props = {
    post: Post;

    onEmojiClick: (emoji: Emoji) => void;
}

export default function AddReactionButton({
    post,
    onEmojiClick,
}: Props) {
    const intl = useIntl();

    const canAdd = useSelector((state: GlobalState) => {
        return canAddReactions(state, post.channel_id);
    });

    const [showEmojiPicker, setShowEmojiPicker] = useState(false);

    const handleEmojiClick = useCallback((emoji: Emoji) => {
        onEmojiClick(emoji);
        setShowEmojiPicker(false);
    }, [onEmojiClick]);

    const {
        emojiPicker,
        getReferenceProps,
        setReference,
    } = useEmojiPicker({
        showEmojiPicker,
        setShowEmojiPicker,

        onEmojiClick: handleEmojiClick,
    });

    const ariaLabel = intl.formatMessage({id: 'reaction.add.ariaLabel', defaultMessage: 'Add a reaction'});

    return (
        <span className='emoji-picker__container'>
            <Gate hasPermission={canAdd}>
                <WithTooltip title={ariaLabel}>
                    <button
                        id={`addReaction-${post.id}`}
                        ref={setReference}
                        aria-label={ariaLabel}
                        className={classNames('Reaction Reaction__add', {
                            'Reaction__add--open': showEmojiPicker,
                        })}
                        {...getReferenceProps()}
                    >
                        <AddReactionIcon/>
                    </button>
                </WithTooltip>
            </Gate>
            {emojiPicker}
        </span>
    );
}
