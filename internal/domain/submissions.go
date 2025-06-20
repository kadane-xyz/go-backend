package domain

import (
	"time"

	"github.com/google/uuid"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Submission struct {
	ID            string               `json:"id"`
	Token         string               `json:"token"`
	Stdout        string               `json:"stdout"`
	Time          string               `json:"time"`
	Memory        int32                `json:"memory"`
	Stderr        string               `json:"stderr"`
	CompileOutput string               `json:"compileOutput"`
	Message       string               `json:"message"`
	Status        sql.SubmissionStatus `json:"status"`
	Language      string               `json:"language"`
	// custom fields
	AccountID       string      `json:"accountId"`
	SubmittedCode   string      `json:"submittedCode"`
	SubmittedStdin  string      `json:"submittedStdin"`
	ProblemID       int32       `json:"problemId"`
	CreatedAt       time.Time   `json:"createdAt"`
	Starred         bool        `json:"starred"`
	FailedTestCase  RunTestCase `json:"failedTestCase"`
	PassedTestCases int32       `json:"passedTestCases"`
	TotalTestCases  int32       `json:"totalTestCases"`
}

type SubmissionRequest struct {
	Language   string `json:"language"`
	SourceCode string `json:"sourceCode"`
	ProblemID  int32  `json:"problemId"`
}

type sqlc.Get

type SubmissionCreateParams struct {
	ID              uuid.UUID
	Status          string
	Memory          int32
	Time            string
	Stdout          string
	Stderr          string
	CompileOutput   string
	Message         string
	LanguageID      int32
	LanguageName    string
	AccountID       string
	ProblemID       int32
	SubmittedCode   string
	FailedTestCase  []byte
	PassedTestCases int32
	TotalTestCases  int32
}
