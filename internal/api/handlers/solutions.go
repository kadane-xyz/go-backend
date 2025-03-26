package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/database/sql"
)

type Solutions struct {
	Id        int64    `json:"id"`
	Username  string   `json:"username"`
	Title     string   `json:"title"`
	Date      string   `json:"date"`
	Tags      []string `json:"tags"`
	Body      string   `json:"body"`
	Votes     int      `json:"votes"`
	ProblemId int64    `json:"problemId"`
}

type CreateSolutionRequest struct {
	ProblemId int64    `json:"problemId"`
	Title     string   `json:"title"`
	Tags      []string `json:"tags"`
	Body      string   `json:"body"`
}

type UpdateSolutionRequest struct {
	Title string   `json:"title"`
	Body  string   `json:"body"`
	Tags  []string `json:"tags"`
}

type SolutionResponse struct {
	Data SolutionsData `json:"data"`
}

type SolutionsData struct {
	Id              int64            `json:"id"`
	Body            string           `json:"body,omitempty"`
	Comments        int32            `json:"comments"`
	Date            pgtype.Timestamp `json:"date"`
	Tags            []string         `json:"tags"`
	Title           string           `json:"title"`
	Username        string           `json:"username,omitempty"`
	Level           int32            `json:"level,omitempty"`
	AvatarUrl       string           `json:"avatarUrl,omitempty"`
	Votes           int32            `json:"votes"`
	CurrentUserVote sql.VoteType     `json:"currentUserVote"`
	Starred         bool             `json:"starred"`
}

