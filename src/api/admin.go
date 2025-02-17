package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/judge0"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type AdminValidation struct {
	IsAdmin bool `json:"isAdmin"`
}

type AdminValidationResponse struct {
	Data AdminValidation `json:"data"`
}

// GET: /admin/validate
func (h *Handler) GetAdminValidation(w http.ResponseWriter, r *http.Request) {
	admin := GetClientAdmin(w, r)

	response := AdminValidationResponse{
		Data: AdminValidation{
			IsAdmin: admin,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type AdminProblemRequest struct {
	Solution  map[string]string `json:"solution"` // ["language": "sourceCode"]
	TestCases []TestCase        `json:"testCases"`
}

type AdminProblemResponse struct {
	Data map[string]RunResult `json:"data"`
}

// POST: /admin/problems/run
// Make sure to check test cases for each language
func (h *Handler) CreateAdminProblemRun(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	var runRequest AdminProblemRequest
	err = json.NewDecoder(r.Body).Decode(&runRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid run data format")
		return
	}

	// check map for missing values
	for language, sourceCode := range runRequest.Solution {
		// Check if source code is missing
		if sourceCode == "" {
			apierror.SendError(w, http.StatusBadRequest, "Missing source code")
			return
		}

		// Check if language is valid
		lang := string(sql.ProblemLanguage(language))
		if language == "" || language != lang {
			apierror.SendError(w, http.StatusBadRequest, "Invalid language: "+language)
			return
		}
	}

	solutionRuns := make(map[string][]judge0.Submission) // Store all judge0 submission inputs for each language
	solutionTestCases := []TestCase{}                    // Store all judge0 submission test cases inputs

	// Handle all judge0 submission test cases inputs
	for _, testCase := range runRequest.TestCases {
		var testCaseInput []TestCaseInput

		// Handle both empty array and populated array cases
		testCaseInput = append(testCaseInput, testCase.Input...)

		solutionTestCases = append(solutionTestCases, TestCase{
			Input:  testCaseInput,
			Output: testCase.Output,
		})
	}

	// Create judge0 submission inputs
	for language, sourceCode := range runRequest.Solution {
		for _, testCase := range solutionTestCases {
			solutionRun := TemplateCreate(TemplateInput{
				Language:     language,
				SourceCode:   sourceCode,
				FunctionName: "myNewProblem",
				TestCase:     testCase,
			})
			solutionRuns[language] = append(solutionRuns[language], solutionRun)
		}
	}

	// Validate submissions before sending
	if len(solutionRuns) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Failed to create runs")
		return
	}

	problemId := 0

	var responseData AdminProblemResponse
	responseData.Data = make(map[string]RunResult)

	for language, _ := range solutionRuns {
		runResponses, err := h.Judge0Client.CreateSubmissionBatchAndWait(solutionRuns[language])
		if err != nil {
			apierror.SendError(w, http.StatusInternalServerError, "Failed to create solution submission")
			return
		}

		var dbSubmission sql.CreateSubmissionParams
		var testCases []RunTestCase

		// Compare outputs for each test case
		for i, solutionResp := range runResponses {
			// store user code test case results
			testCases = append(testCases, RunTestCase{
				Time:           solutionResp.Time,
				Memory:         int(solutionResp.Memory),
				Status:         sql.SubmissionStatus(solutionResp.Status.Description),
				Output:         solutionResp.Stdout,
				ExpectedOutput: runRequest.TestCases[i].Output, // solution code output
			})

			// check stdout before using status and remove spaces from array elements
			if strings.Contains(solutionResp.Stdout, "[") {
				solutionResp.Stdout = strings.ReplaceAll(solutionResp.Stdout, " ", "")
			}

			// remove newlines from stdout
			if strings.Contains(solutionResp.Stdout, "\n") {
				solutionResp.Stdout = strings.ReplaceAll(solutionResp.Stdout, "\n", "")
			}

			// First check if both executions were successful
			if solutionResp.Status.Description != "Accepted" || solutionResp.Stdout != runRequest.TestCases[i].Output {
				// store user code test case results
				testCases[i].Status = sql.SubmissionStatus("Wrong Answer")

				// Test case failed - outputs don't match
				dbSubmission, err = SummarizeSubmissionResponses(userId, int32(problemId), runRequest.Solution[language], runResponses[i:i+1])
				if err != nil {
					apierror.SendError(w, http.StatusInternalServerError, "Failed to process submission")
					return
				}
				// Update status to Wrong Answer
				dbSubmission.Status = sql.SubmissionStatus("Wrong Answer")
				break // Stop at first mismatch
			}
		}

		// If all test cases passed, create submission with averaged results
		if dbSubmission.Status == "" {
			dbSubmission, err = SummarizeSubmissionResponses(userId, int32(problemId), runRequest.Solution[language], runResponses)
			if err != nil {
				apierror.SendError(w, http.StatusInternalServerError, "Failed to process submission")
				return
			}
		}

		convertedLanguage := judge0.LanguageIDToLanguage(int(dbSubmission.LanguageID))
		//dbSubmission.LanguageName = language // Add language name to submission

		// create submission in db
		/*_, err = h.PostgresQueries.CreateSubmission(r.Context(), dbSubmission)
		if err != nil {
			apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
			return
		}*/

		submissionId := uuid.UUID(dbSubmission.ID.Bytes)

		var responseState string
		if dbSubmission.Status == "Accepted" {
			responseState = "Accepted"
		} else {
			responseState = "Wrong Answer"
		}

		response := RunResult{
			Id:        submissionId.String(),
			TestCases: testCases,
			Time:      dbSubmission.Time,                   // average of test case times
			Status:    sql.SubmissionStatus(responseState), // if all test cases passed, then Accepted, otherwise Wrong Answer
			Language:  convertedLanguage,
			AccountID: userId,
			ProblemID: int32(problemId),
			CreatedAt: time.Now(),
		}

		responseData.Data[convertedLanguage] = response
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
