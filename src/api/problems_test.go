package api

import (
	"net/http"
	"testing"
)

func TestGetProblem(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get problem",
			urlParams:      map[string]string{"problemId": "1"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/problems/{problemId}", nil)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetProblem)
		})
	}
}

func TestGetProblems(t *testing.T) {
	testCases := []TestingCase{
		{
			name: "Get problems",
			queryParams: map[string]string{
				"titleSearch": "",
				"difficulty":  "",
				"sort":        "",
				"order":       "",
				"page":        "",
				"perPage":     "",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get problems with title search",
			queryParams: map[string]string{
				"titleSearch": "Two Sum",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get problems with difficulty search",
			queryParams: map[string]string{
				"difficulty": "medium",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get problems with sort",
			queryParams: map[string]string{
				"sort": "alpha",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get problems with sort",
			queryParams: map[string]string{
				"sort": "index",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get problems with sort and order",
			queryParams: map[string]string{
				"sort":  "index",
				"order": "desc",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get problems with sort, order and pagination",
			queryParams: map[string]string{
				"sort":    "index",
				"order":   "desc",
				"page":    "1",
				"perPage": "10",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get problems with sort, order and pagination",
			queryParams: map[string]string{
				"sort":    "index",
				"order":   "desc",
				"page":    "2",
				"perPage": "5",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/problems", nil)
			request = applyQueryParams(request, testCase.queryParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetProblemsRoute)
		})
	}
}
