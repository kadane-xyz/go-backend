package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	Id        string               `json:"id"`
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

func SummarizeSubmissionResponses(userId string, problemId int32, sourceCode string, submissionResponses []judge0.SubmissionResult) (sql.CreateSubmissionParams, error) {
	// Calculate averages from all submission responses
	var totalMemory int
	var totalTime float64
	var failedSubmission *Submission

	// First pass: check for any failures and collect totals
	for _, resp := range submissionResponses {
		if resp.Status.Description != "Accepted" {
			failedSubmission = &Submission{
				Status:        sql.SubmissionStatus(resp.Status.Description),
				Memory:        resp.Memory,
				Time:          resp.Time,
				Stdout:        resp.Stdout,
				Stderr:        resp.Stderr,
				CompileOutput: resp.CompileOutput,
				Message:       resp.Message,
				Language:      judge0.LanguageIDToLanguage(int(resp.Language.ID)),
			}
			break
		}

		totalMemory += resp.Memory
		if timeVal, err := strconv.ParseFloat(resp.Time, 64); err == nil {
			totalTime += timeVal
		}
	}

	// Create the averaged submission
	count := len(submissionResponses)
	lastResp := submissionResponses[len(submissionResponses)-1]

	// If any test failed, use its details, otherwise use averages
	avgSubmission := Submission{
		Status:        sql.SubmissionStatus(lastResp.Status.Description),
		Memory:        int(totalMemory / count),
		Time:          fmt.Sprintf("%.3f", totalTime/float64(count)),
		Stdout:        lastResp.Stdout,
		Stderr:        lastResp.Stderr,
		CompileOutput: lastResp.CompileOutput,
		Message:       lastResp.Message,
	}
	languageID := lastResp.Language.ID
	languageName := lastResp.Language.Name

	if failedSubmission != nil {
		avgSubmission = *failedSubmission
	}

	return sql.CreateSubmissionParams{
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		AccountID:     userId,
		ProblemID:     int32(problemId),
		SubmittedCode: sourceCode,
		Status:        avgSubmission.Status,
		Stdout:        avgSubmission.Stdout,
		Time:          avgSubmission.Time,
		Memory:        int32(avgSubmission.Memory),
		Stderr:        avgSubmission.Stderr,
		CompileOutput: avgSubmission.CompileOutput,
		Message:       avgSubmission.Message,
		LanguageID:    int32(languageID),
		LanguageName:  languageName,
	}, nil
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

// handleTestCases handles the test cases for a problem and appends user test cases to the public test cases
func (h *Handler) handleTestCases(r *http.Request, runRequest RunRequest) (*[]TestCase, *apierror.APIError) {
	var publicTestCases []TestCase // Public test cases include kadane test cases and user test cases

	// Get public test cases
	problemTestCases, err := h.PostgresQueries.GetProblemTestCases(r.Context(), sql.GetProblemTestCasesParams{
		ProblemID:  int32(runRequest.ProblemID),
		Visibility: string(sql.VisibilityPublic),
	})
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to get problem")
	}

	// Add kadane test cases to public test cases
	for _, testCase := range problemTestCases {
		var testCaseInput []TestCaseInput // store test case inputs

		// Handle both empty array and populated array cases
		switch input := testCase.Input.(type) {
		case []interface{}:
			for _, item := range input {
				inputMap := item.(map[string]interface{})
				testCaseInput = append(testCaseInput, TestCaseInput{
					Value: inputMap["value"].(string),
					Type:  TestCaseType(inputMap["type"].(string)), // Use TestCaseType instead of sql.ProblemTestCaseType
				})
			}
		default:
			// Empty array or null case - use empty slice
			testCaseInput = []TestCaseInput{}
		}

		publicTestCases = append(publicTestCases, TestCase{
			Input:  testCaseInput,
			Output: testCase.Output,
		})
	}

	publicTestCases = append(publicTestCases, runRequest.TestCases...) // Add user test cases to public test cases

	return &publicTestCases, nil
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

func (h *Handler) handleJudge0Responses(userId string, runRequest RunRequest, testCases []TestCase, judge0Responses []judge0.SubmissionResult) (*sql.CreateSubmissionParams, []RunTestCase, *apierror.APIError) {
	var runTestCases []RunTestCase
	var dbSubmission sql.CreateSubmissionParams
	var err error

	// Check judge0 responses
	for i, judge0Response := range judge0Responses {
		// store user code test case results
		runTestCases = append(runTestCases, RunTestCase{
			Time:           judge0Response.Time,
			Memory:         int(judge0Response.Memory),
			Status:         sql.SubmissionStatus(judge0Response.Status.Description),
			Output:         judge0Response.Stdout,
			CompileOutput:  judge0Response.CompileOutput,
			ExpectedOutput: testCases[i].Output, // solution code output
		})

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
			runTestCases[i].Status = sql.SubmissionStatus("Wrong Answer")

			// Test case failed - outputs don't match
			dbSubmission, err = SummarizeSubmissionResponses(userId, int32(runRequest.ProblemID), runRequest.SourceCode, judge0Responses[i:i+1])
			if err != nil {
				return nil, nil, apierror.NewError(http.StatusInternalServerError, "Failed to process submission")
			}
			// Update status to Wrong Answer
			dbSubmission.Status = sql.SubmissionStatus("Wrong Answer")
			break // Stop at first mismatch
		}
	}

	// If all test cases passed, create submission with averaged results
	if dbSubmission.Status == "" {
		dbSubmission, err = SummarizeSubmissionResponses(userId, int32(runRequest.ProblemID), runRequest.SourceCode, judge0Responses)
		if err != nil {
			return nil, nil, apierror.NewError(http.StatusInternalServerError, "Failed to process submission")
		}
	}

	dbSubmission.LanguageName = judge0.LanguageIDToLanguage(int(dbSubmission.LanguageID)) // Add language name to submission

	return &dbSubmission, runTestCases, nil
}

