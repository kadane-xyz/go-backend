package handlers

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
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/judge0"
	"kadane.xyz/go-backend/v2/internal/judge0tmpl"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type SubmissionHandler struct {
	repo         *repository.SQLSubmissionsRepository
	judge0client *judge0.Judge0Client
}

func NewSubmissionHandler(repo *repository.SQLSubmissionsRepository, judge0client *judge0.Judge0Client) *SubmissionHandler {
	return &SubmissionHandler{repo: repo, judge0client: judge0client}
}

// ValidateSubmissionRequest validates a submission request
func ValidateSubmissionRequest(request domain.SubmissionRequest) *errors.ApiError {
	problemId := request.ProblemID
	if problemId == 0 {
		return errors.NewApiError(http.StatusBadRequest, "Missing problem ID")
	}

	// Check if language is valid
	lang := string(sql.ProblemLanguage(request.Language))
	if request.Language == "" || request.Language != lang {
		return errors.NewApiError(nil, http.StatusBadRequest, "Invalid language: "+request.Language)
	}

	if request.SourceCode == "" {
		return errors.NewApiError(http.StatusBadRequest, "Missing source code")
	}

	return nil
}

// FetchProblemAndTestCases retrieves problem details and test cases
func (h *SubmissionHandler) FetchProblemAndTestCases(ctx context.Context, problemID int32, userID string) (sql.GetProblemRow, []TestCase, error) {
	// Get problem details
	problem, err := h.repo.GetProblem(ctx, sql.GetProblemParams{
		ProblemID: problemID,
		UserID:    userID,
	})
	if err != nil {
		return sql.GetProblemRow{}, nil, errors.NewApiError(nil, http.StatusInternalServerError, "Failed to get problem")
	}

	// Get problem test cases
	problemTestCases, err := h.repo.GetProblemTestCases(ctx, sql.GetProblemTestCasesParams{
		ProblemID: problemID,
	})
	if err != nil {
		return sql.GetProblemRow{}, nil, errors.NewApiError(http.StatusInternalServerError, "Failed to get problem solution")
	}

	if len(problemTestCases) == 0 {
		return sql.GetProblemRow{}, nil, errors.NewApiError(nil, http.StatusBadRequest, "No test cases found")
	}

	// Convert database test cases to our internal format
	var testCases []domain.TestCase
	for _, testCase := range problemTestCases {
		var testCaseInput []domain.TestCaseInput

		// Handle both empty array and populated array cases
		switch input := testCase.Input.(type) {
		case []any:
			for _, item := range input {
				inputMap := item.(map[string]any)
				testCaseInput = append(testCaseInput, domain.TestCaseInput{
					Name:  inputMap["name"].(string),
					Value: inputMap["value"].(string),
					Type:  domain.TestCaseType(inputMap["type"].(string)),
				})
			}
		default:
			// Empty array or null case - use empty slice
			testCaseInput = []domain.TestCaseInput{}
		}

		testCases = append(testCases, domain.TestCase{
			Input:  testCaseInput,
			Output: testCase.Output,
		})
	}

	if len(testCases) == 0 {
		return sql.GetProblemRow{}, nil, errors.NewApiError(nil, http.StatusBadRequest, "No test cases found")
	}

	return problem, testCases, nil
}

