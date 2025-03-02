package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

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

type CreateProblemResponse struct {
	Data CreateProblemData `json:"data"`
}

func (h *Handler) GetProblemsValidateRequest(w http.ResponseWriter, r *http.Request) (sql.GetProblemsFilteredPaginatedParams, *apierror.APIError) {
	titleSearch := strings.TrimSpace(r.URL.Query().Get("titleSearch"))
	sortType := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortType == "" {
		sortType = string(sql.ProblemSortIndex)
	} else if sortType != string(sql.ProblemSortAlpha) && sortType != string(sql.ProblemSortIndex) {
		return sql.GetProblemsFilteredPaginatedParams{}, apierror.NewError(http.StatusBadRequest, "Invalid sort")
	}

	order := strings.TrimSpace(r.URL.Query().Get("order"))
	if order == "" {
		order = string(sql.SortDirectionAsc)
	} else if order != string(sql.SortDirectionAsc) && order != string(sql.SortDirectionDesc) {
		return sql.GetProblemsFilteredPaginatedParams{}, apierror.NewError(http.StatusBadRequest, "Invalid order")
	}

	var page int32
	pageStr := r.URL.Query().Get("page")
	pageInt, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil {
		page = 1
	} else {
		page = int32(pageInt)
	}

	var perPage int32
	perPageStr := r.URL.Query().Get("perPage")
	perPageInt, err := strconv.ParseInt(perPageStr, 10, 32)
	if err != nil {
		perPage = 10
	} else {
		perPage = int32(perPageInt)
	}

	var difficulty string
	difficultyStr := r.URL.Query().Get("difficulty")
	if difficultyStr == string(sql.ProblemDifficultyEasy) ||
		difficultyStr == string(sql.ProblemDifficultyMedium) ||
		difficultyStr == string(sql.ProblemDifficultyHard) {
		difficulty = difficultyStr
	} else {
		difficulty = ""
	}

	return sql.GetProblemsFilteredPaginatedParams{
		Title:         titleSearch,
		Difficulty:    difficulty,
		Sort:          sql.ProblemSort(sortType),
		SortDirection: sql.SortDirection(order),
		PerPage:       perPage,
		Page:          page,
	}, nil
}

func (h *Handler) GetProblems(ctx context.Context, w http.ResponseWriter, params sql.GetProblemsFilteredPaginatedParams) (ProblemPaginationResponse, *apierror.APIError) {
	problems, err := h.PostgresQueries.GetProblemsFilteredPaginated(ctx, params)
	if err != nil {
		return ProblemPaginationResponse{}, apierror.NewError(http.StatusInternalServerError, "Failed to get problems")
	}

	if len(problems) == 0 {
		return ProblemPaginationResponse{}, apierror.NewError(http.StatusNotFound, "No problems found")
	}

	totalCount := problems[0].TotalCount
	if totalCount == 0 {
		return ProblemPaginationResponse{}, apierror.NewError(http.StatusNotFound, "No problems found")
	}

	lastPage := (totalCount + params.PerPage - 1) / params.PerPage

	if lastPage == 0 {
		lastPage = 1
	}

	// check if page is out of bounds
	if params.Page < 1 || params.Page > lastPage {
		return ProblemPaginationResponse{}, apierror.NewError(http.StatusBadRequest, "Page out of bounds")
	}

	responseData := []Problem{}

	for _, problem := range problems {
		codeMap := InterfaceToMap(problem.Code)
		responseData = append(responseData, Problem{
			ID:            problem.ID,
			Title:         problem.Title,
			Description:   problem.Description.String,
			FunctionName:  problem.FunctionName,
			Tags:          problem.Tags,
			Difficulty:    problem.Difficulty,
			Code:          codeMap,
			Hints:         problem.Hints,
			Points:        problem.Points,
			Solution:      problem.Solutions,
			TestCases:     problem.TestCases,
			Starred:       problem.Starred,
			Solved:        problem.Solved,
			TotalAttempts: problem.TotalAttempts,
			TotalCorrect:  problem.TotalCorrect,
		})
	}

	// Return an empty array if no matches (status 200)
	return ProblemPaginationResponse{
		Data: responseData,
		Pagination: Pagination{
			Page:      params.Page,
			PerPage:   params.PerPage,
			DataCount: totalCount,
			LastPage:  lastPage,
		},
	}, nil
}

