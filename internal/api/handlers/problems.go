package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/errors"
)

type ProblemHandler struct {
	repo *repository.ProblemsRepository
}

func NewProblemHandler(repo *repository.ProblemsRepository) *ProblemHandler {
	return &ProblemHandler{repo: repo}
}

func (h *ProblemHandler) GetProblemsValidateRequest(w http.ResponseWriter, r *http.Request) (sql.GetProblemsFilteredPaginatedParams, *errors.ApiError) {
	titleSearch := strings.TrimSpace(r.URL.Query().Get("titleSearch"))
	sortType := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortType == "" {
		sortType = string(sql.ProblemSortIndex)
	} else if sortType != string(sql.ProblemSortAlpha) && sortType != string(sql.ProblemSortIndex) {
		return sql.GetProblemsFilteredPaginatedParams{}, errors.NewApiError(nil, http.StatusBadRequest, "Invalid sort")
	}

	order := strings.TrimSpace(r.URL.Query().Get("order"))
	if order == "" {
		order = string(sql.SortDirectionAsc)
	} else if order != string(sql.SortDirectionAsc) && order != string(sql.SortDirectionDesc) {
		return sql.GetProblemsFilteredPaginatedParams{}, errors.NewApiError(nil, http.StatusBadRequest, "Invalid order")
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

func (h *ProblemHandler) GetProblems(ctx context.Context, w http.ResponseWriter, params sql.GetProblemsFilteredPaginatedParams) (ProblemPaginationResponse, *errors.ApiError) {
	problems, err := h.accessor.GetProblemsFilteredPaginated(ctx, params)
	if err != nil {
		return ProblemPaginationResponse{}, errors.NewApiError(err, http.StatusInternalServerError, "Failed to get problems")
	}

	if len(problems) == 0 {
		return ProblemPaginationResponse{}, errors.NewApiError(nil, http.StatusNotFound, "No problems found")
	}

	totalCount := problems[0].TotalCount
	if totalCount == 0 {
		return ProblemPaginationResponse{}, errors.NewApiError(nil, http.StatusNotFound, "No problems found")
	}

	lastPage := (totalCount + params.PerPage - 1) / params.PerPage

	if lastPage == 0 {
		lastPage = 1
	}

	// check if page is out of bounds
	if params.Page < 1 || params.Page > lastPage {
		return ProblemPaginationResponse{}, errors.NewApiError(nil, http.StatusBadRequest, "Page out of bounds")
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
func (h *ProblemHandler) GetProblemsRoute(w http.ResponseWriter, r *http.Request) {
	params, apiErr := h.GetProblemsValidateRequest(w, r)
	if apiErr != nil {
		apiErr.Send(w)
		return
	}

	response, apiErr := h.GetProblems(r.Context(), w, params)
	if apiErr != nil {
		apiErr.Send(w)
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, response)
}

func CreateProblemRequestValidate(request ProblemRequest) *errors.ApiError {
	// Check problem fields
	if request.Title == "" || request.Description == "" || request.FunctionName == "" || len(request.Solutions) == 0 {
		return errors.NewApiError(nil, http.StatusBadRequest, "Title, description, function name, and solution are required")
	}

	if len(request.Code) == 0 {
		return errors.NewApiError(nil, http.StatusBadRequest, "At least one code is required")
	}

	if request.Points < 0 {
		return errors.NewApiError(nil, http.StatusBadRequest, "Points must be greater than 0")
	}

	if len(request.Solutions) == 0 {
		return errors.NewApiError(nil, http.StatusBadRequest, "Solution is required")
	}

	return nil
}

func (h *ProblemHandler) CreateProblem(request ProblemRequest) (*CreateProblemData, *errors.ApiError) {
	testCaseDescriptions := []string{}
	testCaseVisibilities := []sql.Visibility{}
	testCaseOutputIndices := []int32{}
	testCaseOutputValues := []string{}
	testCaseInputIndices := []int32{}
	testCaseInputNames := []string{}
	testCaseInputValues := []string{}
	testCaseInputTypes := []sql.ProblemTestCaseType{}
	for i, testCase := range request.TestCases {
		testCaseDescriptions = append(testCaseDescriptions, testCase.Description)
		testCaseVisibilities = append(testCaseVisibilities, sql.Visibility(testCase.Visibility))
		for _, input := range testCase.Input {
			if input.Name == "" || input.Value == "" || input.Type == "" {
				return nil, errors.NewApiError(nil, http.StatusBadRequest, "Input name, value, and type are required")
			}
			testCaseInputNames = append(testCaseInputNames, input.Name)
			testCaseInputValues = append(testCaseInputValues, input.Value)
			testCaseInputTypes = append(testCaseInputTypes, sql.ProblemTestCaseType(input.Type))
			testCaseInputIndices = append(testCaseInputIndices, int32(i))
		}
		testCaseOutputValues = append(testCaseOutputValues, testCase.Output)
		testCaseOutputIndices = append(testCaseOutputIndices, int32(i))
	}

	problemID, err := h.accessor.CreateProblem(context.Background(), sql.CreateProblemParams{
		Title:                 request.Title,
		Description:           request.Description,
		FunctionName:          request.FunctionName,
		Points:                request.Points,
		Tags:                  request.Tags,
		Difficulty:            sql.ProblemDifficulty(request.Difficulty),
		TestCaseDescriptions:  testCaseDescriptions,
		TestCaseVisibilities:  testCaseVisibilities,
		TestCaseOutputIndices: testCaseOutputIndices,
		TestCaseOutputValues:  testCaseOutputValues,
		TestCaseInputIndices:  testCaseInputIndices,
		TestCaseInputNames:    testCaseInputNames,
		TestCaseInputValues:   testCaseInputValues,
		TestCaseInputTypes:    testCaseInputTypes,
	})
	if err != nil {
		return nil, errors.NewApiError(err, http.StatusInternalServerError, "Failed to create problem")
	}

	/*for _, hint := range request.Hints {
		err = h.accessor.CreateProblemHint(context.Background(), sql.CreateProblemHintParams{
			ProblemID:   problemID,
			Description: hint.Description,
			Answer:      hint.Answer,
		})
		if err != nil {
			return nil, errors.NewApiError(err, http.StatusInternalServerError, "Failed to create hint")
		}
	}

	for language, code := range request.Code {
		err = h.accessor.CreateProblemCode(context.Background(), sql.CreateProblemCodeParams{
			ProblemID: problemID,
			Language:  sql.ProblemLanguage(language),
			Code:      code,
		})
		if err != nil {
			return nil, errors.NewApiError(err, http.StatusInternalServerError, "Failed to create code")
		}
	}

	for _, testCase := range request.TestCases {
		testCaseID, err := h.accessor.CreateProblemTestCase(context.Background(), sql.CreateProblemTestCaseParams{
			Description: testCase.Description,
			ProblemID:   problemID,
			Visibility:  sql.Visibility(testCase.Visibility),
		})
		if err != nil {
			return nil, errors.NewApiError(err, http.StatusInternalServerError, "Failed to create test case")
		}

		for _, input := range testCase.Input {
			_, err = h.accessor.CreateProblemTestCaseInput(context.Background(), sql.CreateProblemTestCaseInputParams{
				ProblemTestCaseID: testCaseID.ID,
				Value:             input.Value,
				Type:              sql.ProblemTestCaseType(input.Type),
				Name:              input.Name,
			})
			if err != nil {
				return nil, errors.NewApiError(err, http.StatusInternalServerError, "Failed to create test case input")
			}
		}

		_, err = h.accessor.CreateProblemTestCaseOutput(context.Background(), sql.CreateProblemTestCaseOutputParams{
			ProblemTestCaseID: testCaseID.ID,
			Value:             testCase.Output,
		})
		if err != nil {
			return nil, errors.NewApiError(err, http.StatusInternalServerError, "Failed to create test case output")
		}
	}

	for language, code := range request.Solutions {
		_, err = h.accessor.CreateProblemSolution(context.Background(), sql.CreateProblemSolutionParams{
			ProblemID: problemID,
			Language:  sql.ProblemLanguage(language),
			Code:      code,
		})
		if err != nil {
			return nil, errors.NewApiError(err, http.StatusInternalServerError, "Failed to create solution")
		}
	}*/

	return &CreateProblemData{
		ProblemID: problemID,
	}, nil
}

func ValidateGetProblem(r *http.Request) (sql.GetProblemParams, *errors.ApiError) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return sql.GetProblemParams{}, errors.NewApiError(err, http.StatusInternalServerError, "Failed to get user ID")
	}

	problemId := chi.URLParam(r, "problemId")
	problemIdInt, err := strconv.ParseInt(problemId, 10, 32)
	if err != nil {
		return sql.GetProblemParams{}, errors.NewApiError(err, http.StatusBadRequest, "Invalid problem ID")
	}

	return sql.GetProblemParams{
		UserID:    userId,
		ProblemID: int32(problemIdInt),
	}, nil
}

// GET: /problems/{problemId}
func (h *ProblemHandler) GetProblem(w http.ResponseWriter, r *http.Request) {
	params, apiErr := ValidateGetProblem(r)
	if apiErr != nil {
		apiErr.Send(w)
		return
	}

	problem, err := h.accessor.GetProblem(context.Background(), params)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	// test cases should not contain visibility on response
	codeMap := InterfaceToMap(problem.Code)
	response := Problem{
		ID:            problem.ID,
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
	}

	httputils.SendJSONResponse(w, http.StatusOK, response)
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
