package domain

import (
	"encoding/json"
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Problem struct {
	Id            int32                 `json:"id"`
	Title         string                `json:"title"`
	CreatedAt     time.Time             `json:"createdAt"`
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

type ProblemTestCase struct {
	Id          int32
	ProblemId   int32
	Description string
	CreatedAt   time.Time
	Visibility  sql.Visibility
	Input       TestCaseInput
	Output      string
}

type ProblemGetParams struct {
	UserId    string `json:"userId"`
	ProblemId int32  `json:"problemId"`
}

type ProblemsGetParams struct {
	UserId     string
	Title      string
	Difficulty sql.ProblemDifficulty
	Sort       sql.ProblemSort
	Order      sql.SortDirection
	PerPage    int32
	Page       int32
}

type ProblemTestCasesGetParams struct {
	ProblemId  int32
	Visibility sql.Visibility
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

func FromSQLGetProblemRow(row sql.GetProblemRow) (*Problem, error) {
	hints := []ProblemHint{}
	err := json.Unmarshal(row.Hints.([]byte), &hints)
	if err != nil {
		return nil, err
	}

	codes := []ProblemCode{}
	err = json.Unmarshal(row.Code.([]byte), &codes)
	if err != nil {
		return nil, err
	}

	solutions := []Solution{}
	err = json.Unmarshal(row.Solutions.([]byte), &solutions)
	if err != nil {
		return nil, err
	}

	testCase := []TestCase{}
	err = json.Unmarshal(row.TestCases.([]byte), &testCase)
	if err != nil {
		return nil, err
	}

	return &Problem{
		Id:            row.ID,
		Title:         row.Title,
		Description:   nullHandler(row.Description),
		FunctionName:  row.FunctionName,
		Points:        row.Points,
		CreatedAt:     row.CreatedAt.Time,
		Difficulty:    row.Difficulty,
		Tags:          row.Tags,
		Code:          codes,
		Hints:         hints,
		TestCases:     testCase,
		Solution:      solutions,
		Starred:       row.Starred,
		Solved:        row.Solved,
		TotalAttempts: row.TotalAttempts,
		TotalCorrect:  row.TotalCorrect,
	}, nil
}

func FromSQLGetProblemsFilteredPaginated(rows []sql.GetProblemsFilteredPaginatedRow) ([]*Problem, error) {
	problems := []*Problem{}

	for i, row := range rows {
		problem, err := FromSQLGetProblemRow(sql.GetProblemRow(row))
		if err != nil {
			return nil, err
		}

		problems[i] = problem
	}

	return problems, nil
}

func FromSQLGetStarredProblemsRow(rows []sql.GetStarredProblemsRow) []*StarredProblem {
	problems := []*StarredProblem{}

	for i, row := range problems {
		problem := StarredProblem(*row)
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
		Codes:          codes,
		Solutions:      solutions,
		TestCase:       testCase,
		TestCaseInputs: testCaseInput,
		TestCaseOutput: testCaseOutput,
	}, nil
}

func FromSQLGetProblemTestCases(rows []sql.GetProblemTestCasesRow) ([]*ProblemTestCase, error) {
	testCases := []*ProblemTestCase{}

	for i, row := range rows {
		input := TestCaseInput{}
		if row.Input != nil {
			err := json.Unmarshal(row.Input.([]byte), &input)
			if err != nil {
				return nil, err
			}
		}

		testCases[i] = &ProblemTestCase{
			Id:          row.ID,
			ProblemId:   nullHandler(row.ProblemID),
			Description: row.Description,
			CreatedAt:   row.CreatedAt.Time,
			Visibility:  sql.Visibility(row.Visibility),
			Input:       input,
			Output:      row.Output,
		}
	}

	return testCases, nil
}
