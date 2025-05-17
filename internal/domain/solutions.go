package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Solution struct {
	ID              int32        `json:"id"`
	Title           string       `json:"title"`
	Date            time.Time    `json:"date"`
	Tags            []string     `json:"tags"`
	Body            string       `json:"body"`
	Votes           int32        `json:"votes"`
	ProblemID       int32        `json:"problemId"`
	Username        string       `json:"username"`
	AvatarUrl       string       `json:"avatarUrl"`
	Level           int32        `json:"level"`
	CommentCount    int32        `json:"commentsCount"`
	VotesCount      int32        `json:"votesCount"`
	CurrentUserVote sql.VoteType `json:"currentUserVote"`
	Starred         bool         `json:"starred"`
	TotalCount      int32        `json:"totalCount"`
}

type CreateSolutionRequest struct {
	ProblemID int32    `json:"problemId"`
	Title     string   `json:"title"`
	Tags      []string `json:"tags"`
	Body      string   `json:"body"`
}

type UpdateSolutionRequest struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}

type SolutionGetParams struct {
	ID     int32  `json:"id"`
	UserID string `json:"userId"`
}

type SolutionsGetParams struct {
	ProblemID     int32
	Tags          []string
	Title         string
	UserId        string
	Page          int32
	PerPage       int32
	Sort          string
	SortDirection sql.SortDirection
}

type SolutionsCreateParams struct {
	UserID    string
	Title     string
	Tags      []string
	Body      string
	ProblemID *int32
}

type SolutionsUpdateParams struct {
	SolutionID *int32
	UserID     string
	Title      string
	Tags       []string
	Body       string
}

type VoteSolutionsParams struct {
	UserId     string
	SolutionId int32
	Vote       sql.VoteType
}

func FromSQLGetSolutionRow(row sql.GetSolutionRow) *Solution {
	return &Solution{
		ID:              row.Solution.ID,
		Title:           row.Solution.Title,
		Date:            row.Solution.CreatedAt.Time,
		Tags:            row.Solution.Tags,
		Body:            row.Solution.Body,
		Votes:           nullHandler(row.Solution.Votes),
		ProblemID:       nullHandler(row.Solution.ProblemID),
		Username:        row.UserUsername,
		AvatarUrl:       nullHandler(row.UserAvatarUrl),
		Level:           row.UserLevel,
		CommentCount:    row.CommentsCount,
		VotesCount:      row.VotesCount,
		CurrentUserVote: row.UserVote,
		Starred:         row.Starred,
	}
}

func FromSQLGetSolutionsRow(rows []sql.GetSolutionsRow) []*Solution {
	solutions := []*Solution{}

	for i, row := range rows {
		solutions[i] = &Solution{
			ID:              row.Solution.ID,
			Title:           row.Solution.Title,
			Date:            row.Solution.CreatedAt.Time,
			Tags:            row.Solution.Tags,
			Body:            row.Solution.Body,
			Votes:           nullHandler(row.Solution.Votes),
			ProblemID:       nullHandler(row.Solution.ProblemID),
			Username:        row.UserUsername,
			AvatarUrl:       nullHandler(row.UserAvatarUrl),
			Level:           row.UserLevel,
			CommentCount:    row.CommentsCount,
			VotesCount:      row.VotesCount,
			CurrentUserVote: row.UserVote,
			Starred:         row.Starred,
			TotalCount:      row.TotalCount,
		}
	}

	return solutions
}
