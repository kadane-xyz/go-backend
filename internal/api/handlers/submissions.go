package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/judge0"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Submission struct {
	Id            string               `json:"id"`
	Token         string               `json:"token"`
	Stdout        string               `json:"stdout"`
	Time          string               `json:"time"`
	Memory        int                  `json:"memory"`
	Stderr        string               `json:"stderr"`
	CompileOutput string               `json:"compileOutput"`
	Message       string               `json:"message"`
	Status        sql.SubmissionStatus `json:"status"`
	Language      string               `json:"language"`
	// custom fields
	AccountID       string      `json:"accountId"`
	SubmittedCode   string      `json:"submittedCode"`
	SubmittedStdin  string      `json:"submittedStdin"`
	ProblemID       int32       `json:"problemId"`
	CreatedAt       time.Time   `json:"createdAt"`
	Starred         bool        `json:"starred"`
	FailedTestCase  RunTestCase `json:"failedTestCase,omitempty"`
	PassedTestCases int32       `json:"passedTestCases"`
	TotalTestCases  int32       `json:"totalTestCases"`
}

type SubmissionRequest struct {
	Language   string `json:"language"`
	SourceCode string `json:"sourceCode"`
	ProblemID  int32  `json:"problemId"`
}

type SubmissionResponse struct {
	Data Submission `json:"data"`
}

type SubmissionsResponse struct {
	Data []Submission `json:"data"`
}

// ValidateSubmissionRequest validates a submission request
func ValidateSubmissionRequest(request SubmissionRequest) *APIError {
	problemId := request.ProblemID
	if problemId == 0 {
		return NewError(http.StatusBadRequest, "Missing problem ID")
	}

	// Check if language is valid
	lang := string(sql.ProblemLanguage(request.Language))
	if request.Language == "" || request.Language != lang {
		return NewError(http.StatusBadRequest, "Invalid language: "+request.Language)
	}

	if request.SourceCode == "" {
		return NewError(http.StatusBadRequest, "Missing source code")
	}

	return nil
}

// FetchProblemAndTestCases retrieves problem details and test cases
func (h *Handler) FetchProblemAndTestCases(ctx context.Context, problemID int32, userID string) (sql.GetProblemRow, []TestCase, *APIError) {
	// Get problem details
	problem, err := h.PostgresQueries.GetProblem(ctx, sql.GetProblemParams{
		ProblemID: problemID,
		UserID:    userID,
	})
	if err != nil {
		return sql.GetProblemRow{}, nil, NewError(http.StatusInternalServerError, "Failed to get problem")
	}

	// Get problem test cases
	problemTestCases, err := h.PostgresQueries.GetProblemTestCases(ctx, sql.GetProblemTestCasesParams{
		ProblemID: problemID,
	})
	if err != nil {
		return sql.GetProblemRow{}, nil, NewError(http.StatusInternalServerError, "Failed to get problem solution")
	}

	if len(problemTestCases) == 0 {
		return sql.GetProblemRow{}, nil, NewError(http.StatusBadRequest, "No test cases found")
	}

	// Convert database test cases to our internal format
	var testCases []TestCase
	for _, testCase := range problemTestCases {
		var testCaseInput []TestCaseInput

		// Handle both empty array and populated array cases
		switch input := testCase.Input.(type) {
		case []any:
			for _, item := range input {
				inputMap := item.(map[string]any)
				testCaseInput = append(testCaseInput, TestCaseInput{
					Name:  inputMap["name"].(string),
					Value: inputMap["value"].(string),
					Type:  TestCaseType(inputMap["type"].(string)),
				})
			}
		default:
			// Empty array or null case - use empty slice
			testCaseInput = []TestCaseInput{}
		}

		testCases = append(testCases, TestCase{
			Input:  testCaseInput,
			Output: testCase.Output,
		})
	}

	if len(testCases) == 0 {
		return sql.GetProblemRow{}, nil, NewError(http.StatusBadRequest, "No test cases found")
	}

	return problem, testCases, nil
}

