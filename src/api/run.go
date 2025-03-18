package api

import (
	"encoding/json"
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

// SummarizeRunResponses summarizes the run responses for use in parent RunResult object
func SummarizeRunResponses(userId string, problemId int32, sourceCode string, expectedOutput []string, submissionResponses []judge0.SubmissionResult) (RunResult, error) {
	// Calculate averages from all submission responses
	var totalMemory int
	var totalTime float64
	var runResult RunResult // failed run response if any

	// First pass: check for any failures and collect totals
	for i, resp := range submissionResponses {
		runResult.TestCases = append(runResult.TestCases, RunTestCase{
			Time:           resp.Time,
			Memory:         resp.Memory,
			Status:         sql.SubmissionStatus(resp.Status.Description),
			Output:         resp.Stdout,
			CompileOutput:  resp.CompileOutput,
			ExpectedOutput: expectedOutput[i], // add expected output to test case back
		})

		// Set status to accepted if no status is set
		if runResult.Status == "" {
			runResult.Status = sql.SubmissionStatus("Accepted")
		}

		// Set status to wrong answer if any test case is wrong
		if resp.Status.Description != "Accepted" {
			runResult.Status = sql.SubmissionStatus("Wrong Answer")
		}

		totalMemory += resp.Memory
		if timeVal, err := strconv.ParseFloat(resp.Time, 64); err == nil {
			totalTime += timeVal
		}
	}

	// Create the averaged submission
	count := len(submissionResponses)
	lastResp := submissionResponses[len(submissionResponses)-1]

	response := RunResult{
		Status:    runResult.Status,
		Memory:    int32(totalMemory / count),
		Time:      fmt.Sprintf("%.3f", totalTime/float64(count)),
		Language:  judge0.LanguageIDToLanguage(int(lastResp.Language.ID)),
		TestCases: runResult.TestCases,
		AccountID: userId,
		ProblemID: problemId,
		CreatedAt: time.Now(),
	}

	return response, nil
}

func RunRequestValidate(w http.ResponseWriter, r *http.Request) (RunRequest, *apierror.APIError) {
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

func (h *Handler) handleProblem(r *http.Request, userId string, runRequest RunRequest) (sql.GetProblemRow, *apierror.APIError) {
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

func (h *Handler) handleJudge0Submissions(runRequest RunRequest, testCases []TestCase, problem sql.GetProblemRow) ([]judge0.Submission, *apierror.APIError) {
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

func (h *Handler) handleJudge0Responses(userId string, runRequest RunRequest, testCases []TestCase, judge0Responses []judge0.SubmissionResult) (*RunResult, *apierror.APIError) {
	var runResult RunResult     // run result
	var expectedOutput []string // expected output for each test case
	var err error

	// Check judge0 responses
	for i, judge0Response := range judge0Responses {
		// store user code test case results
		runResult.TestCases = append(runResult.TestCases, RunTestCase{
			Time:           judge0Response.Time,
			Memory:         int(judge0Response.Memory),
			Status:         sql.SubmissionStatus(judge0Response.Status.Description),
			Input:          testCases[i].Input,
			Output:         judge0Response.Stdout,
			CompileOutput:  judge0Response.CompileOutput,
			ExpectedOutput: testCases[i].Output, // solution code output
		})

		// add expected output to array for summarizing
		expectedOutput = append(expectedOutput, testCases[i].Output)

		// check stdout before using status and remove spaces from array elements
		if strings.Contains(judge0Response.Stdout, "[") {
			judge0Response.Stdout = strings.ReplaceAll(judge0Response.Stdout, " ", "")
		}

		// remove newlines from stdout
		if strings.Contains(judge0Response.Stdout, "\n") {
			judge0Response.Stdout = strings.ReplaceAll(judge0Response.Stdout, "\n", "")
		}

		// First check if both executions were successful
		if judge0Response.Status.Description != "Accepted" || judge0Response.Stdout != testCases[i].Output {
			// store user code test case results
			runResult.TestCases[i].Status = sql.SubmissionStatus("Wrong Answer")

			// Update run result status
			runResult.Status = sql.SubmissionStatus("Wrong Answer")
		}
	}

	finalRunResult, err := SummarizeRunResponses(userId, int32(runRequest.ProblemID), runRequest.SourceCode, expectedOutput, judge0Responses)
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to process submission")
	}

	// If all test cases passed, create submission with averaged results
	if runResult.Status == "" {
		runResult = finalRunResult
	}

	return &runResult, nil
}

// Runs should check public test cases first and then user test cases
func (h *Handler) CreateRun(r *http.Request, userId string, runRequest RunRequest) (*RunResultResponse, *apierror.APIError) {
	// Get problem
	problem, apiErr := h.handleProblem(r, userId, runRequest)
	if apiErr != nil {
		return nil, apiErr
	}

	// Create submissions for judge0 for each test case
	judge0Submissions, apiErr := h.handleJudge0Submissions(runRequest, runRequest.TestCases, problem)
	if apiErr != nil {
		return nil, apiErr
	}

	// Send submissions to judge0
	judge0Responses, err := h.Judge0Client.CreateSubmissionBatchAndWait(judge0Submissions)
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create solution submission")
	}

	// Get submission results and test cases
	runResult, apiErr := h.handleJudge0Responses(userId, runRequest, runRequest.TestCases, judge0Responses)
	if apiErr != nil {
		return nil, apiErr
	}

	// Create response
	response := RunResultResponse{
		Data: runResult,
	}

	return &response, nil
}

// POST: /runs
func (h *Handler) CreateRunRoute(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	// Validate request body
	body, apiErr := RunRequestValidate(w, r)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	response, apiErr := h.CreateRun(r, userId, body)
	if apiErr != nil {
		apierror.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
