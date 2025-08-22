// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {UserProfile} from '@mattermost/types/users';
import { mapValues } from 'lodash';

import {Preferences} from 'mattermost-redux/constants';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

// const MENTION_REGEX = /@([^<]+)<x-name>@([^<]+)<\/x-name>/g;
const USERNAME_REGEX = /@([a-zA-Z0-9.\-_]+)/g;
const TAG_REGEX = /<x-name>.*?<\/x-name>/g;

// export const convertToDisplayName = (mapValue: string): string => {
//     return mapValue.replace(new RegExp(MENTION_REGEX.source, 'g'), (_, username, displayName) => {
//         return `@${displayName}`;
//     });
// };
export const convertToDisplayName = (rawValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    return rawValue.replace(USERNAME_REGEX, (match, username) => {
        const displayName = getUserDisplayName(username, usersByUsername, teammateNameDisplay);
        return displayName ? `@${displayName}` : match;
    });
}

// export const convertToRawValue = (mapValue: string): string => {
//     return mapValue.replace(new RegExp(MENTION_REGEX.source, 'g'), (_, username) => {
//         return `@${username}`;
//     });
// };

export const convertToRawValue = (rawValue: string, inputValue: string): string => {
    return rawValue.replace(USERNAME_REGEX, (match, username) => {
        return `@${username}`;
    });
}

// export const initializeToMapValue = (rawValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
//     return rawValue.replace(USERNAME_REGEX, (match, username) => {
//         // const displayName = getUserDisplayName(username, usersByUsername, teammateNameDisplay);
//         const user = usersByUsername[username];
//         const displayName = displayUsername(user, teammateNameDisplay, false);
//         return displayName ? createMentionTag(username, displayName) : match;
//     });
// };

const getUserDisplayName = (username: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    const user = usersByUsername[username];

    if (user) {
        const displayName = displayUsername(user, teammateNameDisplay, false);
        return displayName || username;
    }

    return username;
};

// const createMentionTag = (username: string, displayName: string): string => {
//     return `@${username}<x-name>@${displayName}</x-name>`;
// };

// export const convertToMapValue = (inputValue: string, mapValue: string): string => {
//     if (!mapValue) {
//         return inputValue;
//     }

//     const mentionMappings = extractMentionMappings(mapValue);

//     if (mentionMappings.length === 0) {
//         return inputValue;
//     }

//     return applyMentionMappings(inputValue, mentionMappings);
// };

// const extractMentionMappings = (mapValue: string): Array<{ fullMatch: string; username: string; displayName: string }> => {
//     const mappings: Array<{ fullMatch: string; username: string; displayName: string }> = [];
//     const regex = new RegExp(MENTION_REGEX.source, 'g');
//     let match;

//     while ((match = regex.exec(mapValue)) !== null) {
//         mappings.push({
//             fullMatch: match[0],
//             username: match[1],
//             displayName: match[2],
//         });
//     }

//     return mappings;
// };

// const applyMentionMappings = (inputValue: string, mappings: Array<{ username: string; displayName: string }>): string => {
//     let result = inputValue;
//     const replacedPositions = new Set<number>();

//     for (const mapping of mappings) {
//         const displayNamePattern = `@${mapping.displayName}`;
//         const replacement = createMentionTag(mapping.username, mapping.displayName);

//         result = replaceFirstUnprocessed(result, displayNamePattern, replacement, replacedPositions);
//     }

//     return result;
// };

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

// export const generateMapValue = (username: string, displayName: string, mapValue: string, inputValue: string): string => {
//     console.log('generateMapValue called with:', {username, displayName, mapValue, inputValue});
//     let result = inputValue.replace(`@${username}`, createMentionTag(username, displayName));

//     const mentionMappings = extractMentionMappings(mapValue);
    
//     const replacedPositions = new Set<number>();
    
//     for (const mapping of mentionMappings) {
//         const displayNamePattern = `@${mapping.displayName}`;
//         const replacement = createMentionTag(mapping.username, mapping.displayName);
        
//         result = replaceFirstUnprocessed(result, displayNamePattern, replacement, replacedPositions);
//     }

//     console.log('Final result after generateMapValue:', result);
    
//     return result;
// };

export const extractMentionMapMappings = (mapValue: string): Array<{ fullMatch: string; username: string; }> => {
    const mappings: Array<{ fullMatch: string; username: string; }> = [];
    // mapValue = @user-123<x-name>@田中 太郎</x-name> こんにちは
    // の場合は、{fullMatch: '@user-123<x-name>@田中 太郎</x-name>', username: 'user-123'}
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

export const convertRawValue = (mapValue: string, inputValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    const mentionMappings = extractMentionMapMappings(mapValue);
    console.log('convertRawValue called with:', 'mentionMappings', mentionMappings, 'mapValue', mapValue);

    let result = inputValue
    const replacedPositions = new Set<number>();

    for (const mapping of mentionMappings) {
        const user = usersByUsername[mapping.username];
        const displayName = displayUsername(user, teammateNameDisplay, false);
        const replacement = mapping.username

        // userが存在しない場合はなにもしない
        if (!user) {
            continue;
        }

        // メンションの前後にスペースまたは日本語文字があるかチェック
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

    console.log('convertRawValue result', result)

    return result;
};

export const generateRawValue = (rawValue: string, inputValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME): string => {
    const mentionMappings = extractMentionRawMappings(rawValue);
    console.log('generateRawValue called with:', 'mentionMappings', mentionMappings, 'mapValue', rawValue);

    let result = inputValue
    const replacedPositions = new Set<number>();

    for (const mapping of mentionMappings) {
        const user = usersByUsername[mapping.username];
        const displayName = displayUsername(user, teammateNameDisplay, false);
        const replacement = mapping.username

        // userが存在しない場合はなにもしない
        if (!user) {
            continue;
        }

        result = replaceFirstUnprocessed(result, displayName, replacement, replacedPositions);
    }

    console.log('Final result after generateRawValue:', result);

    return result;
};

export const generateMapValue = (rawValue: string, usersByUsername: Record<string, UserProfile> = {}, teammateNameDisplay = Preferences.DISPLAY_PREFER_USERNAME ): string => {
    // 1. rawValue = @user-123 Hi, How are you. @user-456
    // 2. result = @user-123<x-name>@田中 太郎</x-name> Hi, How are you. @user-456<x-name>@田中 次郎</x-name>
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
