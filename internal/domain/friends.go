package domain

import "time"

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
