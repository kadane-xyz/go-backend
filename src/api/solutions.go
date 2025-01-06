package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
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

type SolutionResponse struct {
	Data SolutionsData `json:"data"`
}

type SolutionsData struct {
	Id              int64            `json:"id"`
	Body            string           `json:"body,omitempty"`
	Comments        int64            `json:"comments"`
	Date            pgtype.Timestamp `json:"date"`
	Tags            []string         `json:"tags"`
	Title           string           `json:"title"`
	Username        string           `json:"username,omitempty"`
	Level           int32            `json:"level,omitempty"`
	AvatarUrl       string           `json:"avatarUrl,omitempty"`
	Votes           int64            `json:"votes"`
	CurrentUserVote sql.VoteType     `json:"currentUserVote"`
	Starred         bool             `json:"starred"`
}

type SolutionsResponse struct {
	Data       []SolutionsData `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// GET: /solutions
func (h *Handler) GetSolutions(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for solutions retrieval")
		return
	}

	problemId := r.URL.Query().Get("problemId")
	if problemId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing problemId for solutions retrieval")
		return
	}

	id, err := strconv.ParseInt(problemId, 10, 64)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problemId format for solutions retrieval")
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
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	// Handle perPage
	perPage, err := strconv.ParseInt(r.URL.Query().Get("perPage"), 10, 64)
	if err != nil || perPage < 1 {
		perPage = 10 // Default to 10 items per page
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

	// Calculate offset
	offset := (page - 1) * perPage

	solutions, err := h.PostgresQueries.GetSolutionsPaginated(r.Context(), sql.GetSolutionsPaginatedParams{
		ProblemID:     pgtype.Int8{Int64: id, Valid: true},
		Tags:          tagsArray,
		Title:         titleSearch,
		PerPage:       int32(perPage),
		Page:          int32(offset),
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

	totalCount, err := h.PostgresQueries.GetSolutionsCount(r.Context(), sql.GetSolutionsCountParams{
		ProblemID: pgtype.Int8{Int64: id, Valid: true},
		Title:     titleSearch,
		Tags:      tagsArray,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "error getting total count")
		return
	}

	// Prepare response
	var solutionsData []SolutionsData
	for _, solution := range solutions {
		if err != nil {
			apierror.SendError(w, http.StatusInternalServerError, "error getting level")
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
			Level:           solution.UserLevel.Int32,
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// POST: /
func (h *Handler) CreateSolution(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for solution creation")
		return
	}

	var solution Solutions
	err := json.NewDecoder(r.Body).Decode(&solution)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid solution data format")
		return
	}

	if solution.Title == "" || solution.Body == "" || solution.ProblemId <= 0 {
		apierror.SendError(w, http.StatusBadRequest, "Missing required fields for solution creation")
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
		apierror.SendError(w, http.StatusInternalServerError, "Error creating solution in database")
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GET: /{solutionId}
func (h *Handler) GetSolution(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for solution retrieval")
		return
	}

	solutionId := chi.URLParam(r, "solutionId")
	if solutionId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing solutionId for solution retrieval")
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid solutionId format for solution retrieval")
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
		Level:           solution.UserLevel.Int32,
		AvatarUrl:       solution.UserAvatarUrl.String,
		Votes:           solution.VotesCount,
		CurrentUserVote: solution.UserVote,
		Starred:         solution.Starred,
	}

	response := SolutionResponse{
		Data: solutionData,
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PUT: /{solutionId}
func (h *Handler) UpdateSolution(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	// Handle problemId query parameter
	solutionId := chi.URLParam(r, "solutionId")
	// If problemId is empty, set idPg as NULL
	if solutionId == "" {
		http.Error(w, "solutionId is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		http.Error(w, "solutionId must be an integer", http.StatusBadRequest)
		return
	}

	var solutionRequest sql.UpdateSolutionParams
	if err := json.NewDecoder(r.Body).Decode(&solutionRequest); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if solutionRequest.Title == "" && solutionRequest.Body == "" && len(solutionRequest.Tags) > 0 {
		http.Error(w, "at least one field must be provided", http.StatusBadRequest)
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
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// DELETE: /{solutionId}
func (h *Handler) DeleteSolution(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	// Handle problemId query parameter
	solutionId := chi.URLParam(r, "solutionId")
	// If problemId is empty, set idPg as NULL
	if solutionId == "" {
		http.Error(w, "solutionId is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		http.Error(w, "solutionId must be an integer", http.StatusBadRequest)
		return
	}

	// Get solutions from db by idPg
	err = h.PostgresQueries.DeleteSolution(r.Context(), sql.DeleteSolutionParams{
		ID:     id,
		UserID: pgtype.Text{String: userId, Valid: true},
	})
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

type VoteRequest struct {
	Vote string `json:"vote"`
}

// PATCH: /{solutionId}/vote
func (h *Handler) VoteSolution(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	// Extract solutionId from URL parameters
	solutionId := chi.URLParam(r, "solutionId")
	if solutionId == "" {
		http.Error(w, "solutionId is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		http.Error(w, "solutionId must be an integer", http.StatusBadRequest)
		return
	}

	// Decode the request body into VoteRequest struct
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if req.Vote == "" {
		http.Error(w, "vote is required", http.StatusBadRequest)
		return
	}

	// Check if the solution exists
	_, err = h.PostgresQueries.GetSolution(r.Context(), sql.GetSolutionParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		http.Error(w, "solution not found", http.StatusBadRequest)
		return
	}

	// Prepare parameters to get the existing vote
	solutionArgs := sql.GetSolutionVoteParams{
		SolutionID: id,
		UserID:     userId,
	}

	// Get the existing vote
	existingVote, err := h.PostgresQueries.GetSolutionVote(r.Context(), solutionArgs)
	if err != nil {
		// If there's no existing vote and the requested vote is 'none', nothing to do
		if req.Vote == "none" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Insert the new vote
		insertArgs := sql.InsertSolutionVoteParams{
			UserID:     userId,
			SolutionID: id,
			Vote:       sql.VoteType(req.Vote),
		}
		if err := h.PostgresQueries.InsertSolutionVote(r.Context(), insertArgs); err != nil {
			http.Error(w, "error inserting vote", http.StatusInternalServerError)
			return
		}
	} else {
		// Handle vote update or deletion
		if req.Vote == "none" {
			// Delete the existing vote
			deleteArgs := sql.DeleteSolutionVoteParams{
				UserID:     userId,
				SolutionID: id,
			}
			if err := h.PostgresQueries.DeleteSolutionVote(r.Context(), deleteArgs); err != nil {
				http.Error(w, "error deleting vote", http.StatusInternalServerError)
				return
			}
		} else if existingVote != sql.VoteType(req.Vote) {
			// Update the vote if it's different
			updateArgs := sql.UpdateSolutionVoteParams{
				UserID:     userId,
				SolutionID: id,
				Vote:       sql.VoteType(req.Vote),
			}
			if err := h.PostgresQueries.UpdateSolutionVote(r.Context(), updateArgs); err != nil {
				http.Error(w, "error updating vote", http.StatusInternalServerError)
				return
			}
		} else {
			// Vote is the same; no action needed
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// Send the updated solution as the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
