package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	ProblemID int       `json:"problemId"`
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
	var failedSubmission *SubmissionResult

	// First pass: check for any failures and collect totals
	for _, resp := range submissionResponses {
		if resp.Status.Description != "Accepted" {
			failedSubmission = &SubmissionResult{
				Status:        sql.SubmissionStatus(resp.Status.Description),
				Memory:        resp.Memory,
				Time:          resp.Time,
				Stdout:        resp.Stdout,
				Stderr:        resp.Stderr,
				CompileOutput: resp.CompileOutput,
				Message:       resp.Message,
				Language:      LanguageInfo{ID: int(resp.Language.ID), Name: resp.Language.Name},
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
	avgSubmission := SubmissionResult{
		Status:        sql.SubmissionStatus(lastResp.Status.Description),
		Memory:        int(totalMemory / count),
		Time:          fmt.Sprintf("%.3f", totalTime/float64(count)),
		Stdout:        lastResp.Stdout,
		Stderr:        lastResp.Stderr,
		CompileOutput: lastResp.CompileOutput,
		Message:       lastResp.Message,
		Language: LanguageInfo{
			ID:   int(lastResp.Language.ID),
			Name: lastResp.Language.Name,
		},
	}

	if failedSubmission != nil {
		avgSubmission = *failedSubmission
	}

	return sql.CreateSubmissionParams{
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		AccountID:     userId,
		ProblemID:     int32(problemId),
		SubmittedCode: sourceCode,
		Status:        avgSubmission.Status,
		Stdout:        pgtype.Text{String: avgSubmission.Stdout, Valid: true},
		Time:          pgtype.Text{String: avgSubmission.Time, Valid: true},
		Memory:        pgtype.Int4{Int32: int32(avgSubmission.Memory), Valid: true},
		Stderr:        pgtype.Text{String: avgSubmission.Stderr, Valid: true},
		CompileOutput: pgtype.Text{String: avgSubmission.CompileOutput, Valid: true},
		Message:       pgtype.Text{String: avgSubmission.Message, Valid: true},
		LanguageID:    int32(avgSubmission.Language.ID),
		LanguageName:  avgSubmission.Language.Name,
	}, nil
}

func (h *Handler) CreateRun(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
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

	problemId := runRequest.ProblemID

	//base64 decode
	decodedSourceCode, err := base64.StdEncoding.DecodeString(runRequest.SourceCode)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid source code")
		return
	}

	languageID := judge0.LanguageToLanguageID(runRequest.Language)

	problemCode, err := h.PostgresQueries.GetProblemCode(r.Context(), pgtype.Int4{Int32: int32(problemId), Valid: true})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	// Validate test cases
	if len(runRequest.TestCases) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "At least one test case is required")
		return
	}

	// create solution runs and custom user runs to compare output
	solutionRuns := []judge0.Submission{}
	userRuns := []judge0.Submission{}

	// create solution runs and custom user runs to compare output
	for _, testCase := range runRequest.TestCases {
		// for each test case input
		for _, input := range testCase.Input {
			solutionRun := judge0.Submission{
				LanguageID:     languageID,
				SourceCode:     []byte(problemCode.Code),
				Stdin:          []byte(input),
				ExpectedOutput: []byte(testCase.Output),
				Wait:           true,
			}
			solutionRuns = append(solutionRuns, solutionRun)

			userRun := judge0.Submission{
				LanguageID:     languageID,
				SourceCode:     decodedSourceCode,
				Stdin:          []byte(input),
				ExpectedOutput: []byte(testCase.Output),
				Wait:           true,
			}
			userRuns = append(userRuns, userRun)
		}
	}

	// Validate submissions before sending
	if len(solutionRuns) == 0 || len(userRuns) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Failed to create submissions")
		return
	}

	// Create and run both batches
	solutionResponses, err := h.Judge0Client.CreateSubmissionBatchAndWait(solutionRuns)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create solution submission")
		return
	}

	userResponses, err := h.Judge0Client.CreateSubmissionBatchAndWait(userRuns)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create user submission")
		return
	}

	var dbSubmission sql.CreateSubmissionParams
	var testCases []RunTestCase

	// Compare outputs for each test case
	for i := 0; i < len(solutionResponses); i++ {
		solutionResp := solutionResponses[i]
		userResp := userResponses[i]

		// store user code test case results
		testCases = append(testCases, RunTestCase{
			Time:           userResp.Time,
			Memory:         int(userResp.Memory),
			Status:         sql.SubmissionStatus(userResp.Status.Description),
			Output:         userResp.Stdout,
			ExpectedOutput: solutionResp.Stdout, // solution code output
		})

		// First check if both executions were successful
		if solutionResp.Status.Description != "Accepted" {
			continue // Skip this test case if solution failed
		}

		if userResp.Status.Description != "Accepted" {
			// User code failed to execute properly
			dbSubmission, err = SummarizeSubmissionResponses(userId, int32(problemId), runRequest.SourceCode, userResponses[i:i+1])
			if err != nil {
				apierror.SendError(w, http.StatusInternalServerError, "Failed to process submission")
				return
			}
			break // Stop at first failure
		}

		// Both executions were successful, compare outputs for this pair
		if solutionResp.Stdout != userResp.Stdout {
			// Test case failed - outputs don't match
			dbSubmission, err = SummarizeSubmissionResponses(userId, int32(problemId), runRequest.SourceCode, userResponses[i:i+1])
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
		dbSubmission, err = SummarizeSubmissionResponses(userId, int32(problemId), runRequest.SourceCode, userResponses)
		if err != nil {
			apierror.SendError(w, http.StatusInternalServerError, "Failed to process submission")
			return
		}
	}

	// create submission in db
	_, err = h.PostgresQueries.CreateSubmission(r.Context(), dbSubmission)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	submissionId := uuid.UUID(dbSubmission.ID.Bytes)

	language := judge0.LanguageIDToLanguage(int(dbSubmission.LanguageID))

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
			Time:      dbSubmission.Time.String,            // average of test case times
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
