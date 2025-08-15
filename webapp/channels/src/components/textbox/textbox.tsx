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

import type {MentionKey} from 'utils/text_formatting';
import * as Utils from 'utils/utils';

import type {TextboxElement} from './index';

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
    mentionKeys?: MentionKey[];
};

const VISIBLE = {visibility: 'visible'};
const HIDDEN = {visibility: 'hidden'};

interface TextboxState {
    displayValue: string; // UI display value (username→fullname converted)
    rawValueNew: string; // Server submission value (username format)
    rawValue: string; // Server submission value (username format)
    selectedMentions: Record<string, string>; // Mapping: @username -> @displayName (mentions explicitly selected by user)
}

export default class Textbox extends React.PureComponent<Props, TextboxState> {
    private readonly suggestionProviders: Provider[];
    private readonly wrapper: React.RefObject<HTMLDivElement>;
    private readonly message: React.RefObject<SuggestionBoxComponent>;
    private readonly preview: React.RefObject<HTMLDivElement>;
    private readonly textareaRef: React.RefObject<HTMLTextAreaElement>;

    // Memoization cache for displayNameToUsername map
    private displayNameMapCache: {
        usersByUsername: Record<string, UserProfile> | null;
        teammateNameDisplay: string | null;
        map: Record<string, string[]>;
        sortedDisplayNames: string[];
    } = {
            usersByUsername: null,
            teammateNameDisplay: null,
            map: {},
            sortedDisplayNames: [],
        };

    state: TextboxState = {
        displayValue: '', // UI display value (username→fullname converted)
        rawValueNew: '', // Server submission value (username format)
        rawValue: '', // Server submission value (username format)
        selectedMentions: {}, // Mapping: displayName -> username (mentions explicitly selected by user)
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

        // Initialize state - set displayValue and rawValue from props.value
        this.state = {
            displayValue: this.convertToDisplayName(props.value),
            rawValue: props.value,
            rawValueNew: props.value,
            selectedMentions: {},
        };
    }

    /**
     * Convert username (@user) to fullname/nickname (@Full Name)
     */
    convertToDisplayName = (text: string): string => {
        // @John Doe Hi, How are you, @John Smith(user-5678)
        // 上記のようなtextを受け取った場合は、@John Doe Hi, How are you, @John Smith Hi に変換する
        // selectedMentionsを参照して、変換後のテキストを生成する
        
        const {usersByUsername = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME} = this.props;
        const {selectedMentions = {}} = this.state;
        
        let result = text;
        
        // First, handle @displayName(username) patterns - remove the (username) part
        result = result.replace(/@([^()]+)\(([^)]+)\)/g, (match, displayName, username) => {
            return `@${displayName}`;
        });
        
        // Then convert @username to @displayName using selectedMentions or usersByUsername
        result = result.replace(/@([a-zA-Z0-9.\-_]+)/g, (match, username) => {
            // First check selectedMentions
            const selectedDisplayName = selectedMentions[`@${username}`];
            if (selectedDisplayName) {
                return selectedDisplayName; // selectedDisplayName already includes @
            }
            
            // Then check usersByUsername
            const user = usersByUsername[username];
            if (user) {
                const displayName = displayUsername(user, teammateNameDisplay, false);
                if (displayName && displayName !== username) {
                    return `@${displayName}`;
                }
            }
            
            return match; // Return original if no conversion found
        });
        
