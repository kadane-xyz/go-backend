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
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type CommentHandler struct {
	repo          repository.CommentsRepository
	solutionsRepo repository.SolutionsRepository
}

func NewCommentHandler(repo repository.CommentsRepository, solutionsRepo repository.SolutionsRepository) *CommentHandler {
	return &CommentHandler{repo: repo, solutionsRepo: solutionsRepo}
}

// GET: /comments
func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	// Get the solutionId from the query parameters
	solutionId := r.URL.Query().Get("solutionId")
	if solutionId == "" {
		return errors.NewApiError(nil, "Missing solutionId for comment retrieval", http.StatusBadRequest)
	}

	// Convert solutionId to int64
	id, err := strconv.ParseInt(solutionId, 10, 32)
	if err != nil {
		return errors.NewApiError(nil, "Invalid solutionId format for comment retrieval", http.StatusBadRequest)
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
		UserID:        claims.UserID,
		Sort:          sort,
		SortDirection: order,
	})
	if err != nil {
		httputils.EmptyDataArrayResponse(w)
		return nil
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
				Votes:      dbComment.Votes,
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

	return nil
}

// POST: /comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	comment, err := httputils.DecodeJSONRequest[domain.CommentCreateRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	// Validate input
	if comment.SolutionId == 0 || comment.Body == "" {
		return errors.NewApiError(nil, "missing required fields for comment creation", http.StatusBadRequest)
	}

	// Check if solution exists
	_, err = h.solutionsRepo.GetSolutionById(r.Context(), int32(comment.SolutionId))
	if err != nil {
		return errors.HandleDatabaseError(err, "solution")
	}

	// create comment
	_, err = h.repo.CreateComment(r.Context(), domain.CommentCreateParams{
		SolutionID: comment.SolutionId,
		ParentID:   comment.ParentId,
		UserID:     claims.UserID,
		Body:       comment.Body,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "comment")
	}

	httputils.SendJSONResponse(w, http.StatusCreated, nil)

	return nil
}

// GET: /comments/{commentId}
func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		return errors.NewApiError(nil, "Missing commentId for comment retrieval", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.NewApiError(err, "Invalid commentId format for comment retrieval", http.StatusBadRequest)
	}

	comment, err := h.repo.GetComment(r.Context(), sql.GetCommentParams{
		ID:     id,
		UserID: claims.UserID,
	})
	if err != nil {
		httputils.EmptyDataResponse(w)
		return nil
	}

	httputils.SendJSONResponse(w, http.StatusOK, comment)

	return nil
}

// PUT: /comments/{commentId}
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		return errors.NewApiError(nil, "Missing commentId", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.NewApiError(err, "Invalid commentId format", http.StatusBadRequest)
	}

	comment, err := httputils.DecodeJSONRequest[domain.CommentUpdateRequest](r)
	if err != nil {
		return errors.NewApiError(nil, "validation", http.StatusBadRequest)
	}

	// Validate input
	if comment.Body == "" {
		return errors.NewApiError(nil, "Body is required", http.StatusBadRequest)
	}

	_, err = h.repo.UpdateComment(r.Context(), sql.UpdateCommentParams{
		ID:     id,
		Body:   comment.Body,
		UserID: claims.UserID, // Check if the user is the owner of the comment
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "comment")
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

// DELETE: /comments/{commentId}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		return errors.NewApiError(nil, "Missing commentId", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.NewApiError(err, "Invalid commentId format", http.StatusBadRequest)
	}

	err = h.repo.DeleteComment(r.Context(), sql.DeleteCommentParams{
		ID:     id,
		UserID: claims.UserID, // Check if the user is the owner of the comment
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "Failed to delete comment")
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

// PATCH: /{commentId}/vote
func (h *CommentHandler) VoteComment(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	// Extract commentId from URL parameters
	commentId := chi.URLParam(r, "commentId")
	if commentId == "" {
		return errors.NewApiError(nil, "Missing commentId", http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.NewApiError(nil, "Invalid commentId format", http.StatusBadRequest)
	}

	// Decode the request body into VoteRequest struct
	req, err := httputils.DecodeJSONRequest[domain.VoteRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if req.Vote == "" {
		return errors.NewApiError(nil, "Vote is required", http.StatusBadRequest)
	}

	// Check if the comment exists
	_, err = h.repo.GetCommentById(r.Context(), id)
	if err != nil {
		return errors.HandleDatabaseError(err, "comment")
	}

	err = h.repo.VoteComment(r.Context(), sql.VoteCommentParams{
		UserID:    claims.UserID,
		CommentID: id,
		Vote:      sql.VoteType(req.Vote),
	})
	if err != nil {
		return errors.NewApiError(err, "Error voting on comment", http.StatusBadRequest)
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}
