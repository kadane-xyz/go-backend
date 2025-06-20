package handlers

import (
	"context"
	"fmt"
	"net/http"

	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type ProblemHandler struct {
	repo repository.ProblemsRepository
}

func NewProblemHandler(repo repository.ProblemsRepository) *ProblemHandler {
	return &ProblemHandler{repo: repo}
}

func (h *ProblemHandler) GetProblemsValidateRequest(r *http.Request) (*domain.ProblemsGetParams, error) {
	titleSearch, err := httputils.GetQueryParam(r, "titleSearch", false)
	if err != nil {
		return nil, err
	}

	sort, err := httputils.GetQueryParam(r, "sort", false)
	if err != nil {
		fmt.Println("error sort")
		return nil, err
	}
	if *sort == "" {
		*sort = string(sql.ProblemSortIndex)
	} else if *sort != string(sql.ProblemSortAlpha) && *sort != string(sql.ProblemSortIndex) {
		return nil, errors.NewApiError(nil, "Invalid sort", http.StatusBadRequest)
	}

	order, err := httputils.GetQueryParamOrder(r, false)
	if err != nil {
		fmt.Println("error order")
		return nil, err
	}

	page, err := httputils.GetQueryParamInt32(r, "page", false)
	if err != nil {
		fmt.Println("page error")
		return nil, err
	}

	perPage, err := httputils.GetQueryParamInt32(r, "perPage", false)
	if err != nil {
		fmt.Println("perPage")
		return nil, err
	}

	// validate difficulty
	difficulty, err := httputils.GetQueryParam(r, "difficulty", false)
	if err != nil {
		fmt.Println("difficulty")
		return nil, err
	}

	return &domain.ProblemsGetParams{
		Title:      *titleSearch,
		Difficulty: sql.ProblemDifficulty(*difficulty),
		Sort:       sql.ProblemSort(*sort),
		Order:      sql.SortDirection(*order),
		PerPage:    perPage,
		Page:       page,
	}, nil
}

// GET: /problems
func (h *ProblemHandler) GetProblems(w http.ResponseWriter, r *http.Request) error {
	params, err := h.GetProblemsValidateRequest(r)
	if err != nil {
		fmt.Println("test validate request")
		return err
	}

	if params == nil {
		return nil
	}

	problems, err := h.repo.GetProblems(r.Context(), params)
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

	// convert code to map format
	for _, problem := range problems {
		codeMap := InterfaceToMap(problem.Code)
		problem.Code = codeMap
	}

	// Return an empty array if no matches (status 200)
	httputils.SendJSONPaginatedResponse(w, http.StatusOK,
		problems,
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

	problemId, err := httputils.GetURLParamInt32(r, "problemId")
	if err != nil {
		return nil, err
	}

	return &domain.ProblemGetParams{
		UserID:    claims.UserID,
		ProblemID: problemId,
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

	// convert code to map format
	codeMap := InterfaceToMap(problem.Code)
	problem.Code = codeMap

	httputils.SendJSONResponse(w, http.StatusOK, problem)

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
