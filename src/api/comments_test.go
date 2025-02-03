package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"kadane.xyz/go-backend/v2/src/sql/sql"
)

func TestGetComment(t *testing.T) {
	baseReq := newTestRequest(t, "GET", "/comments/", nil)

	testCases := []struct {
		name              string
		commentIdUrlParam string
		expectedStatus    int
	}{
		{
			name:              "Get comment",
			commentIdUrlParam: "1",
			expectedStatus:    http.StatusOK,
		},
		{
			name:              "Get comment invalid id",
			commentIdUrlParam: "a",
			expectedStatus:    http.StatusBadRequest,
		},
		{
			name:              "Get comment not found",
			commentIdUrlParam: "0",
			expectedStatus:    http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"commentId": testcase.commentIdUrlParam})
			executeTestRequest(t, req, testcase.expectedStatus, handler.GetComment)
		})
	}
}

func TestGetComments(t *testing.T) {
	baseReq := newTestRequest(t, "GET", "/comments", nil)

	testCases := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
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
			name:           "Get comments sort by time order asc",
			queryParams:    map[string]string{"solutionId": "1", "sortBy": "time", "order": "asc"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get comments sort by time order desc",
			queryParams:    map[string]string{"solutionId": "1", "sortBy": "time", "order": "desc"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get comments order asc",
			queryParams:    map[string]string{"solutionId": "1", "order": "asc"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get comments order desc",
			queryParams:    map[string]string{"solutionId": "1", "order": "desc"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyQueryParams(req, testcase.queryParams)
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.GetComments)
		})
	}
}

func TestCreateComment(t *testing.T) {
	baseReq := newTestRequest(t, "POST", "/comments", nil)

	parentId := int64(7)

	testCases := []struct {
		name           string
		input          CommentCreateRequest
		expectedStatus int
	}{
		{
			name: "Create comment",
			input: CommentCreateRequest{
				SolutionId: 2,
				Body:       "This is a test comment",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Create child comment",
			input: CommentCreateRequest{
				SolutionId: 2,
				Body:       "This is a test child comment",
				ParentId:   &parentId,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Create comment missing solution id",
			input: CommentCreateRequest{
				Body: "This is a test comment",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create comment missing body",
			input: CommentCreateRequest{
				SolutionId: 1,
				Body:       "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	// run in order
	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			body, err := json.Marshal(testcase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := baseReq.Clone(baseReq.Context())
			req.Body = io.NopCloser(bytes.NewReader(body))
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testcase.expectedStatus, handler.CreateComment)
		})
	}
}

// run in order after TestCreateComment
func TestUpdateComment(t *testing.T) {
	testCases := []struct {
		name              string
		commentIdUrlParam string
		input             CommentUpdateRequest
		expectedStatus    int
	}{
		{
			name:              "Update comment",
			commentIdUrlParam: "6",
			input:             CommentUpdateRequest{Body: "This is a test comment"},
			expectedStatus:    http.StatusNoContent,
		},
		{
			name:              "Update comment invalid id",
			commentIdUrlParam: "a",
			input:             CommentUpdateRequest{Body: "This is a test comment"},
			expectedStatus:    http.StatusBadRequest,
		},
		{
			name:              "Update comment not found",
			commentIdUrlParam: "0",
			input:             CommentUpdateRequest{Body: "This is a test comment"},
			expectedStatus:    http.StatusInternalServerError,
		},
		{
			name:              "Update comment empty body",
			commentIdUrlParam: "1",
			input:             CommentUpdateRequest{Body: ""},
			expectedStatus:    http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			body, err := json.Marshal(testcase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := newTestRequest(t, "PUT", "/comments/", bytes.NewReader(body))
			req = applyRouteParams(req, map[string]string{"commentId": testcase.commentIdUrlParam})

			executeTestRequest(t, req, testcase.expectedStatus, handler.UpdateComment)
		})
	}
}

// run in order after TestUpdateComment and TestCreateComment
func TestDeleteComment(t *testing.T) {
	testCases := []struct {
		name              string
		commentIdUrlParam string
		expectedStatus    int
	}{
		{
			name:              "Delete comment",
			commentIdUrlParam: "6",
			expectedStatus:    http.StatusNoContent,
		},
		{
			name:              "Delete comment invalid id",
			commentIdUrlParam: "a",
			expectedStatus:    http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			req := newTestRequest(t, "DELETE", "/comments/", nil)
			req = applyRouteParams(req, map[string]string{"commentId": testcase.commentIdUrlParam})

			executeTestRequest(t, req, testcase.expectedStatus, handler.DeleteComment)
		})
	}

}

// run in order after TestDeleteComment
func TestVoteComment(t *testing.T) {
	testCases := []struct {
		name              string
		commentIdUrlParam string
		voteRequest       VoteRequest
		expectedStatus    int
	}{
		{
			name:              "Vote comment up",
			commentIdUrlParam: "1",
			voteRequest:       VoteRequest{Vote: sql.VoteTypeUp},
			expectedStatus:    http.StatusNoContent,
		},
		{
			name:              "Vote comment down",
			commentIdUrlParam: "1",
			voteRequest:       VoteRequest{Vote: sql.VoteTypeDown},
			expectedStatus:    http.StatusNoContent,
		},
		{
			name:              "Vote comment up empty id",
			commentIdUrlParam: "",
			voteRequest:       VoteRequest{Vote: sql.VoteTypeUp},
			expectedStatus:    http.StatusBadRequest,
		},
		{
			name:              "Vote comment up invalid id",
			commentIdUrlParam: "a",
			voteRequest:       VoteRequest{Vote: sql.VoteTypeUp},
			expectedStatus:    http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			body, err := json.Marshal(testcase.voteRequest)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := newTestRequest(t, "PATCH", "/comments/"+testcase.commentIdUrlParam+"/vote", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = applyRouteParams(req, map[string]string{"commentId": testcase.commentIdUrlParam})

			executeTestRequest(t, req, testcase.expectedStatus, handler.VoteComment)
		})
	}
}
