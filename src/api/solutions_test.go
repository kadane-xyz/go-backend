package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
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
	baseReq := newTestRequest(t, http.MethodGet, "/solutions/1", nil)

	testCases := []struct {
		name               string
		solutionIdUrlParam string
		expectedStatus     int
	}{
		{
			name:               "Get solution",
			solutionIdUrlParam: "1",
			expectedStatus:     http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"solutionId": testCase.solutionIdUrlParam})

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetSolution)
		})
	}
}

func TestGetSolutions(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/solutions", nil)
	testCases := []struct {
		name           string
		queryParams    TestGetSolutionsQueryParams
		expectedStatus int
	}{
		{
			name: "Get solutions",
			queryParams: TestGetSolutionsQueryParams{
				TitleSearch: "",
				ProblemId:   "1",
				Tags:        []string{"array", "hash table"},
				Page:        "1",
				PerPage:     "10",
				Sort:        "time",
				Order:       "desc",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			queryParams := map[string]string{
				"problemId": testCase.queryParams.ProblemId,
				"title":     testCase.queryParams.TitleSearch,
				"tags":      strings.Join(testCase.queryParams.Tags, ","),
				"page":      testCase.queryParams.Page,
				"perPage":   testCase.queryParams.PerPage,
				"sort":      testCase.queryParams.Sort,
				"order":     testCase.queryParams.Order,
			}
			req = applyQueryParams(req, queryParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetSolutions)
		})
	}
}

func TestCreateSolution(t *testing.T) {
	testCases := []struct {
		name           string
		input          CreateSolutionRequest
		expectedStatus int
	}{
		{
			name: "Create solution",
			input: CreateSolutionRequest{
				Title:     "Solution 1",
				Body:      "Solution 1 body",
				ProblemId: 1,
				Tags:      []string{"array", "hash table"},
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := newTestRequest(t, http.MethodPost, "/solutions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.CreateSolution)
		})
	}
}

func TestUpdateSolution(t *testing.T) {
	testCases := []struct {
		name               string
		solutionIdUrlParam string
		input              UpdateSolutionRequest
		expectedStatus     int
	}{
		{
			name:               "Update solution",
			solutionIdUrlParam: "1",
			input: UpdateSolutionRequest{
				Title: "Solution 1",
				Body:  "Solution 1 body",
				Tags:  []string{"array", "hash table"},
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := newTestRequest(t, http.MethodPut, "/solutions/1", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = applyRouteParams(req, map[string]string{"solutionId": testCase.solutionIdUrlParam})

			executeTestRequest(t, req, testCase.expectedStatus, handler.UpdateSolution)
		})
	}
}

func TestDeleteSolution(t *testing.T) {
	testCases := []struct {
		name               string
		solutionIdUrlParam string
		expectedStatus     int
	}{
		{
			name:               "Delete solution",
			solutionIdUrlParam: "2",
			expectedStatus:     http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequest(t, http.MethodDelete, "/solutions/2", nil)
			req = applyRouteParams(req, map[string]string{"solutionId": testCase.solutionIdUrlParam})

			executeTestRequest(t, req, testCase.expectedStatus, handler.DeleteSolution)
		})
	}
}

func TestVoteSolution(t *testing.T) {
	testCases := []struct {
		name               string
		solutionIdUrlParam string
		input              VoteRequest
		expectedStatus     int
	}{
		{
			name:               "Vote solution",
			solutionIdUrlParam: "4",
			input:              VoteRequest{Vote: "up"},
			expectedStatus:     http.StatusNoContent,
		},
		{
			name:               "Vote solution",
			solutionIdUrlParam: "4",
			input:              VoteRequest{Vote: "down"},
			expectedStatus:     http.StatusNoContent,
		},
		{
			name:               "Remove vote",
			solutionIdUrlParam: "4",
			input:              VoteRequest{Vote: "none"},
			expectedStatus:     http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := newTestRequest(t, http.MethodPatch, "/solutions/4/vote", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = applyRouteParams(req, map[string]string{"solutionId": testCase.solutionIdUrlParam})

			executeTestRequest(t, req, testCase.expectedStatus, handler.VoteSolution)
		})
	}
}
