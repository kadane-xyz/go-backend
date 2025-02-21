package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

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
	FunctionName string            `json:"functionName"`
	Solution     map[string]string `json:"solution"` // ["language": "sourceCode"]
	TestCases    []TestCase        `json:"testCases"`
}

type AdminProblemRunResult struct {
	TestCases []RunTestCase        `json:"testCases"`
	Status    sql.SubmissionStatus `json:"status"` // Accepted, Wrong Answer, etc
	CreatedAt time.Time            `json:"createdAt"`
}

type AdminProblemData struct {
	Runs        map[string]AdminProblemRunResult `json:"runs"`
	Status      sql.SubmissionStatus             `json:"status"`
	AccountID   string                           `json:"accountId"`
	CompletedAt time.Time                        `json:"completedAt"`
}

type AdminProblemResponse struct {
	Data AdminProblemData `json:"data"`
}

// POST: /admin/problems/run
// Make sure to check test cases for each language
func (h *Handler) CreateAdminProblemRun(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Decode request body
	var runRequest AdminProblemRequest
	err = json.NewDecoder(r.Body).Decode(&runRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid run data format")
		return
	}

	if runRequest.FunctionName == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing function name")
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

		// Check if function name is valid
		if !strings.Contains(sourceCode, runRequest.FunctionName) {
			apierror.SendError(w, http.StatusBadRequest, "Function name not found in "+language+" source code")
			return
		}
	}

	solutionRuns := make(map[string][]judge0.Submission) // Store all judge0 submission inputs for each language

	// Create judge0 submission inputs by combining test case handling and template creation.
	for language, sourceCode := range runRequest.Solution {
		for _, testCase := range runRequest.TestCases {
			var testCaseInput []TestCaseInput

			// Append testCase.Input values (handles both empty and populated arrays)
			testCaseInput = append(testCaseInput, testCase.Input...)

			// Create the submission for this test case and language
			solutionRun := TemplateCreate(TemplateInput{
				Language:     language,
				SourceCode:   sourceCode,
				FunctionName: runRequest.FunctionName,
				TestCase: TestCase{
					Input:  testCaseInput,
					Output: testCase.Output,
				},
			})
			solutionRuns[language] = append(solutionRuns[language], solutionRun)
		}
	}

	// Validate submissions before sending
	if len(solutionRuns) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Failed to create runs")
		return
	}

	// Initialize response data
	var responseData AdminProblemResponse
	responseData.Data = AdminProblemData{
		Runs: make(map[string]AdminProblemRunResult),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create submissions for each language
	for language := range solutionRuns {
		wg.Add(1)
		go func(language string, submissions []judge0.Submission) {
			defer wg.Done()
			runResponses, _ := h.Judge0Client.CreateSubmissionBatchAndWait(submissions)
			/*if err != nil {
				apierror.SendError(w, http.StatusInternalServerError, "Failed to create solution submission for language: "+language)
				continue
			}*/

			// We'll store test cases in a local slice
			var localTestCases []RunTestCase

			// Compare outputs for each test case
			for i, solutionResp := range runResponses {
				testCase := RunTestCase{
					Time:           solutionResp.Time,
					Memory:         int(solutionResp.Memory),
					Status:         sql.SubmissionStatus(solutionResp.Status.Description),
					Output:         solutionResp.Stdout,
					ExpectedOutput: runRequest.TestCases[i].Output,
				}

				if strings.Contains(solutionResp.Stdout, "[") {
					solutionResp.Stdout = strings.ReplaceAll(solutionResp.Stdout, " ", "")
					testCase.Output = strings.ReplaceAll(testCase.Output, " ", "")
				}
				if strings.Contains(solutionResp.Stdout, "\n") {
					solutionResp.Stdout = strings.ReplaceAll(solutionResp.Stdout, "\n", "")
					testCase.Output = strings.ReplaceAll(testCase.Output, "\n", "")
				}

				if solutionResp.Status.Description != "Accepted" || solutionResp.Stdout != runRequest.TestCases[i].Output {
					testCase.Status = sql.SubmissionStatus("Wrong Answer")
				}

				localTestCases = append(localTestCases, testCase)
			}

			// Determine overall status for this language
			var responseState string
			for _, testCase := range localTestCases {
				if testCase.Status == "Wrong Answer" {
					responseState = "Wrong Answer"
					break
				} else if testCase.Status == "Accepted" {
					responseState = "Accepted"
				}
			}

			// Package the results in a local variable
			result := AdminProblemRunResult{
				TestCases: localTestCases,
				Status:    sql.SubmissionStatus(responseState),
				CreatedAt: time.Now(),
			}

			// Protect map access with the mutex during write
			mu.Lock()
			responseData.Data.Runs[language] = result
			mu.Unlock()
		}(language, solutionRuns[language])
	}

	wg.Wait()

	// Determine the overall status of all runs
	var status string
	for _, run := range responseData.Data.Runs {
		if run.Status == "Wrong Answer" {
			status = "Wrong Answer"
			break
		} else if run.Status == "Accepted" {
			status = "Accepted"
		}
	}

	// Set response values
	responseData.Data.AccountID = userId
	responseData.Data.Status = sql.SubmissionStatus(status)
	responseData.Data.CompletedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
