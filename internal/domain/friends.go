package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/database/sql"
)

type Friend struct {
	Id         string    `json:"id"`
	Username   string    `json:"username"`
	AvatarUrl  string    `json:"avatarUrl"`
	Location   string    `json:"location"`
	Level      int32     `json:"level"`
	AcceptedAt time.Time `json:"acceptedAt"`
}

type FriendRequestRequest struct {
	FriendName string `json:"friendName"`
}

type FriendsResponse struct {
	Data []Friend `json:"data"`
}

type FriendRequestsResponse struct {
	Data []FriendRequest `json:"data"`
}

type FriendRequest struct {
	FriendId   string    `json:"friendId"`
	FriendName string    `json:"friendName"`
	AvatarUrl  string    `json:"avatarUrl"`
	Level      int32     `json:"level"`
	CreatedAt  time.Time `json:"createdAt"`
	Location   string    `json:"location"`
}

func FromSQLFriendsByUsernameRows(rows []sql.GetFriendsByUsernameRow) []Friend {
	var friends []Friend
	for _, row := range rows {
		friends = append(friends, Friend{
			Id:         row.FriendID,
			Username:   row.FriendUsername,
			AvatarUrl:  row.AvatarUrl,
			Level:      row.Level,
			Location:   row.Location,
			AcceptedAt: row.AcceptedAt.Time,
		})
	}
	return friends
}