// Runs should check public test cases first and then user test cases
func (h *Handler) CreateRun(r *http.Request, userId string, runRequest RunRequest) (*RunResultResponse, *apierror.APIError) {
	// Get problem
	problem, apiErr := h.handleProblem(r, userId, runRequest)
	if apiErr != nil {
		return nil, apiErr
	}

	// Get public test cases and user test cases
	publicTestCases, apiErr := h.handleTestCases(r, runRequest)
	if apiErr != nil {
		return nil, apiErr
	}

	// Create submissions for judge0 for each test case
	judge0Submissions, apiErr := h.handleJudge0Submissions(runRequest, *publicTestCases, problem)
	if apiErr != nil {
		return nil, apiErr
	}

	// Send submissions to judge0
	judge0Responses, err := h.Judge0Client.CreateSubmissionBatchAndWait(judge0Submissions)
	if err != nil {
		return nil, apierror.NewError(http.StatusInternalServerError, "Failed to create solution submission")
	}

	// Get submission results
	dbRunRecord, runTestCases, apiErr := h.handleJudge0Responses(userId, runRequest, *publicTestCases, judge0Responses)
	if apiErr != nil {
		return nil, apiErr
	}

	// Get response state
	var responseState string
	if dbRunRecord.Status == "Accepted" {
		responseState = "Accepted"
	} else {
		responseState = "Wrong Answer"
	}

	// Create response
	response := RunResultResponse{
		Data: &RunResult{
			Id:        uuid.UUID(dbRunRecord.ID.Bytes).String(),
			TestCases: runTestCases,
			Time:      dbRunRecord.Time, // average of test case times
			Memory:    dbRunRecord.Memory,
			Status:    sql.SubmissionStatus(responseState), // if all test cases passed, then Accepted, otherwise Wrong Answer
			Language:  dbRunRecord.LanguageName,
			AccountID: userId,
			ProblemID: int32(runRequest.ProblemID),
			CreatedAt: time.Now(),
		},
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
