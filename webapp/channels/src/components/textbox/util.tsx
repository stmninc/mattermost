import React from 'react';
import type {UserProfile} from '@mattermost/types/users';
import {Preferences} from 'mattermost-redux/constants';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

import * as Utils from 'utils/utils';

const MENTION_REGEX = /@([^<]+)<x-name>@([^<]+)<\/x-name>/g;
const USERNAME_REGEX = /@([a-zA-Z0-9.\-_]+)/g;

export const initializeMapValueFromInputValue = (inputValue: string, usersByUsername: Record<string, UserProfile> | undefined, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    if (!usersByUsername) {
        return inputValue;
    }
    return inputValue.replace(USERNAME_REGEX, (match, username) => {
        const user = usersByUsername[username];
        if (user) {
            const displayUserName = displayUsername(user, teammateNameDisplay, false);
            return createMentionTag(username, displayUserName);
        }
        return match
    });
}

export const convertToDisplayValueFromMapValue = (mapValue: string): string => {
    return mapValue.replace(new RegExp(MENTION_REGEX.source, 'g'), (_, username, displayName) => {
        return `@${displayName}`;
    });
}

export const updateStateWhenSuggestionSelected = (
    item: any, 
    inputValue: string, 
    rawValue: string, 
    usersByUsername: Record<string, UserProfile> | undefined, 
    teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME,
    setState: (state: any, callback?: () => void) => void,
    textBox?: HTMLInputElement | HTMLTextAreaElement | null,
) => {
    if (!usersByUsername) {
        return;
    }
    if (item && item.username && item.type !== 'mention_groups') {
        const newRawValue = generateRawValue(rawValue, inputValue, usersByUsername, teammateNameDisplay);
        const newMapValue = generateMapValue(newRawValue, usersByUsername, teammateNameDisplay);
        const newDisplayValue = convertToDisplayValueFromMapValue(newMapValue);

        // textBoxのvalueを新しいdisplayValueで更新
        if (textBox && textBox.value !== newDisplayValue) {
            textBox.value = newDisplayValue;
        }

        const cursorPosition = calculateCursorPositionAfterMention(
            inputValue,
            item.username,
            displayUsername(item, teammateNameDisplay, false)
        );

        setState({
            rawValue: newRawValue,
            mapValue: newMapValue,
            displayValue: newDisplayValue,
        }, () => {
            window.requestAnimationFrame(() => {
                if (textBox) {
                    Utils.setCaretPosition(textBox, cursorPosition);
                }
            });
        });
    }
}

const createMentionTag = (username: string, displayName: string): string => {
    return `@${username}<x-name>@${displayName}</x-name>`;
};

export const generateRawValue = (rawValue: string, inputValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    const mentionMappings = extractMentionRawMappings(rawValue);

    let result = inputValue
    const replacedPositions = new Set<number>();

    for (const mapping of mentionMappings) {
        const user = usersByUsername[mapping.username];
        const displayName = displayUsername(user, teammateNameDisplay, false);
        const replacement = mapping.username

        if (!user) {
            continue;
        }

        result = replaceFirstUnprocessed(result, displayName, replacement, replacedPositions);
    }

    return result;
};

export const updateStateWhenOnChanged = (mapValue: string, usersByUsername: Record<string, UserProfile> | undefined, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME, setState: (state: any) => void, e: React.ChangeEvent<HTMLInputElement>, onChange: (event: React.ChangeEvent<HTMLInputElement>) => void) => {
    const inputValue = e.target.value;

    if (!usersByUsername) {
        return;
    }

    const newRawValue = convertRawValue(mapValue, inputValue, usersByUsername, teammateNameDisplay);
    const newMapValue = generateMapValue(newRawValue, usersByUsername, teammateNameDisplay);
    const newDisplayValue = convertToDisplayValueFromMapValue(newMapValue);
    const newMentionHighlights = calculateMentionPositions(newMapValue, newDisplayValue);

    setState({
        mapValue: newMapValue,
        rawValue: newRawValue,
        displayValue: newDisplayValue,
        mentionHighlights: newMentionHighlights
    });

    const syntheticEvent = {
        ...e,
        target: {
            ...e.target,
            value: newRawValue,
        },
    } as React.ChangeEvent<HTMLInputElement>;

    onChange(syntheticEvent);
};

export const resetState = (prevProps: any, setState: (state: any) => void, currentChannelId: string, value: string, usersByUsername: Record<string, UserProfile> | undefined, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME) => {
    if (prevProps.channelId !== currentChannelId) {
        setState({
            mapValue: '',
            displayValue: '',
            rawValue: '',
            mentionHighlights: [],
        });
    }

    if (prevProps.value !== value && value.length > 0 && prevProps.value.length === 0 && usersByUsername) {
            const mapValue = generateMapValue(value, usersByUsername, teammateNameDisplay);
            const displayValue = initializeMapValueFromInputValue(value, usersByUsername, teammateNameDisplay);

            setState({
                rawValue: value,
                mapValue: mapValue,
                displayValue: displayValue,
                mentionHighlights: calculateMentionPositions(mapValue, displayValue)
            });
        }

    if (prevProps.value !== value && value.length === 0 && prevProps.value.length > 0) {
        setState({
            rawValue: "",
            mapValue: "",
            displayValue: "",
            mentionHighlights: [],
        });
    }
};

