// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/request"
	"github.com/mattermost/mattermost/server/v8/channels/store"
)

func TestEngageChatStore(t *testing.T, rctx request.CTX, ss store.Store, s SqlStore) {
	t.Run("HasDMGMChannelMemberWithEngageAdmin", func(t *testing.T) {
		t.Run("SystemRolesMatch", func(t *testing.T) { testHasDMGMChannelMemberWithEngageAdminMatch(t, rctx, ss) })
		t.Run("SystemRolesNoMatch", func(t *testing.T) { testHasDMGMChannelMemberWithEngageAdminNoMatch(t, rctx, ss) })
		t.Run("NonexistentChannel", func(t *testing.T) { testHasDMGMChannelMemberWithEngageAdminNonexistentChannel(t, rctx, ss) })
		t.Run("DirectChannel", func(t *testing.T) { testHasDMGMChannelMemberWithEngageAdminDirectChannel(t, rctx, ss) })
		t.Run("MultipleMembers", func(t *testing.T) { testHasDMGMChannelMemberWithEngageAdminMultipleMembers(t, rctx, ss) })
	})
}

// engageChatCreateTeam creates a team for test use.
func engageChatCreateTeam(t *testing.T, ss store.Store) *model.Team {
	t.Helper()
	team, err := ss.Team().Save(&model.Team{
		DisplayName: "Test Team",
		Name:        NewTestID(),
		Email:       MakeEmail(),
		Type:        model.TeamOpen,
	})
	require.NoError(t, err)
	return team
}

// engageChatCreateUser creates a user with the given system roles.
func engageChatCreateUser(t *testing.T, rctx request.CTX, ss store.Store, roles string) *model.User {
	t.Helper()
	user := &model.User{
		Email:    MakeEmail(),
		Username: "u_" + model.NewId(),
		Roles:    roles,
	}
	user, err := ss.User().Save(rctx, user)
	require.NoError(t, err)
	return user
}

// engageChatCreateChannel creates a channel with the given type on the given team.
func engageChatCreateChannel(t *testing.T, rctx request.CTX, ss store.Store, teamID string, channelType model.ChannelType) *model.Channel {
	t.Helper()
	channel, err := ss.Channel().Save(rctx, &model.Channel{
		TeamId:      teamID,
		DisplayName: "Test Channel",
		Name:        NewTestID(),
		Type:        channelType,
	}, -1)
	require.NoError(t, err)
	return channel
}

// engageChatAddChannelMember adds the user as a member of the channel.
func engageChatAddChannelMember(t *testing.T, rctx request.CTX, ss store.Store, channelID, userID string) {
	t.Helper()
	_, err := ss.Channel().SaveMember(rctx, &model.ChannelMember{
		ChannelId:   channelID,
		UserId:      userID,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
		SchemeUser:  true,
	})
	require.NoError(t, err)
}

// engageChatAddTeamMember adds the user as a team member with the given explicit roles.
func engageChatAddTeamMember(t *testing.T, rctx request.CTX, ss store.Store, teamID, userID, explicitRoles string) {
	t.Helper()
	_, err := ss.Team().SaveMember(rctx, &model.TeamMember{
		TeamId:        teamID,
		UserId:        userID,
		SchemeUser:    true,
		ExplicitRoles: explicitRoles,
	}, -1)
	require.NoError(t, err)
}

// --- Test cases ---

func testHasDMGMChannelMemberWithEngageAdminMatch(t *testing.T, rctx request.CTX, ss store.Store) {
	team := engageChatCreateTeam(t, ss)

	// Create a user with system_engage_admin role.
	user := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId+" "+model.SystemEngageAdmin)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user.Id)) }()

	// Create a GM channel and add the user as a member.
	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypeGroup)
	engageChatAddTeamMember(t, rctx, ss, team.Id, user.Id, "")
	engageChatAddChannelMember(t, rctx, ss, channel.Id, user.Id)

	// Should find the user with SystemEngageAdmin.
	result, err := ss.EngageChat().HasDMGMChannelMemberWithEngageAdmin(channel.Id)
	require.NoError(t, err)
	assert.True(t, result)
}

func testHasDMGMChannelMemberWithEngageAdminNoMatch(t *testing.T, rctx request.CTX, ss store.Store) {
	team := engageChatCreateTeam(t, ss)

	// Create a regular user WITHOUT system_engage_admin.
	user := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user.Id)) }()

	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypeGroup)
	engageChatAddTeamMember(t, rctx, ss, team.Id, user.Id, "")
	engageChatAddChannelMember(t, rctx, ss, channel.Id, user.Id)

	// Should NOT find any match.
	result, err := ss.EngageChat().HasDMGMChannelMemberWithEngageAdmin(channel.Id)
	require.NoError(t, err)
	assert.False(t, result)
}

func testHasDMGMChannelMemberWithEngageAdminNonexistentChannel(t *testing.T, rctx request.CTX, ss store.Store) {
	// A non-existent channel ID → should return false with no error.
	result, err := ss.EngageChat().HasDMGMChannelMemberWithEngageAdmin(model.NewId())
	require.NoError(t, err)
	assert.False(t, result)
}

func testHasDMGMChannelMemberWithEngageAdminDirectChannel(t *testing.T, rctx request.CTX, ss store.Store) {
	// Create two users — one with system_engage_admin, one without.
	user1 := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId+" "+model.SystemEngageAdmin)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user1.Id)) }()

	user2 := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, user2.Id)) }()

	// Create a DM channel between the two users.
	dm := model.Channel{
		DisplayName: "DM",
		Name:        model.GetDMNameFromIds(user1.Id, user2.Id),
		Type:        model.ChannelTypeDirect,
	}
	m1 := model.ChannelMember{UserId: user1.Id, NotifyProps: model.GetDefaultChannelNotifyProps()}
	m2 := model.ChannelMember{UserId: user2.Id, NotifyProps: model.GetDefaultChannelNotifyProps()}

	savedChannel, err := ss.Channel().SaveDirectChannel(rctx, &dm, &m1, &m2)
	require.NoError(t, err)

	// Should find user1 who has SystemEngageAdmin.
	result, hErr := ss.EngageChat().HasDMGMChannelMemberWithEngageAdmin(savedChannel.Id)
	require.NoError(t, hErr)
	assert.True(t, result)
}

func testHasDMGMChannelMemberWithEngageAdminMultipleMembers(t *testing.T, rctx request.CTX, ss store.Store) {
	// Verify that if only one of multiple members has the role, it still returns true.
	team := engageChatCreateTeam(t, ss)

	userWithRole := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId+" "+model.SystemEngageAdmin)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, userWithRole.Id)) }()

	userWithoutRole := engageChatCreateUser(t, rctx, ss, model.SystemUserRoleId)
	defer func() { require.NoError(t, ss.User().PermanentDelete(rctx, userWithoutRole.Id)) }()

	channel := engageChatCreateChannel(t, rctx, ss, team.Id, model.ChannelTypeGroup)

	engageChatAddTeamMember(t, rctx, ss, team.Id, userWithRole.Id, "")
	engageChatAddTeamMember(t, rctx, ss, team.Id, userWithoutRole.Id, "")
	engageChatAddChannelMember(t, rctx, ss, channel.Id, userWithRole.Id)
	engageChatAddChannelMember(t, rctx, ss, channel.Id, userWithoutRole.Id)

	// Should find the member with the role.
	result, err := ss.EngageChat().HasDMGMChannelMemberWithEngageAdmin(channel.Id)
	require.NoError(t, err)
	assert.True(t, result)
}
