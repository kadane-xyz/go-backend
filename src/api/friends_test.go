package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetFriends(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/friends", nil)

	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Get friends",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetFriends)
		})
	}
}

func TestGetFriendRequestsSent(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/friends/requests/sent", nil)

	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Get friend requests sent",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetFriendRequestsSent)
		})
	}
}

func TestGetFriendRequestsReceived(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/friends/requests/received", nil)

	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "Get friend requests received",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetFriendRequestsReceived)
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodPost, "/friends/requests/accept", nil)

	testCases := []struct {
		name           string
		input          FriendRequest
		expectedStatus int
	}{
		{
			name: "Accept friend request",
			input: FriendRequest{
				FriendName: "janesmith",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := baseReq.Clone(baseReq.Context())
			req.Body = io.NopCloser(bytes.NewReader(body))
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.AcceptFriendRequest)
		})
	}
}

func TestBlockFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodPost, "/friends/requests/block", nil)

	testCases := []struct {
		name           string
		input          FriendRequest
		expectedStatus int
	}{
		{
			name: "Block friend request",
			input: FriendRequest{
				FriendName: "bobjohnson",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := baseReq.Clone(baseReq.Context())
			req.Body = io.NopCloser(bytes.NewReader(body))
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.BlockFriendRequest)
		})
	}
}

func TestUnblockFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodPost, "/friends/requests/unblock", nil)

	testCases := []struct {
		name           string
		input          FriendRequest
		expectedStatus int
	}{
		{
			name: "Unblock friend request",
			input: FriendRequest{
				FriendName: "bobjohnson",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := baseReq.Clone(baseReq.Context())
			req.Body = io.NopCloser(bytes.NewReader(body))
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.UnblockFriendRequest)
		})
	}
}

func TestDeleteFriend(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodDelete, "/friends", nil)

	testCases := []struct {
		name             string
		usernameUrlParam string
		expectedStatus   int
	}{
		{
			name:             "Delete friend",
			usernameUrlParam: "bobjohnson",
			expectedStatus:   http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"username": testCase.usernameUrlParam})
			req = applyQueryParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.DeleteFriend)
		})
	}
}

func TestGetFriendsUsername(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/friends/username/", nil)

	testCases := []struct {
		name             string
		usernameUrlParam string
		expectedStatus   int
	}{
		{
			name:             "Get friends username",
			usernameUrlParam: "janesmith",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Get friends username not found",
			usernameUrlParam: "notfound",
			expectedStatus:   http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"username": testCase.usernameUrlParam})
			req = applyQueryParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetFriendsUsername)
		})
	}
}

func TestDeleteFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodDelete, "/friends/requests", nil)

	testCases := []struct {
		name             string
		usernameUrlParam string
		expectedStatus   int
	}{
		{
			name:             "Delete friend request",
			usernameUrlParam: "bobjohnson",
			expectedStatus:   http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"username": testCase.usernameUrlParam})
			req = applyQueryParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.DeleteFriendRequest)
		})
	}
}
