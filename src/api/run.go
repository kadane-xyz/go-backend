package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/judge0"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type RunRequest struct {
	Language   string     `json:"language"`
	SourceCode string     `json:"sourceCode"`
	ProblemID  int        `json:"problemId"`
	TestCases  []TestCase `json:"testCases"`
}

type RunResponse struct {
	Data *judge0.SubmissionResponse `json:"data"`
}

type RunTestCase struct {
	Time           string               `json:"time"`
	Memory         int                  `json:"memory"`
	Status         sql.SubmissionStatus `json:"status"` // Accepted, Wrong Answer, etc
	Input          []TestCaseInput      `json:"input,omitempty"`
	Output         string               `json:"output"`         // User code output
	CompileOutput  string               `json:"compileOutput"`  // Compile output
	ExpectedOutput string               `json:"expectedOutput"` // Solution code output
}

type RunResult struct {
	Id        string               `json:"id,omitempty"`
	Language  string               `json:"language"`
	Time      string               `json:"time"`
	Memory    int32                `json:"memory"`
	TestCases []RunTestCase        `json:"testCases"`
	Status    sql.SubmissionStatus `json:"status"` // Accepted, Wrong Answer, etc
	// Our custom fields
	AccountID string    `json:"accountId"`
	ProblemID int32     `json:"problemId"`
	CreatedAt time.Time `json:"createdAt"`
}

type RunResultResponse struct {
	Data *RunResult `json:"data"`
}

type RunsResponse struct {
	Data []RunResult `json:"data"`
}

