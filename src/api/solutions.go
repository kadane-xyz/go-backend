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
	ID    int      `json:"id"`
	Email string   `json:"email"`
	Title string   `json:"title"`
	Date  string   `json:"date"`
	Tags  []string `json:"tags"`
	Body  string   `json:"body"`
	//Comments []Comments `json:"comments"`
	Votes int `json:"votes"`
}

// GET: /
func (h *Handler) GetSolutions(w http.ResponseWriter, r *http.Request) {
	var idPg pgtype.Int8

	// Handle problemID query parameter
	problemID := r.URL.Query().Get("problemID")
	// If problemID is empty, set idPg as NULL
	if problemID == "" {
		idPg = pgtype.Int8{
			Valid: false,
		}
	} else {
		id, err := strconv.ParseInt(problemID, 10, 64)
		if err != nil {
			http.Error(w, "problemID must be an integer", http.StatusBadRequest)
			return
		}

		idPg = pgtype.Int8{
			Int64: id,
			Valid: true,
		}
	}

	// Get solutions from db by idPg
	solutions, err := h.PostgresQueries.GetSolutionsWithCommentsCount(r.Context(), idPg)
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	// Marshal solutions to JSON
	solutionsJSON, err := json.Marshal(solutions)
	if err != nil {
		http.Error(w, "error marshalling solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(solutionsJSON)
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

	// Insert solution into db
	_, err = h.PostgresQueries.CreateSolution(r.Context(), sql.CreateSolutionParams{
		Email: pgtype.Text{
			String: solution.Email,
			Valid:  true,
		},
		Title: solution.Title,
		Tags:  solution.Tags,
		Body:  solution.Body,
	})
	if err != nil {
		http.Error(w, "error creating solution", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "solution created"}`))
}

// GET: /{solutionID}
func (h *Handler) GetSolution(w http.ResponseWriter, r *http.Request) {
	// Handle problemID query parameter
	solutionID := chi.URLParam(r, "solutionID")
	// If problemID is empty, set idPg as NULL
	if solutionID == "" {
		http.Error(w, "solutionID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(solutionID, 10, 64)
	if err != nil {
		http.Error(w, "solutionID must be an integer", http.StatusBadRequest)
		return
	}

	// Get solutions from db by idPg
	solutions, err := h.PostgresQueries.GetSolution(r.Context(), id)
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	// Marshal solutions to JSON
	solutionsJSON, err := json.Marshal(solutions)
	if err != nil {
		http.Error(w, "error marshalling solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(solutionsJSON)
}

// PUT: /{solutionID}
func (h *Handler) UpdateSolution(w http.ResponseWriter, r *http.Request) {
	// Handle problemID query parameter
	solutionID := chi.URLParam(r, "solutionID")
	// If problemID is empty, set idPg as NULL
	if solutionID == "" {
		http.Error(w, "solutionID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(solutionID, 10, 64)
	if err != nil {
		http.Error(w, "solutionID must be an integer", http.StatusBadRequest)
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
	solutions, err := h.PostgresQueries.UpdateSolution(r.Context(), solutionArgs)
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	// Marshal solutions to JSON
	solutionsJSON, err := json.Marshal(solutions)
	if err != nil {
		http.Error(w, "error marshalling solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(solutionsJSON)
}

// DELETE: /{solutionID}
func (h *Handler) DeleteSolution(w http.ResponseWriter, r *http.Request) {
	// Handle problemID query parameter
	solutionID := chi.URLParam(r, "solutionID")
	// If problemID is empty, set idPg as NULL
	if solutionID == "" {
		http.Error(w, "solutionID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(solutionID, 10, 64)
	if err != nil {
		http.Error(w, "solutionID must be an integer", http.StatusBadRequest)
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

// PATCH: /{solutionID}/vote
func (h *Handler) VoteSolution(w http.ResponseWriter, r *http.Request) {
	// Handle problemID query parameter
	solutionID := chi.URLParam(r, "solutionID")
	// If problemID is empty, set idPg as NULL
	if solutionID == "" {
		http.Error(w, "solutionID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(solutionID, 10, 64)
	if err != nil {
		http.Error(w, "solutionID must be an integer", http.StatusBadRequest)
		return
	}

	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Vote == "" {
		http.Error(w, "email and vote are required", http.StatusBadRequest)
		return
	}

	//add checks for valid email and solution id

	solutionArgs := sql.GetSolutionVoteParams{
		Email:      req.Email,
		SolutionID: id,
	}

	existingVote, err := h.PostgresQueries.GetSolutionVote(r.Context(), solutionArgs)
	if err != nil {
		insertArgs := sql.InsertSolutionVoteParams{
			Email:      req.Email,
			SolutionID: id,
			Vote:       sql.VoteType(req.Vote),
		}
		if err := h.PostgresQueries.InsertSolutionVote(r.Context(), insertArgs); err != nil {
			http.Error(w, "error inserting vote", http.StatusInternalServerError)
			return
		}
	} else if existingVote != sql.VoteType(req.Vote) {
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
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.PostgresQueries.UpVoteSolution(r.Context(), id)
	if err != nil {
		http.Error(w, "error getting solutions", http.StatusInternalServerError)
		return
	}

	// Write solutionsJSON to response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
