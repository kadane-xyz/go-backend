package api

import (
	"net/http"
	"testing"
)

func TestGetStarredProblems(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get starred problems",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/starred/problems", nil)
			executeTestRequest(t, request, testCase.expectedStatus, handler.GetStarredProblems)
		})
	}
}

func TestGetStarredSolutions(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get starred solutions",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/starred/solutions", nil)
			executeTestRequest(t, request, testCase.expectedStatus, handler.GetStarredSolutions)
		})
	}
}

func TestGetStarredSubmissions(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get starred submissions",
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/starred/submissions", nil)
			executeTestRequest(t, request, testCase.expectedStatus, handler.GetStarredSubmissions)
		})
	}
}

func TestPutStarProblem(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Put starred problem",
			urlParams:      map[string]string{"problemId": "1"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Problem is not starred",
			urlParams:      map[string]string{"problemId": "1"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequestWithBody(t, http.MethodPut, "/starred/problems", testCase.body)
			req = applyURLParams(req, testCase.urlParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.PutStarProblem)
		})
	}
}

func TestPutStarSolution(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Solution is starred",
			urlParams:      map[string]string{"solutionId": "4"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Solution is not starred",
			urlParams:      map[string]string{"solutionId": "1"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequestWithBody(t, http.MethodPut, "/starred/solutions", testCase.body)
			req = applyURLParams(req, testCase.urlParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.PutStarSolution)
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
