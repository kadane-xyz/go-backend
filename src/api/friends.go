package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Friend struct {
	Id         string    `json:"id"`
	Username   string    `json:"username"`
	AvatarUrl  string    `json:"avatarUrl"`
	Location   string    `json:"location"`
	Level      int32     `json:"level"`
	AcceptedAt time.Time `json:"acceptedAt"`
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
	AvatarUrl  string    `json:"avatarUrl"`
	Level      int32     `json:"level"`
	CreatedAt  time.Time `json:"createdAt"`
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
			Id:         friend.FriendID,
			Username:   friend.FriendUsername,
			AvatarUrl:  friend.AvatarUrl,
			Level:      friend.Level,
			Location:   friend.Location,
			AcceptedAt: friend.AcceptedAt.Time,
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

// GET: /friends/requests/sent
// GetFriendRequestsSent gets all friend requests sent
func (h *Handler) GetFriendRequestsSent(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	friendRequests, err := h.PostgresQueries.GetFriendRequestsSent(r.Context(), userId)
	if err != nil {
		EmptyDataArrayResponse(w)
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

// GET: /friends/requests/received
// GetFriendRequestsReceived gets all friend requests received
func (h *Handler) GetFriendRequestsReceived(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		http.Error(w, "Missing user id", http.StatusBadRequest)
		return
	}

	friendRequests, err := h.PostgresQueries.GetFriendRequestsReceived(r.Context(), userId)
	if err != nil {
		EmptyDataArrayResponse(w)
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

// GET: /friends/username/{username}
// GetFriendsUsername gets all friends by username
func (h *Handler) GetFriendsUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing username")
		return
	}

	_, err := h.PostgresQueries.GetAccountByUsername(r.Context(), sql.GetAccountByUsernameParams{
		Username:          username,
		IncludeAttributes: false,
	})
	if err != nil {
		apierror.SendError(w, http.StatusNotFound, "User not found")
		return
	}

	friends, err := h.PostgresQueries.GetFriendsByUsername(r.Context(), username)
	if err != nil {
		EmptyDataArrayResponse(w)
		return
	}

	responseData := make([]Friend, len(friends))
	for i := range friends {
		responseData[i] = Friend{
			Id:         friends[i].FriendID,
			Username:   friends[i].FriendUsername,
			AvatarUrl:  friends[i].AvatarUrl,
			Level:      friends[i].Level,
			Location:   friends[i].Location,
			AcceptedAt: friends[i].AcceptedAt.Time,
		}
	}

	response := FriendsResponse{
		Data: responseData,
	}

	json.NewEncoder(w).Encode(response)
}
