package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Comment struct {
	ID         int64      `json:"id"`
	SolutionId int32      `json:"solutionId"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	Votes      int32      `json:"votes"`
	ParentId   *int64     `json:"parentId,omitempty"`
	Children   []*Comment `json:"children,omitempty"` // For nested child comments
}

type CommentRelation struct {
	Comment
	Children        []*CommentRelation `json:"children,omitempty"` // For nested child comments
	Username        string             `json:"username"`
	AvatarUrl       string             `json:"avatarUrl"`
	Level           int32              `json:"level"`
	CurrentUserVote int32              `json:"currentUserVote"`
}

type CommentCreateRequest struct {
	SolutionId int32  `json:"solutionId"`
	Body       string `json:"body"`
	ParentId   *int64 `json:"parentId,omitempty"`
}

type CommentCreateParams struct {
	UserID     string `json:"userId"`
	Body       string `json:"body"`
	SolutionID int32  `json:"solutionId"`
	ParentID   *int64 `json:"parentId"`
}

type CommentUpdateRequest struct {
	Body string `json:"body"`
}

func FromSQLComment(row sql.Comment) *Comment {
	comment := Comment{
		ID:         row.ID,
		SolutionId: row.SolutionID,
		Body:       row.Body,
		CreatedAt:  row.CreatedAt.Time,
		Children:   []*Comment{},
		ParentId:   nullHandler(&row.ParentID),
		Votes:      nullHandler(row.Votes),
	}
	return &comment
}

func FromSQLGetCommentRow(row sql.GetCommentRow) (*CommentRelation, error) {
	comment := CommentRelation{
		Comment: Comment{
			ID:         row.ID,
			SolutionId: row.SolutionID,
			Body:       row.Body,
			CreatedAt:  row.CreatedAt.Time,
			Children:   []*Comment{},
			ParentId:   nullHandler(&row.ParentID),
			Votes:      nullHandler(row.Votes),
		},
		AvatarUrl:       nullHandler(row.UserAvatarUrl),
		Username:        row.UserUsername,
		Level:           row.UserLevel,
		CurrentUserVote: row.VotesCount,
	}

	return &comment, nil
}

func FromSQLGetCommentsRow(row []sql.GetCommentsRow) []*CommentRelation {
	comments := []*CommentRelation{}
	for i, comment := range row {
		domainComment, err := FromSQLGetCommentRow(sql.GetCommentRow(comment))
		if err != nil {
			return nil
		}

		if domainComment != nil {
			comments[i] = domainComment
		}
	}

	return comments
}

// Reuse function FromSQLGetCommentsRow because type matches
func FromSQLGetCommentsSorted(row []sql.GetCommentsSortedRow) []*CommentRelation {
	comments := []sql.GetCommentsRow{}
	for i, comment := range row {
		comments[i] = sql.GetCommentsRow(comment)
	}
	return FromSQLGetCommentsRow(comments)
}
