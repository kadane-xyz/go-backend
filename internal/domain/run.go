package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type RunRequest struct {
	Language   string     `json:"language"`
	SourceCode string     `json:"sourceCode"`
	ProblemID  int32      `json:"problemId"`
	TestCases  []TestCase `json:"testCases"`
}

type RunTestCase struct {
	Time           string               `json:"time"`
	Memory         int                  `json:"memory"`
	Status         sql.SubmissionStatus `json:"status"` // Accepted, Wrong Answer, etc
	Input          []TestCaseInput      `json:"input,omitempty"`
	Output         string               `json:"output"`         // User code output
	CompileOutput  string               `json:"compileOutput"`  // Compile output
	ExpectedOutput string               `json:"expectedOutput"` // Solution code output
}

type RunResult struct {
	ID        string               `json:"id,omitempty"`
	Language  string               `json:"language"`
	Time      string               `json:"time"`
	Memory    int32                `json:"memory"`
	TestCases []RunTestCase        `json:"testCases"`
	Status    sql.SubmissionStatus `json:"status"` // Accepted, Wrong Answer, etc
	// Our custom fields
	AccountID string    `json:"accountId"`
	ProblemID int32     `json:"problemId"`
	CreatedAt time.Time `json:"createdAt"`
}
