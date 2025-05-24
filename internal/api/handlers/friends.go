package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type FriendHandler struct {
	repo        repository.FriendRepository
	accountRepo repository.AccountRepository
}

func NewFriendHandler(repo repository.FriendRepository, accountRepo repository.AccountRepository) *FriendHandler {
	return &FriendHandler{repo: repo, accountRepo: accountRepo}
}

// GET: /friends
// GetFriends gets all friends
func (h *FriendHandler) GetFriends(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	friends, err := h.repo.GetFriends(r.Context(), claims.UserID)
	if err != nil {
		httputils.EmptyDataArrayResponse(w)
		return nil
	}

	httputils.SendJSONResponse(w, http.StatusOK, friends)

	return nil
}

// POST: /friends
// CreateFriendRequest creates a friend request
func (h *FriendHandler) CreateFriendRequest(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	friendRequest, err := httputils.DecodeJSONRequest[domain.FriendRequestRequest](r)
	if err != nil {
		return errors.NewApiError(err, "Invalid request body", http.StatusBadRequest)
	}

	if friendRequest.FriendName == "" {
		return errors.NewApiError(nil, "Missing friend name", http.StatusBadRequest)
	}

	// Get current user's username
	currentUser, err := h.accountRepo.GetAccount(r.Context(), &domain.AccountGetParams{
		ID:                claims.UserID,
		IncludeAttributes: false,
		UsernamesFilter:   []string{},
		LocationsFilter:   []string{},
		Sort:              "",
		SortDirection:     "",
	})
	if err != nil {
		return errors.NewApiError(err, "Error getting current user", http.StatusInternalServerError)
	}

	// Check if the friend name is the same as the current user's username
	if friendRequest.FriendName == currentUser.Username {
		return errors.NewApiError(nil, "Cannot add yourself as a friend", http.StatusBadRequest)
	}

	// Check if the friend request already exists
	friendRequestStatus, err := h.repo.GetFriendRequestStatus(r.Context(), &domain.FriendRequesStatusParams{
		UserID:     claims.UserID,
		FriendName: friendRequest.FriendName,
	})
	if friendRequestStatus != nil {
		return errors.NewApiError(err, "Friend relationship already exists", http.StatusBadRequest)
	}

	err = h.repo.CreateFriendRequest(r.Context(), &domain.FriendBlockParams{
		UserID:     claims.UserID,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "friend request")
	}

	httputils.SendJSONResponse(w, http.StatusCreated, nil)

	return nil
}

// GET: /friends/requests/sent
// GetFriendRequestsSent gets all friend requests sent
func (h *FriendHandler) GetFriendRequestsSent(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	friendRequests, err := h.repo.GetFriendRequestsSent(r.Context(), claims.UserID)
	if err != nil {
		httputils.EmptyDataArrayResponse(w)
		return err
	}

	httputils.SendJSONResponse(w, http.StatusOK, friendRequests)

	return nil
}

// GET: /friends/requests/received
// GetFriendRequestsReceived gets all friend requests received
func (h *FriendHandler) GetFriendRequestsReceived(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	friendRequests, err := h.repo.GetFriendRequestReceived(r.Context(), claims.UserID)
	if err != nil {
		httputils.EmptyDataArrayResponse(w)
		return err
	}

	httputils.SendJSONResponse(w, http.StatusOK, friendRequests)

	return nil
}

// POST: /friends/requests/accept
// AcceptFriendRequest accepts a friend request
func (h *FriendHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	friendRequest, err := httputils.DecodeJSONRequest[domain.FriendRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	err = h.repo.AcceptFriendRequest(r.Context(), &domain.FriendBlockParams{
		UserID:     claims.UserID,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		return errors.NewApiError(err, "Error accepting friend request", http.StatusInternalServerError)
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

// POST: /friends/requests/block
// BlockFriendRequest blocks a friend request
func (h *FriendHandler) BlockFriendRequest(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	friendRequest, err := httputils.DecodeJSONRequest[domain.FriendRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	err = h.repo.BlockFriend(r.Context(), &domain.FriendBlockParams{
		UserID:     claims.UserID,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "block friend")
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

// POST: /friends/requests/unblock
// UnblockFriendRequest unblocks a friend request
func (h *FriendHandler) UnblockFriendRequest(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	friendRequest, err := httputils.DecodeJSONRequest[domain.FriendRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	err = h.repo.UnblockFriend(r.Context(), &domain.FriendBlockParams{
		UserID:     claims.UserID,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "unblock friend")
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

// DELETE: /friends or /friends/requests/deny
// DeleteFriend deletes a friend request or a friend
func (h *FriendHandler) DeleteFriend(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		return errors.NewApiError(nil, "Missing username", http.StatusBadRequest)
	}

	err = h.repo.DeleteFriendship(r.Context(), &domain.FriendBlockParams{
		UserID:     claims.UserID,
		FriendName: username,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "delete friendship")
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}

// GET: /friends/username/{username}
// GetFriendsUsername gets all friends by username
func (h *FriendHandler) GetFriendsUsername(w http.ResponseWriter, r *http.Request) error {
	username, err := httputils.GetURLParam(r, "username")
	if err != nil {
		return err
	}

	_, err = h.accountRepo.GetAccountByUsername(r.Context(), sql.GetAccountByUsernameParams{
		Username:          *username,
		IncludeAttributes: false,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "get account by username")
	}

	friend, err := h.repo.GetFriendByUsername(r.Context(), *username)
	if err != nil {
		httputils.EmptyDataArrayResponse(w)
		return nil
	}

	httputils.SendJSONResponse(w, http.StatusOK, friend)

	return nil
}

// DELETE: /friends/requests
// DeleteFriendRequest deletes a friend request
func (h *FriendHandler) DeleteFriendRequest(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	username, err := httputils.GetURLParam(r, "username")
	if err != nil {
		return nil
	}

	err = h.repo.DeleteFriendship(r.Context(), &domain.FriendBlockParams{
		UserID:     claims.UserID,
		FriendName: *username,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "delete friendship")
	}

	httputils.SendJSONResponse(w, http.StatusNoContent, nil)

	return nil
}
