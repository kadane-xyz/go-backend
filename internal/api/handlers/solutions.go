package handlers

import (
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

type SolutionsHandler struct {
	repo repository.SolutionsRepository
}

func NewSolutionsHandler(repo repository.SolutionsRepository) *SolutionsHandler {
	return &SolutionsHandler{repo: repo}
}

func ValidateGetSolutions(r *http.Request, userId string) (*domain.SolutionsGetParams, error) {
	problemId := r.URL.Query().Get("problemId")
	if problemId == "" {
		return nil, errors.NewApiError(nil, "Missing problemId for solutions retrieval", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(problemId, 10, 32)
	if err != nil {
		return nil, errors.NewApiError(err, "Invalid problemId format for solutions retrieval", http.StatusBadRequest)
	}

	titleSearch := r.URL.Query().Get("titleSearch")
	if titleSearch == "" {
		titleSearch = ""
	}

	var tagsArray []string
	tags := r.URL.Query().Get("tags")
	if tags != "" {
		tagsArray = strings.Split(tags, ",")
	}

	// Handle pagination
	var page int32
	pageStr := r.URL.Query().Get("page")
	pageInt, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil {
		page = 1
	} else {
		page = int32(pageInt)
	}

	// Handle perPage
	var perPage int32
	perPageStr := r.URL.Query().Get("perPage")
	perPageInt, err := strconv.ParseInt(perPageStr, 10, 32)
	if err != nil {
		perPage = 10
	} else {
		perPage = int32(perPageInt)
	}

	// Handle sort
	sort := r.URL.Query().Get("sort")
	switch sort {
	case "time":
		sort = "created_at"
	default:
		sort = "votes"
	}

	// Handle order
	order := r.URL.Query().Get("order")
	if order == "asc" {
		order = "ASC"
	} else if order == "desc" {
		order = "DESC"
	} else {
		order = "DESC"
	}

	return &domain.SolutionsGetParams{
		ProblemID:     int32(id),
		Tags:          tagsArray,
		Title:         titleSearch,
		Page:          page,
		PerPage:       perPage,
		Sort:          sort,
		SortDirection: sql.SortDirection(order),
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
	solutionId := chi.URLParam(r, "solutionId")
	if solutionId == "" {
		return nil, errors.NewApiError(nil, "Missing solutionId for solution retrieval", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(solutionId, 10, 32)
	if err != nil {
		return nil, errors.NewApiError(err, "Invalid solutionId format for solution retrieval", http.StatusBadRequest)
	}

	return &domain.SolutionGetParams{
		ID:     int32(id),
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
	solutionId := chi.URLParam(r, "solutionId")
	// If problemId is empty, set idPg as NULL
	if solutionId == "" {
		return nil, errors.NewApiError(nil, "solutionId is required", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(solutionId, 10, 32)
	if err != nil {
		return nil, errors.NewApiError(err, "solutionId must be an integer", http.StatusBadRequest)
	}
	solutionID := int32(id)

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
	solutionId := chi.URLParam(r, "solutionId")
	// If problemId is empty, set idPg as NULL
	if solutionId == "" {
		return errors.NewApiError(nil, "solutionId is required", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(solutionId, 10, 32)
	if err != nil {
		return errors.NewApiError(err, "solutionId must be an integer", http.StatusBadRequest)
	}

	// Get solutions from db by idPg
	err = h.repo.DeleteSolution(r.Context(), claims.UserID, int32(id))
	if err != nil {
		return errors.HandleDatabaseError(err, "delete solution")
	}

	// Write solutionsJSON to response
	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

func validateVoteSolutionRequest(r *http.Request, userId string) (*domain.VoteSolutionsParams, error) {
	// Extract solutionId from URL parameters
	solutionId := chi.URLParam(r, "solutionId")
	if solutionId == "" {
		return nil, errors.NewApiError(nil, "Missing solutionId for solution retrieval", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(solutionId, 10, 32)
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
		SolutionId: int32(id),
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
