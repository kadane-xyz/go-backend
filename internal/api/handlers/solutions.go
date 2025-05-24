package handlers

import (
	"net/http"

	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type SolutionsHandler struct {
	repo repository.SolutionsRepository
}

func NewSolutionsHandler(repo repository.SolutionsRepository) *SolutionsHandler {
	return &SolutionsHandler{repo: repo}
}

func ValidateGetSolutions(r *http.Request, userId string) (*domain.SolutionsGetParams, error) {
	problemId, err := httputils.GetQueryParamInt32(r, "problemId")
	if err != nil {
		return nil, err
	}

	titleSearch, err := httputils.GetQueryParam(r, "titleSearch")
	if err != nil {
		return nil, err
	}

	tags, err := httputils.GetQueryParamStringArray(r, "tags")
	if err != nil {
		return nil, err
	}

	// Handle pagination
	page, err := httputils.GetQueryParamInt32(r, "page")
	if err != nil {
		return nil, err
	}

	// Handle perPage
	perPage, err := httputils.GetQueryParamInt32(r, "perPage")
	if err != nil {
		return nil, err
	}

	// Handle sort
	sort, err := httputils.GetQueryParam(r, "sort")
	if err != nil {
		return nil, err
	}
	switch *sort {
	case "time":
		*sort = "created_at"
	default:
		*sort = "votes"
	}

	// Handle order
	order, err := httputils.GetQueryParamOrder(r)
	if err != nil {
		return nil, err
	}

	return &domain.SolutionsGetParams{
		ProblemID:     problemId,
		Tags:          tags,
		Title:         *titleSearch,
		Page:          page,
		PerPage:       perPage,
		Sort:          *sort,
		SortDirection: sql.SortDirection(*order),
		UserId:        userId,
	}, nil
}

// GET: /solutions
func (h *SolutionsHandler) GetSolutions(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	params, err := ValidateGetSolutions(r, claims.UserID)
	if err != nil {
		return err
	}

	solutions, err := h.repo.GetSolutions(r.Context(), params)
	if err != nil {
		httputils.EmptyDataArrayResponse(w) // { data: [] }
		return nil
	}

	if len(solutions) == 0 {
		httputils.EmptyDataArrayResponse(w) // { data: [] }
		return nil
	}

	totalCount := solutions[0].TotalCount
	if totalCount == 0 {
		return errors.NewAppError(nil, "No solutions found", http.StatusNotFound)
	}

	// Prepare response
	for i, solution := range solutions {
		// If tags is nil, set it to an empty array
		if solution.Tags == nil {
			solutions[i].Tags = []string{}
		}

		// If preview is not true, include the body
		if r.URL.Query().Get("preview") != "true" {
			solutions[i].Body = solution.Body
		}
	}

	// Calculate last page
	lastPage := (totalCount + params.PerPage - 1) / params.PerPage

	// Write solutionsJSON to response
	httputils.SendJSONPaginatedResponse(w, http.StatusOK,
		solutions, domain.Pagination{
			Page:      params.Page,
			PerPage:   params.PerPage,
			DataCount: totalCount,
			LastPage:  lastPage,
		},
	)

	return nil
}

func validateCreateSolutionRequest(r *http.Request, userId string) (*domain.SolutionsCreateParams, error) {
	solution, err := httputils.DecodeJSONRequest[domain.SolutionsCreateParams](r)
	if err != nil {
		return nil, err
	}

	if solution.Title == "" || solution.Body == "" || *solution.ProblemID <= 0 {
		return nil, errors.NewApiError(nil, "Missing required fields for solution creation", http.StatusBadRequest)
	}

	return &domain.SolutionsCreateParams{
		UserID:    userId,
		Title:     solution.Title,
		Tags:      solution.Tags,
		Body:      solution.Body,
		ProblemID: solution.ProblemID,
	}, nil
}

// POST: /
func (h *SolutionsHandler) CreateSolution(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	params, err := validateCreateSolutionRequest(r, claims.UserID)
	if err != nil {
		return err
	}

	// Insert solution into db
	err = h.repo.CreateSolution(r.Context(), params)
	if err != nil {
		return errors.HandleDatabaseError(err, "create solution")
	}

	// Write response
	httputils.SendJSONResponse(w, http.StatusCreated, nil)

	return nil
}

func validateGetSolutionRequest(r *http.Request, userId string) (*domain.SolutionGetParams, error) {
	solutionId, err := httputils.GetURLParamInt32(r, "solutionId")
	if err != nil {
		return nil, err
	}

	return &domain.SolutionGetParams{
		ID:     solutionId,
		UserID: userId,
	}, nil
}

// GET: /{solutionId}
func (h *SolutionsHandler) GetSolution(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	params, err := validateGetSolutionRequest(r, claims.UserID)
	if err != nil {
		return err
	}

	// Get solutions from db by idPg
	solution, err := h.repo.GetSolution(r.Context(), params)
	if err != nil {
		httputils.EmptyDataResponse(w) // { data: {} }
		return nil
	}

	// Write solutionsJSON to response
	httputils.SendJSONDataResponse(w, http.StatusOK, solution)

	return nil
}

func validateUpdateSolutionRequest(r *http.Request, userID string) (*domain.SolutionsUpdateParams, error) {
	// Handle problemId query parameter
	solutionID, err := httputils.GetURLParamInt32(r, "solutionId")
	if err != nil {
		return nil, err
	}

	solutionRequest, err := httputils.DecodeJSONRequest[domain.UpdateSolutionRequest](r)
	if err != nil {
		return nil, err
	}

	if solutionRequest.Title == "" && solutionRequest.Body == "" && len(solutionRequest.Tags) > 0 {
		return nil, errors.NewApiError(nil, "at least one field must be provided", http.StatusBadRequest)
	}

	return &domain.SolutionsUpdateParams{
		SolutionID: &solutionID,
		UserID:     userID,
		Title:      solutionRequest.Title,
		Body:       solutionRequest.Body,
		Tags:       solutionRequest.Tags,
	}, nil
}

// PUT: /{solutionId}
func (h *SolutionsHandler) UpdateSolution(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	params, err := validateUpdateSolutionRequest(r, claims.UserID)
	if err != nil {
		return err
	}

	// Get solutions from db by idPg
	err = h.repo.UpdateSolution(r.Context(), params)
	if err != nil {
		return errors.HandleDatabaseError(err, "update solution")
	}

	// Write solutionsJSON to response
	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

// DELETE: /{solutionId}
func (h *SolutionsHandler) DeleteSolution(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	// Handle problemId query parameter
	solutionID, err := httputils.GetURLParamInt32(r, "solutionId")
	if err != nil {
		return err
	}

	// Get solutions from db by idPg
	err = h.repo.DeleteSolution(r.Context(), claims.UserID, solutionID)
	if err != nil {
		return errors.HandleDatabaseError(err, "delete solution")
	}

	// Write solutionsJSON to response
	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

func validateVoteSolutionRequest(r *http.Request, userId string) (*domain.VoteSolutionsParams, error) {
	// Extract solutionId from URL parameters
	solutionId, err := httputils.GetURLParamInt32(r, "solutionId")
	if err != nil {
		return nil, errors.NewApiError(err, "Invalid solutionId format for solution retrieval", http.StatusBadRequest)
	}

	// Decode the request body into VoteRequest struct
	req, err := httputils.DecodeJSONRequest[domain.VoteRequest](r)
	if err != nil {
		return nil, errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if req.Vote == "" {
		return nil, errors.NewApiError(nil, "Vote is required", http.StatusBadRequest)
	}

	return &domain.VoteSolutionsParams{
		UserId:     userId,
		SolutionId: solutionId,
		Vote:       req.Vote,
	}, nil
}

// PATCH: /{solutionId}/vote
func (h *SolutionsHandler) VoteSolution(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	params, err := validateVoteSolutionRequest(r, claims.UserID)
	if err != nil {
		return err
	}

	err = h.repo.VoteSolution(r.Context(), params)
	if err != nil {
		return errors.HandleDatabaseError(err, "Error voting on solution")
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}
