package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Friend struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	AvatarUrl  string    `json:"avatarUrl"`
	Location   string    `json:"location"`
	Level      int32     `json:"level"`
	AcceptedAt time.Time `json:"acceptedAt"`
}

type FriendRequestRequest struct {
	FriendName string `json:"friendName"`
}

type FriendRequest struct {
	FriendID   string    `json:"friendId"`
	FriendName string    `json:"friendName"`
	AvatarUrl  string    `json:"avatarUrl"`
	Level      int32     `json:"level"`
	CreatedAt  time.Time `json:"createdAt"`
	Location   string    `json:"location"`
}

type FriendshipStatus = sql.FriendshipStatus

type FriendParams struct {
	FriendName string `json:"friendName"`
	UserID     string `json:"userId"`
}

type FriendRequesStatusParams = FriendParams
type FriendRequestCreateParams = FriendParams
type FriendRequestAcceptParams = FriendParams
type FriendBlockParams = FriendParams
type FriendUnblockParams = FriendParams
type FriendshipDeleteParams = FriendParams

func FromSQLFromSQLGetFriendsRow(rows []sql.GetFriendsRow) []*Friend {
	friends := make([]*Friend, len(rows))
	for i, row := range rows {
		friends[i] = &Friend{
			ID:         row.FriendID,
			Username:   row.FriendUsername,
			AvatarUrl:  row.AvatarUrl,
			Level:      row.Level,
			Location:   row.Location,
			AcceptedAt: row.AcceptedAt.Time,
		}
	}

	return friends
}

func FromSQLFriendsByUsernameRows(rows []sql.GetFriendsByUsernameRow) []*Friend {
	friends := make([]*Friend, len(rows))
	for _, row := range rows {
		friends = append(friends, &Friend{
			ID:         row.FriendID,
			Username:   row.FriendUsername,
			AvatarUrl:  row.AvatarUrl,
			Level:      row.Level,
			Location:   row.Location,
			AcceptedAt: row.AcceptedAt.Time,
		})
	}
	return friends
}

// sql.GetFriendRequestRow shared by Get FriendRequest routes
func FromSQLGetFriendRequestRow(row sql.GetFriendRequestsReceivedRow) *FriendRequest {
	return &FriendRequest{
		FriendID:   row.FriendID,
		FriendName: row.FriendUsername,
		AvatarUrl:  row.AvatarUrl,
		Level:      row.Level,
		Location:   row.Location,
		CreatedAt:  row.CreatedAt.Time,
	}
}

func FromSQLGetFriendRequestsSentRows(rows []sql.GetFriendRequestsSentRow) []*FriendRequest {
	friends := []*FriendRequest{}
	for i, row := range rows {
		friends[i] = FromSQLGetFriendRequestRow(sql.GetFriendRequestsReceivedRow(row))
	}
	return friends
}

func FromSQLGetFriendRequestsReceivedRows(rows []sql.GetFriendRequestsReceivedRow) []*FriendRequest {
	friends := []*FriendRequest{}
	for i, row := range rows {
		friends[i] = FromSQLGetFriendRequestRow(row)
	}
	return friends
}
