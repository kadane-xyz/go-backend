package api

import (
	"encoding/json"
	"net/http"
	"time"

	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Friend struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatar_url"`
	Location  string `json:"location"`
	Level     int32  `json:"level"`
}

type FriendRequestRequest struct {
	FriendName string `json:"friendName"`
}

type FriendsResponse struct {
	Data []Friend `json:"data"`
}

type FriendRequest struct {
	FriendId   string    `json:"friendId"`
	FriendName string    `json:"friendName"`
	AvatarUrl  string    `json:"avatar_url"`
	Level      int32     `json:"level"`
	CreatedAt  time.Time `json:"created_at"`
	Location   string    `json:"location"`
}

type FriendRequestsResponse struct {
	Data []FriendRequest `json:"data"`
}

// GET: /friends
// GetFriends gets all friends
func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	friends, err := h.PostgresQueries.GetFriends(r.Context(), userId)
	if err != nil {
		EmptyDataArrayResponse(w)
		return
	}

	friendsResponseData := []Friend{}
	for _, friend := range friends {
		friendsResponseData = append(friendsResponseData, Friend{
			Id:        friend.FriendID,
			Username:  friend.FriendUsername,
			AvatarUrl: friend.AvatarUrl,
			Level:     friend.Level,
			Location:  friend.Location,
		})
	}

	friendsResponse := FriendsResponse{
		Data: friendsResponseData,
	}

	json.NewEncoder(w).Encode(friendsResponse)
}

// POST: /friends
// CreateFriendRequest creates a friend request
func (h *Handler) CreateFriendRequest(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	var friendRequest FriendRequestRequest
	err := json.NewDecoder(r.Body).Decode(&friendRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if friendRequest.FriendName == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing friend name")
		return
	}

	// Get current user's username
	currentUser, err := h.PostgresQueries.GetAccount(r.Context(), sql.GetAccountParams{
		ID:                userId,
		IncludeAttributes: false,
		UsernamesFilter:   []string{},
		LocationsFilter:   []string{},
		Sort:              "",
		SortDirection:     "",
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error getting current user")
		return
	}

	// Check if the friend name is the same as the current user's username
	if friendRequest.FriendName == currentUser.Username {
		apierror.SendError(w, http.StatusBadRequest, "Cannot add yourself as a friend")
		return
	}

	// Check if the friend request already exists
	friendRequestStatus, _ := h.PostgresQueries.GetFriendRequestStatus(r.Context(), sql.GetFriendRequestStatusParams{
		UserID:     userId,
		FriendName: friendRequest.FriendName,
	})
	if friendRequestStatus != "" {
		apierror.SendError(w, http.StatusBadRequest, "Friend relationship already exists")
		return
	}

	err = h.PostgresQueries.CreateFriendRequest(r.Context(), sql.CreateFriendRequestParams{
		UserID:     userId,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Error creating friend request")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET: /friends/requests
// GetFriendRequests gets all friend requests
func (h *Handler) GetFriendRequests(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	friendRequests, err := h.PostgresQueries.GetFriendRequests(r.Context(), userId)
	if err != nil {
		http.Error(w, "Error getting friend requests", http.StatusInternalServerError)
		return
	}

	friendRequestsResponseData := []FriendRequest{}
	for _, friendRequest := range friendRequests {
		friendRequestsResponseData = append(friendRequestsResponseData, FriendRequest{
			FriendId:   friendRequest.FriendID,
			FriendName: friendRequest.FriendUsername,
			AvatarUrl:  friendRequest.AvatarUrl,
			Level:      friendRequest.Level,
			CreatedAt:  friendRequest.CreatedAt.Time,
			Location:   friendRequest.Location,
		})
	}

	friendRequestsResponse := FriendRequestsResponse{
		Data: friendRequestsResponseData,
	}

	json.NewEncoder(w).Encode(friendRequestsResponse)
}

// POST: /friends/requests/accept
// AcceptFriendRequest accepts a friend request
func (h *Handler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	var friendRequest FriendRequest
	err := json.NewDecoder(r.Body).Decode(&friendRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.PostgresQueries.AcceptFriendRequest(r.Context(), sql.AcceptFriendRequestParams{
		UserID:     userId,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		http.Error(w, "Error accepting friend request", http.StatusInternalServerError)
		return
	}
}

// POST: /friends/requests/block
// BlockFriendRequest blocks a friend request
func (h *Handler) BlockFriendRequest(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	var friendRequest FriendRequest
	err := json.NewDecoder(r.Body).Decode(&friendRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.PostgresQueries.BlockFriend(r.Context(), sql.BlockFriendParams{
		UserID:     userId,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		http.Error(w, "Error blocking friend request", http.StatusInternalServerError)
		return
	}
}

// POST: /friends/requests/unblock
// UnblockFriendRequest unblocks a friend request
func (h *Handler) UnblockFriendRequest(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	var friendRequest FriendRequest
	err := json.NewDecoder(r.Body).Decode(&friendRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.PostgresQueries.UnblockFriend(r.Context(), sql.UnblockFriendParams{
		UserID:     userId,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		http.Error(w, "Error unblocking friend request", http.StatusInternalServerError)
		return
	}
}

// DELETE: /friends or /friends/requests/deny
// DeleteFriend deletes a friend request or a friend
func (h *Handler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	var friendRequest FriendRequest
	err := json.NewDecoder(r.Body).Decode(&friendRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.PostgresQueries.DeleteFriendship(r.Context(), sql.DeleteFriendshipParams{
		UserID:     userId,
		FriendName: friendRequest.FriendName,
	})
	if err != nil {
		http.Error(w, "Error deleting friend request", http.StatusInternalServerError)
		return
	}
}
