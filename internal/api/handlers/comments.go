package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
)

type CommentHandler struct {
	repo          *repository.SQLCommentsRepository
	solutionsRepo *repository.SQLSolutionsRepository
}

func NewCommentHandler(repo *repository.SQLCommentsRepository, solutionsRepo *repository.SQLSolutionsRepository) *CommentHandler {
	return &CommentHandler{repo: repo, solutionsRepo: solutionsRepo}
}

// GET: /comments
func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Get the solutionId from the query parameters
	solutionId := r.URL.Query().Get("solutionId")
	if solutionId == "" {
		errors.SendError(w, http.StatusBadRequest, "Missing solutionId for comment retrieval")
		return
	}

	// Convert solutionId to int64
	id, err := strconv.ParseInt(solutionId, 10, 64)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Invalid solutionId format for comment retrieval")
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

	dbComments, err := h.repo.GetCommentsSorted(r.Context(), sql.GetCommentsSortedParams{
		SolutionID:    id,
		UserID:        userId,
		Sort:          sort,
		SortDirection: order,
	})
	if err != nil {
		httputils.EmptyDataArrayResponse(w)
		return
	}

	// Create a map to hold all comments by ID
	commentMap := make(map[int64]*domain.CommentRelation, len(dbComments))

	// Create a slice to maintain the order of top-level comments
	var topLevelComments []*domain.CommentRelation

	// First pass: Create CommentsData objects
	for _, dbComment := range dbComments {
		comment := &domain.CommentRelation{
			Comment: domain.Comment{
				ID:         dbComment.ID,
				SolutionId: dbComment.SolutionId,
				Body:       dbComment.Body,
				CreatedAt:  dbComment.CreatedAt,
				Votes:      int32(dbComment.CurrentUserVote),
				Children:   []*domain.Comment{},
			},
			CurrentUserVote: dbComment.CurrentUserVote,
			Username:        dbComment.Username,
			AvatarUrl:       dbComment.AvatarUrl,
			Level:           dbComment.Level,
		}
		commentMap[comment.ID] = comment

		if dbComment.ParentId != nil {
			topLevelComments = append(topLevelComments, comment)
		}
	}

	// Second pass: Build the comment tree
	for _, dbComment := range dbComments {
		if dbComment.ParentId != nil {
			parentId := dbComment.ParentId
			if parent, exists := commentMap[*parentId]; exists {
				parent.Children = append(parent.Children, commentMap[dbComment.ID])
			}
		}
	}

	response := topLevelComments

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}

// POST: /comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	comment, apiErr := httputils.DecodeJSONRequest[domain.CommentCreateRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	// Validate input
	if comment.SolutionId == 0 || comment.Body == "" {
		errors.SendError(w, http.StatusBadRequest, "Missing required fields for comment creation")
		return
	}

	// Check if solution exists
	_, err = h.solutionsRepo.GetSolutionById(r.Context(), comment.SolutionId)
	if err != nil {
		errors.SendError(w, http.StatusNotFound, "Solution not found")
		return
	}

	// create comment
	_, err = h.repo.CreateComment(r.Context(), domain.CommentCreateParams{
		SolutionID: comment.SolutionId,
		ParentID:   comment.ParentId,
		UserID:     userId,
		Body:       comment.Body,
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	httputils.SendJSONResponse(w, http.StatusCreated, nil)
}

// GET: /comments/{commentId}
func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		errors.SendError(w, http.StatusBadRequest, "Missing commentId for comment retrieval")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Invalid commentId format for comment retrieval")
		return
	}

	comment, err := h.repo.GetComment(r.Context(), sql.GetCommentParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		httputils.EmptyDataResponse(w)
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, comment)
}

// PUT: /comments/{commentId}
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		errors.SendError(w, http.StatusBadRequest, "Missing commentId")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Invalid commentId format")
		return
	}

	comment, apiErr := httputils.DecodeJSONRequest[domain.CommentUpdateRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	// Validate input
	if comment.Body == "" {
		errors.SendError(w, http.StatusBadRequest, "Body is required")
		return
	}

	_, err = h.repo.UpdateComment(r.Context(), sql.UpdateCommentParams{
		ID:     id,
		Body:   comment.Body,
		UserID: userId, // Check if the user is the owner of the comment
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)
}

// DELETE: /comments/{commentId}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		errors.SendError(w, http.StatusBadRequest, "Missing commentId")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Invalid commentId format")
		return
	}

	err = h.repo.DeleteComment(r.Context(), sql.DeleteCommentParams{
		ID:     id,
		UserID: userId, // Check if the user is the owner of the comment
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)
}

// PATCH: /{commentId}/vote
func (h *CommentHandler) VoteComment(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Extract commentId from URL parameters
	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		errors.SendError(w, http.StatusBadRequest, "Missing commentId")
		return
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Invalid commentId format")
		return
	}

	// Decode the request body into VoteRequest struct
	req, apiErr := httputils.DecodeJSONRequest[domain.VoteRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	if req.Vote == "" {
		errors.SendError(w, http.StatusBadRequest, "Vote is required")
		return
	}

	// Check if the comment exists
	_, err = h.repo.GetCommentById(r.Context(), id)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Comment not found")
		return
	}

	err = h.repo.VoteComment(r.Context(), sql.VoteCommentParams{
		UserID:    userId,
		CommentID: id,
		Vote:      sql.VoteType(req.Vote),
	})
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Error voting on comment")
		return
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)
}
