package integration_test

import (
	"net/http"
	"testing"

	"kadane.xyz/go-backend/v2/internal/database/sql"
)

func TestGetComment(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get comment",
			urlParams:      map[string]string{"commentId": "1"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get comment invalid id",
			urlParams:      map[string]string{"commentId": "a"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get comment not found",
			urlParams:      map[string]string{"commentId": "0"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/comments/", nil)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetComment)
		})
	}
}

func TestGetComments(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get comments",
			queryParams:    map[string]string{"solutionId": "1"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get comments not found",
			queryParams:    map[string]string{"solutionId": ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Get comments sort by time order asc",
			queryParams: map[string]string{
				"solutionId": "1",
				"sortBy":     "time",
				"order":      "asc",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get comments sort by time order desc",
			queryParams: map[string]string{
				"solutionId": "1",
				"sortBy":     "time",
				"order":      "desc",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get comments order asc",
			queryParams: map[string]string{
				"solutionId": "1",
				"order":      "asc",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get comments order desc",
			queryParams: map[string]string{
				"solutionId": "1",
				"order":      "desc",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/comments", nil)
			request = applyQueryParams(request, testCase.queryParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetComments)
		})
	}
}

func TestCreateComment(t *testing.T) {
	parentId := int64(7)

	testCases := []TestingCase{
		{
			name: "Create comment",
			body: map[string]any{
				"solutionId": 2,
				"body":       "This is a test comment",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Create child comment",
			body: map[string]any{
				"solutionId": 2,
				"body":       "This is a test child comment",
				"parentId":   &parentId,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Create comment missing solution id",
			body: map[string]any{
				"body": "This is a test comment",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create comment missing body",
			body: map[string]any{
				"solutionId": 1,
				"body":       "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	// run in order
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPost, "/comments", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.CreateComment)
		})
	}
}

// run in order after TestCreateComment
func TestUpdateComment(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Update comment",
			urlParams:      map[string]string{"commentId": "6"},
			body:           map[string]any{"body": "This is a test comment"},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Update comment invalid id",
			urlParams:      map[string]string{"commentId": "a"},
			body:           map[string]any{"body": "This is a test comment"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Update comment not found",
			urlParams:      map[string]string{"commentId": "0"},
			body:           map[string]any{"body": "This is a test comment"},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Update comment empty body",
			urlParams:      map[string]string{"commentId": "1"},
			body:           map[string]any{"body": ""},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPut, "/comments/", testCase.body)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.UpdateComment)
		})
	}
}

// run in order after TestUpdateComment and TestCreateComment
func TestDeleteComment(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Delete comment",
			urlParams:      map[string]string{"commentId": "6"},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Delete comment invalid id",
			urlParams:      map[string]string{"commentId": "a"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := newTestRequest(t, http.MethodDelete, "/comments/", nil)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.DeleteComment)
		})
	}

}

// run in order after TestDeleteComment
func TestVoteComment(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Vote comment up",
			urlParams:      map[string]string{"commentId": "1"},
			body:           map[string]any{"vote": sql.VoteTypeUp},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Vote comment down",
			urlParams:      map[string]string{"commentId": "1"},
			body:           map[string]any{"vote": sql.VoteTypeDown},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Vote comment up empty id",
			urlParams:      map[string]string{"commentId": ""},
			body:           map[string]any{"vote": sql.VoteTypeUp},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Vote comment up invalid id",
			urlParams:      map[string]string{"commentId": "a"},
			body:           map[string]any{"vote": sql.VoteTypeUp},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPatch, "/comments/vote", testCase.body)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.VoteComment)
		})
	}
}
