package domain

import (
	"github.com/jackc/pgx/v5/pgtype"
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

type SolutionResponse struct {
	Data SolutionsData `json:"data"`
}

type SolutionsData struct {
	Id              int64            `json:"id"`
	Body            string           `json:"body,omitempty"`
	Comments        int32            `json:"comments"`
	Date            pgtype.Timestamp `json:"date"`
	Tags            []string         `json:"tags"`
	Title           string           `json:"title"`
	Username        string           `json:"username,omitempty"`
	Level           int32            `json:"level,omitempty"`
	AvatarUrl       string           `json:"avatarUrl,omitempty"`
	Votes           int32            `json:"votes"`
	CurrentUserVote sql.VoteType     `json:"currentUserVote"`
	Starred         bool             `json:"starred"`
}

type SolutionsResponse struct {
	Data       []SolutionsData `json:"data"`
	Pagination Pagination      `json:"pagination"`
}
