package domain

import (
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Solution struct {
	Id        int64    `json:"id"`
	Username  string   `json:"username"`
	Title     string   `json:"title"`
	Date      string   `json:"date"`
	Tags      []string `json:"tags"`
	Body      string   `json:"body"`
	Votes     int      `json:"votes"`
	ProblemId int64    `json:"problemId"`
}

type SolutionRelations struct {
	Solution
	CommentCount    int32        `json:"commentsCount"`
	VotesCount      int32        `json:"votesCount"`
	CurrentUserVote sql.VoteType `json:"currentUserVote"`
	Starred         bool         `json:"starred"`
}

type CreateSolutionRequest struct {
	ProblemId int64    `json:"problemId"`
	Title     string   `json:"title"`
	Tags      []string `json:"tags"`
	Body      string   `json:"body"`
}

type UpdateSolutionRequest struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}

type SolutionsGetParams struct {
	ProblemId     int32
	Tags          []string
	Title         string
	PerPage       int32
	Page          int32
	Sort          string
	SortDirection sql.SortDirection
	UserId        string
}

type SolutionsUpdateParams struct {
	ID     int32    `json:"id"`
	UserID string   `json:"userId"`
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Tags   []string `json:"tags"`
}

type VoteSolutionsParams struct {
	UserId     string
	SolutionId int32
	Vote       sql.VoteType
}

func FromSQLGetSolutionsRow(rows []sql.GetSolutionRow) ([]*SolutionRelations, error) {
	solutions := []*SolutionRelations{}

	for _, row := range rows {
		solution := SolutionRelation{}
	}
}
