// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	sq "github.com/mattermost/squirrel"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/v8/channels/store"
)

type SqlEngageChatStore struct {
	*SqlStore
}

func newSqlEngageChatStore(sqlStore *SqlStore) store.EngageChatStore {
	s := &SqlEngageChatStore{
		SqlStore: sqlStore,
	}
	return s
}

func (s *SqlEngageChatStore) HasDMGMChannelMemberWithEngageAdmin(channelID string) (bool, error) {
	subQuery := s.getQueryBuilder().Select("1").
		From("ChannelMembers cm").
		Join("Channels c ON cm.ChannelId = c.Id").
		Join("Users u ON cm.UserId = u.Id AND u.DeleteAt = 0").
		Where(sq.Eq{"cm.ChannelId": channelID}).
		Where(sq.Eq{"c.Type": []string{string(model.ChannelTypeDirect), string(model.ChannelTypeGroup)}}).
		Where(sq.Expr("CONCAT(' ', u.Roles, ' ') LIKE ?", "% "+model.SystemEngageAdmin+" %")).
		Limit(1)

	subQueryString, args, err := subQuery.ToSql()
	if err != nil {
		return false, errors.Wrap(err, "failed to build query")
	}

	existsQuery := "SELECT EXISTS(" + subQueryString + ")"

	var exists bool
	err = s.GetReplica().Get(&exists, existsQuery, args...)
	if err != nil {
		return false, errors.Wrap(err, "failed to query for channel member with role")
	}

	return exists, nil
}

// HasDMChannelBotMember checks if a Direct Message channel has any bot as a member.
func (s *SqlEngageChatStore) HasDMChannelBotMember(channelID string) (bool, error) {
	subQuery := s.getQueryBuilder().Select("1").
		From("ChannelMembers cm").
		Join("Channels c ON cm.ChannelId = c.Id").
		Join("Bots b ON cm.UserId = b.UserId AND b.DeleteAt = 0").
		Where(sq.Eq{"cm.ChannelId": channelID}).
		Where(sq.Eq{"c.Type": string(model.ChannelTypeDirect)}).
		Limit(1)

	subQueryString, args, err := subQuery.ToSql()
	if err != nil {
		return false, errors.Wrap(err, "failed to build query")
	}

	existsQuery := "SELECT EXISTS(" + subQueryString + ")"

	var exists bool
	err = s.GetReplica().Get(&exists, existsQuery, args...)
	if err != nil {
		return false, errors.Wrap(err, "failed to query for bot channel member")
	}

	return exists, nil
}
