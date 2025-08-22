// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import classNames from 'classnames';
import React from 'react';
import type {ChangeEvent, ElementType, FocusEvent, KeyboardEvent, MouseEvent} from 'react';
import {FormattedMessage} from 'react-intl';

import type {Channel} from '@mattermost/types/channels';
import type {Group} from '@mattermost/types/groups';
import type {UserProfile} from '@mattermost/types/users';

import {Preferences} from 'mattermost-redux/constants';
import type {ActionResult} from 'mattermost-redux/types/actions';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

import AutosizeTextarea from 'components/autosize_textarea';
import PostMarkdown from 'components/post_markdown';
import AtMentionProvider from 'components/suggestion/at_mention_provider';
import ChannelMentionProvider from 'components/suggestion/channel_mention_provider';
import AppCommandProvider from 'components/suggestion/command_provider/app_provider';
import CommandProvider from 'components/suggestion/command_provider/command_provider';
import EmoticonProvider from 'components/suggestion/emoticon_provider';
import type Provider from 'components/suggestion/provider';
import SuggestionBox from 'components/suggestion/suggestion_box';
import type SuggestionBoxComponent from 'components/suggestion/suggestion_box/suggestion_box';
import SuggestionList from 'components/suggestion/suggestion_list';
import {convertRawValue, convertToDisplayName, generateMapValue, generateRawValue} from 'components/textbox/util';

import * as Utils from 'utils/utils';

import type {TextboxElement} from './index';
import { su } from 'i18n/langmap';

const ALL = ['all'];

export type Props = {
    id: string;
    channelId: string;
    rootId?: string;
    tabIndex?: number;
    value: string;
    onChange: (e: ChangeEvent<TextboxElement>) => void;
    onKeyPress: (e: KeyboardEvent<any>) => void;
    onComposition?: () => void;
    onHeightChange?: (height: number, maxHeight: number) => void;
    onWidthChange?: (width: number) => void;
    createMessage: string;
    onKeyDown?: (e: KeyboardEvent<TextboxElement>) => void;
    onMouseUp?: (e: React.MouseEvent<TextboxElement>) => void;
    onKeyUp?: (e: React.KeyboardEvent<TextboxElement>) => void;
    onBlur?: (e: FocusEvent<TextboxElement>) => void;
    onFocus?: (e: FocusEvent<TextboxElement>) => void;
    supportsCommands?: boolean;
    handlePostError?: (message: JSX.Element | null) => void;
    onPaste?: (e: ClipboardEvent) => void;
    suggestionList?: React.ComponentProps<typeof SuggestionBox>['listComponent'];
    suggestionListPosition?: React.ComponentProps<typeof SuggestionList>['position'];
    alignWithTextbox?: boolean;
    emojiEnabled?: boolean;
    characterLimit: number;
    disabled?: boolean;
    badConnection?: boolean;
    currentUserId: string;
    currentTeamId: string;
    preview?: boolean;
    autocompleteGroups: Group[] | null;
    delayChannelAutocomplete: boolean;
    actions: {
        autocompleteUsersInChannel: (prefix: string, channelId: string) => Promise<ActionResult>;
        autocompleteChannels: (term: string, success: (channels: Channel[]) => void, error: () => void) => Promise<ActionResult>;
        searchAssociatedGroupsForReference: (prefix: string, teamId: string, channelId: string | undefined) => Promise<{ data: any }>;
    };
    useChannelMentions: boolean;
    inputComponent?: ElementType;
    openWhenEmpty?: boolean;
    priorityProfiles?: UserProfile[];
    hasLabels?: boolean;
    hasError?: boolean;
    isInEditMode?: boolean;
    usersByUsername?: Record<string, UserProfile>;
    teammateNameDisplay?: string;
};

const VISIBLE = {visibility: 'visible'};
const HIDDEN = {visibility: 'hidden'};

