// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {UserProfile} from '@mattermost/types/users';

import {Preferences} from 'mattermost-redux/constants';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

const MENTION_REGEX = /@([^<]+)<x-name>@([^<]+)<\/x-name>/g;
const USERNAME_REGEX = /@([a-zA-Z0-9.\-_]+)/g;

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

export const convertToMapValue = (inputValue: string, mapValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    // if (!mapValue && mapValue.length !== 0) {
    //     return inputValue;
    // }

    // inputValueとmapValueの差分を取得する
    // const diff = 

    // // ユーザー名をアルファベット順にソートする
    // let sortedUsersByUsername = Object.values(usersByUsername).sort((a, b) => a.username.localeCompare(b.username));
    // // displayNameの重複を除去
    // const seenDisplayNames = new Set<string>();
    // sortedUsersByUsername = sortedUsersByUsername.filter(user => {
    //     const displayName = displayUsername(user, teammateNameDisplay, false);
    //     if (seenDisplayNames.has(displayName)) {
    //         return false;
    //     }
    //     seenDisplayNames.add(displayName);
    //     return true;
    // });

    // console.log('mapValue', mapValue)

    // for (const user of sortedUsersByUsername) {
    //     const displayName = displayUsername(user, teammateNameDisplay, false);
    //     // mapValueにdisplayNameが含まれる場合は、userNameに置換する
    //         const displayNameWithSpace = '@' + displayName + ' ';
    //         if (inputValue.includes(displayNameWithSpace)) {
    //             console.log('inputValue', inputValue, 'mapValue', mapValue, 'displayName', displayName);
    //             mapValue = mapValue.replace(`@${displayName}`, createMentionTag(user.username, displayName) + ' ');
    //         }
    // }

    console.log('convertToMapValue', 'mapValue', mapValue, 'inputValue', inputValue);

    const mentionMappings = extractMentionMappings(mapValue);

    console.log('mentionMappings', mentionMappings)


    if (mentionMappings.length === 0) {
        return inputValue;
    }

    const appliedMapValue = applyMentionMappings(inputValue, mentionMappings);

    console.log('applyMentionMappings', appliedMapValue);

    return appliedMapValue;
};

export const generateMapValue = (usernameMention: string, displayNameMention: string, mapValue: string, inputValue: string): string => {
    const mentionMappings = extractMentionMappings(mapValue);

    let result = inputValue;

    for (const mapping of mentionMappings) {
        const displayNamePattern = `@${mapping.displayName}`;
        if (result.includes(displayNamePattern)) {
            result = result.replace(displayNamePattern, mapping.fullMatch);
        }
    }

    if (result.includes(usernameMention)) {
        const newMentionTag = `${usernameMention}<x-name>${displayNameMention}</x-name>`;
        result = result.replace(new RegExp(`${usernameMention}(?!<x-name>)`, 'g'), newMentionTag);
    }

    return result;
};

const extractMentionMappings = (mapValue: string): Array<{ fullMatch: string; username: string; displayName: string }> => {
    const mappings: Array<{ fullMatch: string; username: string; displayName: string }> = [];
    const regex = new RegExp(MENTION_REGEX.source, 'g');
    let match;

    while ((match = regex.exec(mapValue)) !== null) {
        const endIndex = regex.lastIndex;
        const nextChar = mapValue[endIndex] || '';
        if (nextChar !== ' ') {
            continue;
        }
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

            for (let i = foundIndex; i < foundIndex + replacement.length; i++) {
                processedPositions.add(i);
            }

            return result;
        }
        searchIndex = foundIndex + 1;
    }

    return text;
};
