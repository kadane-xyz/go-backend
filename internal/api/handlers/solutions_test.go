package server_test

import (
	"net/http"
	"testing"
)

type TestGetSolutionsQueryParams struct {
	TitleSearch string
	Tags        []string
	Page        string
	PerPage     string
	Sort        string
	Order       string
	ProblemId   string
}

func TestGetSolution(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get solution",
			urlParams:      map[string]string{"solutionId": "1"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/solutions/{solutionId}", nil)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetSolution)
		})
	}
}

func TestGetSolutions(t *testing.T) {
	testCases := []TestingCase{
		{
			name: "Get solutions",
			queryParams: map[string]string{
				"title":     "",
				"problemId": "1",
				"tags":      "array,hash table",
				"page":      "1",
				"perPage":   "10",
				"sort":      "time",
				"order":     "desc",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/solutions", nil)
			request = applyQueryParams(request, testCase.queryParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetSolutions)
		})
	}
}

func TestCreateSolution(t *testing.T) {
	testCases := []TestingCase{
		{
			name: "Create solution",
			body: map[string]any{
				"title":     "Solution 1",
				"body":      "Solution 1 body",
				"problemId": 1,
				"tags":      `["array", "hash table"]`,
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequestWithBody(t, http.MethodPost, "/solutions", testCase.body)

			executeTestRequest(t, req, testCase.expectedStatus, handler.CreateSolution)
		})
	}
}

func TestUpdateSolution(t *testing.T) {
	testCases := []TestingCase{
		{
			name:      "Update solution",
			urlParams: map[string]string{"solutionId": "1"},
			body: map[string]any{
				"title": "Solution 1",
				"body":  "Solution 1 body",
				"tags":  `["array", "hash table"]`,
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequestWithBody(t, http.MethodPut, "/solutions/1", testCase.body)
			req = applyURLParams(req, testCase.urlParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.UpdateSolution)
		})
	}
}

func TestDeleteSolution(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Delete solution",
			urlParams:      map[string]string{"solutionId": "2"},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequest(t, http.MethodDelete, "/solutions/2", nil)
			req = applyURLParams(req, testCase.urlParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.DeleteSolution)
		})
	}
}

func TestVoteSolution(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Vote solution",
			urlParams:      map[string]string{"solutionId": "4"},
			body:           map[string]any{"vote": "up"},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Vote solution",
			urlParams:      map[string]string{"solutionId": "4"},
			body:           map[string]any{"vote": "down"},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Remove vote",
			urlParams:      map[string]string{"solutionId": "4"},
			body:           map[string]any{"vote": "none"},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequestWithBody(t, http.MethodPatch, "/solutions/4/vote", testCase.body)
			req = applyURLParams(req, testCase.urlParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.VoteSolution)
		})
	}
}
