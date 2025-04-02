package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Submission struct {
	Id            string               `json:"id"`
	Token         string               `json:"token"`
	Stdout        string               `json:"stdout"`
	Time          string               `json:"time"`
	Memory        int                  `json:"memory"`
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
	FailedTestCase  RunTestCase `json:"failedTestCase,omitempty"`
	PassedTestCases int32       `json:"passedTestCases"`
	TotalTestCases  int32       `json:"totalTestCases"`
}
