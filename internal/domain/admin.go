package domain

import (
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type AdminValidation struct {
	IsAdmin bool `json:"isAdmin"`
}

type AdminProblemRunResult struct {
	TestCase  RunTestCase          `json:"testCase"`
	Status    sql.SubmissionStatus `json:"status"` // Accepted, Wrong Answer, etc
	CreatedAt time.Time            `json:"createdAt"`
}

type AdminProblemData struct {
	Runs        map[string]AdminProblemRunResult `json:"runs"`
	Status      sql.SubmissionStatus             `json:"status"`
	CompletedAt time.Time                        `json:"completedAt"`
}

type CreateAdminProblemData struct {
	ProblemID int32 `json:"problemId"`
}

type AdminProblem struct {
	Problem  Problem           `json:"problem"`
	Solution map[string]string `json:"solution,omitempty"` // ["language": "sourceCode"]
}

type AdminProblemRunRequest struct {
	FunctionName string            `json:"functionName"`
	Solutions    map[string]string `json:"solutions"`
	TestCase     TestCase          `json:"testCase"`
}

type AdminProblemCreateRequest struct {
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	FunctionName string               `json:"functionName"`
	Tags         []string             `json:"tags"`
	Difficulty   string               `json:"difficulty"`
	Code         ProblemRequestCode   `json:"code"`
	Hints        []ProblemRequestHint `json:"hints"`
	Points       int32                `json:"points"`
	Solutions    map[string]string    `json:"solutions"` // ["language": "sourceCode"]
	TestCases    []TestCase           `json:"testCases"`
}

type AdminProblemCreateParams struct {
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	FunctionName string               `json:"functionName"`
	Tags         []string             `json:"tags"`
	Difficulty   string               `json:"difficulty"`
	Code         ProblemRequestCode   `json:"code"`
	Hints        []ProblemRequestHint `json:"hints"`
	Points       int32                `json:"points"`
	Solutions    map[string]string    `json:"solutions"` // ["language": "sourceCode"]
	TestCases    []TestCase           `json:"testCases"`
}
