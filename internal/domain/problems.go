package domain

import "kadane.xyz/go-backend/v2/internal/database/sql"

type Problem struct {
	ID            int32                 `json:"id"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	FunctionName  string                `json:"functionName"`
	Tags          []string              `json:"tags"`
	Difficulty    sql.ProblemDifficulty `json:"difficulty"`
	Code          interface{}           `json:"code"`
	Hints         interface{}           `json:"hints"`
	Points        int32                 `json:"points"`
	Solution      interface{}           `json:"solution,omitempty"`
	TestCases     interface{}           `json:"testCases"`
	Starred       bool                  `json:"starred"`
	Solved        bool                  `json:"solved"`
	TotalAttempts int32                 `json:"totalAttempts"`
	TotalCorrect  int32                 `json:"totalCorrect"`
}
