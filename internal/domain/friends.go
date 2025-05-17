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

func FromSQLFriendsByUsernameRows(rows []sql.GetFriendsByUsernameRow) []*Friend {
	friends := []*Friend{}
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

func FromSQLGetFriendRequestsReceivedRows(rows []sql.GetFriendRequestsReceivedRow) []*FriendRequest {
	friends := []*FriendRequest{}
	for i, row := range rows {
		friends[i] = &FriendRequest{
			FriendID:   row.FriendID,
			FriendName: row.FriendUsername,
			AvatarUrl:  row.AvatarUrl,
			Level:      row.Level,
			CreatedAt:  row.CreatedAt.Time,
			Location:   row.Location,
		}
	}
	return friends
}
