package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/judge0"
	"kadane.xyz/go-backend/v2/internal/judge0tmpl"
)

type AdminHandler struct {
	repo           *repository.AdminRepository
	judge0         *judge0.Judge0Client
	problemHandler *ProblemHandler
}

func NewAdminHandler(repo *repository.AdminRepository, judge0 *judge0.Judge0Client, problemHandler *ProblemHandler) *AdminHandler {
	return &AdminHandler{repo: repo, judge0: judge0, problemHandler: problemHandler}
}

func (h *AdminHandler) ProblemRun(runRequest domain.AdminProblemRunRequest) (domain.AdminProblemData, *errors.ApiError) {
	solutionRuns := make(map[string][]judge0.Submission) // Store all judge0 submission inputs for each language

	// Create judge0 submission inputs by combining test case handling and template creation.
	for language, sourceCode := range runRequest.Solutions {
		// Create the submission for this test case and language
		solutionRun := judge0tmpl.TemplateCreate(judge0tmpl.TemplateInput{
			Language:     language,
			SourceCode:   sourceCode,
			FunctionName: runRequest.FunctionName,
			TestCase:     runRequest.TestCase,
		})
		solutionRuns[language] = append(solutionRuns[language], solutionRun)
	}

	// Validate submissions before sending
	if len(solutionRuns) == 0 {
		return domain.AdminProblemData{}, errors.NewApiError(nil, http.StatusBadRequest, "Failed to create runs")
	}

	// Initialize response data
	var responseData domain.AdminProblemData
	responseData.Runs = make(map[string]domain.AdminProblemRunResult)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create submissions for each language
	for language := range solutionRuns {
		wg.Add(1)
		go func(language string, submissions []judge0.Submission) {
			defer wg.Done()
			runResponses, _ := h.judge0.CreateSubmissionBatchAndWait(submissions)
			/*if err != nil {
				SendError(w, http.StatusInternalServerError, "Failed to create solution submission for language: "+language)
				continue
			}*/

			var localTestCase domain.RunTestCase

			// Compare outputs for each test case
			for _, solutionResp := range runResponses {
				testCase := domain.RunTestCase{
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
			result := domain.AdminProblemRunResult{
				TestCase:  localTestCase,
				Status:    sql.SubmissionStatus(responseState),
				CreatedAt: time.Now(),
			}

			// Protect map access with the mutex during write
			mu.Lock()
			responseData.Runs[language] = result
			mu.Unlock()
		}(language, solutionRuns[language])
	}

	wg.Wait()

	// Determine the overall status of all runs
	var status string
	for _, run := range responseData.Runs {
		if run.Status == "Wrong Answer" {
			status = "Wrong Answer"
			break
		} else if run.Status == "Accepted" {
			status = "Accepted"
		}
	}

	// Set response values
	responseData.Status = sql.SubmissionStatus(status)
	responseData.CompletedAt = time.Now()

	return responseData, nil
}

func ProblemRunRequestValidate(runRequest domain.AdminProblemRunRequest) *errors.ApiError {
	if runRequest.FunctionName == "" {
		return errors.NewApiError(nil, http.StatusBadRequest, "Missing function name")
	}

	// check map for missing values
	for language, sourceCode := range runRequest.Solutions {
		// Check if source code is missing
		if sourceCode == "" {
			return errors.NewApiError(nil, http.StatusBadRequest, "Missing source code")
		}

		// Check if language is valid
		lang := string(sql.ProblemLanguage(language))
		if language == "" || language != lang {
			return errors.NewApiError(nil, http.StatusBadRequest, "Invalid language: "+language)
		}

		// Check if function name is valid
		if !strings.Contains(sourceCode, runRequest.FunctionName) {
			return errors.NewApiError(nil, http.StatusBadRequest, "Correct function name: "+runRequest.FunctionName+" not found in "+language+" source code")
		}
	}

	return nil
}

// GET: /admin/validate
func (h *AdminHandler) GetAdminValidation(w http.ResponseWriter, r *http.Request) {
	admin := httputils.GetClientAdmin(w, r)

	response := domain.AdminValidation{
		IsAdmin: admin,
	}

	httputils.SendJSONResponse(w, http.StatusOK, response)
}

// GET: /admin/problems
func (h *AdminHandler) GetAdminProblems(w http.ResponseWriter, r *http.Request) {
	problems, apiErr := h.repo.GetAdminProblems(r.Context())
	if apiErr != nil {
		errors.SendError(w, http.StatusInternalServerError, apiErr.Error())
		return
	}

	var adminProblems []domain.AdminProblem

	for _, problem := range problems {
		// Create a map to store language->code mapping
		solutionMap := make(map[string]string)

		// Handle the solutions data which is already unmarshaled as []interface{}
		if problem.Solutions == nil {
			continue
		}

		// Check if it's already a slice of interfaces
		solutionsArray := problem.Solutions.([]interface{})
		for _, solutionItem := range solutionsArray {
			// Each solution item should be a map[string]interface{}
			solutionMap_, ok := solutionItem.(map[string]interface{})
			if !ok {
				continue
			}

			language, languageOk := solutionMap_["language"].(string)
			code, codeOk := solutionMap_["code"].(string)

			if languageOk && codeOk {
				solutionMap[language] = code
			}
		}

		adminProblems = append(adminProblems, domain.AdminProblem{
			Problem: domain.Problem{
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

	httputils.SendJSONResponse(w, http.StatusOK, adminProblems)
}

// POST: /admin/problems
func (h *AdminHandler) CreateAdminProblem(w http.ResponseWriter, r *http.Request) {
	request, apiErr := httputils.DecodeJSONRequest[domain.ProblemRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	apiErr = CreateProblemRequestValidate(request)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	// Test problem test cases against solutions in each language
	for i, testCase := range request.TestCases {
		responseData, apiErr := h.ProblemRun(domain.AdminProblemRunRequest{
			FunctionName: request.FunctionName,
			Solutions:    request.Solutions,
			TestCase:     testCase,
		})
		if apiErr != nil {
			errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
			return
		}

		// Check if any test cases fail
		if responseData.Status != "Accepted" {
			errors.SendError(w, http.StatusBadRequest, "Wrong answer for test case: "+testCase.Input[i].Name)
			return
		}
	}

	// Create problem in database if all test cases pass
	problemID, dbErr := h.problemHandler.CreateProblem(r.Context(), request)
	if dbErr != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to create problem")
		return
	}

	response := domain.CreateAdminProblemData{
		ProblemID: problemID,
	}

	httputils.SendJSONResponse(w, http.StatusCreated, response)
}

// POST: /admin/problems/run
// Make sure to check test cases for each language
func (h *AdminHandler) CreateAdminProblemRun(w http.ResponseWriter, r *http.Request) {
	runRequest, apiErr := httputils.DecodeJSONRequest[domain.AdminProblemRunRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	// Validate request
	apiErr = ProblemRunRequestValidate(runRequest)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	// Run problems against judge0
	responseData, apiErr := h.ProblemRun(runRequest)
	if apiErr != nil {
		errors.SendError(w, apiErr.Error.StatusCode, apiErr.Error.Message)
		return
	}

	httputils.SendJSONResponse(w, http.StatusOK, responseData)
}