interface TextboxState {
    mapValue: string;
    displayValue: string; // UI display value (username→fullname converted)
    rawValue: string; // Server submission value (username format)
    mentionHighlights: Array<{start: number; end: number; username: string}>; // 追加
}

export default class Textbox extends React.PureComponent<Props, TextboxState> {
    private readonly suggestionProviders: Provider[];
    private readonly wrapper: React.RefObject<HTMLDivElement>;
    private readonly message: React.RefObject<SuggestionBoxComponent>;
    private readonly preview: React.RefObject<HTMLDivElement>;
    private readonly textareaRef: React.RefObject<HTMLTextAreaElement>;

    state: TextboxState = {
        mapValue: '',
        displayValue: '', // UI display value (username→fullname converted)
        rawValue: '', // Server submission value (username format)
        mentionHighlights: [], // 追加
    };

    static defaultProps = {
        supportsCommands: true,
        inputComponent: AutosizeTextarea,
        suggestionList: SuggestionList,
    };

    constructor(props: Props) {
        super(props);

        this.suggestionProviders = [];

        if (props.supportsCommands) {
            this.suggestionProviders.push(new AppCommandProvider({
                teamId: this.props.currentTeamId,
                channelId: this.props.channelId,
                rootId: this.props.rootId,
            }));
        }

        this.suggestionProviders.push(
            new AtMentionProvider({
                textboxId: this.props.id,
                currentUserId: this.props.currentUserId,
                channelId: this.props.channelId,
                autocompleteUsersInChannel: (prefix: string) => this.props.actions.autocompleteUsersInChannel(prefix, this.props.channelId),
                useChannelMentions: this.props.useChannelMentions,
                autocompleteGroups: this.props.autocompleteGroups,
                searchAssociatedGroupsForReference: (prefix: string) => this.props.actions.searchAssociatedGroupsForReference(prefix, this.props.currentTeamId, this.props.channelId),
                priorityProfiles: this.props.priorityProfiles,
            }),
            new ChannelMentionProvider(props.actions.autocompleteChannels, props.delayChannelAutocomplete),
            new EmoticonProvider(),
        );

        if (props.supportsCommands) {
            this.suggestionProviders.push(new CommandProvider({
                teamId: this.props.currentTeamId,
                channelId: this.props.channelId,
                rootId: this.props.rootId,
            }));
        }

        this.checkMessageLength(props.value);
        this.wrapper = React.createRef();
        this.message = React.createRef();
        this.preview = React.createRef();
        this.textareaRef = React.createRef();

        console.log('this.props.isInEditMode', this.props.isInEditMode)

        // Initialize state - set displayValue and rawValue from props.value
        // const mapValue = initializeToMapValue(props.value, props.usersByUsername, props.teammateNameDisplay);
        const initialDisplayValue = convertToDisplayName(props.value, props.usersByUsername, props.teammateNameDisplay);
        const initialMapValue = generateMapValue(props.value, props.usersByUsername, props.teammateNameDisplay);
        
        this.state = {
            displayValue: initialDisplayValue,
            rawValue: props.value,
            mapValue: initialMapValue,
            mentionHighlights: this.calculateMentionPositions(initialMapValue, initialDisplayValue), // 修正
        };
    }

    /**
     * Get raw value for server submission (username format)
     */
    getRawValue = () => {
        return this.state.rawValue;
    };

    /**
     * Get display value for UI (fullname format)
     */
    getDisplayValue = () => {
        return this.state.displayValue;
    };

