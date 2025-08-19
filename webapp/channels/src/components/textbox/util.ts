// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {UserProfile} from '@mattermost/types/users';

import {Preferences} from 'mattermost-redux/constants';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

const MENTION_REGEX = /@([^<]+)<x-name>@([^<]+)<\/x-name>/g;
const USERNAME_REGEX = /@([a-zA-Z0-9.\-_]+)/g;
const TAG_REGEX = /<x-name>.*?<\/x-name>/g;

export const convertToDisplayName = (mapValue: string): string => {
    return mapValue.replace(new RegExp(MENTION_REGEX.source, 'g'), (_, username, displayName) => {
        return `@${displayName}`;
    });
};

export const convertToRawValue = (mapValue: string): string => {
    return mapValue.replace(new RegExp(MENTION_REGEX.source, 'g'), (_, username) => {
        return `@${username}`;
    });
};

export const initializeToMapValue = (rawValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    return rawValue.replace(USERNAME_REGEX, (match, username) => {
        const displayName = getUserDisplayName(username, usersByUsername, teammateNameDisplay);
        return displayName ? createMentionTag(username, displayName) : match;
    });
};

const getUserDisplayName = (username: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    const user = usersByUsername[username];

    if (user) {
        const displayName = displayUsername(user, teammateNameDisplay, false);
        return displayName || username;
    }

    return username;
};

const createMentionTag = (username: string, displayName: string): string => {
    return `@${username}<x-name>@${displayName}</x-name>`;
};

export const convertToMapValue = (inputValue: string, mapValue: string): string => {
    if (!mapValue) {
        return inputValue;
    }

    const mentionMappings = extractMentionMappings(mapValue);

    if (mentionMappings.length === 0) {
        return inputValue;
    }

    return applyMentionMappings(inputValue, mentionMappings);
};

const extractMentionMappings = (mapValue: string): Array<{ fullMatch: string; username: string; displayName: string }> => {
    const mappings: Array<{ fullMatch: string; username: string; displayName: string }> = [];
    const regex = new RegExp(MENTION_REGEX.source, 'g');
    let match;

    while ((match = regex.exec(mapValue)) !== null) {
        mappings.push({
            fullMatch: match[0],
            username: match[1],
            displayName: match[2],
        });
    }

    return mappings;
};

const applyMentionMappings = (inputValue: string, mappings: Array<{ username: string; displayName: string }>): string => {
    let result = inputValue;
    const replacedPositions = new Set<number>();

    for (const mapping of mappings) {
        const displayNamePattern = `@${mapping.displayName}`;
        const replacement = createMentionTag(mapping.username, mapping.displayName);

        result = replaceFirstUnprocessed(result, displayNamePattern, replacement, replacedPositions);
    }

    return result;
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

export const generateMapValue = (usernameMention: string, displayNameMention: string, mapValue: string, inputValue: string): string => {
    const insertionPoint = inputValue.indexOf(usernameMention);

    if (insertionPoint === -1) {
        return mapValue;
    }

    // Sync mapValue with inputValue if needed
    const syncedMapValue = syncMapValueWithInput(mapValue, inputValue);
    const adjustedInsertionPoint = calculateInsertionPoint(insertionPoint, syncedMapValue);

    return insertMentionTag(syncedMapValue, adjustedInsertionPoint, usernameMention, displayNameMention);
};

const syncMapValueWithInput = (mapValue: string, inputValue: string): string => {
    const mapTextContent = mapValue.replace(TAG_REGEX, '');

    if (mapTextContent.length < inputValue.length) {
        const additionalText = inputValue.slice(mapTextContent.length);
        return mapValue + additionalText;
    }

    return mapValue;
};

const calculateInsertionPoint = (baseInsertionPoint: number, mapValue: string): number => {
    const adjustedPoint = baseInsertionPoint;
    const existingMentions = mapValue.match(TAG_REGEX) || [];

    let tagOffset = 0;
    for (const mention of existingMentions) {
        const mentionStart = mapValue.indexOf(mention, tagOffset);
        if (mentionStart < adjustedPoint + tagOffset) {
            tagOffset += mention.length - mention.replace(/<\/?x-name>/g, '').length;
        } else {
            break;
        }
    }

    return adjustedPoint + tagOffset;
};

const insertMentionTag = (mapValue: string, insertionPoint: number, usernameMention: string, displayNameMention: string): string => {
    const before = mapValue.slice(0, insertionPoint + usernameMention.length);
    const after = mapValue.slice(insertionPoint + usernameMention.length);
    const nameTag = `<x-name>${displayNameMention}</x-name>`;

    return before + nameTag + after;
};

