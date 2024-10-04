package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Comment struct {
	ID         int64            `json:"id"`
	SolutionId int64            `json:"solutionId"`
	Username   string           `json:"username"`
	Body       string           `json:"body"`
	CreatedAt  pgtype.Timestamp `json:"createdAt"`
	Votes      pgtype.Int4      `json:"votes"`
	ParentId   pgtype.Int8      `json:"parentId,omitempty"`
	Children   []*Comment       `json:"children,omitempty"` // For nested child comments
}

// GET: /comments
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	// Get the solutionId from the query parameters
	solutionId := r.URL.Query().Get("solutionId")
	if solutionId == "" {
		http.Error(w, "Missing solutionId", http.StatusBadRequest)
		return
	}

	// Convert solutionId to int64
	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		http.Error(w, "solutionId must be an integer", http.StatusBadRequest)
		return
	}

	// Get all comments associated with the given solutionId
	dbComments, err := h.PostgresQueries.GetComments(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create maps to track all comments and top-level comments
	allComments := make(map[int64]*Comment)
	var topLevelComments []*Comment

	for _, dbComment := range dbComments {
		comment := &Comment{
			ID:         dbComment.ID,
			SolutionId: dbComment.SolutionID,
			Username:   dbComment.Username,
			Body:       dbComment.Body,
			CreatedAt:  dbComment.CreatedAt,
			Votes:      dbComment.Votes,
			ParentId:   dbComment.ParentID,
		}

		allComments[comment.ID] = comment
	}

	for _, comment := range allComments {
		if !comment.ParentId.Valid {
			topLevelComments = append(topLevelComments, comment)
		} else {
			if parent, ok := allComments[comment.ParentId.Int64]; ok {
				parent.Children = append(parent.Children, comment)
			}
		}
	}

	response := map[string]interface{}{
		"data": topLevelComments,
	}

	// Marshal the structured comments into JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// POST: /comments
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate input
	if comment.SolutionId == 0 || comment.Username == "" || comment.Body == "" {
		http.Error(w, "SolutionId, Username, and Body are required", http.StatusBadRequest)
		return
	}

	// Check if solution exists
	_, err = h.PostgresQueries.GetSolution(r.Context(), comment.SolutionId)
	if err != nil {
		http.Error(w, "Solution not found", http.StatusNotFound)
		return
	}

	// Check if parent comment exists if ParentId is provided
	if comment.ParentId.Valid {
		_, err := h.PostgresQueries.GetComment(r.Context(), comment.ParentId.Int64)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Parent comment not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error checking parent comment: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	// create comment
	_, err = h.PostgresQueries.CreateComment(r.Context(), sql.CreateCommentParams{
		SolutionID: comment.SolutionId,
		ParentID:   comment.ParentId,
		Username:   comment.Username,
		Body:       comment.Body,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GET: /comments/{commentId}
func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request) {
	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		http.Error(w, "Missing commentId", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		http.Error(w, "commentId must be an integer", http.StatusBadRequest)
		return
	}

	comment, err := h.PostgresQueries.GetComment(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Comment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := map[string]interface{}{
		"data": comment,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// PUT: /comments/{commentId}
func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		http.Error(w, "Missing commentId", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		http.Error(w, "commentId must be an integer", http.StatusBadRequest)
		return
	}

	var comment Comment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate input
	if comment.Username == "" || comment.Body == "" {
		http.Error(w, "Username and Body are required", http.StatusBadRequest)
		return
	}

	_, err = h.PostgresQueries.UpdateComment(r.Context(), sql.UpdateCommentParams{
		ID:   id,
		Body: comment.Body,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Comment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// DELETE: /comments/{commentId}
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		http.Error(w, "Missing commentId", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		http.Error(w, "commentId must be an integer", http.StatusBadRequest)
		return
	}

	err = h.PostgresQueries.DeleteComment(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Comment not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// PATCH: /{commentId}/vote
func (h *Handler) VoteComment(w http.ResponseWriter, r *http.Request) {
	// Extract commentId from URL parameters
	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		http.Error(w, "commentId is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		http.Error(w, "commentId must be an integer", http.StatusBadRequest)
		return
	}

	// Decode the request body into VoteRequest struct
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Vote == "" {
		http.Error(w, "username and vote are required", http.StatusBadRequest)
		return
	}

	// Validate the vote value
	validVotes := map[string]bool{"up": true, "down": true, "none": true}
	if !validVotes[req.Vote] {
		http.Error(w, "invalid vote type", http.StatusBadRequest)
		return
	}

	// Validate username and comment ID
	// Check if the user exists
	/*_, err = h.PostgresQueries.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}*/

	// Check if the comment exists
	_, err = h.PostgresQueries.GetComment(r.Context(), id)
	if err != nil {
		http.Error(w, "comment not found", http.StatusBadRequest)
		return
	}

	// Prepare parameters to get the existing vote
	commentArgs := sql.GetCommentVoteParams{
		Username: pgtype.Text{
			String: req.Username,
			Valid:  true,
		},
		CommentID: pgtype.Int8{
			Int64: id,
			Valid: true,
		},
	}

	// Get the existing vote
	existingVote, err := h.PostgresQueries.GetCommentVote(r.Context(), commentArgs)
	if err != nil {
		// No existing vote
		if req.Vote == "none" {
			// Nothing to delete; return OK
			w.WriteHeader(http.StatusOK)
			return
		}
		// Insert the new vote
		insertArgs := sql.InsertCommentVoteParams{
			Username: pgtype.Text{
				String: req.Username,
				Valid:  true,
			},
			CommentID: pgtype.Int8{
				Int64: id,
				Valid: true,
			},
			Vote: sql.VoteType(req.Vote),
		}
		if err := h.PostgresQueries.InsertCommentVote(r.Context(), insertArgs); err != nil {
			http.Error(w, "error inserting vote", http.StatusInternalServerError)
			return
		}
	} else {
		// Existing vote found
		if req.Vote == "none" {
			// Delete the existing vote
			deleteArgs := sql.DeleteCommentVoteParams{
				Username: pgtype.Text{
					String: req.Username,
					Valid:  true,
				},
				CommentID: pgtype.Int8{
					Int64: id,
					Valid: true,
				},
			}
			if err := h.PostgresQueries.DeleteCommentVote(r.Context(), deleteArgs); err != nil {
				http.Error(w, "error deleting vote", http.StatusInternalServerError)
				return
			}
		} else if existingVote != sql.VoteType(req.Vote) {
			// Update the vote if it's different
			updateArgs := sql.UpdateCommentVoteParams{
				Username: pgtype.Text{
					String: req.Username,
					Valid:  true,
				},
				CommentID: pgtype.Int8{
					Int64: id,
					Valid: true,
				},
				Vote: sql.VoteType(req.Vote),
			}
			if err := h.PostgresQueries.UpdateCommentVote(r.Context(), updateArgs); err != nil {
				http.Error(w, "error updating vote", http.StatusInternalServerError)
				return
			}
		} else {
			// Vote is the same; no action needed
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// Send the updated comment as the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
