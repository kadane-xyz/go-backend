package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/api/responses"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/judge0"
	"kadane.xyz/go-backend/v2/internal/judge0tmpl"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type SubmissionHandler struct {
	repo         repository.SubmissionsRepository
	problemsRepo repository.ProblemsRepository
	judge0client *judge0.Judge0Client
}

func NewSubmissionHandler(repo repository.SubmissionsRepository, problemsRepo repository.ProblemsRepository, judge0client *judge0.Judge0Client) *SubmissionHandler {
	return &SubmissionHandler{repo: repo, problemsRepo: problemsRepo, judge0client: judge0client}
}

func validateCreateSubmissionRequest(r *http.Request) (*domain.SubmissionCreateRequest, error) {
	request, err := httputils.DecodeJSONRequest[domain.SubmissionCreateRequest](r)
	if err != nil {
		return nil, errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	problemId := request.ProblemID
	if problemId == 0 {
		return nil, errors.NewApiError(nil, "Missing problem ID", http.StatusBadRequest)
	}

	// Check if language is valid
	lang := string(sql.ProblemLanguage(request.Language))
	if request.Language == "" || request.Language != lang {
		return nil, errors.NewApiError(nil, "Invalid language: "+request.Language, http.StatusBadRequest)
	}

	if request.SourceCode == "" {
		return nil, errors.NewApiError(nil, "Missing source code", http.StatusBadRequest)
	}

	return &domain.SubmissionCreateRequest{
		Language:   request.Language,
		SourceCode: request.SourceCode,
		ProblemID:  request.ProblemID,
	}, nil
}

type testResults struct {
	passedTestCases  int32
	failedTestCase   *domain.RunTestCase
	failedSubmission *domain.Submission
	totalMemory      int32
	totalTime        time.Duration
}

// EvaluateTestResults processes judge0 responses and finds any failures
func EvaluateTestResults(testCases []*domain.TestCase, responses []*judge0.SubmissionResult) (*testResults, error) {
	var testResults testResults
	for i, resp := range responses {
		language := judge0.LanguageIDToLanguage(int(resp.Language.ID))

		// Normalize outputs before comparison
		actualOutput := resp.Stdout
		expectedOutput := testCases[i].Output

		// handle time as time.Duration
		time, err := time.ParseDuration(resp.Time)
		if err != nil {
			return nil, err
		}

		timeStr := time.String()

		// Check for failures
		if actualOutput == "" || actualOutput != expectedOutput || resp.CompileOutput != "" {
			var submissionStatus sql.SubmissionStatus
			if resp.Status.Description != "Accepted" {
				submissionStatus = sql.SubmissionStatus(resp.Status.Description)
			} else {
				submissionStatus = sql.SubmissionStatusWrongAnswer
			}

			testResults.failedSubmission = &domain.Submission{
				Status:        submissionStatus,
				Memory:        int32(resp.Memory),
				Time:          timeStr,
				Stdout:        resp.Stdout,
				Stderr:        resp.Stderr,
				CompileOutput: resp.CompileOutput,
				Message:       resp.Message,
				Language:      language,
			}

			testResults.failedTestCase = &domain.RunTestCase{
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

		testResults.passedTestCases++

		testResults.totalMemory += int32(resp.Memory)
		testResults.totalTime += time
	}

	return &testResults, nil
}

func (h *SubmissionHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	request, err := validateCreateSubmissionRequest(r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	// get problem
	problem, err := h.problemsRepo.GetProblem(r.Context(), &domain.ProblemGetParams{
		ProblemID: request.ProblemID,
		UserID:    claims.UserID,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "get problem")
	}

	// get problem test cases all visibilities
	problemTestCases, err := h.problemsRepo.GetProblemTestCases(r.Context(), &domain.ProblemTestCasesGetParams{
		ProblemID: request.ProblemID,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "get problem test cases")
	}

	// return error if no problem test cases returned
	if len(problemTestCases) == 0 {
		return errors.NewApiError(nil, "No test cases found", http.StatusBadRequest)
	}

	submissions := make([]*judge0.Submission, len(problemTestCases))
	for i, testCase := range problemTestCases {
		submission := judge0tmpl.TemplateCreate(judge0tmpl.TemplateInput{
			Language:     request.Language,
			SourceCode:   request.SourceCode,
			FunctionName: problem.FunctionName,
			TestCase:     *testCase,
			Problem:      *problem,
		})
		submissions[i] = &submission
	}

	if len(submissions) == 0 {
		return errors.NewApiError(nil, "Failed to create submissions", http.StatusInternalServerError)
	}

	// Submit to judge0
	submissionResponses, err := h.judge0client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		return errors.NewAppError(err, "Failed to create submission", http.StatusInternalServerError)
	}

	if len(submissionResponses) == 0 {
		return errors.NewApiError(nil, "Failed to create submissions", http.StatusBadRequest)
	}

	// Evaluate the test results
	testResults, err := EvaluateTestResults(problemTestCases, submissionResponses)
	if err != nil {
		return err
	}

	// Verify passed test cases count
	if testResults.passedTestCases > int32(len(problemTestCases)) {
		return errors.NewAppError(nil, "Passed test cases greater than total test cases", http.StatusInternalServerError)
	}

	// Create the averaged submission
	count := int32(len(submissionResponses))
	lastResp := submissionResponses[len(submissionResponses)-1]

	// If any test failed, use its details, otherwise use averages
	memory := int32(testResults.totalMemory / count)
	avgDuration := testResults.totalTime / time.Duration(count)
	avgDuration = avgDuration.Truncate(time.Millisecond)
	avgDurationStr := avgDuration.String()

	// store averaged results to save to database
	submissionCreateID := uuid.New()

	var failedTestCaseByte []byte

	// convert failedSubmission to []byte if exists
	if testResults.failedSubmission != nil {
		failedTestCaseByte, err = json.Marshal(testResults.failedTestCase)
		if err != nil {
			return err
		}
	}

	submissionCreateParams := domain.SubmissionCreateParams{
		ID:              submissionCreateID,
		Status:          lastResp.Status.Description,
		Memory:          memory,
		Time:            avgDurationStr,
		Stdout:          lastResp.Stdout,
		Stderr:          lastResp.Stderr,
		CompileOutput:   lastResp.CompileOutput,
		Message:         lastResp.Message,
		LanguageID:      int32(lastResp.Language.ID),
		LanguageName:    lastResp.Language.Name,
		AccountID:       claims.UserID,
		ProblemID:       request.ProblemID,
		SubmittedCode:   request.SourceCode,
		FailedTestCase:  failedTestCaseByte,
		PassedTestCases: testResults.passedTestCases,
		TotalTestCases:  int32(len(problemTestCases)),
	}

	// Save to database
	err = h.repo.CreateSubmission(r.Context(), &domain.SubmissionCreateParams{})
	if err != nil {
		return errors.HandleDatabaseError(err, "create submission")
	}

	// Prepare response
	response, err := responses.FromDomainSubmissionCreateParamsToApiSubmissionResponse(&submissionCreateParams)
	if err != nil {
		return err
	}

	httputils.SendJSONResponse(w, http.StatusOK, response)

	return nil
}

func validateGetSubmissionRequest(r *http.Request, userId string) (*domain.SubmissionGetParams, error) {
	token := chi.URLParam(r, "token")
	if token == "" {
		return nil, errors.NewApiError(nil, "Missing submission ID", http.StatusBadRequest)
	}

	idUUID, err := uuid.Parse(token)
	if err != nil {
		return nil, errors.NewApiError(err, "Invalid submission ID", http.StatusBadRequest)
	}

	return &domain.SubmissionGetParams{
		UserID:       userId,
		SubmissionID: idUUID,
	}, nil
}

func (h *SubmissionHandler) GetSubmission(w http.ResponseWriter, r *http.Request) error {
	// Get userid from middleware context
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	params, err := validateGetSubmissionRequest(r, claims.UserID)
	if err != nil {
		return err
	}

	result, err := h.repo.GetSubmission(r.Context(), params)
	if err != nil {
		httputils.EmptyDataResponse(w) // { data: {} }
		return nil
	}

	httputils.SendJSONResponse(w, http.StatusOK, result)

	return nil
}

func (h *SubmissionHandler) GetSubmissionsByUsername(w http.ResponseWriter, r *http.Request) error {
	params, err := validateGetSubmissionsByUsernameRequest(r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	// Get submissions from database
	submissions, err := h.repo.GetSubmissionsByUsername(r.Context(), params)
	if err != nil {
		httputils.EmptyDataArrayResponse(w) // { data: [] }
		return nil
	}

	httputils.SendJSONResponse(w, http.StatusOK, submissions)

	return nil
}

// ExtractSubmissionQueryParams processes and validates query parameters
// New helper function for FetchSubmissionsByUsername
func validateGetSubmissionsByUsernameRequest(r *http.Request) (*domain.SubmissionsGetByUsernameParams, error) {
	username := chi.URLParam(r, "username")
	if username == "" {
		return nil, errors.NewApiError(nil, "Missing username", http.StatusBadRequest)
	}

	// Process problem ID
	id := r.URL.Query().Get("problemId")
	var problemID int32
	if id != "" {
		i, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return nil, err
		}
		if i < 1 {
			return nil, errors.NewAppError(nil, "page number is less than 1", http.StatusBadRequest)
		}
		problemID = int32(i)
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
			return nil, errors.NewApiError(nil, "Invalid status parameter", http.StatusBadRequest)
		}
	}

	// Process order
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "DESC"
	} else if order == "asc" {
		order = "ASC"
	} else if order == "desc" {
		order = "DESC"
	} else {
		order = "DESC" // Default
	}

	sort := r.URL.Query().Get("sort")
	if sort == "runtime" {
		sort = "time"
	} else if sort == "memory" {
		sort = "memory"
	} else if sort == "created" {
		sort = "created_at"
	} else {
		sort = "created_at" // Default to sorting by creation time
	}

	page := r.URL.Query().Get("page")
	var pageInt int32
	if page != "" {
		p, err := strconv.ParseInt(page, 10, 32)
		if err != nil {
			return nil, err
		}
		if p < 1 {
			return nil, errors.NewAppError(nil, "page number is less than 1", http.StatusBadRequest)
		}
		pageInt = int32(p)
	}

	perPage := r.URL.Query().Get("perPage")
	var perPageInt int32
	if page != "" {
		pp, err := strconv.ParseInt(perPage, 10, 32)
		if err != nil {
			return nil, err
		}
		if pp < 1 {
			return nil, errors.NewAppError(nil, "page number is less than 1", http.StatusBadRequest)
		}
		perPageInt = int32(pp)
	}

	return &domain.SubmissionsGetByUsernameParams{
		Username:  username,
		ProblemID: problemID,
		Status:    sql.SubmissionStatus(status),
		Sort:      sql.ProblemSort(sort),
		Order:     sql.SortDirection(order),
		Page:      pageInt,
		PerPage:   perPageInt,
	}, nil
}
