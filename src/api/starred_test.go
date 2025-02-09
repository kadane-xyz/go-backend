package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func TestGetStarredProblems(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/starred/problems", nil)

	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{name: "Get starred problems", expectedStatus: http.StatusOK},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			executeTestRequest(t, req, testcase.expectedStatus, handler.GetStarredProblems)
		})
	}
}

func TestGetStarredSolutions(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/starred/solutions", nil)

	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{name: "Get starred solutions", expectedStatus: http.StatusOK},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			executeTestRequest(t, req, testcase.expectedStatus, handler.GetStarredSolutions)
		})
	}
}

func TestGetStarredSubmissions(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/starred/submissions", nil)

	testCases := []struct {
		name           string
		expectedStatus int
	}{
		{name: "Get starred submissions", expectedStatus: http.StatusOK},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			executeTestRequest(t, req, testcase.expectedStatus, handler.GetStarredSubmissions)
		})
	}
}

func TestPutStarProblem(t *testing.T) {
	testCases := []struct {
		name           string
		problemId      StarProblemRequest
		expectedStatus int
	}{
		{name: "Problem is starred", problemId: StarProblemRequest{ProblemID: 1}, expectedStatus: http.StatusOK},
		{name: "Problem is not starred", problemId: StarProblemRequest{ProblemID: 1}, expectedStatus: http.StatusOK},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testcase.problemId)
			if err != nil {
				t.Fatalf("Failed to marshal problem request: %v", err)
			}

			req := newTestRequest(t, http.MethodPut, "/starred/problems", bytes.NewBuffer(body))
			req = applyRouteParams(req, map[string]string{"problemId": strconv.Itoa(int(testcase.problemId.ProblemID))})

			executeTestRequest(t, req, testcase.expectedStatus, handler.PutStarProblem)
		})
	}
}

func TestPutStarSolution(t *testing.T) {
	testCases := []struct {
		name           string
		solutionId     StarSolutionRequest
		expectedStatus int
	}{
		{name: "Solution is starred", solutionId: StarSolutionRequest{SolutionID: 4}, expectedStatus: http.StatusOK},
		{name: "Solution is not starred", solutionId: StarSolutionRequest{SolutionID: 1}, expectedStatus: http.StatusOK},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testcase.solutionId)
			if err != nil {
				t.Fatalf("Failed to marshal solution request: %v", err)
			}

			req := newTestRequest(t, http.MethodPut, "/starred/solutions", bytes.NewBuffer(body))
			req = applyRouteParams(req, map[string]string{"solutionId": strconv.Itoa(int(testcase.solutionId.SolutionID))})

			executeTestRequest(t, req, testcase.expectedStatus, handler.PutStarSolution)
		})
	}
}

/*func TestPutStarSubmission(t *testing.T) {
	testCases := []struct {
		name           string
		submissionId   StarSubmissionRequest
		expectedStatus int
	}{
		{name: "Submission is starred", submissionId: StarSubmissionRequest{SubmissionID: "1"}, expectedStatus: http.StatusOK},
		{name: "Submission is not starred", submissionId: StarSubmissionRequest{SubmissionID: "1"}, expectedStatus: http.StatusOK},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testcase.submissionId)
			if err != nil {
				t.Fatalf("Failed to marshal submission request: %v", err)
			}

			req := newTestRequest(t, http.MethodPut, "/starred/submissions", bytes.NewBuffer(body))
			req = applyRouteParams(req, map[string]string{"submissionId": testcase.submissionId.SubmissionID})

			executeTestRequest(t, req, testcase.expectedStatus, handler.PutStarSubmission)
		})
	}
}*/
