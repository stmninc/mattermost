package app

import (
	"encoding/json"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

// SendNotificationCallEnd sends a notification to mobile app users when a call ends
// This function is intended to be called from UpdatePost
func (a *App) SendNotificationCallEnd(c request.CTX, post *model.Post) *model.AppError {
	c.Logger().Info("SendNotificationCallEnd called",
		mlog.String("post_id", post.Id),
		mlog.String("post_type", post.Type),
		mlog.String("post_props", model.StringInterfaceToJSON(post.Props)))

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

	currentUserId := c.Session().UserId
	if currentUserId == "" {
		c.Logger().Error("No current user for call notification",
			mlog.String("channel_id", post.ChannelId))
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
		SenderId:   currentUserId,
	}

	for _, member := range channelMembers {
		// Don't send notification to the user who ended the call
		if member.UserId == currentUserId {
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

			tmpMessage := notification.DeepCopy()
			deviceID := session.DeviceId
			if notification.SubType == model.PushSubTypeCalls && session.VoipDeviceId != "" {
				deviceID = session.VoipDeviceId
			}
			tmpMessage.SetDeviceIdAndPlatform(deviceID)

			c.Logger().Info("Preparing to send call end notification",
				mlog.String("user_id", member.UserId),
				mlog.String("session_id", session.Id),
				mlog.String("device_id", deviceID),
				mlog.String("platform", tmpMessage.Platform))

      // Don't send notification if platform is iOS React Native because call notifications are not supported
			if session.VoipDeviceId == "" && tmpMessage.Platform == model.PushNotifyAppleReactNative {
				continue
			}

			tmpMessage.AckId = model.NewId()

			notificationJSON, _ := json.Marshal(tmpMessage)
			c.Logger().Debug("Sending call end notification to session",
				mlog.String("user_id", member.UserId),
				mlog.String("session_id", session.Id),
				mlog.String("device_id", deviceID),
				mlog.String("notification", string(notificationJSON)))

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
