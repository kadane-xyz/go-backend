package server_test

import (
	"net/http"
	"testing"
)

func TestGetFriends(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get friends",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/friends", nil)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetFriends)
		})
	}
}

func TestGetFriendRequestsSent(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get friend requests sent",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/friends/requests/sent", nil)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetFriendRequestsSent)
		})
	}
}

func TestGetFriendRequestsReceived(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get friend requests received",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/friends/requests/received", nil)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetFriendRequestsReceived)
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Accept friend request",
			body:           map[string]any{"friendName": "janesmith"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPost, "/friends/requests/accept", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.AcceptFriendRequest)
		})
	}
}

func TestBlockFriendRequest(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Block friend request",
			body:           map[string]any{"friendName": "bobjohnson"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPost, "/friends/requests/block", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.BlockFriendRequest)
		})
	}
}

func TestUnblockFriendRequest(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Unblock friend request",
			body:           map[string]any{"friendName": "bobjohnson"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPost, "/friends/requests/unblock", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.UnblockFriendRequest)
		})
	}
}

func TestDeleteFriend(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Delete friend",
			body:           map[string]any{"friendName": "bobjohnson"},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodDelete, "/friends", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.DeleteFriend)
		})
	}
}

func TestGetFriendsUsername(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get friends username",
			body:           map[string]any{"username": "janesmith"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get friends username not found",
			body:           map[string]any{"username": "notfound"},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodGet, "/friends/username", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetFriendsUsername)
		})
	}
}

func TestDeleteFriendRequest(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Delete friend request",
			body:           map[string]any{"username": "bobjohnson"},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodDelete, "/friends/requests", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.DeleteFriendRequest)
		})
	}
}
