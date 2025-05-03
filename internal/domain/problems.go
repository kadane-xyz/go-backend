package domain

import (
	"encoding/json"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

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

type ProblemGetParams struct {
	UserId    string `json:"userId"`
	ProblemId int64  `json:"problemId"`
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

type ProblemCreateRequest struct {
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

type ProblemCreate struct {
	Id             int32           `json:"id"`
	Hints          []ProblemHint   `json:"hints"`
	Codes          []ProblemCode   `json:"problemCode"`
	Solutions      []Solution      `json:"solution"`
	TestCase       []TestCase      `json:"testCase"`
	TestCaseInputs []TestCaseInput `json:"testCaseInput"`
	TestCaseOutput []TestCaseInput `json:"testCaseOutput"`
}

type ProblemCreateParams struct {
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

type CreateProblemData struct {
	ProblemID string `json:"problemId"`
}

func FromSQLGetStarredProblemsRow(rows []sql.GetStarredProblemsRow) []*Problem {
	problems := []*Problem{}

	for i, row := range problems {
		problem := Problem(*row)
		problems[i] = &problem
	}

	return problems
}

func FromSQLCreateProblemRow(row sql.CreateProblemRow) (*ProblemCreate, error) {
	hints := []ProblemHint{}
	err := json.Unmarshal(row.TestCase, &hints)
	if err != nil {
		return nil, err
	}

	codes := []ProblemCode{}
	err = json.Unmarshal(row.Hints.([]byte), &codes)
	if err != nil {
		return nil, err
	}

	solutions := []Solution{}
	err = json.Unmarshal(row.Solutions.([]byte), &solutions)
	if err != nil {
		return nil, err
	}

	testCase := []TestCase{}
	err = json.Unmarshal(row.TestCase, &testCase)
	if err != nil {
		return nil, err
	}

	testCaseInput := []TestCaseInput{}
	err = json.Unmarshal(row.TestCaseInputs.([]byte), &testCaseInput)
	if err != nil {
		return nil, err
	}

	testCaseOutput := []TestCaseInput{}
	err = json.Unmarshal(row.TestCaseOutput, &testCaseOutput)
	if err != nil {
		return nil, err
	}

	return &ProblemCreate{
		Id:             row.ProblemID,
		Hints:          hints,
		Solutions:      solutions,
		TestCase:       testCase,
		TestCaseInputs: testCaseInput,
		TestCaseOutput: testCaseOutput,
	}, nil
}
