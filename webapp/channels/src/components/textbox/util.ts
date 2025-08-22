import type {UserProfile} from '@mattermost/types/users';
import {Preferences} from 'mattermost-redux/constants';
import {displayUsername} from 'mattermost-redux/utils/user_utils';

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

const createMentionTag = (username: string, displayName: string): string => {
    return `@${username}<x-name>@${displayName}</x-name>`;
};
