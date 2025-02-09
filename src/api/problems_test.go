package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetProblem(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/problems/{problemId}", nil)

	testCases := []struct {
		name              string
		problemIdUrlParam string
		expectedStatus    int
	}{
		{
			name:              "Get problem",
			problemIdUrlParam: "1",
			expectedStatus:    http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"problemId": testCase.problemIdUrlParam})

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetProblem)
		})
	}
}

func TestGetProblems(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/problems", nil)

	testCases := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
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

			req := baseReq.Clone(baseReq.Context())
			req = applyQueryParams(req, testCase.queryParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetProblems)
		})
	}
}

func TestCreateProblem(t *testing.T) {
	testCases := []struct {
		name           string
		body           ProblemRequest
		expectedStatus int
	}{
		{
			name: "Create problem",
			body: ProblemRequest{
				Title:        "Test Problem",
				Description:  "This is a test problem",
				FunctionName: "testFunction",
				Tags:         []string{"test", "problem"},
				Difficulty:   "easy",
				Code: map[string]string{
					"go":         "package main\n\nfunc TestProblem()",
					"python":     "def test_problem():\n    pass",
					"javascript": "function testProblem() {\n    // test code\n}",
					"java":       "public class TestProblem {\n    public static void main(String[] args) {\n        // test code\n    }\n}",
					"cpp":        "int testProblem() {\n    // test code\n}",
					"typescript": "function testProblem() {\n    // test code\n}",
				},
				Hints: []ProblemRequestHint{
					{
						Description: "This is a test hint",
						Answer:      "This is a test answer",
					},
				},
				Solution: "This is a test solution",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.body)
			if err != nil {
				t.Fatalf("Failed to marshal body: %v", err)
			}

			req := newTestRequest(t, http.MethodPost, "/problems", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			executeTestRequest(t, req, testCase.expectedStatus, handler.CreateProblem)
		})
	}
}
