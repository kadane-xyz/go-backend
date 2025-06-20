package domain

import (
	"encoding/json"
	"time"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Problem struct {
	ID            int32                 `json:"id"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	FunctionName  string                `json:"functionName"`
	Points        int32                 `json:"points"`
	CreatedAt     time.Time             `json:"createdAt"`
	Difficulty    sql.ProblemDifficulty `json:"difficulty"`
	Tags          []string              `json:"tags"`
	Code          any                   `json:"code"`
	Hints         any                   `json:"hints"`
	Solution      any                   `json:"solution,omitempty"`
	TestCases     any                   `json:"testCases"`
	Starred       bool                  `json:"starred"`
	Solved        bool                  `json:"solved"`
	TotalAttempts int32                 `json:"totalAttempts"`
	TotalCorrect  int32                 `json:"totalCorrect"`
	TotalCount    int32                 `json:"totalCount"`
}

type ProblemGetParams struct {
	UserID    string `json:"userId"`
	ProblemID int32  `json:"problemId"`
}

type ProblemsGetParams struct {
	UserID     string
	Title      string
	Difficulty sql.ProblemDifficulty
	Sort       sql.ProblemSort
	Order      sql.SortDirection
	PerPage    int32
	Page       int32
}

type ProblemTestCasesGetParams struct {
	ProblemID  int32
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
	ID             int32           `json:"id"`
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

type ProblemSolution = ProblemCode

type ProblemRequestHint struct {
	Description string `json:"description"`
	Answer      string `json:"answer"`
}

type ProblemRequestCode map[string]string

type CreateProblemData struct {
	ProblemID string `json:"problemId"`
}

func FromSQLProblem(row sql.Problem) *Problem {
	return &Problem{
		ID:           row.ID,
		Title:        row.Title,
		Description:  row.Description.String,
		FunctionName: row.FunctionName,
		Points:       row.Points,
		CreatedAt:    row.CreatedAt.Time,
		Difficulty:   row.Difficulty,
		Tags:         row.Tags,
	}
}

// handles problem hints interface{} type
func hintsHandler(hints any) ([]*ProblemHint, error) {
	return jsonArrayHandler[ProblemHint](hints.([]byte))
}

// handles problem codes interface{} type
func codesHandler(codes any) ([]*ProblemCode, error) {
	return jsonArrayHandler[ProblemCode](codes.([]byte))
}

// handles problem solutions interface{} type
func solutionsHandler(solutions any) ([]*ProblemSolution, error) {
	return jsonArrayHandler[ProblemSolution](solutions.([]byte))
}

// handles problem testCases interface{} type
func testCaseHandler(testCases any) ([]*TestCase, error) {
	return jsonArrayHandler[TestCase](testCases.([]byte))
}

func FromSQLGetProblemRow(row sql.GetProblemRow) (*Problem, error) {
	hints, err := hintsHandler(row.Hints)
	if err != nil {
		return nil, err
	}

	codes, err := codesHandler(row.Code)
	if err != nil {
		return nil, err
	}

	solutions, err := solutionsHandler(row.Solutions)
	if err != nil {
		return nil, err
	}

	testCase, err := testCaseHandler(row.TestCases)
	if err != nil {
		return nil, err
	}

	return &Problem{
		ID:            row.ID,
		Title:         row.Title,
		Description:   row.Description.String,
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

func FromSQLGetProblemsRow(rows []sql.GetProblemsRow) ([]*Problem, error) {
	problems := []*Problem{}

	for i, row := range rows {
		hints, err := hintsHandler(row.Hints)
		if err != nil {
			return nil, err
		}

		codes, err := codesHandler(row.Code)
		if err != nil {
			return nil, err
		}

		solutions, err := solutionsHandler(row.Solutions)
		if err != nil {
			return nil, err
		}

		testCase, err := testCaseHandler(row.TestCases)
		if err != nil {
			return nil, err
		}

		problems[i] = &Problem{
			ID:            row.ID,
			Title:         row.Title,
			Description:   row.Description.String,
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
			TotalCount:    row.TotalCount,
		}
	}

	return problems, nil
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
		ID:             row.ProblemID,
		Hints:          hints,
		Codes:          codes,
		Solutions:      solutions,
		TestCase:       testCase,
		TestCaseInputs: testCaseInput,
		TestCaseOutput: testCaseOutput,
	}, nil
}

func FromSQLGetProblemTestCases(rows []sql.GetProblemTestCasesRow) ([]*TestCase, error) {
	testCases := []*TestCase{}

	for i, row := range rows {
		input := TestCaseInput{}
		if row.Input != nil {
			err := json.Unmarshal(row.Input.([]byte), &input)
			if err != nil {
				return nil, err
			}
		}

		testCases[i] = &TestCase{
			Description: row.Description,
			Visibility:  sql.Visibility(row.Visibility),
			Input:       []TestCaseInput{input},
			Output:      row.Output,
		}
	}

	return testCases, nil
}