        return result;
    };

    /**
     * Get memoized displayNameToUsername map and sorted display names
     * Only rebuilds when usersByUsername or teammateNameDisplay changes
     */
    private getDisplayNameMap = (): {map: Record<string, string[]>; sortedDisplayNames: string[]} => {
        const {usersByUsername = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME} = this.props;

        // Check if cache is valid
        if (this.displayNameMapCache.usersByUsername === usersByUsername &&
            this.displayNameMapCache.teammateNameDisplay === teammateNameDisplay) {
            return {
                map: this.displayNameMapCache.map,
                sortedDisplayNames: this.displayNameMapCache.sortedDisplayNames,
            };
        }

        // Rebuild cache
        const displayNameToUsername: Record<string, string[]> = {};
        Object.entries(usersByUsername).forEach(([username, user]) => {
            const displayName = displayUsername(user, teammateNameDisplay, false);

            if (displayName && displayName !== username) {
                if (!displayNameToUsername[displayName]) {
                    displayNameToUsername[displayName] = [];
                }
                displayNameToUsername[displayName].push(username);
            }
        });

        const sortedDisplayNames = Object.keys(displayNameToUsername).sort((a, b) => b.length - a.length);

        // Update cache
        this.displayNameMapCache = {
            usersByUsername,
            teammateNameDisplay,
            map: displayNameToUsername,
            sortedDisplayNames,
        };

        return {
            map: displayNameToUsername,
            sortedDisplayNames,
        };
    };

    /**
     * Convert fullname/nickname (@Full Name) to username (@user)
     */
    convertToRawValue = (text: string): string => {
        console.log('text:', text);
        const {selectedMentions = {}} = this.state;

        // Get memoized display name map
        const {map: displayNameToUsername, sortedDisplayNames} = this.getDisplayNameMap();

        let result = text;
        for (const displayName of sortedDisplayNames) {
            const escapedDisplayName = displayName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            const regex = new RegExp(`@${escapedDisplayName}(?=\\s|$|[^\\w])`, 'g');

            result = result.replace(regex, () => {
                if (selectedMentions[displayName]) {
                    return `@${selectedMentions[displayName]}`;
                }
                const usernames = displayNameToUsername[displayName];
                return `@${usernames[0]}`;
            });
        }

        return result;
    };

    convertToRawValueNew = (inputValue: string, newRawValue: string): string => {
        // inputValue = @John Doe Hi, How are you, @user-5678
        // newRawValue = @John Doe(user-1234) Hi
        // 上記のような入力を受け取った場合は、@John Doe Hi, How are you, @John Smith(user-5678) に変換する
        // selectedMentions = {
        //   'John Doe': 'user-1234',
        //   'John Smith': 'user-5678'
        // }
        
        const {usersByUsername = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME} = this.props;
        const {selectedMentions = {}} = this.state;

        console.log('Selected Mentions:', selectedMentions);

        let result = inputValue;

        // Handle mentions in the format @username by converting them to @displayname(username)
        result = result.replace(/@([a-zA-Z0-9.\-_]+)/g, (match, username) => {
            const user = usersByUsername[username];
            if (user) {
                const displayName = displayUsername(user, teammateNameDisplay, false);
                // If the username and display name are different, use the format @displayname(username)
                if (displayName && displayName !== username) {
                    return `@${displayName}(${username})`;
                }
            }
            return match;
        });

        // Handle mentions that are already in display name format by adding the username in parentheses
        // Get memoized display name map
        const {map: displayNameToUsername, sortedDisplayNames} = this.getDisplayNameMap();

        for (const displayName of sortedDisplayNames) {
            const escapedDisplayName = displayName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            // Match @displayName that is not already in the format @displayName(username)
            const regex = new RegExp(`@${escapedDisplayName}(?!\\([^)]+\\))(?=\\s|$|[^\\w])`, 'g');

            result = result.replace(regex, () => {
                // Find the username from selectedMentions by looking for entries where the value matches this display name
                let selectedUsername = null;
                for (const [key, value] of Object.entries(selectedMentions)) {
                    if (value === `@${displayName}`) {
                        selectedUsername = key.replace('@', ''); // Remove @ prefix from key
                        break;
                    }
                }
                
                if (selectedUsername) {
                    return `@${displayName}(${selectedUsername})`;
                }
                // Otherwise, use the first username from the mapping
                const usernames = displayNameToUsername[displayName];
                if (usernames && usernames.length > 0) {
                    return `@${displayName}(${usernames[0]})`;
                }
                return `@${displayName}`;
            });
        }

        return result;
    }

    /**
     * Get raw value for server submission (username format)
     */
    getRawValue = () => {
        return this.state.rawValue;
    };

    getRawValueNew = () => {
        return this.state.rawValueNew;
    };

    /**
     * Get display value for UI (fullname format)
     */
    getDisplayValue = () => {
        return this.state.displayValue;
    };

    // convertToRawValue
    /// - inputValue を @displayname(username) に変換する。例: @john.doe Hi  → @John Doe(user-1234) Hi
    /// - 引数は、inputValue と newRawValue を受け取る。差分を確認して、テキストのマージとメンションがあれば変換して追加する。
    /// - newRawValue に値を格納する
    /// - Submit 時に利用する。その際は、@username に変換する。例: @John Doe(user-1234) Hi → user-1234 Hi
    // convertToDisplayName
    /// - newRawValue を @displayname のみに変換する。例: @John Doe (user-1234) Hi → @John Doe Hi
    /// - newDisplayValue に値を格納する
    /// - PostMarkdown で利用する

    handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const inputValue = e.target.value;

        const newRawValue = this.convertToRawValueNew(inputValue, this.getRawValueNew());

        console.log('New Raw Value (username format):', newRawValue);

        // Update raw value (username format)
        const newRawValueNew = this.convertToRawValue(inputValue);

        // Update display value (fullname format)
        const newDisplayValue = this.convertToDisplayName(newRawValue);

        console.log('Raw Value:', newRawValue);
        console.log('Display Value:', newDisplayValue);

        this.setState({
            rawValueNew: newRawValueNew,
            rawValue: newRawValue,
            displayValue: newDisplayValue,
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

            // Update state when props.value changes
            this.setState({
                rawValueNew: this.props.value,
                rawValue: this.props.value,
                displayValue: this.convertToDisplayName(this.props.value),
            });
        }

        // Recalculate displayValue when usersByUsername or teammateNameDisplay changes
        if (prevProps.usersByUsername !== this.props.usersByUsername ||
            prevProps.teammateNameDisplay !== this.props.teammateNameDisplay) {
            this.setState({
                displayValue: this.convertToDisplayName(this.state.rawValue),
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
     * Handles when a mention suggestion is selected
     * Stores information about mentions explicitly selected by the user
     */
    handleSuggestionSelected = (item: any) => {
        // Only process items from AtMentionProvider
        if (item && item.username && item.type !== 'mention_groups') {
            const displayName = displayUsername(item, this.props.teammateNameDisplay || Preferences.DISPLAY_PREFER_USERNAME, false);
            const username = item.username;

            // Save the selected mention information to state (username -> displayName)
            this.setState((prevState) => ({
                selectedMentions: {
                    ...prevState.selectedMentions,
                    [`@${username}`]: `@${displayName}`,
                },
            }));
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
                        message={this.state.displayValue}
                        channelId={this.props.channelId}
                        options={{
                            mentionHighlight: true,
                            atMentions: true,
                            mentionKeys: this.props.mentionKeys,
                        }}
                        imageProps={{hideUtilities: true}}
                    />
                </div>
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
                />
            </div>
        );
    }
}
