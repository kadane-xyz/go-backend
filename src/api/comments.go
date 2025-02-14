package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Comment struct {
	ID         int64      `json:"id"`
	SolutionId int64      `json:"solutionId"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	Votes      int32      `json:"votes"`
	ParentId   int64      `json:"parentId,omitempty"`
	Children   []*Comment `json:"children,omitempty"` // For nested child comments
}

type CommentCreateRequest struct {
	SolutionId int64  `json:"solutionId"`
	Body       string `json:"body"`
	ParentId   *int64 `json:"parentId,omitempty"`
}

type CommentUpdateRequest struct {
	Body string `json:"body"`
}

type CommentResponse struct {
	Data CommentsData `json:"data"`
}

type CommentsData struct {
	ID              int64           `json:"id"`
	SolutionId      int64           `json:"solutionId"`
	Username        string          `json:"username"`
	AvatarUrl       string          `json:"avatarUrl,omitempty"`
	Level           int32           `json:"level"`
	Body            string          `json:"body"`
	CreatedAt       time.Time       `json:"createdAt"`
	Votes           int32           `json:"votes"`
	ParentId        *int64          `json:"parentId,omitempty"`
	Children        []*CommentsData `json:"children"` // For nested child comments
	CurrentUserVote sql.VoteType    `json:"currentUserVote"`
}

type CommentsResponse struct {
	Data []*CommentsData `json:"data"`
}

// GET: /comments
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Get the solutionId from the query parameters
	solutionId := r.URL.Query().Get("solutionId")
	if solutionId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing solutionId for comment retrieval")
		return
	}

	// Convert solutionId to int64
	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid solutionId format for comment retrieval")
		return
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

	dbComments, err := h.PostgresQueries.GetCommentsSorted(r.Context(), sql.GetCommentsSortedParams{
		SolutionID:    id,
		UserID:        userId,
		Sort:          sort,
		SortDirection: order,
	})
	if err != nil {
		EmptyDataArrayResponse(w)
		return
	}

	// Create a map to hold all comments by ID
	commentMap := make(map[int64]*CommentsData, len(dbComments))

	// Create a slice to maintain the order of top-level comments
	var topLevelComments []*CommentsData

	// First pass: Create CommentsData objects
	for _, dbComment := range dbComments {
		comment := &CommentsData{
			ID:              dbComment.ID,
			SolutionId:      dbComment.SolutionID,
			Username:        dbComment.UserUsername,
			AvatarUrl:       dbComment.UserAvatarUrl.String,
			Level:           dbComment.UserLevel.Int32,
			Body:            dbComment.Body,
			CreatedAt:       dbComment.CreatedAt.Time,
			Votes:           int32(dbComment.VotesCount),
			Children:        []*CommentsData{},
			CurrentUserVote: dbComment.UserVote,
		}
		commentMap[comment.ID] = comment

		if !dbComment.ParentID.Valid {
			topLevelComments = append(topLevelComments, comment)
		}
	}

	// Second pass: Build the comment tree
	for _, dbComment := range dbComments {
		if dbComment.ParentID.Valid {
			parentId := dbComment.ParentID.Int64
			if parent, exists := commentMap[parentId]; exists {
				parent.Children = append(parent.Children, commentMap[dbComment.ID])
			}
		}
	}

	response := CommentsResponse{
		Data: topLevelComments,
	}

	// Set JSON response headers and encode the response directly
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST: /comments
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	var comment CommentCreateRequest
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid comment data format")
		return
	}

	// Validate input
	if comment.SolutionId == 0 || comment.Body == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing required fields for comment creation")
		return
	}

	// Check if solution exists
	_, err = h.PostgresQueries.GetSolutionById(r.Context(), comment.SolutionId)
	if err != nil {
		apierror.SendError(w, http.StatusNotFound, "Solution not found")
		return
	}

	// Check if parent comment exists if ParentId is provided
	if comment.ParentId != nil {
		_, err := h.PostgresQueries.GetCommentById(r.Context(), *comment.ParentId)
		if err != nil {
			apierror.SendError(w, http.StatusNotFound, "Parent comment not found")
			return
		}
	}

	var parentId pgtype.Int8
	if comment.ParentId != nil {
		parentId = pgtype.Int8{Int64: *comment.ParentId, Valid: true}
	} else {
		parentId = pgtype.Int8{Valid: false}
	}

	// create comment
	_, err = h.PostgresQueries.CreateComment(r.Context(), sql.CreateCommentParams{
		SolutionID: comment.SolutionId,
		ParentID:   parentId,
		UserID:     userId,
		Body:       comment.Body,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GET: /comments/{commentId}
func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing commentId for comment retrieval")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid commentId format for comment retrieval")
		return
	}

	comment, err := h.PostgresQueries.GetComment(r.Context(), sql.GetCommentParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		EmptyDataResponse(w)
		return
	}

	commentData := CommentsData{
		ID:              comment.ID,
		SolutionId:      comment.SolutionID,
		Username:        comment.UserUsername,
		AvatarUrl:       comment.UserAvatarUrl.String,
		Level:           comment.UserLevel.Int32,
		Body:            comment.Body,
		CreatedAt:       comment.CreatedAt.Time,
		Votes:           comment.Votes.Int32,
		ParentId:        &comment.ParentID.Int64,
		Children:        []*CommentsData{},
		CurrentUserVote: comment.UserVote,
	}

	response := CommentResponse{
		Data: commentData,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to marshal comment response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// PUT: /comments/{commentId}
func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing commentId")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid commentId format")
		return
	}

	var comment CommentUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid comment data format")
		return
	}

	// Validate input
	if comment.Body == "" {
		apierror.SendError(w, http.StatusBadRequest, "Body is required")
		return
	}

	_, err = h.PostgresQueries.UpdateComment(r.Context(), sql.UpdateCommentParams{
		ID:     id,
		Body:   comment.Body,
		UserID: userId, // Check if the user is the owner of the comment
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// DELETE: /comments/{commentId}
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing commentId")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid commentId format")
		return
	}

	err = h.PostgresQueries.DeleteComment(r.Context(), sql.DeleteCommentParams{
		ID:     id,
		UserID: userId, // Check if the user is the owner of the comment
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// PATCH: /{commentId}/vote
func (h *Handler) VoteComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Extract commentId from URL parameters
	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing commentId")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid commentId format")
		return
	}

	// Decode the request body into VoteRequest struct
	var req VoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Error decoding request body")
		return
	}

	if req.Vote == "" {
		apierror.SendError(w, http.StatusBadRequest, "Vote is required")
		return
	}

	// Check if the comment exists
	_, err = h.PostgresQueries.GetCommentById(r.Context(), id)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Comment not found")
		return
	}

	err = h.PostgresQueries.VoteComment(r.Context(), sql.VoteCommentParams{
		UserID:    userId,
		CommentID: id,
		Vote:      sql.VoteType(req.Vote),
	})
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Error voting on comment")
		return
	}

	// Send the updated comment as the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
