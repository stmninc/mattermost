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
