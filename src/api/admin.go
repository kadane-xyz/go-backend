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

type AdminProblemRunRequest struct {
	FunctionName string            `json:"functionName"`
	Solutions    map[string]string `json:"solutions"` // ["language": "sourceCode"]
	TestCase     TestCase          `json:"testCase"`
}

type AdminProblemRunResult struct {
	TestCase  RunTestCase          `json:"testCase"`
	Status    sql.SubmissionStatus `json:"status"` // Accepted, Wrong Answer, etc
	CreatedAt time.Time            `json:"createdAt"`
}

type AdminProblemData struct {
	Runs        map[string]AdminProblemRunResult `json:"runs"`
	Status      sql.SubmissionStatus             `json:"status"`
	CompletedAt time.Time                        `json:"completedAt"`
}

type AdminProblemResponse struct {
	Data AdminProblemData `json:"data"`
}

type CreateAdminProblemData struct {
	ProblemID string `json:"problemId"`
}

type CreateAdminProblemResponse struct {
	Data CreateAdminProblemData `json:"data"`
}

func (h *Handler) ProblemRun(runRequest AdminProblemRunRequest) (AdminProblemResponse, *apierror.APIError) {
	solutionRuns := make(map[string][]judge0.Submission) // Store all judge0 submission inputs for each language

	// Create judge0 submission inputs by combining test case handling and template creation.
	for language, sourceCode := range runRequest.Solutions {
		// Create the submission for this test case and language
		solutionRun := TemplateCreate(TemplateInput{
			Language:     language,
			SourceCode:   sourceCode,
			FunctionName: runRequest.FunctionName,
			TestCase:     runRequest.TestCase,
		})
		solutionRuns[language] = append(solutionRuns[language], solutionRun)
	}

	// Validate submissions before sending
	if len(solutionRuns) == 0 {
		return AdminProblemResponse{}, apierror.NewError(http.StatusBadRequest, "Failed to create runs")
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

			var localTestCase RunTestCase

			// Compare outputs for each test case
			for _, solutionResp := range runResponses {
				testCase := RunTestCase{
					Time:           solutionResp.Time,
					Memory:         int(solutionResp.Memory),
					Status:         sql.SubmissionStatus(solutionResp.Status.Description),
					Output:         solutionResp.Stdout,
					CompileOutput:  solutionResp.CompileOutput,
					ExpectedOutput: runRequest.TestCase.Output,
				}

				if strings.Contains(solutionResp.Stdout, "[") {
					solutionResp.Stdout = strings.ReplaceAll(solutionResp.Stdout, " ", "")
					testCase.Output = strings.ReplaceAll(testCase.Output, " ", "")
				}
				if strings.Contains(solutionResp.Stdout, "\n") {
					solutionResp.Stdout = strings.ReplaceAll(solutionResp.Stdout, "\n", "")
					testCase.Output = strings.ReplaceAll(testCase.Output, "\n", "")
				}

				if solutionResp.Status.Description != "Accepted" || solutionResp.Stdout != runRequest.TestCase.Output {
					testCase.Status = sql.SubmissionStatus("Wrong Answer")
				}

				localTestCase = testCase
			}

			// Determine overall status for this language
			var responseState string
			if localTestCase.Status == "Wrong Answer" {
				responseState = "Wrong Answer"
			} else if localTestCase.Status == "Accepted" {
				responseState = "Accepted"
			}

			// Package the results in a local variable
			result := AdminProblemRunResult{
				TestCase:  localTestCase,
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
	responseData.Data.Status = sql.SubmissionStatus(status)
	responseData.Data.CompletedAt = time.Now()

	return responseData, nil
}

func ProblemRunRequestValidate(runRequest AdminProblemRunRequest) *apierror.APIError {
	if runRequest.FunctionName == "" {
		return apierror.NewError(http.StatusBadRequest, "Missing function name")
	}

	// check map for missing values
	for language, sourceCode := range runRequest.Solutions {
		// Check if source code is missing
		if sourceCode == "" {
			return apierror.NewError(http.StatusBadRequest, "Missing source code")
		}

		// Check if language is valid
		lang := string(sql.ProblemLanguage(language))
		if language == "" || language != lang {
			return apierror.NewError(http.StatusBadRequest, "Invalid language: "+language)
		}

		// Check if function name is valid
		if !strings.Contains(sourceCode, runRequest.FunctionName) {
			return apierror.NewError(http.StatusBadRequest, "Function name not found in "+language+" source code")
		}
	}

	return nil
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

// POST: /admin/problems
func (h *Handler) CreateAdminProblem(w http.ResponseWriter, r *http.Request) {
	request, apiErr := DecodeJSONRequest[ProblemRequest](r)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	apiErr = CreateProblemRequestValidate(request)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	// Test problem test cases against solutions in each language
	for i, testCase := range request.TestCases {
		responseData, apiErr := h.ProblemRun(AdminProblemRunRequest{
			FunctionName: request.FunctionName,
			Solutions:    request.Solutions,
			TestCase:     testCase,
		})
		if apiErr != nil {
			apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
			return
		}

		// Check if any test cases fail
		if responseData.Data.Status == "Wrong Answer" {
			apierror.SendError(w, http.StatusBadRequest, "Wrong answer for test case: "+testCase.Input[i].Name)
			return
		}
	}

	// Create problem in database if all test cases pass
	problemID, apiErr := h.CreateProblem(request)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	response := CreateAdminProblemResponse{
		Data: CreateAdminProblemData{
			ProblemID: problemID.Data.ProblemID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// POST: /admin/problems/run
// Make sure to check test cases for each language
func (h *Handler) CreateAdminProblemRun(w http.ResponseWriter, r *http.Request) {
	runRequest, apiErr := DecodeJSONRequest[AdminProblemRunRequest](r)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	// Validate request
	apiErr = ProblemRunRequestValidate(runRequest)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	// Run problems against judge0
	responseData, apiErr := h.ProblemRun(runRequest)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}
