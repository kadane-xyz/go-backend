package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Solutions struct {
	Id    int64    `json:"id"`
	Email string   `json:"email"`
	Title string   `json:"title"`
	Date  string   `json:"date"`
	Tags  []string `json:"tags"`
	Body  string   `json:"body"`
	//Comments []Comments `json:"comments"`
	Votes     int         `json:"votes"`
	ProblemId pgtype.Int8 `json:"problemId"`
}

// GET: /solutions
func (h *Handler) GetSolutions(w http.ResponseWriter, r *http.Request) {
	var id *int64

	// Handle problemId query parameter
	problemId := r.URL.Query().Get("problemId")
	// If problemId is empty, set idPg as NULL
	if problemId == "" {
		id = nil
		http.Error(w, "problemId is required", http.StatusBadRequest)
		return
	} else {
		parsedId, err := strconv.ParseInt(problemId, 10, 64)
		if err != nil {
			http.Error(w, "problemId must be an integer", http.StatusBadRequest)
			return
		}
		id = &parsedId
	}

	// Get solutions from db by idPg
	solutions, err := h.PostgresQueries.GetSolutionsWithCommentsCount(r.Context(), pgtype.Int8{
		Int64: *id,
		Valid: id != nil,
	})
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data": solutions,
	}

	// Marshal solutions to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "error marshalling solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// POST: /
func (h *Handler) CreateSolution(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var solution Solutions
	err := json.NewDecoder(r.Body).Decode(&solution)
	if err != nil {
		log.Println(err)
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if solution.Email == "" || solution.Title == "" || solution.Body == "" || !solution.ProblemId.Valid {
		http.Error(w, "email, title, body, and problemId are required", http.StatusBadRequest)
		return
	}

	// Insert solution into db
	_, err = h.PostgresQueries.CreateSolution(r.Context(), sql.CreateSolutionParams{
		Email: pgtype.Text{
			String: solution.Email,
			Valid:  true,
		},
		Title:     solution.Title,
		Tags:      solution.Tags,
		Body:      solution.Body,
		ProblemID: solution.ProblemId,
	})
	if err != nil {
		http.Error(w, "error creating solution", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GET: /{solutionId}
func (h *Handler) GetSolution(w http.ResponseWriter, r *http.Request) {
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
	solutions, err := h.PostgresQueries.GetSolution(r.Context(), id)
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data": solutions,
	}

	// Marshal solutions to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "error marshalling solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// PUT: /{solutionId}
func (h *Handler) UpdateSolution(w http.ResponseWriter, r *http.Request) {
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
		ID:    id,
		Title: solutionRequest.Title,
		Body:  solutionRequest.Body,
		Tags:  solutionRequest.Tags,
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
	err = h.PostgresQueries.DeleteSolution(r.Context(), id)
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

type VoteRequest struct {
	Email string `json:"email"`
	Vote  string `json:"vote"`
}

// PATCH: /{solutionId}/vote
func (h *Handler) VoteSolution(w http.ResponseWriter, r *http.Request) {
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

	if req.Email == "" || req.Vote == "" {
		http.Error(w, "email and vote are required", http.StatusBadRequest)
		return
	}

	// Validate email and solution Id
	// Check if the user exists
	/*_, err = h.PostgresQueries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}*/

	// Check if the solution exists
	_, err = h.PostgresQueries.GetSolution(r.Context(), id)
	if err != nil {
		http.Error(w, "solution not found", http.StatusBadRequest)
		return
	}

	// Prepare parameters to get the existing vote
	solutionArgs := sql.GetSolutionVoteParams{
		Email:      req.Email,
		SolutionID: id,
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
			Email:      req.Email,
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
				Email:      req.Email,
				SolutionID: id,
			}
			if err := h.PostgresQueries.DeleteSolutionVote(r.Context(), deleteArgs); err != nil {
				http.Error(w, "error deleting vote", http.StatusInternalServerError)
				return
			}
		} else if existingVote != sql.VoteType(req.Vote) {
			// Update the vote if it's different
			updateArgs := sql.UpdateSolutionVoteParams{
				Email:      req.Email,
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