// PrepareSubmissions creates Judge0 submissions for each test case
func (h *SubmissionHandler) PrepareSubmissions(request domain.SubmissionRequest, testCases []domain.TestCase, problem sql.GetProblemRow) ([]judge0.Submission, error) {
	var submissions []judge0.Submission

	for _, testCase := range testCases {
		submissionRun := judge0tmpl.TemplateCreate(judge0tmpl.TemplateInput{
			Language:     request.Language,
			SourceCode:   request.SourceCode,
			FunctionName: problem.FunctionName,
			TestCase:     testCase,
			Problem: domain.Problem{
				Title:       problem.Title,
				Description: problem.Description,
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
		return nil, errors.NewApiError(nil, "Failed to create submissions", http.StatusInternalServerError)
	}

	return submissions, nil
}

// EvaluateTestResults processes judge0 responses and finds any failures
func EvaluateTestResults(testCases []domain.TestCase, responses []judge0.SubmissionResult) (int32, domain.RunTestCase, *domain.Submission, int, float64) {
	var totalMemory int
	var totalTime float64
	var failedSubmission *domain.Submission
	var passedTestCases int32
	var failedTestCase domain.RunTestCase

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

			failedSubmission = &domain.Submission{
				Status:        submissionStatus,
				Memory:        resp.Memory,
				Time:          resp.Time,
				Stdout:        resp.Stdout,
				Stderr:        resp.Stderr,
				CompileOutput: resp.CompileOutput,
				Message:       resp.Message,
				Language:      language,
			}

			failedTestCase = domain.RunTestCase{
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
func CreateDatabaseSubmission(userId string, problem sql.GetProblemRow, request domain.SubmissionRequest,
	submission domain.Submission, failedTestCase domain.RunTestCase,
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
func (h *SubmissionHandler) ProcessSubmission(ctx context.Context, request domain.SubmissionRequest, userId string) (*domain.SubmissionResponse, error) {
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
	submissionResponses, err := h.judge0client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		return nil, errors.NewAppError(error, "Failed to create submission", http.StatusInternalServerError)
	}

	if len(submissionResponses) == 0 {
		return nil, errors.NewApiError(nil, "Failed to create submissions", http.StatusBadRequest)
	}

	// Evaluate the test results
	passedTestCases, failedTestCase, failedSubmission, totalMemory, totalTime :=
		EvaluateTestResults(testCases, submissionResponses)

	// Verify passed test cases count
	if passedTestCases > totalTestCases {
		return nil, errors.NewAppError(nil, "Passed test cases greater than total test cases", http.StatusInternalServerError)
	}

	// Create the averaged submission
	count := len(submissionResponses)
	lastResp := submissionResponses[len(submissionResponses)-1]

	// If any test failed, use its details, otherwise use averages
	memory := totalMemory / count
	avgSubmission := domain.Submission{
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
		return nil, errors.HandleDatabaseError(err, "Failed to create submission")
	}

	// Save to database
	_, err = h.repo.CreateSubmission(ctx, dbSubmission)
	if err != nil {
		return nil, errors.HandleDatabaseError(err, "create submission")
	}

	// Prepare response
	language := judge0.LanguageIDToLanguage(int(lastLanguageID))

	response := domain.Submission{
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
	}

	return &response, nil
}

func (h *SubmissionHandler) CreateSubmissionRoute(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	request, err := httputils.DecodeJSONRequest[domain.SubmissionRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	err = ValidateSubmissionRequest(request)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	response, apiErr := h.ProcessSubmission(r.Context(), request, claims.UserID)
	if apiErr != nil {
		return errors.NewAppError(err, "process submission", http.StatusInternalServerError)
	}

	httputils.SendJSONResponse(w, http.StatusOK, response)

	return nil
}

func (h *SubmissionHandler) GetSubmission(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		return errors.NewApiError(nil, "Missing submission ID", http.StatusBadRequest)
	}

	idUUID, err := uuid.Parse(token)
	if err != nil {
		return errors.NewApiError(err, "Invalid submission ID", http.StatusBadRequest)
	}

	result, err := h.repo.GetSubmissionByID(r.Context(), sql.GetSubmissionByIDParams{
		ID:     pgtype.UUID{Bytes: idUUID, Valid: true},
		UserID: claims.UserID,
	})
	if err != nil {
		httputils.EmptyDataResponse(w) // { data: {} }
		return nil
	}

	httputils.SendJSONResponse(w, http.StatusOK, result)

	return nil
}

func (h *SubmissionHandler) GetSubmissionsByUsername(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		return errors.NewApiError(nil, "Missing username", http.StatusBadRequest)
	}

	// Extract and validate query parameters
	queryParams, err := ExtractSubmissionQueryParams(r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	// Get submissions from database
	submissions, err := h.repo.GetSubmissionsByUsername(r.Context(), sql.GetSubmissionsByUsernameParams{
		Username:      username,
		ProblemID:     int32(queryParams.problemId),
		Sort:          queryParams.sort,
		SortDirection: queryParams.order,
		Status:        sql.SubmissionStatus(queryParams.status),
		UserID:        claims.UserID,
	})
	if err != nil {
		httputils.EmptyDataArrayResponse(w) // { data: [] }
		return nil
	}

	// Transform database results to API response
	submissionResults, err := TransformSubmissionResults(submissions)
	if err != nil {
		return errors.NewAppError(err, "submission result error", http.StatusInternalServerError)
	}

	httputils.SendJSONResponse(w, http.StatusOK, submissionResults)

	return nil
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
func ExtractSubmissionQueryParams(r *http.Request) (SubmissionQueryParams, error) {
	result := SubmissionQueryParams{}

	// Process problem ID
	id := r.URL.Query().Get("problemId")
	if id != "" {
		problemId, err := strconv.Atoi(id)
		if err != nil {
			return result, errors.NewApiError(nil, http.StatusBadRequest, "Invalid problem ID")
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
			return result, errors.NewApiError(nil, http.StatusBadRequest, "Invalid status parameter")
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
func TransformSubmissionResults(submissions []sql.GetSubmissionsByUsernameRow) ([]domain.Submission, error) {
	submissionResults := make([]domain.Submission, 0, len(submissions))

	for _, submission := range submissions {
		submissionId := uuid.UUID(submission.ID.Bytes)
		submissionFailedTestCase := domain.RunTestCase{}

		err := json.Unmarshal(submission.FailedTestCase, &submissionFailedTestCase)
		if err != nil {
			return nil, errors.NewAppError(err, "Failed to unmarshal failed test case", http.StatusInternalServerError)
		}

		submissionResults = append(submissionResults, domain.Submission{
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