// PrepareSubmissions creates Judge0 submissions for each test case
func (h *Handler) PrepareSubmissions(request SubmissionRequest, testCases []TestCase, problem sql.GetProblemRow) ([]judge0.Submission, *APIError) {
	var submissions []judge0.Submission

	for _, testCase := range testCases {
		submissionRun := TemplateCreate(TemplateInput{
			Language:     request.Language,
			SourceCode:   request.SourceCode,
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
		submissions = append(submissions, submissionRun)
	}

	if len(submissions) == 0 {
		return nil, NewError(http.StatusBadRequest, "Failed to create submissions")
	}

	return submissions, nil
}

// EvaluateTestResults processes judge0 responses and finds any failures
func EvaluateTestResults(testCases []TestCase, responses []judge0.SubmissionResult) (int32, RunTestCase, *Submission, int, float64) {
	var totalMemory int
	var totalTime float64
	var failedSubmission *Submission
	var passedTestCases int32
	var failedTestCase RunTestCase

	for i, resp := range responses {
		language := judge0.LanguageIDToLanguage(int(resp.Language.ID))

		// Normalize outputs before comparison
		actualOutput := NormalizeOutput(resp.Stdout)
		expectedOutput := NormalizeOutput(testCases[i].Output)

		// Check for failures
		if actualOutput == "" || actualOutput != expectedOutput || resp.CompileOutput != "" {
			var submissionStatus sql.SubmissionStatus
			if resp.Status.Description != "Accepted" {
				submissionStatus = sql.SubmissionStatus(resp.Status.Description)
			} else {
				submissionStatus = sql.SubmissionStatusWrongAnswer
			}

			failedSubmission = &Submission{
				Status:        submissionStatus,
				Memory:        resp.Memory,
				Time:          resp.Time,
				Stdout:        resp.Stdout,
				Stderr:        resp.Stderr,
				CompileOutput: resp.CompileOutput,
				Message:       resp.Message,
				Language:      language,
			}

			failedTestCase = RunTestCase{
				Time:           resp.Time,
				Memory:         resp.Memory,
				Status:         submissionStatus,
				Input:          testCases[i].Input,
				Output:         resp.Stdout,
				CompileOutput:  resp.CompileOutput,
				ExpectedOutput: testCases[i].Output,
			}

			break
		}

		passedTestCases++

		totalMemory += resp.Memory
		if timeVal, err := strconv.ParseFloat(resp.Time, 64); err == nil {
			totalTime += timeVal
		}
	}

	return passedTestCases, failedTestCase, failedSubmission, totalMemory, totalTime
}

// CreateDatabaseSubmission prepares the database record for a submission
func CreateDatabaseSubmission(userId string, problem sql.GetProblemRow, request SubmissionRequest,
	submission Submission, failedTestCase RunTestCase,
	passedTestCases, totalTestCases int32, languageID int32, languageName string,
	submissionId uuid.UUID) (sql.CreateSubmissionParams, error) {
	failedTestCaseJson, err := json.Marshal(failedTestCase)
	if err != nil {
		return sql.CreateSubmissionParams{}, err
	}

	return sql.CreateSubmissionParams{
		ID:              pgtype.UUID{Bytes: submissionId, Valid: true},
		AccountID:       userId,
		ProblemID:       problem.ID,
		SubmittedCode:   request.SourceCode,
		Status:          submission.Status,
		Stdout:          submission.Stdout,
		Time:            submission.Time,
		Memory:          int32(submission.Memory),
		Stderr:          submission.Stderr,
		CompileOutput:   submission.CompileOutput,
		Message:         submission.Message,
		LanguageID:      languageID,
		LanguageName:    languageName,
		FailedTestCase:  failedTestCaseJson,
		PassedTestCases: passedTestCases,
		TotalTestCases:  totalTestCases,
	}, nil
}

// ProcessSubmission handles the submission workflow
func (h *Handler) ProcessSubmission(ctx context.Context, request SubmissionRequest, userId string) (*SubmissionResponse, *APIError) {
	// Fetch problem and test cases
	problem, testCases, apiErr := h.FetchProblemAndTestCases(ctx, request.ProblemID, userId)
	if apiErr != nil {
		return nil, apiErr
	}

	// Prepare submissions for judge0
	submissions, apiErr := h.PrepareSubmissions(request, testCases, problem)
	if apiErr != nil {
		return nil, apiErr
	}

	// Total test cases count
	totalTestCases := int32(len(testCases))

	// Submit to judge0
	submissionResponses, err := h.Judge0Client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		return nil, NewError(http.StatusInternalServerError, "Failed to create submission")
	}

	if len(submissionResponses) == 0 {
		return nil, NewError(http.StatusBadRequest, "Failed to create submissions")
	}

	// Evaluate the test results
	passedTestCases, failedTestCase, failedSubmission, totalMemory, totalTime :=
		EvaluateTestResults(testCases, submissionResponses)

	// Verify passed test cases count
	if passedTestCases > totalTestCases {
		return nil, NewError(http.StatusInternalServerError, "Passed test cases greater than total test cases")
	}

	// Create the averaged submission
	count := len(submissionResponses)
	lastResp := submissionResponses[len(submissionResponses)-1]

	// If any test failed, use its details, otherwise use averages
	memory := totalMemory / count
	avgSubmission := Submission{
		Status:        sql.SubmissionStatus(lastResp.Status.Description),
		Memory:        memory,
		Time:          fmt.Sprintf("%.3f", totalTime/float64(count)),
		Stdout:        lastResp.Stdout,
		Stderr:        lastResp.Stderr,
		CompileOutput: lastResp.CompileOutput,
		Message:       lastResp.Message,
	}

	// Use failed submission if available
	if failedSubmission != nil {
		avgSubmission = *failedSubmission
	}

	// Store language ID and name for db
	lastLanguageID := int32(lastResp.Language.ID)
	lastLanguageName := lastResp.Language.Name

	// Create database record
	submissionId := uuid.New()
	dbSubmission, err := CreateDatabaseSubmission(
		userId, problem, request, avgSubmission,
		failedTestCase, passedTestCases, totalTestCases,
		lastLanguageID, lastLanguageName, submissionId)
	if err != nil {
		return nil, NewError(http.StatusInternalServerError, "Failed to create submission")
	}

	// Save to database
	_, err = h.PostgresQueries.CreateSubmission(ctx, dbSubmission)
	if err != nil {
		return nil, NewError(http.StatusInternalServerError, "Failed to create submission")
	}

	// Prepare response
	language := judge0.LanguageIDToLanguage(int(lastLanguageID))

	response := SubmissionResponse{
		Data: Submission{
			Id:              submissionId.String(),
			Stdout:          avgSubmission.Stdout,
			Time:            avgSubmission.Time,
			Memory:          avgSubmission.Memory,
			Stderr:          avgSubmission.Stderr,
			CompileOutput:   avgSubmission.CompileOutput,
			Message:         avgSubmission.Message,
			Status:          avgSubmission.Status,
			Language:        language,
			AccountID:       userId,
			SubmittedCode:   request.SourceCode,
			SubmittedStdin:  "",
			ProblemID:       request.ProblemID,
			CreatedAt:       time.Now(),
			FailedTestCase:  failedTestCase,
			PassedTestCases: passedTestCases,
			TotalTestCases:  totalTestCases,
		},
	}

	return &response, nil
}

func (h *Handler) CreateSubmissionRoute(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	request, apiErr := DecodeJSONRequest[SubmissionRequest](r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	apiErr = ValidateSubmissionRequest(request)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	response, apiErr := h.ProcessSubmission(r.Context(), request, userId)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	SendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		SendError(w, http.StatusBadRequest, "Missing submission ID")
		return
	}

	idUUID, err := uuid.Parse(token)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	result, err := h.PostgresQueries.GetSubmissionByID(r.Context(), sql.GetSubmissionByIDParams{
		ID:     pgtype.UUID{Bytes: idUUID, Valid: true},
		UserID: userId,
	})
	if err != nil {
		EmptyDataResponse(w) // { data: {} }
		return
	}

	response := SubmissionResponse{
		Data: Submission{
			Id:             token,
			Stdout:         result.Stdout.String,
			Time:           result.Time.String,
			Memory:         int(result.Memory.Int32),
			Stderr:         result.Stderr.String,
			CompileOutput:  result.CompileOutput.String,
			Message:        result.Message.String,
			Status:         result.Status,
			Language:       judge0.LanguageIDToLanguage(int(result.LanguageID)),
			AccountID:      userId,
			SubmittedCode:  result.SubmittedCode,
			SubmittedStdin: result.SubmittedStdin.String,
			ProblemID:      result.ProblemID,
			CreatedAt:      result.CreatedAt.Time,
			Starred:        result.Starred,
		},
	}

	SendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) GetSubmissionsByUsername(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		SendError(w, http.StatusBadRequest, "Missing username")
		return
	}

	// Extract and validate query parameters
	queryParams, apiErr := ExtractSubmissionQueryParams(r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	// Get submissions from database
	submissions, err := h.PostgresQueries.GetSubmissionsByUsername(r.Context(), sql.GetSubmissionsByUsernameParams{
		Username:      username,
		ProblemID:     int32(queryParams.problemId),
		Sort:          queryParams.sort,
		SortDirection: queryParams.order,
		Status:        sql.SubmissionStatus(queryParams.status),
		UserID:        userId,
	})
	if err != nil {
		EmptyDataArrayResponse(w) // { data: [] }
		return
	}

	// Transform database results to API response
	submissionResults, apiErr := TransformSubmissionResults(submissions)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	response := SubmissionsResponse{
		Data: submissionResults,
	}

	SendJSONResponse(w, http.StatusOK, response)
}

// QueryParams holds processed query parameters
type SubmissionQueryParams struct {
	problemId int
	status    string
	order     string
	sort      string
}

// ExtractSubmissionQueryParams processes and validates query parameters
// New helper function for FetchSubmissionsByUsername
func ExtractSubmissionQueryParams(r *http.Request) (SubmissionQueryParams, *APIError) {
	result := SubmissionQueryParams{}

	// Process problem ID
	id := r.URL.Query().Get("problemId")
	if id != "" {
		problemId, err := strconv.Atoi(id)
		if err != nil {
			return result, NewError(http.StatusBadRequest, "Invalid problem ID")
		}
		result.problemId = problemId
	}

	// Process status
	status := r.URL.Query().Get("status")
	if status != "" {
		// Check if status is valid
		validStatuses := []sql.SubmissionStatus{
			sql.SubmissionStatusAccepted,
			sql.SubmissionStatusWrongAnswer,
			sql.SubmissionStatusTimeLimitExceeded,
			sql.SubmissionStatusCompilationError,
			sql.SubmissionStatusRuntimeErrorSIGSEGV,
			sql.SubmissionStatusRuntimeErrorSIGXFSZ,
			sql.SubmissionStatusRuntimeErrorSIGFPE,
			sql.SubmissionStatusRuntimeErrorSIGABRT,
			sql.SubmissionStatusRuntimeErrorNZEC,
			sql.SubmissionStatusRuntimeErrorOther,
			sql.SubmissionStatusInternalError,
			sql.SubmissionStatusExecFormatError,
		}

		isValid := false
		for _, validStatus := range validStatuses {
			if sql.SubmissionStatus(status) == validStatus {
				isValid = true
				break
			}
		}

		if !isValid {
			return result, NewError(http.StatusBadRequest, "Invalid status parameter")
		}
		result.status = status
	}

	// Process order
	order := r.URL.Query().Get("order")
	if order == "" {
		result.order = "DESC"
	} else if order == "asc" {
		result.order = "ASC"
	} else if order == "desc" {
		result.order = "DESC"
	} else {
		result.order = "DESC" // Default
	}

	// Process sort
	sort := r.URL.Query().Get("sort")
	if sort == "runtime" {
		result.sort = "time"
	} else if sort == "memory" {
		result.sort = "memory"
	} else if sort == "created" {
		result.sort = "created_at"
	} else {
		result.sort = "created_at" // Default to sorting by creation time
	}

	return result, nil
}

// TransformSubmissionResults converts database results to API response format
func TransformSubmissionResults(submissions []sql.GetSubmissionsByUsernameRow) ([]Submission, *APIError) {
	submissionResults := make([]Submission, 0, len(submissions))

	for _, submission := range submissions {
		submissionId := uuid.UUID(submission.ID.Bytes)
		submissionFailedTestCase := RunTestCase{}

		err := json.Unmarshal(submission.FailedTestCase, &submissionFailedTestCase)
		if err != nil {
			return nil, NewError(http.StatusInternalServerError, "Failed to unmarshal failed test case")
		}

		submissionResults = append(submissionResults, Submission{
			Id:              submissionId.String(),
			Stdout:          submission.Stdout.String,
			Time:            submission.Time.String,
			Memory:          int(submission.Memory.Int32),
			Stderr:          submission.Stderr.String,
			CompileOutput:   submission.CompileOutput.String,
			Message:         submission.Message.String,
			Status:          submission.Status,
			Language:        judge0.LanguageIDToLanguage(int(submission.LanguageID)),
			AccountID:       submission.AccountID,
			SubmittedCode:   submission.SubmittedCode,
			SubmittedStdin:  submission.SubmittedStdin.String,
			ProblemID:       submission.ProblemID,
			CreatedAt:       submission.CreatedAt.Time,
			Starred:         submission.Starred,
			FailedTestCase:  submissionFailedTestCase,
			PassedTestCases: submission.PassedTestCases.Int32,
			TotalTestCases:  submission.TotalTestCases.Int32,
		})
	}

	return submissionResults, nil
}
