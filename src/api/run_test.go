package api

import (
	"net/http"
	"testing"
)

// TestCreateRun tests the POST /runs endpoint
func TestCreateRun(t *testing.T) {
	testingCases := []TestingCase{
		{
			name: "Create run",
			body: RunRequest{
				Language:   "go",
				SourceCode: `func twoSum(nums []int, target int) []int { m := make(map[int]int); for i, num := range nums { m[num] = i }; for i, num := range nums { if j, ok := m[target-num]; ok && j != i { return []int{i, j} } }; return []int{} }`,
				ProblemID:  1,
				TestCases: []TestCase{
					{
						Input: []TestCaseInput{
							{
								Name:  "nums",
								Type:  "int[]",
								Value: "[2, 7, 11, 15]",
							},
							{
								Name:  "target",
								Type:  "int",
								Value: "9",
							},
						},
						Output: "[0, 1]",
					},
					{
						Input: []TestCaseInput{
							{
								Name:  "nums",
								Type:  "int[]",
								Value: "[3, 2, 4]",
							},
							{
								Name:  "target",
								Type:  "int",
								Value: "6",
							},
						},
						Output: "[1, 2]",
					},
					{
						Input: []TestCaseInput{
							{
								Name:  "nums",
								Type:  "int[]",
								Value: "[3, 3]",
							},
							{
								Name:  "target",
								Type:  "int",
								Value: "6",
							},
						},
						Output: "[0, 1]",
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testingCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Build request
			request := newTestRequestWithBody(t, http.MethodPost, "/runs", tc.body)

			// Execute request
			executeTestRequest(t, request, tc.expectedStatus, handler.CreateRunRoute)
		})
	}
}
