package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	Points       int                  `json:"points"`
	Solutions    map[string]string    `json:"solutions"` // ["language": "sourceCode"]
	TestCases    []TestCase           `json:"testCases"`
}

type Problem struct {
	ID            int                   `json:"id"`
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
	TotalAttempts int64                 `json:"totalAttempts"`
	TotalCorrect  int64                 `json:"totalCorrect"`
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

type Sort string

const (
	SortAlpha Sort = "alpha"
	SortIndex Sort = "index"
)

// GET: /problems
func (h *Handler) GetProblems(w http.ResponseWriter, r *http.Request) {
	titleSearch := strings.TrimSpace(r.URL.Query().Get("titleSearch"))
	difficulty := strings.TrimSpace(r.URL.Query().Get("difficulty"))
	sortType := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortType == "" {
		sortType = string(SortIndex)
	} else if sortType != string(SortAlpha) && sortType != string(SortIndex) {
		apierror.SendError(w, http.StatusBadRequest, "Invalid sort")
		return
	}

	order := strings.TrimSpace(r.URL.Query().Get("order"))
	if order == "" {
		order = "asc"
	} else if order != "asc" && order != "desc" {
		apierror.SendError(w, http.StatusBadRequest, "Invalid order")
		return
	}

	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.ParseInt(r.URL.Query().Get("perPage"), 10, 64)
	if err != nil || perPage < 1 {
		perPage = 10
	}

	if difficulty != "" {
		valid := (difficulty == string(sql.ProblemDifficultyEasy) ||
			difficulty == string(sql.ProblemDifficultyMedium) ||
			difficulty == string(sql.ProblemDifficultyHard))
		if !valid {
			apierror.SendError(w, http.StatusBadRequest, "Invalid difficulty")
			return
		}
	}

	problems, err := h.PostgresQueries.GetProblemsFilteredPaginated(r.Context(), sql.GetProblemsFilteredPaginatedParams{
		Title:         titleSearch,
		Difficulty:    difficulty,
		Sort:          sortType,
		SortDirection: order,
		PerPage:       int32(perPage),
		Page:          int32(page),
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problems")
		return
	}

	if len(problems) == 0 {
		apierror.SendError(w, http.StatusNotFound, "No problems found")
		return
	}

	totalCount := int64(problems[0].TotalCount)
	if totalCount == 0 {
		apierror.SendError(w, http.StatusNotFound, "No problems found")
		return
	}

	lastPage := (int64(totalCount) + perPage - 1) / perPage

	if lastPage == 0 {
		lastPage = 1
	}

	// check if page is out of bounds
	if page < 1 || page > lastPage {
		apierror.SendError(w, http.StatusBadRequest, "Page out of bounds")
		return
	}

	responseData := []Problem{}

	for _, problem := range problems {
		codeMap := InterfaceToMap(problem.Code)
		responseData = append(responseData, Problem{
			ID:            int(problem.ID),
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
	response := ProblemPaginationResponse{
		Data: responseData,
		Pagination: Pagination{
			Page:      page,
			PerPage:   perPage,
			DataCount: totalCount,
			LastPage:  lastPage,
		},
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

func (h *Handler) CreateProblem(request ProblemRequest) (*int32, *apierror.APIError) {
	problemID, err := h.PostgresQueries.CreateProblem(context.Background(), sql.CreateProblemParams{
		Title:        request.Title,
		Description:  pgtype.Text{String: request.Description, Valid: true},
		FunctionName: request.FunctionName,
		Points:       int32(request.Points),
		Tags:         request.Tags,
		Difficulty:   sql.ProblemDifficulty(request.Difficulty),
	})
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create problem")
	}

	for _, hint := range request.Hints {
		err = h.PostgresQueries.CreateProblemHint(context.Background(), sql.CreateProblemHintParams{
			ProblemID:   pgtype.Int4{Int32: int32(problemID), Valid: true},
			Description: hint.Description,
			Answer:      hint.Answer,
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create hint")
		}
	}

	for language, code := range request.Code {
		err = h.PostgresQueries.CreateProblemCode(context.Background(), sql.CreateProblemCodeParams{
			ProblemID: pgtype.Int4{Int32: int32(problemID), Valid: true},
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
			ProblemID:   pgtype.Int4{Int32: int32(problemID), Valid: true},
			Visibility:  sql.Visibility(testCase.Visibility),
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create test case")
		}

		for _, input := range testCase.Input {
			_, err = h.PostgresQueries.CreateProblemTestCaseInput(context.Background(), sql.CreateProblemTestCaseInputParams{
				ProblemTestCaseID: pgtype.Int4{Int32: int32(testCaseID.ID), Valid: true},
				Value:             input.Value,
				Type:              sql.ProblemTestCaseType(input.Type),
				Name:              input.Name,
			})
			if err != nil {
				return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create test case input")
			}
		}

		_, err = h.PostgresQueries.CreateProblemTestCaseOutput(context.Background(), sql.CreateProblemTestCaseOutputParams{
			ProblemTestCaseID: pgtype.Int4{Int32: int32(testCaseID.ID), Valid: true},
			Value:             testCase.Output,
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create test case output")
		}
	}

	for language, code := range request.Solutions {
		_, err = h.PostgresQueries.CreateProblemSolution(context.Background(), sql.CreateProblemSolutionParams{
			ProblemID: pgtype.Int4{Int32: int32(problemID), Valid: true},
			Language:  sql.ProblemLanguage(language),
			Code:      code,
		})
		if err != nil {
			return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create solution")
		}
	}

	return &problemID, nil
}

// POST: /problems
func (h *Handler) CreateProblemRoute(w http.ResponseWriter, r *http.Request) {
	admin := GetClientAdmin(w, r)
	if !admin {
		apierror.SendError(w, http.StatusForbidden, "You are not authorized to create problems")
		return
	}

	request, apiErr := DecodeJSONRequest[ProblemRequest](r)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	apiErr = CreateProblemRequestValidate(request)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	problemID, apiErr := h.CreateProblem(request)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	response := CreateAdminProblemResponse{
		Data: CreateAdminProblemData{
			ProblemID: strconv.Itoa(int(*problemID)),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GET: /problems/{problemId}
func (h *Handler) GetProblem(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	id := chi.URLParam(r, "problemId")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	problem, err := h.PostgresQueries.GetProblem(context.Background(), sql.GetProblemParams{
		ProblemID: int32(idInt),
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
			ID:            int(problem.ID),
			Title:         problem.Title,
			FunctionName:  problem.FunctionName,
			Description:   problem.Description.String,
			Tags:          problem.Tags,
			Difficulty:    problem.Difficulty,
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
			if codeMap, ok := entry.(map[string]interface{}); ok {
				if language, ok := codeMap["language"].(string); ok {
					if code, ok := codeMap["code"].(string); ok {
						response[language] = code
					}
				}
			}
		}
	}

	return response
}