    /**
     * Calculate mention positions in the text
     */
    private calculateMentionPositions = (mapValue: string, displayValue: string): Array<{start: number; end: number; username: string}> => {
        const positions: Array<{start: number; end: number; username: string}> = [];
        
        // mapValueからメンション情報を抽出
        const mapMentionRegex = /@([a-zA-Z0-9.\-_]+)<x-name>@([^<]+)<\/x-name>/g;
        let mapMatch;
        
        while ((mapMatch = mapMentionRegex.exec(mapValue)) !== null) {
            const username = mapMatch[1];
            const displayName = mapMatch[2];
            
            // displayValue内でこのメンションの位置を探す
            const displayMentionPattern = `@${displayName}`;
            let searchStartIndex = 0;
            let displayIndex = displayValue.indexOf(displayMentionPattern, searchStartIndex);
            
            while (displayIndex !== -1) {
                // 既に処理された位置でないかチェック
                const isAlreadyProcessed = positions.some(pos => 
                    displayIndex >= pos.start && displayIndex < pos.end
                );
                
                if (!isAlreadyProcessed) {
                    positions.push({
                        start: displayIndex,
                        end: displayIndex + displayMentionPattern.length,
                        username: username
                    });
                    break; // 最初の一致のみを処理
                }
                
                // 次の一致を探す
                searchStartIndex = displayIndex + 1;
                displayIndex = displayValue.indexOf(displayMentionPattern, searchStartIndex);
            }
        }
        
        return positions;
    };

    handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const inputValue = e.target.value;
        // const newMapValue = convertToMapValue(inputValue, this.state.mapValue);
        console.log('inputValue', inputValue, 'rawValue', this.state.rawValue, this.state.mapValue);
        const newRawValue = convertRawValue(this.state.mapValue, inputValue, this.props.usersByUsername, this.props.teammateNameDisplay);
        const newMapValue = generateMapValue(newRawValue, this.props.usersByUsername, this.props.teammateNameDisplay);
        const newDisplayValue = convertToDisplayName(newRawValue, this.props.usersByUsername, this.props.teammateNameDisplay);

        // メンション位置を計算
        const mentionHighlights = this.calculateMentionPositions(newMapValue, newDisplayValue);

        this.setState({
            mapValue: newMapValue,
            rawValue: newRawValue,
            displayValue: newDisplayValue,
            mentionHighlights, // 追加
        });

        // Pass raw value (username format) to parent component
        const syntheticEvent = {
            ...e,
            target: {
                ...e.target,
                value: newRawValue,
            },
        } as React.ChangeEvent<HTMLInputElement>;

