package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Starred struct {
	ID      interface{}
	Starred bool
}

type StarredProblem struct {
	ID          int
	Title       string
	Description string
	Tags        []string
	Difficulty  string
	Points      int
	Starred     bool
}

type StarredSolution struct {
	Id        int64
	Username  string
	Title     string
	Date      pgtype.Timestamp
	Tags      []string
	Body      string
	Votes     int32
	ProblemId int64
	Starred   bool
}

type StarredSubmission struct {
	Id            pgtype.UUID
	Token         string
	Stdout        string
	Time          string
	Memory        int32
	Stderr        string
	CompileOutput string
	Message       string
	Status        sql.SubmissionStatus
	Language      string
	// custom fields
	AccountID      string
	SubmittedCode  string
	SubmittedStdin string
	ProblemID      int32
	CreatedAt      time.Time
	Starred        bool
}

type StarredRequest struct {
	ID any `json:"id"` // can be int32 or string
}
