package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/judge0"
	"kadane.xyz/go-backend/v2/internal/judge0tmpl"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type RunHandler struct {
	repo         repository.ProblemsRepository
	judge0Client *judge0.Judge0Client
}

func NewRunHandler(repo repository.ProblemsRepository, judge0Client *judge0.Judge0Client) *RunHandler {
	return &RunHandler{repo: repo, judge0Client: judge0Client}
}

// AggregateTestResults aggregates the results of multiple test runs
// Renamed from SummarizeRunResponses to better reflect its purpose
func (h *RunHandler) AggregateTestResults(userId string, problemId int32, sourceCode string, expectedOutput []string, submissionResponses []judge0.SubmissionResult) (domain.RunResult, error) {
	// Initialize the result structure
	runResult := domain.RunResult{
		AccountID: userId,
		ProblemID: problemId,
		CreatedAt: time.Now(),
		Status:    sql.SubmissionStatusAccepted, // Default to Accepted, will override if any failures
		TestCases: make([]domain.RunTestCase, 0, len(submissionResponses)),
	}

	var totalMemory int
	var totalTime float64

	// Process all test cases
	for i, resp := range submissionResponses {
		testCaseStatus := sql.SubmissionStatus(resp.Status.Description)

		// Add test case to results
		runResult.TestCases = append(runResult.TestCases, domain.RunTestCase{
			Time:           resp.Time,
			Memory:         resp.Memory,
			Status:         testCaseStatus,
			Output:         resp.Stdout,
			CompileOutput:  resp.CompileOutput,
			ExpectedOutput: expectedOutput[i],
		})

		// If any test case failed, update the overall status
		if testCaseStatus != sql.SubmissionStatusAccepted {
			runResult.Status = testCaseStatus
		}

		// Collect metrics for averaging
		totalMemory += resp.Memory
		if timeVal, err := strconv.ParseFloat(resp.Time, 64); err == nil {
			totalTime += timeVal
		}
	}

	// Calculate averages
	count := len(submissionResponses)
	lastResp := submissionResponses[len(submissionResponses)-1]

	// Set the averaged metrics
	runResult.Memory = int32(totalMemory / count)
	runResult.Time = fmt.Sprintf("%.3f", totalTime/float64(count))
	runResult.Language = judge0.LanguageIDToLanguage(int(lastResp.Language.ID))

	return runResult, nil
}

// ValidateRunRequest validates the incoming run request
// Renamed from RunRequestValidate to follow verb-noun convention
func (h *RunHandler) ValidateRunRequest(w http.ResponseWriter, r *http.Request) (*domain.RunRequest, error) {
	runRequest, err := httputils.DecodeJSONRequest[domain.RunRequest](r)
	if err != nil {
		return nil, err
	}

	if runRequest.ProblemID == 0 {
		return nil, errors.NewApiError(nil, "Missing problem ID", http.StatusBadRequest)
	}

	// Check if language is valid
	if runRequest.Language != "" {
		lang := string(sql.ProblemLanguage(runRequest.Language))
		if runRequest.Language != lang {
			return nil, errors.NewApiError(nil, "Invalid language: "+runRequest.Language, http.StatusBadRequest)
		}
	}

	if runRequest.SourceCode == "" {
		return nil, errors.NewApiError(nil, "Missing source code", http.StatusBadRequest)
	}

	return &runRequest, nil
}

// FetchAndValidateProblem gets problem details and validates against request
func (h *RunHandler) FetchAndValidateProblem(r *http.Request, userId string, runRequest domain.RunRequest) (*domain.Problem, error) {
	// Get problem
	problem, err := h.repo.GetProblem(sql.GetProblemParams{
		ProblemID: int32(runRequest.ProblemID),
		UserID:    userId,
	})
	if err != nil {
		return nil, errors.HandleDatabaseError(err, "get problem")
	}

	// Check if function name is valid
	if !strings.Contains(runRequest.SourceCode, problem.FunctionName) {
		return nil, errors.NewApiError(http.StatusBadRequest, "Correct function name: "+problem.FunctionName+" not found in "+runRequest.Language+" source code")
	}

	return &problem, nil
}

// PrepareJudge0Submissions creates submissions for Judge0 from test cases
func (h *RunHandler) PrepareJudge0Submissions(runRequest domain.RunRequest, testCases []domain.TestCase, problem sql.GetProblemRow) ([]judge0.Submission, error) {
	var judge0Submissions []judge0.Submission // submissions for judge0 to run

	for _, testCase := range testCases {
		description := ""
		if problem.Description != nil {
			description = *problem.Description
		}
		solutionRun := judge0tmpl.TemplateCreate(judge0tmpl.TemplateInput{
			Language:     runRequest.Language,
			SourceCode:   runRequest.SourceCode,
			FunctionName: problem.FunctionName,
			TestCase:     testCase,
			Problem: domain.Problem{
				Title:       problem.Title,
				Description: description,
				Tags:        problem.Tags,
				Difficulty:  problem.Difficulty,
				Hints:       problem.Hints,
				Points:      problem.Points,
				Solved:      problem.Solved,
			},
		})
		judge0Submissions = append(judge0Submissions, solutionRun)
	}

	// Validate submissions before sending
	if len(judge0Submissions) == 0 {
		return nil, errors.NewApiError(nil, "Failed to create runs", http.StatusInternalServerError)
	}

	return judge0Submissions, nil
}

