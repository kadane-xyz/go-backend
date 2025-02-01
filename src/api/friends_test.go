package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetFriends(t *testing.T) {
	baseReq := newTestRequest(t, "GET", "/friends", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.GetFriends)
		})
	}
}

func TestGetFriendRequestsSent(t *testing.T) {
	baseReq := newTestRequest(t, "GET", "/friends/requests/sent", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.GetFriendRequestsSent)
		})
	}
}

func TestGetFriendRequestsReceived(t *testing.T) {
	baseReq := newTestRequest(t, "GET", "/friends/requests/received", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.GetFriendRequestsReceived)
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, "POST", "/friends/requests/accept", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testcase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := baseReq.Clone(baseReq.Context())
			req.Body = io.NopCloser(bytes.NewReader(body))
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.AcceptFriendRequest)
		})
	}
}

func TestBlockFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, "POST", "/friends/requests/block", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testcase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := baseReq.Clone(baseReq.Context())
			req.Body = io.NopCloser(bytes.NewReader(body))
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.BlockFriendRequest)
		})
	}
}

func TestUnblockFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, "POST", "/friends/requests/unblock", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testcase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := baseReq.Clone(baseReq.Context())
			req.Body = io.NopCloser(bytes.NewReader(body))
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.UnblockFriendRequest)
		})
	}
}

func TestDeleteFriend(t *testing.T) {
	baseReq := newTestRequest(t, "DELETE", "/friends", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"username": testcase.usernameUrlParam})
			req = applyQueryParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.DeleteFriend)
		})
	}
}

func TestGetFriendsUsername(t *testing.T) {
	baseReq := newTestRequest(t, "GET", "/friends/username/", nil)

	testCases := []struct {
		name             string
		usernameUrlParam string
		expectedStatus   int
	}{
		{
			name:             "Get friends username",
			usernameUrlParam: "bobjohnson",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Get friends username not found",
			usernameUrlParam: "notfound",
			expectedStatus:   http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"username": testcase.usernameUrlParam})
			req = applyQueryParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.GetFriendsUsername)
		})
	}
}

func TestDeleteFriendRequest(t *testing.T) {
	baseReq := newTestRequest(t, "DELETE", "/friends/requests", nil)

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
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"username": testcase.usernameUrlParam})
			req = applyQueryParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.DeleteFriendRequest)
		})
	}
}
