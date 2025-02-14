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
	"kadane.xyz/go-backend/v2/src/middleware"
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
	Status         sql.SubmissionStatus `json:"status"`         // Accepted, Wrong Answer, etc
	Output         string               `json:"output"`         // User code output
	ExpectedOutput string               `json:"expectedOutput"` // Solution code output
}

type RunResult struct {
	Id        string               `json:"id"`
	Language  string               `json:"language"`
	Time      string               `json:"time"`
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

// Runs should check public test cases first and then user test cases
// POST: /runs
func (h *Handler) CreateRun(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.ClientTokenKey).(middleware.ClientContext).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for run")
		return
	}

	var runRequest RunRequest
	err := json.NewDecoder(r.Body).Decode(&runRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid run data format")
		return
	}

	if runRequest.ProblemID == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Missing problem ID")
		return
	}

	if runRequest.Language == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing language")
		return
	}

	if runRequest.SourceCode == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing source code")
		return
	}

	problemId := int32(runRequest.ProblemID)

	solutionRuns := []judge0.Submission{}

	problem, err := h.PostgresQueries.GetProblem(r.Context(), sql.GetProblemParams{
		ID:     problemId,
		UserID: userId,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	if problem.FunctionName == "" {
		apierror.SendError(w, http.StatusBadRequest, "Function name is missing from problem")
		return
	}

	problemTestCases, err := h.PostgresQueries.GetProblemTestCases(r.Context(), sql.GetProblemTestCasesParams{
		ProblemID:  problemId,
		Visibility: sql.VisibilityPublic,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	var publicTestCases []TestCase
	for _, testCase := range problemTestCases {
		var testCaseInput []TestCaseInput

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

	runRequest.TestCases = append(runRequest.TestCases, publicTestCases...) // Add public test cases to the end of the user test cases

	// Test Case Runs
	for _, testCase := range runRequest.TestCases {
		// Test Case Runs
		solutionRun := TemplateCreate(TemplateInput{
			Language:     runRequest.Language,
			SourceCode:   runRequest.SourceCode,
			FunctionName: problem.FunctionName,
			TestCases:    testCase,
			Problem: Problem{
				Title:       problem.Title,
				Description: problem.Description.String,
				Tags:        problem.Tags,
				Difficulty:  problem.Difficulty,
				Hints:       problem.HintsJson,
				Points:      problem.Points,
				Solved:      problem.Solved,
			},
		})
		solutionRuns = append(solutionRuns, solutionRun)
	}

	// Validate submissions before sending
	if len(solutionRuns) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Failed to create runs")
		return
	}

	runResponses, err := h.Judge0Client.CreateSubmissionBatchAndWait(solutionRuns)
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
			dbSubmission, err = SummarizeSubmissionResponses(userId, int32(problemId), runRequest.SourceCode, runResponses[i:i+1])
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
		dbSubmission, err = SummarizeSubmissionResponses(userId, int32(problemId), runRequest.SourceCode, runResponses)
		if err != nil {
			apierror.SendError(w, http.StatusInternalServerError, "Failed to process submission")
			return
		}
	}

	language := judge0.LanguageIDToLanguage(int(dbSubmission.LanguageID))
	dbSubmission.LanguageName = language // Add language name to submission

	// create submission in db
	_, err = h.PostgresQueries.CreateSubmission(r.Context(), dbSubmission)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	submissionId := uuid.UUID(dbSubmission.ID.Bytes)

	var responseState string
	if dbSubmission.Status == "Accepted" {
		responseState = "Accepted"
	} else {
		responseState = "Wrong Answer"
	}

	response := RunResultResponse{
		Data: &RunResult{
			Id:        submissionId.String(),
			TestCases: testCases,
			Time:      dbSubmission.Time,                   // average of test case times
			Status:    sql.SubmissionStatus(responseState), // if all test cases passed, then Accepted, otherwise Wrong Answer
			Language:  language,
			AccountID: userId,
			ProblemID: problemId,
			CreatedAt: time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
