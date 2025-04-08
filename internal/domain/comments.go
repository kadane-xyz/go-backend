package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Comment struct {
	ID         int64      `json:"id"`
	SolutionId int64      `json:"solutionId"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	Votes      int32      `json:"votes"`
	ParentId   int64      `json:"parentId,omitempty"`
	Children   []*Comment `json:"children,omitempty"` // For nested child comments
}

type CommentCreateRequest struct {
	SolutionId int64  `json:"solutionId"`
	Body       string `json:"body"`
	ParentId   *int64 `json:"parentId,omitempty"`
}

type CommentUpdateRequest struct {
	Body string `json:"body"`
}

type CommentsData struct {
	ID              int64           `json:"id"`
	SolutionId      int64           `json:"solutionId"`
	Username        string          `json:"username"`
	AvatarUrl       string          `json:"avatarUrl,omitempty"`
	Level           int32           `json:"level"`
	Body            string          `json:"body"`
	CreatedAt       time.Time       `json:"createdAt"`
	Votes           int32           `json:"votes"`
	ParentId        *int64          `json:"parentId,omitempty"`
	Children        []*CommentsData `json:"children"` // For nested child comments
	CurrentUserVote sql.VoteType    `json:"currentUserVote"`
}

type CommentsResponse struct {
	Data []*CommentsData `json:"data"`
}
