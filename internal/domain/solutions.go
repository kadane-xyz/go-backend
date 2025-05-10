package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Solution struct {
	Id        int32     `json:"id"`
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	Tags      []string  `json:"tags"`
	Body      string    `json:"body"`
	Votes     int32     `json:"votes"`
	ProblemId int32     `json:"problemId"`
}

// GetSolution
type SolutionRelations struct {
	Solution
	Username        string       `json:"username"`
	AvatarUrl       string       `json:"avatarUrl"`
	Level           int32        `json:"level"`
	CommentCount    int32        `json:"commentsCount"`
	VotesCount      int32        `json:"votesCount"`
	CurrentUserVote sql.VoteType `json:"currentUserVote"`
	Starred         bool         `json:"starred"`
}

// GetSolutions
type SolutionsRelations struct {
	SolutionRelations
	TotalCount int32 `json:"totalCount"`
}

type CreateSolutionRequest struct {
	ProblemId int32    `json:"problemId"`
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
	Id     int32  `json:"id"`
	UserId string `json:"userId"`
}

type SolutionsGetParams struct {
	ProblemId     int32
	Tags          []string
	Title         string
	UserId        string
	Page          int32
	PerPage       int32
	Sort          string
	SortDirection sql.SortDirection
}

type SolutionsCreateParams struct {
	UserId    string
	Title     string
	Tags      []string
	Body      string
	ProblemId *int32
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

func FromSQLGetSolutionRow(row sql.GetSolutionRow) *SolutionRelations {
	return &SolutionRelations{
		Solution: Solution{
			Id:        row.Solution.ID,
			Title:     row.Solution.Title,
			Date:      row.Solution.CreatedAt.Time,
			Tags:      row.Solution.Tags,
			Body:      row.Solution.Body,
			Votes:     nullHandler(row.Solution.Votes),
			ProblemId: nullHandler(row.Solution.ProblemID),
		},
		Username:        row.UserUsername,
		AvatarUrl:       nullHandler(row.UserAvatarUrl),
		Level:           row.UserLevel,
		CommentCount:    row.CommentsCount,
		VotesCount:      row.VotesCount,
		CurrentUserVote: row.UserVote,
		Starred:         row.Starred,
	}
}

func FromSQLGetSolutionsRow(rows []sql.GetSolutionsRow) []*SolutionsRelations {
	solutions := []*SolutionsRelations{}

	for i, row := range rows {
		solutions[i] = &SolutionsRelations{
			SolutionRelations: SolutionRelations{
				Solution: Solution{
					Id:        row.Solution.ID,
					Title:     row.Solution.Title,
					Date:      row.Solution.CreatedAt.Time,
					Tags:      row.Solution.Tags,
					Body:      row.Solution.Body,
					Votes:     nullHandler(row.Solution.Votes),
					ProblemId: nullHandler(row.Solution.ProblemID),
				},
				Username:        row.UserUsername,
				AvatarUrl:       nullHandler(row.UserAvatarUrl),
				Level:           row.UserLevel,
				CommentCount:    row.CommentsCount,
				VotesCount:      row.VotesCount,
				CurrentUserVote: row.UserVote,
				Starred:         row.Starred,
			},
			TotalCount: row.TotalCount,
		}
	}

	return solutions
}