type SolutionsResponse struct {
	Data       []SolutionsData `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// GET: /solutions
func (h *Handler) GetSolutions(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	problemId := r.URL.Query().Get("problemId")
	if problemId == "" {
		SendError(w, http.StatusBadRequest, "Missing problemId for solutions retrieval")
		return
	}

	id, err := strconv.ParseInt(problemId, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid problemId format for solutions retrieval")
		return
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

	solutions, err := h.PostgresQueries.GetSolutionsPaginated(r.Context(), sql.GetSolutionsPaginatedParams{
		ProblemID:     pgtype.Int8{Int64: id, Valid: true},
		Tags:          tagsArray,
		Title:         titleSearch,
		PerPage:       perPage,
		Page:          page,
		Sort:          sort,
		SortDirection: order,
		UserID:        userId,
	})
	if err != nil {
		EmptyDataArrayResponse(w) // { data: [] }
		return
	}

	if len(solutions) == 0 {
		EmptyDataArrayResponse(w) // { data: [] }
		return
	}

	totalCount := solutions[0].TotalCount
	if totalCount == 0 {
		SendError(w, http.StatusNotFound, "No solutions found")
		return
	}

	// Prepare response
	var solutionsData []SolutionsData
	for _, solution := range solutions {
		// If tags is nil, set it to an empty array
		if solution.Tags == nil {
			solution.Tags = []string{}
		}

		solutionData := SolutionsData{
			Id:              solution.ID,
			Body:            solution.Body,
			Comments:        solution.CommentsCount,
			Date:            solution.CreatedAt,
			Tags:            solution.Tags,
			Title:           solution.Title,
			Username:        solution.UserUsername,
			Level:           solution.UserLevel,
			AvatarUrl:       solution.UserAvatarUrl.String,
			Votes:           solution.VotesCount,
			CurrentUserVote: solution.UserVote,
			Starred:         solution.Starred,
		}

		// If preview is not true, include the body
		if r.URL.Query().Get("preview") != "true" {
			solutionData.Body = solution.Body
		}

		solutionsData = append(solutionsData, solutionData)
	}

	// Calculate last page
	lastPage := (totalCount + perPage - 1) / perPage

	// Final response
	response := SolutionsResponse{
		Data: solutionsData,
		Pagination: Pagination{
			Page:      page,       // Current page
			PerPage:   perPage,    // Items per page
			DataCount: totalCount, // Total items
			LastPage:  lastPage,   // Last page
		},
	}

	// Write solutionsJSON to response
	SendJSONResponse(w, http.StatusOK, response)
}

// POST: /
func (h *Handler) CreateSolution(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	solution, apiErr := DecodeJSONRequest[CreateSolutionRequest](r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if solution.Title == "" || solution.Body == "" || solution.ProblemId <= 0 {
		SendError(w, http.StatusBadRequest, "Missing required fields for solution creation")
		return
	}

	// Insert solution into db
	_, err = h.PostgresQueries.CreateSolution(r.Context(), sql.CreateSolutionParams{
		UserID: pgtype.Text{
			String: userId,
			Valid:  true,
		},
		Title:     solution.Title,
		Tags:      solution.Tags,
		Body:      solution.Body,
		ProblemID: pgtype.Int8{Int64: solution.ProblemId, Valid: true},
	})
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Error creating solution in database")
		return
	}

	// Write response
	SendJSONResponse(w, http.StatusCreated, nil)
}

// GET: /{solutionId}
func (h *Handler) GetSolution(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	solutionId := chi.URLParam(r, "solutionId")
	if solutionId == "" {
		SendError(w, http.StatusBadRequest, "Missing solutionId for solution retrieval")
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid solutionId format for solution retrieval")
		return
	}

	// Get solutions from db by idPg
	solution, err := h.PostgresQueries.GetSolution(r.Context(), sql.GetSolutionParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		EmptyDataResponse(w) // { data: {} }
		return
	}

	// If tags is nil, set it to an empty array
	if solution.Tags == nil {
		solution.Tags = []string{}
	}

	solutionData := SolutionsData{
		Id:              solution.ID,
		Body:            solution.Body,
		Comments:        solution.CommentsCount,
		Date:            solution.CreatedAt,
		Tags:            solution.Tags,
		Title:           solution.Title,
		Username:        solution.UserUsername,
		Level:           solution.UserLevel,
		AvatarUrl:       solution.UserAvatarUrl.String,
		Votes:           solution.VotesCount,
		CurrentUserVote: solution.UserVote,
		Starred:         solution.Starred,
	}

	response := SolutionResponse{
		Data: solutionData,
	}

	// Write solutionsJSON to response
	SendJSONResponse(w, http.StatusOK, response)
}

// PUT: /{solutionId}
func (h *Handler) UpdateSolution(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Handle problemId query parameter
	solutionId := chi.URLParam(r, "solutionId")
	// If problemId is empty, set idPg as NULL
	if solutionId == "" {
		SendError(w, http.StatusBadRequest, "solutionId is required")
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "solutionId must be an integer")
		return
	}

	solutionRequest, apiErr := DecodeJSONRequest[UpdateSolutionRequest](r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if solutionRequest.Title == "" && solutionRequest.Body == "" && len(solutionRequest.Tags) > 0 {
		SendError(w, http.StatusBadRequest, "at least one field must be provided")
		return
	}

	solutionArgs := sql.UpdateSolutionParams{
		ID:     id,
		Title:  solutionRequest.Title,
		Body:   solutionRequest.Body,
		Tags:   solutionRequest.Tags,
		UserID: pgtype.Text{String: userId, Valid: true},
	}

	// Get solutions from db by idPg
	_, err = h.PostgresQueries.UpdateSolution(r.Context(), solutionArgs)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "error getting solutions")
		return
	}

	// Write solutionsJSON to response
	SendJSONResponse(w, http.StatusNoContent, nil)
}

// DELETE: /{solutionId}
func (h *Handler) DeleteSolution(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Handle problemId query parameter
	solutionId := chi.URLParam(r, "solutionId")
	// If problemId is empty, set idPg as NULL
	if solutionId == "" {
		SendError(w, http.StatusBadRequest, "solutionId is required")
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "solutionId must be an integer")
		return
	}

	// Get solutions from db by idPg
	err = h.PostgresQueries.DeleteSolution(r.Context(), sql.DeleteSolutionParams{
		ID:     id,
		UserID: pgtype.Text{String: userId, Valid: true},
	})
	if err != nil {
		SendError(w, http.StatusInternalServerError, "error getting solutions")
		return
	}

	// Write solutionsJSON to response
	SendJSONResponse(w, http.StatusNoContent, nil)
}

// PATCH: /{solutionId}/vote
func (h *Handler) VoteSolution(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Extract solutionId from URL parameters
	solutionId := chi.URLParam(r, "solutionId")
	if solutionId == "" {
		SendError(w, http.StatusBadRequest, "Missing solutionId for solution retrieval")
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid solutionId format for solution retrieval")
		return
	}

	// Decode the request body into VoteRequest struct
	req, apiErr := DecodeJSONRequest[VoteRequest](r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if req.Vote == "" {
		SendError(w, http.StatusBadRequest, "Vote is required")
		return
	}

	// Check if the solution exists
	_, err = h.PostgresQueries.GetSolutionById(r.Context(), id)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Solution not found")
		return
	}

	err = h.PostgresQueries.VoteSolution(r.Context(), sql.VoteSolutionParams{
		UserID:     userId,
		SolutionID: id,
		Vote:       req.Vote,
	})
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Error voting on solution")
		return
	}

	SendJSONResponse(w, http.StatusNoContent, nil)
}
