package api

import (
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

				if solutionResp.Status.Description != "Accepted" ||
					solutionResp.Stdout != runRequest.TestCase.Output ||
					solutionResp.CompileOutput != "" {
					testCase.Status = sql.SubmissionStatus("Wrong Answer")
				}

				localTestCase = testCase
			}

			// Determine overall status for this language
			var responseState string
			if localTestCase.Status != "Accepted" {
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
			return apierror.NewError(http.StatusBadRequest, "Correct function name: "+runRequest.FunctionName+" not found in "+language+" source code")
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

	SendJSONResponse(w, http.StatusOK, response)
}

type AdminProblem struct {
	Problem  Problem           `json:"problem"`
	Solution map[string]string `json:"solution,omitempty"` // ["language": "sourceCode"]
}

type AdminProblemsResponse struct {
	Data []AdminProblem `json:"data"`
}

// GET: /admin/problems
func (h *Handler) GetAdminProblems(w http.ResponseWriter, r *http.Request) {
	problems, apiErr := h.PostgresQueries.GetAdminProblems(r.Context())
	if apiErr != nil {
		apierror.SendError(w, http.StatusInternalServerError, apiErr.Error())
		return
	}

	var adminProblems []AdminProblem

	for _, problem := range problems {
		// Create a map to store language->code mapping
		solutionMap := make(map[string]string)

		// Handle the solutions data which is already unmarshaled as []interface{}
		if problem.Solutions != nil {
			// Check if it's already a slice of interfaces
			if solutionsArray, ok := problem.Solutions.([]interface{}); ok {
				for _, solutionItem := range solutionsArray {
					// Each solution item should be a map[string]interface{}
					if solutionMap_, ok := solutionItem.(map[string]interface{}); ok {
						language, languageOk := solutionMap_["language"].(string)
						code, codeOk := solutionMap_["code"].(string)

						if languageOk && codeOk {
							solutionMap[language] = code
						}
					}
				}
			}
		}

		adminProblems = append(adminProblems, AdminProblem{
			Problem: Problem{
				ID:           problem.ID,
				Title:        problem.Title,
				Description:  problem.Description.String,
				FunctionName: problem.FunctionName,
				Points:       problem.Points,
				Difficulty:   problem.Difficulty,
				Tags:         problem.Tags,
			},
			Solution: solutionMap,
		})
	}

	var response AdminProblemsResponse
	response.Data = adminProblems

	SendJSONResponse(w, http.StatusOK, response)
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
		if responseData.Data.Status != "Accepted" {
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

	SendJSONResponse(w, http.StatusCreated, response)
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

	SendJSONResponse(w, http.StatusOK, responseData)
}
