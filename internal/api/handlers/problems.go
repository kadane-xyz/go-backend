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
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type ProblemHandler struct {
	repo repository.SQLProblemsRepository
}

func NewProblemHandler(repo repository.SQLProblemsRepository) *ProblemHandler {
	return &ProblemHandler{repo: repo}
}

func (h *ProblemHandler) GetProblemsValidateRequest(w http.ResponseWriter, r *http.Request) (*domain.ProblemRequest, error) {
	titleSearch := strings.TrimSpace(r.URL.Query().Get("titleSearch"))
	sortType := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortType == "" {
		sortType = string(sql.ProblemSortIndex)
	} else if sortType != string(sql.ProblemSortAlpha) && sortType != string(sql.ProblemSortIndex) {
		return nil, errors.NewApiError(nil, "Invalid sort", http.StatusBadRequest)
	}

	order := strings.TrimSpace(r.URL.Query().Get("order"))
	if order == "" {
		order = string(sql.SortDirectionAsc)
	} else if order != string(sql.SortDirectionAsc) && order != string(sql.SortDirectionDesc) {
		return nil, errors.NewApiError(nil, "Invalid order", http.StatusBadRequest)
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

	return &domain.ProblemGetParams{
		Title:         titleSearch,
		Difficulty:    difficulty,
		Sort:          sql.ProblemSort(sortType),
		SortDirection: sql.SortDirection(order),
		PerPage:       perPage,
		Page:          page,
	}, nil
}

// GET: /problems
func (h *ProblemHandler) GetProblems(ctx context.Context, w http.ResponseWriter, r *http.Response, params sql.GetProblemsFilteredPaginatedParams) error {
	params, err := h.GetProblemsValidateRequest(w, r)
	if err != nil {
		return err
	}

	problems, err := h.repo.GetProblemsFilteredPaginated(ctx, params)
	if err != nil {
		return errors.HandleDatabaseError(err, "get problems")
	}

	if len(problems) == 0 {
		return errors.NewAppError(err, "No problems found", http.StatusNotFound)
	}

	totalCount := problems[0].TotalCount
	if totalCount == 0 {
		return errors.NewAppError(nil, "No problems found", http.StatusNotFound)
	}

	lastPage := (totalCount + params.PerPage - 1) / params.PerPage

	if lastPage == 0 {
		lastPage = 1
	}

	// check if page is out of bounds
	if params.Page < 1 || params.Page > lastPage {
		return errors.NewApiError(nil, "Page out of bounds", http.StatusBadRequest)
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

func ValidateGetProblem(r *http.Request) (*domain.ProblemGetParams, error) {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return nil, err
	}

	problemId := chi.URLParam(r, "problemId")
	problemIdInt, err := strconv.ParseInt(problemId, 10, 32)
	if err != nil {
		return nil, err
	}

	return &domain.ProblemGetParams{
		UserId:    claims.UserID,
		ProblemId: problemIdInt,
	}, nil
}

// GET: /problems/{problemId}
func (h *ProblemHandler) GetProblem(w http.ResponseWriter, r *http.Request) error {
	params, err := ValidateGetProblem(r)
	if err != nil {
		return err
	}

	problem, err := h.repo.GetProblem(context.Background(), params)
	if err != nil {
		return errors.HandleDatabaseError(err, "get problem")
	}

	// test cases should not contain visibility on response
	codeMap := InterfaceToMap(problem.Code)

	description := ""
	if problem.Description != nil {
		description = *problem.Description
	}

	response := domain.Problem{
		ID:            problem.ID,
		Title:         problem.Title,
		FunctionName:  problem.FunctionName,
		Description:   description,
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

	return nil
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