// ProcessTestCaseResults processes each test case result and determines overall status
func (h *RunHandler) ProcessTestCaseResults(testCases []domain.TestCase, judge0Responses []judge0.SubmissionResult) ([]domain.RunTestCase, map[string]int, []string) {
	testCaseResults := make([]domain.RunTestCase, len(judge0Responses))
	statusMap := make(map[string]int)
	expectedOutput := make([]string, len(testCases))

	for i, resp := range judge0Responses {
		expectedOutput[i] = testCases[i].Output
		actualOutput := resp.Stdout

		// Determine status using helper function
		status := h.DetermineTestCaseStatus(resp, actualOutput, testCases[i].Output)

		// Record test case result
		testCaseResults[i] = domain.RunTestCase{
			Time:           resp.Time,
			Memory:         int(resp.Memory),
			Status:         status,
			Input:          testCases[i].Input,
			Output:         resp.Stdout,
			CompileOutput:  resp.CompileOutput,
			ExpectedOutput: testCases[i].Output,
		}

		// Track status counts
		statusMap[string(status)]++
	}

	return testCaseResults, statusMap, expectedOutput
}

// DetermineTestCaseStatus determines the status of a single test case
func (h *RunHandler) DetermineTestCaseStatus(response judge0.SubmissionResult, actualOutput, expectedOutput string) sql.SubmissionStatus {
	if response.Status.Description != "Accepted" {
		return sql.SubmissionStatus(response.Status.Description)
	}

	normalizedActual := h.NormalizeOutput(actualOutput)
	normalizedExpected := h.NormalizeOutput(expectedOutput)

	if normalizedActual != normalizedExpected || response.CompileOutput != "" {
		return sql.SubmissionStatusWrongAnswer
	}

	return sql.SubmissionStatusAccepted
}

// NormalizeOutput standardizes output for comparison
// New helper function extracted from prior code
func (h *RunHandler) NormalizeOutput(output string) string {
	result := output
	// Remove spaces from array elements
	if strings.Contains(result, "[") {
		result = strings.ReplaceAll(result, " ", "")
	}
	// Remove newlines
	return strings.ReplaceAll(result, "\n", "")
}

// EvaluateRunResults processes judge0 responses and creates final result
func (h *RunHandler) EvaluateRunResults(userId string, runRequest domain.RunRequest, testCases []domain.TestCase, judge0Responses []judge0.SubmissionResult) (*domain.RunResult, error) {
	// Process each test case
	testCaseResults, statusMap, expectedOutput := h.ProcessTestCaseResults(testCases, judge0Responses)

	// Determine overall status
	overallStatus := h.DetermineOverallStatus(statusMap, len(testCases))

	// Get the aggregated results which includes all metadata
	finalResult, err := h.AggregateTestResults(userId, int32(runRequest.ProblemID), runRequest.SourceCode, expectedOutput, judge0Responses)
	if err != nil {
		return nil, errors.NewApiError(err, "Failed to process submission", http.StatusInternalServerError)
	}

	// Always use the aggregated results, but override status and test cases if needed
	if overallStatus != sql.SubmissionStatusAccepted {
		// Keep metadata from finalResult but use our calculated status and test cases
		finalResult.Status = overallStatus
		finalResult.TestCases = testCaseResults
	}

	return &finalResult, nil
}

// DetermineOverallStatus calculates the overall submission status
func (h *RunHandler) DetermineOverallStatus(statusMap map[string]int, totalTestCases int) sql.SubmissionStatus {
	if statusMap[string(sql.SubmissionStatusAccepted)] == totalTestCases {
		return sql.SubmissionStatusAccepted
	}

	var errorStatuses []string
	for status, count := range statusMap {
		if status != string(sql.SubmissionStatusAccepted) && count > 0 {
			errorStatuses = append(errorStatuses, status)
		}
	}
	return sql.SubmissionStatus(strings.Join(errorStatuses, ","))
}

// ExecuteCodeRun processes a code run request end-to-end
func (h *RunHandler) ExecuteCodeRun(r *http.Request, userId string, runRequest domain.RunRequest) (*domain.RunResultResponse, error) {
	// Get problem
	problem, err := h.FetchAndValidateProblem(r, userId, runRequest)
	if err != nil {
		return nil, err
	}

	// Create submissions for judge0
	submissions, err := h.PrepareJudge0Submissions(runRequest, runRequest.TestCases, problem)
	if err != nil {
		return nil, err
	}

	// Send submissions to judge0
	judge0Responses, err := h.judge0Client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		return nil, errors.NewApiError(err, "Failed to create solution submission", http.StatusInternalServerError)
	}

	// Process results
	runResult, err := h.EvaluateRunResults(userId, runRequest, runRequest.TestCases, judge0Responses)
	if err != nil {
		return nil, err
	}

	return &domain.RunResultResponse{Data: runResult}, nil
}

func (h *RunHandler) CreateRunRoute(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	// Validate request body
	body, apiErr := h.ValidateRunRequest(w, r)
	if apiErr != nil {
		return errors.NewApiError(nil, "validation", http.StatusBadRequest)
	}

	response, err := h.ExecuteCodeRun(r, claims.UserID, body)
	if err != nil {
		return errors.NewAppError(err, "execute code run", http.StatusInternalServerError)
	}

	httputils.SendJSONResponse(w, http.StatusCreated, response)

	return nil
}