// GET: /problems
func (h *Handler) GetProblemsRoute(w http.ResponseWriter, r *http.Request) {
	params, apiErr := h.GetProblemsValidateRequest(w, r)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	response, apiErr := h.GetProblems(r.Context(), w, params)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateProblemRequestValidate(request ProblemRequest) *apierror.APIError {
	// Check problem fields
	if request.Title == "" || request.Description == "" || request.FunctionName == "" || len(request.Solutions) == 0 {
		return apierror.NewError(http.StatusBadRequest, "Title, description, function name, and solution are required")
	}

	if len(request.Code) == 0 {
		return apierror.NewError(http.StatusBadRequest, "At least one code is required")
	}

	if request.Points < 0 {
		return apierror.NewError(http.StatusBadRequest, "Points must be greater than 0")
	}

	if len(request.Solutions) == 0 {
		return apierror.NewError(http.StatusBadRequest, "Solution is required")
	}

	return nil
}

func (h *Handler) CreateProblem(request ProblemRequest) (*CreateProblemResponse, *apierror.APIError) {
	problemID, err := h.PostgresQueries.CreateProblem(context.Background(), sql.CreateProblemParams{
		Title:        request.Title,
		Description:  request.Description,
		FunctionName: request.FunctionName,
		Points:       request.Points,
		Tags:         request.Tags,
		Difficulty:   sql.ProblemDifficulty(request.Difficulty),
	})
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create problem")
	}

	for _, hint := range request.Hints {
		err = h.PostgresQueries.CreateProblemHint(context.Background(), sql.CreateProblemHintParams{
			ProblemID:   problemID,
			Description: hint.Description,
			Answer:      hint.Answer,
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create hint")
		}
	}

	for language, code := range request.Code {
		err = h.PostgresQueries.CreateProblemCode(context.Background(), sql.CreateProblemCodeParams{
			ProblemID: problemID,
			Language:  sql.ProblemLanguage(language),
			Code:      code,
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create code")
		}
	}

	for _, testCase := range request.TestCases {
		testCaseID, err := h.PostgresQueries.CreateProblemTestCase(context.Background(), sql.CreateProblemTestCaseParams{
			Description: testCase.Description,
			ProblemID:   problemID,
			Visibility:  sql.Visibility(testCase.Visibility),
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create test case")
		}

		for _, input := range testCase.Input {
			_, err = h.PostgresQueries.CreateProblemTestCaseInput(context.Background(), sql.CreateProblemTestCaseInputParams{
				ProblemTestCaseID: testCaseID.ID,
				Value:             input.Value,
				Type:              sql.ProblemTestCaseType(input.Type),
				Name:              input.Name,
			})
			if err != nil {
				return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create test case input")
			}
		}

		_, err = h.PostgresQueries.CreateProblemTestCaseOutput(context.Background(), sql.CreateProblemTestCaseOutputParams{
			ProblemTestCaseID: testCaseID.ID,
			Value:             testCase.Output,
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create test case output")
		}
	}

	for language, code := range request.Solutions {
		_, err = h.PostgresQueries.CreateProblemSolution(context.Background(), sql.CreateProblemSolutionParams{
			ProblemID: problemID,
			Language:  sql.ProblemLanguage(language),
			Code:      code,
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create solution")
		}
	}

	return &CreateProblemResponse{
		Data: CreateProblemData{
			ProblemID: strconv.Itoa(int(problemID)),
		},
	}, nil
}

// GET: /problems/{problemId}
func (h *Handler) GetProblem(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	var id int32
	idStr := chi.URLParam(r, "problemId")
	idInt, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}
	id = int32(idInt)

	problem, err := h.PostgresQueries.GetProblem(context.Background(), sql.GetProblemParams{
		ProblemID: id,
		UserID:    userId,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	// test cases should not contain visibility on response
	codeMap := InterfaceToMap(problem.Code)
	response := ProblemResponse{
		Data: Problem{
			ID:            problem.ID,
			Title:         problem.Title,
			FunctionName:  problem.FunctionName,
			Description:   problem.Description.String,
			Tags:          problem.Tags,
			Difficulty:    sql.ProblemDifficulty(problem.Difficulty),
			Code:          codeMap,
			Hints:         problem.Hints,
			Points:        problem.Points,
			TestCases:     problem.TestCases,
			Starred:       problem.Starred,
			Solved:        problem.Solved,
			TotalAttempts: problem.TotalAttempts,
			TotalCorrect:  problem.TotalCorrect,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func InterfaceToMap(object interface{}) map[string]string {
	response := make(map[string]string)

	// Convert to array of code entries
	if codeEntries, ok := object.([]interface{}); ok {
		for _, entry := range codeEntries {
			codeMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}

			language, ok := codeMap["language"].(string)
			if !ok {
				continue
			}

			code, ok := codeMap["code"].(string)
			if !ok {
				continue
			}

			response[language] = code
		}
	}

	return response
}