// AggregateTestResults aggregates the results of multiple test runs
// Renamed from SummarizeRunResponses to better reflect its purpose
func AggregateTestResults(userId string, problemId int32, sourceCode string, expectedOutput []string, submissionResponses []judge0.SubmissionResult) (RunResult, error) {
	// Initialize the result structure
	runResult := RunResult{
		AccountID: userId,
		ProblemID: problemId,
		CreatedAt: time.Now(),
		Status:    sql.SubmissionStatusAccepted, // Default to Accepted, will override if any failures
		TestCases: make([]RunTestCase, 0, len(submissionResponses)),
	}

	var totalMemory int
	var totalTime float64

	// Process all test cases
	for i, resp := range submissionResponses {
		testCaseStatus := sql.SubmissionStatus(resp.Status.Description)

		// Add test case to results
		runResult.TestCases = append(runResult.TestCases, RunTestCase{
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
func ValidateRunRequest(w http.ResponseWriter, r *http.Request) (RunRequest, *apierror.APIError) {
	runRequest, apiErr := DecodeJSONRequest[RunRequest](r)
	if apiErr != nil {
		return RunRequest{}, apiErr
	}

	if runRequest.ProblemID == 0 {
		return RunRequest{}, apierror.NewError(http.StatusBadRequest, "Missing problem ID")
	}

	// Check if language is valid
	if runRequest.Language != "" {
		lang := string(sql.ProblemLanguage(runRequest.Language))
		if runRequest.Language != lang {
			return RunRequest{}, apierror.NewError(http.StatusBadRequest, "Invalid language: "+runRequest.Language)
		}
	}

	if runRequest.SourceCode == "" {
		return RunRequest{}, apierror.NewError(http.StatusBadRequest, "Missing source code")
	}

	return runRequest, nil
}

// FetchAndValidateProblem gets problem details and validates against request
// Renamed from handleProblem to be more specific about what it does
func (h *Handler) FetchAndValidateProblem(r *http.Request, userId string, runRequest RunRequest) (sql.GetProblemRow, *apierror.APIError) {
	// Get problem
	problem, err := h.PostgresQueries.GetProblem(r.Context(), sql.GetProblemParams{
		ProblemID: int32(runRequest.ProblemID),
		UserID:    userId,
	})
	if err != nil {
		return sql.GetProblemRow{}, apierror.NewError(http.StatusInternalServerError, "Failed to get problem")
	}

	// Check if function name is valid
	if !strings.Contains(runRequest.SourceCode, problem.FunctionName) {
		return sql.GetProblemRow{}, apierror.NewError(http.StatusBadRequest, "Correct function name: "+problem.FunctionName+" not found in "+runRequest.Language+" source code")
	}

	return problem, nil
}

// PrepareJudge0Submissions creates submissions for Judge0 from test cases
// Renamed from createJudge0Submissions to be more descriptive
func (h *Handler) PrepareJudge0Submissions(runRequest RunRequest, testCases []TestCase, problem sql.GetProblemRow) ([]judge0.Submission, *apierror.APIError) {
	var judge0Submissions []judge0.Submission // submissions for judge0 to run

	for _, testCase := range testCases {
		solutionRun := TemplateCreate(TemplateInput{
			Language:     runRequest.Language,
			SourceCode:   runRequest.SourceCode,
			FunctionName: problem.FunctionName,
			TestCase:     testCase,
			Problem: Problem{
				Title:       problem.Title,
				Description: problem.Description.String,
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
		return nil, apierror.NewError(http.StatusBadRequest, "Failed to create runs")
	}

	return judge0Submissions, nil
}

// ProcessTestCaseResults processes each test case result and determines overall status
// Split from handleJudge0Responses to separate concerns
func ProcessTestCaseResults(testCases []TestCase, judge0Responses []judge0.SubmissionResult) ([]RunTestCase, map[string]int, []string) {
	testCaseResults := make([]RunTestCase, len(judge0Responses))
	statusMap := make(map[string]int)
	expectedOutput := make([]string, len(testCases))

	for i, resp := range judge0Responses {
		expectedOutput[i] = testCases[i].Output
		actualOutput := resp.Stdout

		// Determine status using helper function
		status := DetermineTestCaseStatus(resp, actualOutput, testCases[i].Output)

		// Record test case result
		testCaseResults[i] = RunTestCase{
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
// New helper function extracted from handleJudge0Responses
func DetermineTestCaseStatus(response judge0.SubmissionResult, actualOutput, expectedOutput string) sql.SubmissionStatus {
	if response.Status.Description != "Accepted" {
		return sql.SubmissionStatus(response.Status.Description)
	}

	normalizedActual := NormalizeOutput(actualOutput)
	normalizedExpected := NormalizeOutput(expectedOutput)

	if normalizedActual != normalizedExpected || response.CompileOutput != "" {
		return sql.SubmissionStatusWrongAnswer
	}

	return sql.SubmissionStatusAccepted
}

// NormalizeOutput standardizes output for comparison
// New helper function extracted from prior code
func NormalizeOutput(output string) string {
	result := output
	// Remove spaces from array elements
	if strings.Contains(result, "[") {
		result = strings.ReplaceAll(result, " ", "")
	}
	// Remove newlines
	return strings.ReplaceAll(result, "\n", "")
}

// EvaluateRunResults processes judge0 responses and creates final result
// Renamed from handleJudge0Responses to better reflect its purpose
func (h *Handler) EvaluateRunResults(userId string, runRequest RunRequest, testCases []TestCase, judge0Responses []judge0.SubmissionResult) (*RunResult, *apierror.APIError) {
	// Process each test case
	testCaseResults, statusMap, expectedOutput := ProcessTestCaseResults(testCases, judge0Responses)

	// Determine overall status
	overallStatus := DetermineOverallStatus(statusMap, len(testCases))

	// Get the aggregated results which includes all metadata
	finalResult, err := AggregateTestResults(userId, int32(runRequest.ProblemID), runRequest.SourceCode, expectedOutput, judge0Responses)
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to process submission")
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
// New helper function extracted from EvaluateRunResults
func DetermineOverallStatus(statusMap map[string]int, totalTestCases int) sql.SubmissionStatus {
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
// Renamed from CreateRun to better describe what it does
func (h *Handler) ExecuteCodeRun(r *http.Request, userId string, runRequest RunRequest) (*RunResultResponse, *apierror.APIError) {
	// Get problem
	problem, apiErr := h.FetchAndValidateProblem(r, userId, runRequest)
	if apiErr != nil {
		return nil, apiErr
	}

	// Create submissions for judge0
	submissions, apiErr := h.PrepareJudge0Submissions(runRequest, runRequest.TestCases, problem)
	if apiErr != nil {
		return nil, apiErr
	}

	// Send submissions to judge0
	judge0Responses, err := h.Judge0Client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create solution submission")
	}

	// Process results
	runResult, apiErr := h.EvaluateRunResults(userId, runRequest, runRequest.TestCases, judge0Responses)
	if apiErr != nil {
		return nil, apiErr
	}

	return &RunResultResponse{Data: runResult}, nil
}

func (h *Handler) CreateRunRoute(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Validate request body
	body, apiErr := ValidateRunRequest(w, r)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	response, apiErr := h.ExecuteCodeRun(r, userId, body)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	SendJSONResponse(w, http.StatusCreated, response)
}
