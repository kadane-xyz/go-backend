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
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
)

type ProblemHandler struct {
	repo repository.ProblemsRepository
}

func NewProblemHandler(repo repository.ProblemsRepository) *ProblemHandler {
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

// GET: /problems
func (h *ProblemHandler) GetProblems(ctx context.Context, w http.ResponseWriter, params sql.GetProblemsFilteredPaginatedParams) *errors.ApiError {
	params, apiErr := h.GetProblemsValidateRequest(w, r)
	if apiErr != nil {
		apiErr.Send(w)
	}

	problems, err := h.repo.GetProblemsFilteredPaginated(ctx, params)
	if err != nil {
		return errors.NewApiError(err, http.StatusInternalServerError, "Failed to get problems")
	}

	if len(problems) == 0 {
		return errors.NewApiError(nil, http.StatusNotFound, "No problems found")
	}

	totalCount := problems[0].TotalCount
	if totalCount == 0 {
		return errors.NewApiError(nil, http.StatusNotFound, "No problems found")
	}

	lastPage := (totalCount + params.PerPage - 1) / params.PerPage

	if lastPage == 0 {
		lastPage = 1
	}

	// check if page is out of bounds
	if params.Page < 1 || params.Page > lastPage {
		return errors.NewApiError(nil, http.StatusBadRequest, "Page out of bounds")
	}

	responseData := []domain.Problem{}

	for i, problem := range problems {
		codeMap := InterfaceToMap(problem.Code)
		responseData[i] = domain.Problem{
			ID:            problem.ID,
			Title:         problem.Title,
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
		}
		if problem.Description != nil {
			responseData[i].Description = *problem.Description
		}
	}

	// Return an empty array if no matches (status 200)
	httputils.SendJSONPaginatedResponse(w, http.StatusOK,
		responseData,
		domain.Pagination{
			Page:      params.Page,
			PerPage:   params.PerPage,
			DataCount: totalCount,
			LastPage:  lastPage,
		},
	)

	return nil
}

func ValidateGetProblem(r *http.Request) (sql.GetProblemParams, *errors.ApiError) {
	userId, err := httputils.GetClientUserID(r)
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

	problem, err := h.repo.GetProblem(context.Background(), params)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	// test cases should not contain visibility on response
	codeMap := InterfaceToMap(problem.Code)
	response := domain.Problem{
		ID:            problem.ID,
		Title:         problem.Title,
		FunctionName:  problem.FunctionName,
		Description:   problem.Description,
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
