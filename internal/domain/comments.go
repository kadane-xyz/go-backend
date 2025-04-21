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
	SolutionId int64  `json:"solutionId"`
	Body       string `json:"body"`
	ParentId   *int64 `json:"parentId,omitempty"`
}

type CommentUpdateRequest struct {
	Body string `json:"body"`
}

func FromSQLComment(row sql.Comment) (*Comment, error) {
	return &Comment{
		ID:         row.ID,
		SolutionId: row.SolutionID,
		Body:       row.Body,
		CreatedAt:  row.CreatedAt.Time,
		Votes:      row.Votes.Int32,
		ParentId:   row.ParentID.Int64,
		Children:   []*Comment{},
	}, nil
}

func FromSQLGetCommentRow(row sql.GetCommentRow) (*CommentRelation, error) {
	comment := CommentRelation{
		Comment: Comment{
			ID:         row.ID,
			SolutionId: row.SolutionID,
			Body:       row.Body,
			CreatedAt:  row.CreatedAt.Time,
			Children:   []*Comment{},
		},
		Username:        row.UserUsername,
		Level:           row.UserLevel,
		CurrentUserVote: row.VotesCount,
	}

	if row.ParentID != nil {
		comment.ParentId = row.ParentID
	}

	if row.Votes != nil {
		comment.Votes = *row.Votes
	}
	if row.UserAvatarUrl != nil {
		comment.AvatarUrl = *row.UserAvatarUrl
	}

	return &comment, nil
}

func FromSQLGetCommentsRow(row []sql.GetCommentsRow) ([]*CommentRelation, error) {
	var comments []*CommentRelation
	for i, comment := range row {
		domainComment, err := FromSQLGetCommentRow(sql.GetCommentRow(comment))
		if err != nil {
			return nil, err
		}
		comments[i] = domainComment
	}
}

// Reuse function FromSQLGetCommentsRow because type matches
func FromSQLGetCommentsSorted(row []sql.GetCommentsSortedRow) ([]*CommentRelation, error) {
	comments := []sql.GetCommentsRow{}
	for i, comment := range row {
		comments[i] = sql.GetCommentsRow(comment)
	}
	return FromSQLGetCommentsRow(comments)
}
