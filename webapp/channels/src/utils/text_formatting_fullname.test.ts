// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import EmojiMap from 'utils/emoji_map';
import * as TextFormatting from 'utils/text_formatting';

const emptyEmojiMap = new EmojiMap(new Map());

describe('TextFormatting.FullnameMentions', () => {
    describe('Fullname mentions in autolinkAtMentions', () => {
        test('should handle basic fullname mention', () => {
            const tokens = new Map();
            const result = TextFormatting.autolinkAtMentions('@john smith', tokens);

            console.log('Input: @john smith');
            console.log('Output:', result);
            console.log('Tokens:', Array.from(tokens.entries()));

            expect(result).toContain('$MM_ATMENTION');
            expect(tokens.size).toBeGreaterThan(0);
        });

        test('should handle multiple fullname mentions', () => {
            const tokens = new Map();
            const result = TextFormatting.autolinkAtMentions('@john smith and @jane doe', tokens);

            console.log('Input: @john smith and @jane doe');
            console.log('Output:', result);
            console.log('Tokens:', Array.from(tokens.entries()));

            expect(result).toContain('$MM_ATMENTION');
            expect(tokens.size).toBeGreaterThan(1);
        });

        test('should handle fullname mention with punctuation', () => {
            const tokens = new Map();
            const result = TextFormatting.autolinkAtMentions('@john smith.', tokens);

            console.log('Input: @john smith.');
            console.log('Output:', result);
            console.log('Tokens:', Array.from(tokens.entries()));

            expect(result).toContain('$MM_ATMENTION');
        });

        test('should handle fullname mention in context', () => {
            const tokens = new Map();
            const result = TextFormatting.autolinkAtMentions('Hello @john smith, how are you?', tokens);

            console.log('Input: Hello @john smith, how are you?');
            console.log('Output:', result);
            console.log('Tokens:', Array.from(tokens.entries()));

            expect(result).toContain('$MM_ATMENTION');
        });
    });

    describe('Fullname mentions in formatText', () => {
        test('should convert fullname mention to HTML span', () => {
            const result = TextFormatting.formatText('@john smith', {atMentions: true}, emptyEmojiMap);

            console.log('formatText Input: @john smith');
            console.log('formatText Output:', result);

            expect(result).toContain('<span data-mention="john smith">@john smith</span>');
        });

        test('should handle mixed single and fullname mentions', () => {
            // Use a different pattern that doesn't create conflicts
            const result = TextFormatting.formatText('@user1, @john smith', {atMentions: true}, emptyEmojiMap);

            console.log('formatText Input: @user1, @john smith');
            console.log('formatText Output:', result);

            expect(result).toContain('<span data-mention="user1">@user1</span>');
            expect(result).toContain('<span data-mention="john smith">@john smith</span>');
        });
    });
});
