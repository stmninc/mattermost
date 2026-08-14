// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"github.com/golang-jwt/jwt/v5"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/mattermost/mattermost/server/public/shared/request"
)

const (
	PushSubTypeCallsEnded = "calls_ended"

	// constants originating from the calls plugin
	PostTypeCallsPluginCustomCall = "custom_calls"
	PostPropsCallsPluginEndAt     = "end_at"
)

// SendNotificationCallEnd sends a notification to mobile app users when a call ends
// This function is intended to be called from UpdatePost
func (a *App) SendNotificationCallEnd(rctx request.CTX, post *model.Post) *model.AppError {
	if post.Type != PostTypeCallsPluginCustomCall {
		return nil
	}

	endAt, exists := post.Props[PostPropsCallsPluginEndAt]
	if !exists || endAt == nil {
		return nil
	}

	channel, err := a.GetChannel(rctx, post.ChannelId)
	if err != nil {
		rctx.Logger().Error("Failed to get channel for call notification",
			mlog.String("channel_id", post.ChannelId), mlog.Err(err))
		return err
	}

	if !channel.IsGroupOrDirect() {
		return nil
	}

	channelMembers, err := a.GetChannelMembersPage(rctx, post.ChannelId, 0, model.ChannelGroupMaxUsers)
	if err != nil {
		rctx.Logger().Error("Failed to get channel members for call notification",
			mlog.String("channel_id", post.ChannelId), mlog.Err(err))
		return err
	}

	postUserId := post.UserId
	if postUserId == "" {
		rctx.Logger().Error("Post user ID is empty for call notification",
			mlog.String("post_id", post.Id))
		return nil
	}

	notification := &model.PushNotification{
		Version:     model.PushMessageV2,
		Type:        model.PushTypeClear,
		SubType:     PushSubTypeCallsEnded,
		TeamId:      channel.TeamId,
		ChannelId:   post.ChannelId,
		PostId:      post.Id,
		ChannelName: channel.DisplayName,
	}

	for _, member := range channelMembers {
		// Don't send notification to the user who created the post because they started the call
		if member.UserId == postUserId {
			continue
		}

		sessions, appErr := a.getMobileAppSessions(member.UserId)
		if appErr != nil {
			rctx.Logger().Debug("Failed to get mobile sessions for user",
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
			if session.Props[model.SessionPropOs] == "iOS" && session.VoipDeviceId == "" {
				continue
			}

			tmpMessage := notification.DeepCopy()
			deviceID := session.DeviceId
			tmpMessage.SetDeviceIdAndPlatform(deviceID)
			tmpMessage.AckId = model.NewId()
			signature, err := jwt.NewWithClaims(jwt.SigningMethodES256, pushJWTClaims{
				AckId:    tmpMessage.AckId,
				DeviceId: tmpMessage.DeviceId,
			}).SignedString(a.AsymmetricSigningKey())
			if err != nil {
				a.Log().Error("Notification error",
					mlog.String("ackId", tmpMessage.AckId),
					mlog.String("type", tmpMessage.Type),
					mlog.String("userId", session.UserId),
					mlog.String("postId", tmpMessage.PostId),
					mlog.String("channelId", tmpMessage.ChannelId),
					mlog.String("deviceId", tmpMessage.DeviceId),
					mlog.String("status", err.Error()),
				)
				continue
			}
			tmpMessage.Signature = signature

			if err := a.sendToPushProxy(rctx, tmpMessage, session); err != nil {
				rctx.Logger().Error("Failed to send call end notification to session",
					mlog.String("user_id", member.UserId),
					mlog.String("session_id", session.Id),
					mlog.String("device_id", deviceID),
					mlog.Err(err))
			}
		}
	}
	return nil
}
