// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// SendNotificationCallEnd sends a notification to mobile app users when a call ends
// This function is intended to be called from UpdatePost
func (a *App) SendNotificationCallEnd(c request.CTX, post *model.Post) *model.AppError {
	if post.Type != "custom_calls" {
		return nil
	}

	endAt, exists := post.Props["end_at"]
	if !exists || endAt == nil {
		return nil
	}

	channel, err := a.GetChannel(c, post.ChannelId)
	if err != nil {
		c.Logger().Error("Failed to get channel for call notification",
			mlog.String("channel_id", post.ChannelId), mlog.Err(err))
		return err
	}

	if channel.Type != model.ChannelTypeDirect && channel.Type != model.ChannelTypeGroup {
		return nil
	}

	channelMembers, err := a.GetChannelMembersPage(c, post.ChannelId, 0, model.ChannelGroupMaxUsers)
	if err != nil {
		c.Logger().Error("Failed to get channel members for call notification",
			mlog.String("channel_id", post.ChannelId), mlog.Err(err))
		return err
	}

	postUserId := post.UserId
	if postUserId == "" {
		c.Logger().Error("Post user ID is empty for call notification",
			mlog.String("post_id", post.Id))
		return nil
	}

	notification := &model.PushNotification{
		Version:     model.PushMessageV2,
		Type:        model.PushTypeMessage,
		SubType:     model.PushSubTypeCalls,
		TeamId:      channel.TeamId,
		ChannelId:   post.ChannelId,
		PostId:      post.Id,
		Message:     "Call ended",
		ChannelName: channel.DisplayName,
	}

	for _, member := range channelMembers {
		// Don't send notification to the user who created the post because they started the call
		if member.UserId == postUserId {
			continue
		}

		sessions, appErr := a.getMobileAppSessions(member.UserId)
		if appErr != nil {
			c.Logger().Debug("Failed to get mobile sessions for user",
				mlog.String("user_id", member.UserId), mlog.Err(appErr))
			continue
		}

		if len(sessions) == 0 {
			continue
		}

		for _, session := range sessions {
			if session.IsExpired() {
				continue
			}

			// Do not send notifications to devices that do not ring
			if session.Props["os"] == "iOS" && session.VoipDeviceId == "" {
				continue
			}

			tmpMessage := notification.DeepCopy()
			deviceID := session.DeviceId
			tmpMessage.SetDeviceIdAndPlatform(deviceID)
			tmpMessage.AckId = model.NewId()

			if err := a.sendToPushProxy(tmpMessage, session); err != nil {
				c.Logger().Error("Failed to send call end notification to session",
					mlog.String("user_id", member.UserId),
					mlog.String("session_id", session.Id),
					mlog.String("device_id", deviceID),
					mlog.Err(err))
			}
		}
	}
	return nil
}
