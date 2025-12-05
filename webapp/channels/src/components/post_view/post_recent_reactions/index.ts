// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';
import type {Dispatch} from 'redux';

import type {Emoji} from '@mattermost/types/emojis';

import {canAddReactions} from 'mattermost-redux/selectors/entities/reactions';

import {toggleReaction} from 'actions/post_actions';
import {getEmojiMap} from 'selectors/emojis';
import {getCurrentLocale} from 'selectors/i18n';

import type {GlobalState} from 'types/store';

import PostReaction from './post_recent_reactions';

type OwnProps = {
    channelId?: string;
};

function mapDispatchToProps(dispatch: Dispatch) {
    return {
        actions: bindActionCreators({
            toggleReaction,
        }, dispatch),
    };
}

function mapStateToProps(state: GlobalState, ownProps: OwnProps) {
    const locale = getCurrentLocale(state);
    const emojiMap = getEmojiMap(state);
    const defaultEmojis = [emojiMap.get('thumbsup'), emojiMap.get('grinning'), emojiMap.get('white_check_mark')] as Emoji[];

    const canAdd = ownProps.channelId ? canAddReactions(state, ownProps.channelId) : false;

    return {
        defaultEmojis,
        locale,
        canAddReactions: canAdd,
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(PostReaction);
