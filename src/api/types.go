package api

import "kadane.xyz/go-backend/v2/src/sql/sql"

type Pagination struct {
	Page      int64 `json:"page"`
	PerPage   int64 `json:"perPage"`
	DataCount int64 `json:"dataCount"`
	LastPage  int64 `json:"lastPage"`
}

type TestCase struct {
	Description string         `json:"description"`
	Input       string         `json:"input"`
	Output      string         `json:"output"`
	Visibility  sql.Visibility `json:"visibility"`
}

type FriendshipStatus string

const (
	FriendshipStatusNone            FriendshipStatus = "none"
	FriendshipStatusFriend          FriendshipStatus = "friend"
	FriendshipStatusBlocked         FriendshipStatus = "blocked"
	FriendshipStatusRequestSent     FriendshipStatus = "request_sent"
	FriendshipStatusRequestReceived FriendshipStatus = "request_received"
)