export const calculateMentionPositions = (mapValue: string, displayValue: string): Array<{start: number; end: number; username: string}> => {
    const positions: Array<{start: number; end: number; username: string}> = [];
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
    };
    return positions;
}

export const renderMentionOverlay = (
    textbox: HTMLTextAreaElement | null, 
    mentionHighlights: Array<{start: number; end: number; username: string}>, 
    displayValue: string
): JSX.Element | null => {
    if (!textbox || mentionHighlights.length === 0) {
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
            {renderHighlightedText(mentionHighlights, displayValue)}
        </div>
    );
};

const generateMapValue = (rawValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME ): string => {
    const mentionMappings = extractMentionRawMappings(rawValue);

    let result = rawValue
    const replacedPositions = new Set<number>();

    for (const mapping of mentionMappings) {
        const user = usersByUsername[mapping.username];
        const displayName = displayUsername(user, teammateNameDisplay, false);
        result = replaceFirstUnprocessed(result, mapping.fullMatch, `@${mapping.username}<x-name>@${displayName}</x-name>`, replacedPositions);
    }
    return result;
}

const extractMentionRawMappings = (rawValue: string): Array<{ fullMatch: string; username: string; }> => {
    const mappings: Array<{ fullMatch: string; username: string; }> = [];
    const regex = new RegExp(USERNAME_REGEX.source, 'g');
    let match;

    while ((match = regex.exec(rawValue)) !== null) {
        mappings.push({
            fullMatch: match[0],
            username: match[1]
        });
    }

    return mappings;
};

const replaceFirstUnprocessed = (
    text: string,
    pattern: string,
    replacement: string,
    processedPositions: Set<number>,
): string => {
    let searchIndex = 0;
    let foundIndex = -1;

    while ((foundIndex = text.indexOf(pattern, searchIndex)) !== -1) {
        if (!processedPositions.has(foundIndex)) {
            const result = text.slice(0, foundIndex) + replacement + text.slice(foundIndex + pattern.length);

            // Mark processed positions
            for (let i = foundIndex; i < foundIndex + replacement.length; i++) {
                processedPositions.add(i);
            }

            return result;
        }
        searchIndex = foundIndex + 1;
    }

    return text;
};

const calculateCursorPositionAfterMention = (
    textValue: string,
    username: string,
    displayName: string
): number => {
    const usernameIndex = textValue.indexOf(username);
    if (usernameIndex === -1) {
        return textValue.length;
    }

    const basePosition = usernameIndex + username.length + 1;

    const lengthDifference = displayName.length - username.length;

    return basePosition + lengthDifference;
};

const convertRawValue = (mapValue: string, inputValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    const mentionMappings = extractMentionMapMappings(mapValue);

    let result = inputValue
    const replacedPositions = new Set<number>();

    for (const mapping of mentionMappings) {
        const user = usersByUsername[mapping.username];
        const displayName = displayUsername(user, teammateNameDisplay, false);
        const replacement = mapping.username

        if (!user) {
            continue;
        }

        const mentionPattern = `@${displayName}`;
        const mentionIndex = result.indexOf(mentionPattern);

        if (mentionIndex !== -1) {
            const beforeMentionIndex = mentionIndex - 1;
            const afterMentionIndex = mentionIndex + mentionPattern.length;
            const charBeforeMention = result.charAt(beforeMentionIndex);
            const charAfterMention = result.charAt(afterMentionIndex);

            // メンションの前が文字列の開始またはスペース・改行・日本語文字で、
            // かつメンションの後が文字列の終端またはスペース・改行・日本語文字の場合置換
            const beforeValid = mentionIndex === 0 || /[\s\n\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FAF]/.test(charBeforeMention);
            const afterValid = afterMentionIndex === result.length || /[\s\n\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FAF]/.test(charAfterMention);

            if (beforeValid && afterValid) {
                result = replaceFirstUnprocessed(result, displayName, replacement, replacedPositions);
            }
        }
    }

    return result;
};

const extractMentionMapMappings = (mapValue: string): Array<{ fullMatch: string; username: string; }> => {
    const mappings: Array<{ fullMatch: string; username: string; }> = [];
    const regex = /@([a-zA-Z0-9.\-_]+)<x-name>.*?<\/x-name>/g;
    let match;

    while ((match = regex.exec(mapValue)) !== null) {
        console.log('matchMapValue', match);
        mappings.push({
            fullMatch: match[0],
            username: match[1]
        });
    }

    return mappings;
}

const renderHighlightedText = (mentionHighlights: Array<{start: number; end: number; username: string}>, displayValue: string) => {
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