        this.props.onChange(syntheticEvent);
    };

    updateSuggestions(prevProps: Props) {
        if (this.props.channelId !== prevProps.channelId ||
            this.props.currentUserId !== prevProps.currentUserId ||
            this.props.autocompleteGroups !== prevProps.autocompleteGroups ||
            this.props.useChannelMentions !== prevProps.useChannelMentions ||
            this.props.currentTeamId !== prevProps.currentTeamId ||
            this.props.priorityProfiles !== prevProps.priorityProfiles) {
            // Update channel id for AtMentionProvider.
            for (const provider of this.suggestionProviders) {
                if (provider instanceof AtMentionProvider) {
                    provider.setProps({
                        textboxId: this.props.id,
                        currentUserId: this.props.currentUserId,
                        channelId: this.props.channelId,
                        autocompleteUsersInChannel: (prefix: string) => this.props.actions.autocompleteUsersInChannel(prefix, this.props.channelId),
                        useChannelMentions: this.props.useChannelMentions,
                        autocompleteGroups: this.props.autocompleteGroups,
                        searchAssociatedGroupsForReference: (prefix: string) => this.props.actions.searchAssociatedGroupsForReference(prefix, this.props.currentTeamId, this.props.channelId),
                        priorityProfiles: this.props.priorityProfiles,
                    });
                }
            }
        }

        if (this.props.channelId !== prevProps.channelId ||
            this.props.currentTeamId !== prevProps.currentTeamId ||
            this.props.rootId !== prevProps.rootId) {
            // Update channel id for CommandProvider and AppCommandProvider.
            for (const provider of this.suggestionProviders) {
                if (provider instanceof CommandProvider) {
                    provider.setProps({
                        teamId: this.props.currentTeamId,
                        channelId: this.props.channelId,
                        rootId: this.props.rootId,
                    });
                }
                if (provider instanceof AppCommandProvider) {
                    provider.setProps({
                        teamId: this.props.currentTeamId,
                        channelId: this.props.channelId,
                        rootId: this.props.rootId,
                    });
                }
            }
        }

        if (this.props.delayChannelAutocomplete !== prevProps.delayChannelAutocomplete) {
            for (const provider of this.suggestionProviders) {
                if (provider instanceof ChannelMentionProvider) {
                    provider.setProps({
                        delayChannelAutocomplete: this.props.delayChannelAutocomplete,
                    });
                }
            }
        }

        if (prevProps.value !== this.props.value) {
            this.checkMessageLength(this.props.value);
        }

                if (prevProps.channelId !== this.props.channelId) {
            this.setState({
                rawValue: "",
                mapValue: "",
                displayValue: "",
                mentionHighlights: [], // 追加
            });
        }

        if (prevProps.value !== this.props.value && this.props.value.length > 0 && prevProps.value.length === 0) {
            const mapValue = generateMapValue(this.props.value, this.props.usersByUsername, this.props.teammateNameDisplay);
            const displayValue = convertToDisplayName(this.props.value, this.props.usersByUsername, this.props.teammateNameDisplay);

            this.setState({
                rawValue: this.props.value,
                mapValue: mapValue,
                displayValue: displayValue,
                mentionHighlights: this.calculateMentionPositions(mapValue, displayValue), // 修正
            });
        }

        if (prevProps.value !== this.props.value && this.props.value.length === 0 && prevProps.value.length > 0) {
            this.setState({
                rawValue: "",
                mapValue: "",
                displayValue: "",
                mentionHighlights: [], // 追加
            });
        }
    }

    componentDidUpdate(prevProps: Props) {
        if (!prevProps.preview && this.props.preview) {
            this.preview.current?.focus();
        }
        this.updateSuggestions(prevProps);
    }

    checkMessageLength = (message: string) => {
        if (this.props.handlePostError) {
            if (message.length > this.props.characterLimit) {
                const errorMessage = (
                    <FormattedMessage
                        id='create_post.error_message'
                        defaultMessage='Your message is too long. Character count: {length}/{limit}'
                        values={{
                            length: message.length,
                            limit: this.props.characterLimit,
                        }}
                    />);
                this.props.handlePostError(errorMessage);
            } else {
                this.props.handlePostError(null);
            }
        }
    };

    // adding in the HTMLDivElement to support event handling in preview state
    handleKeyDown = (e: KeyboardEvent<TextboxElement | HTMLDivElement>) => {
        // since we do only handle the sending when in preview mode this is fine to be casted
        this.props.onKeyDown?.(e as KeyboardEvent<TextboxElement>);
    };

    handleMouseUp = (e: MouseEvent<TextboxElement>) => this.props.onMouseUp?.(e);

    handleKeyUp = (e: KeyboardEvent<TextboxElement>) => this.props.onKeyUp?.(e);

    // adding in the HTMLDivElement to support event handling in preview state
    handleBlur = (e: FocusEvent<TextboxElement | HTMLDivElement>) => {
        // since we do only handle the sending when in preview mode this is fine to be casted
        this.props.onBlur?.(e as FocusEvent<TextboxElement>);
    };

    /**
     * Calculate cursor position after mention replacement
     */
    private calculateCursorPositionAfterMention = (
        textValue: string,
        username: string,
        displayName: string
    ): number => {
        const usernameIndex = textValue.indexOf(username);
        if (usernameIndex === -1) {
            return textValue.length;
        }
        
        // Position after username + space character
        const basePosition = usernameIndex + username.length + 1;
        
        // Adjust for the difference in length between username and display name
        const lengthDifference = displayName.length - username.length;
        
        return basePosition + lengthDifference;
    };

    /**
     * Handles when a mention suggestion is selected
     * Stores information about mentions explicitly selected by the user
     */
    handleSuggestionSelected = (item: any) => {
        // Only process items from AtMentionProvider
        if (item && item.username && item.type !== 'mention_groups') {

            const { usersByUsername = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME } = this.props;

            const textBox = this.getInputBox();
            const textBoxValue = textBox.value;
            
            // Calculate cursor position after mention replacement
            const cursorPosition = this.calculateCursorPositionAfterMention(
                textBoxValue,
                item.username,
                displayUsername(item, teammateNameDisplay, false)
            );

            const newRawValue = generateRawValue(this.state.rawValue, textBoxValue, usersByUsername, teammateNameDisplay);
            const newMapValue = generateMapValue(newRawValue, usersByUsername, teammateNameDisplay);
            const newDisplayValue = convertToDisplayName(newRawValue, usersByUsername, teammateNameDisplay);

            // Update the textbox value directly to ensure SuggestionBox displays the correct value
            if (textBox && textBox.value !== newDisplayValue) {
                textBox.value = newDisplayValue;
            }

            this.setState(
                {
                    rawValue: newRawValue,
                    mapValue: newMapValue,
                    displayValue: newDisplayValue,
                    mentionHighlights: this.calculateMentionPositions(newMapValue, newDisplayValue), // 修正
                },
                () => {
                    // Set cursor position after state update
                    window.requestAnimationFrame(() => {
                        Utils.setCaretPosition(textBox, cursorPosition);
                    });
                }
            );
        }
    };

    handleSuggestionsReceived = (suggestions: any[]) => {

        // Use type assertion to access matchedPretext
        const matchedPretext = (suggestions as any).matchedPretext;

        if (suggestions && matchedPretext && this.state.rawValue !== this.state.displayValue && matchedPretext.endsWith(' ')) {
            const convertedDisplayValue = convertToDisplayName(this.state.rawValue, this.props.usersByUsername, this.props.teammateNameDisplay);

            if (convertedDisplayValue === this.state.displayValue) {
                if (this.message.current) {
                    this.message.current.handleEmitClearSuggestions();
                }
            }
        }
    };

    getInputBox = () => {
        const textbox = this.message.current?.getTextbox();
        if (textbox && this.textareaRef.current !== textbox) {
            // Update textareaRef
            (this.textareaRef as any).current = textbox;
        }
        return textbox;
    };

    focus = () => {
        const textbox = this.getInputBox();
        if (textbox) {
            textbox.focus();
            Utils.placeCaretAtEnd(textbox);
            setTimeout(() => {
                Utils.scrollToCaret(textbox);
            });

            // reset character count warning
            this.checkMessageLength(textbox.value);
        }
    };

    blur = () => {
        this.getInputBox()?.blur();
    };

    getStyle = () => {
        return this.props.preview ? HIDDEN : VISIBLE;
    };

    /**
     * Render mention overlay to highlight mentions in blue
     */
    private renderMentionOverlay = () => {
        const textbox = this.getInputBox();
        if (!textbox || this.state.mentionHighlights.length === 0) {
            return null;
        }

        const computedStyle = window.getComputedStyle(textbox);
        const overlayStyle: React.CSSProperties = {
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            pointerEvents: 'none',
            color: 'transparent',
            backgroundColor: 'transparent',
            border: 'transparent',
            fontFamily: computedStyle.fontFamily,
            fontSize: computedStyle.fontSize,
            lineHeight: computedStyle.lineHeight,
            padding: computedStyle.padding,
            whiteSpace: 'pre-wrap',
            wordWrap: 'break-word',
            overflow: 'hidden',
            zIndex: 1,
        };

        return (
            <div style={overlayStyle} className="mention-overlay">
                {this.renderHighlightedText()}
            </div>
        );
    };

    /**
     * Render text with highlighted mentions
     */
    private renderHighlightedText = () => {
        const { displayValue, mentionHighlights } = this.state;
        const parts: JSX.Element[] = [];
        let lastIndex = 0;

        mentionHighlights.forEach((highlight, index) => {
            // 通常のテキスト部分
            if (highlight.start > lastIndex) {
                parts.push(
                    <span key={`text-${index}`} style={{ color: 'transparent' }}>
                        {displayValue.substring(lastIndex, highlight.start)}
                    </span>
                );
            }

            // メンション部分（青色でハイライト）
            parts.push(
                <span 
                    key={`mention-${index}`} 
                    className="mention-highlight"
                    style={{
                        color: Preferences.THEMES.denim.linkColor, // メンションの色をテーマに合わせる
                    }}
                >
                    {displayValue.substring(highlight.start, highlight.end)}
                </span>
            );

            lastIndex = highlight.end;
        });

        // 最後の通常テキスト部分
        if (lastIndex < displayValue.length) {
            parts.push(
                <span key="text-final" style={{ color: 'transparent' }}>
                    {displayValue.substring(lastIndex)}
                </span>
            );
        }

        return parts;
    };

    render() {
        let textboxClassName = 'form-control custom-textarea textbox-edit-area';
        if (this.props.emojiEnabled) {
            textboxClassName += ' custom-textarea--emoji-picker';
        }
        if (this.props.badConnection) {
            textboxClassName += ' bad-connection';
        }
        if (this.props.hasLabels) {
            textboxClassName += ' textarea--has-labels';
        }

        if (this.props.hasError) {
            textboxClassName += ' textarea--has-errors';
        }

        return (
            <div
                ref={this.wrapper}
                className={classNames('textarea-wrapper', {'textarea-wrapper-preview': this.props.preview, 'textarea-wrapper-preview--disabled': Boolean(this.props.preview && this.props.disabled)})}
            >
                <div
                    tabIndex={this.props.tabIndex}
                    ref={this.preview}
                    className={classNames('form-control custom-textarea textbox-preview-area', {'textarea--has-labels': this.props.hasLabels})}
                    onKeyPress={this.props.onKeyPress}
                    onKeyDown={this.handleKeyDown}
                    onBlur={this.handleBlur}
                >
                    <PostMarkdown
                        message={this.state.rawValue}
                        channelId={this.props.channelId}
                        imageProps={{hideUtilities: true}}
                    />
                </div>
                <div style={{position: 'relative'}}>
                <SuggestionBox
                    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                    // @ts-ignore
                    ref={this.message}
                    id={this.props.id}
                    className={textboxClassName}
                    spellCheck='true'
                    placeholder={this.props.createMessage}
                    onChange={this.handleChange}
                    onKeyPress={this.props.onKeyPress}
                    onKeyDown={this.handleKeyDown}
                    onMouseUp={this.handleMouseUp}
                    onKeyUp={this.handleKeyUp}
                    onComposition={this.props.onComposition}
                    onBlur={this.handleBlur}
                    onFocus={this.props.onFocus}
                    onHeightChange={this.props.onHeightChange}
                    onWidthChange={this.props.onWidthChange}
                    onPaste={this.props.onPaste}
                    style={this.getStyle()}
                    inputComponent={this.props.inputComponent}
                    listComponent={this.props.suggestionList}
                    listPosition={this.props.suggestionListPosition}
                    providers={this.suggestionProviders}
                    value={this.state.displayValue}
                    renderDividers={ALL}
                    disabled={this.props.disabled}
                    contextId={this.props.channelId}
                    openWhenEmpty={this.props.openWhenEmpty}
                    alignWithTextbox={this.props.alignWithTextbox}
                    onItemSelected={this.handleSuggestionSelected}
                    onSuggestionsReceived={this.handleSuggestionsReceived}
                />
                {!this.props.preview && this.renderMentionOverlay()}
                </div>
            </div>
        );
    }
}
