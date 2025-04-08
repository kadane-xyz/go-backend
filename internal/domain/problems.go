package domain

import "kadane.xyz/go-backend/v2/internal/database/sql"

type Problem struct {
	ID            int32                 `json:"id"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	FunctionName  string                `json:"functionName"`
	Tags          []string              `json:"tags"`
	Difficulty    sql.ProblemDifficulty `json:"difficulty"`
	Code          any                   `json:"code"`
	Hints         any                   `json:"hints"`
	Points        int32                 `json:"points"`
	Solution      any                   `json:"solution,omitempty"`
	TestCases     any                   `json:"testCases"`
	Starred       bool                  `json:"starred"`
	Solved        bool                  `json:"solved"`
	TotalAttempts int32                 `json:"totalAttempts"`
	TotalCorrect  int32                 `json:"totalCorrect"`
}

type ProblemRequest struct {
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

type ProblemHint struct {
	Description string `json:"description"`
	Answer      string `json:"answer"`
}

type ProblemCode struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type ProblemRequestHint struct {
	Description string `json:"description"`
	Answer      string `json:"answer"`
}

type ProblemRequestCode map[string]string

type ProblemResponse struct {
	Data Problem `json:"data"`
}

type ProblemsResponse struct {
	Data []Problem `json:"data"`
}

type ProblemPaginationResponse struct {
	Data       []Problem  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type CreateProblemData struct {
	ProblemID string `json:"problemId"`
}
