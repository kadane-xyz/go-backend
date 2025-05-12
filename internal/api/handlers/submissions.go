package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	repo         repository.SubmissionsRepository
	problemsRepo repository.ProblemsRepository
	judge0client *judge0.Judge0Client
}

func NewSubmissionHandler(repo repository.SubmissionsRepository, problemsRepo repository.ProblemsRepository, judge0client *judge0.Judge0Client) *SubmissionHandler {
	return &SubmissionHandler{repo: repo, problemsRepo: problemsRepo, judge0client: judge0client}
}

func validateCreateSubmissionRequest(r *http.Request, userId string) (*domain.SubmissionCreateRequest, error) {
	request, err := httputils.DecodeJSONRequest[domain.SubmissionCreateRequest](r)
	if err != nil {
		return nil, errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	problemId := request.ProblemID
	if problemId == 0 {
		return nil, errors.NewApiError(http.StatusBadRequest, "Missing problem ID")
	}

	// Check if language is valid
	lang := string(sql.ProblemLanguage(request.Language))
	if request.Language == "" || request.Language != lang {
		return nil, errors.NewApiError(nil, "Invalid language: "+request.Language, http.StatusBadRequest)
	}

	if request.SourceCode == "" {
		return nil, errors.NewApiError(http.StatusBadRequest, "Missing source code")
	}

	return &domain.SubmissionCreateRequest{
		Language:   request.Language,
		SourceCode: request.SourceCode,
		ProblemID:  request.ProblemID,
	}, nil
}

// FetchProblemAndTestCases retrieves problem details and test cases
func (h *SubmissionHandler) FetchProblemAndTestCases(ctx context.Context, problemID int32, userID string) (*domain.Problem, []domain.TestCase, error) {
	// Get problem details
	problem, err := h.problemsRepo.GetProblem(ctx, &domain.ProblemGetParams{
		ProblemId: problemID,
		UserId:    userID,
	})
	if err != nil {
		return nil, nil, errors.HandleDatabaseError(err, "get problem")
	}

	// Get problem test cases
	problemTestCases, err := h.problemsRepo.GetProblemTestCases(ctx, &domain.ProblemTestCasesGetParams{
		ProblemId: problemID,
	})
	if err != nil {
		return nil, nil, errors.NewApiError(nil, "Failed to get problem solution", http.StatusInternalServerError)
	}

	if len(problemTestCases) == 0 {
		return nil, nil, errors.NewApiError(nil, "No test cases found", http.StatusBadRequest)
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
		return nil, nil, errors.NewApiError(nil, "No test cases found", http.StatusBadRequest)
	}

	return problem, testCases, nil
}

// PrepareSubmissions creates Judge0 submissions for each test case
func (h *SubmissionHandler) PrepareSubmissions(request domain.SubmissionRequest, testCases []domain.TestCase, problem sql.GetProblemRow) ([]judge0.Submission, error) {
	var submissions []judge0.Submission

	for _, testCase := range testCases {
		description := ""
		if problem.Description != nil {
			description = *problem.Description
		}
		submissionRun := judge0tmpl.TemplateCreate(judge0tmpl.TemplateInput{
			Language:     request.Language,
			SourceCode:   request.SourceCode,
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
				Memory:        int32(resp.Memory),
				Time:          resp.Time.Time,
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

type SubmissionCreateParams struct {
	userId          string
	problem         *domain.Problem
	request         domain.SubmissionRequest
	passedTestCases int32
	totalTestCases  int32
	languageId      int32
	languageName    int32
	submissionId    uuid.UUID
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

	// Submit to judge0
	submissionResponses, err := h.judge0client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		return nil, errors.NewAppError(err, "Failed to create submission", http.StatusInternalServerError)
	}

	if len(submissionResponses) == 0 {
		return nil, errors.NewApiError(nil, "Failed to create submissions", http.StatusBadRequest)
	}

	// Evaluate the test results
	passedTestCases, failedTestCase, failedSubmission, totalMemory, totalTime :=
		EvaluateTestResults(testCases, submissionResponses)

	// Total test cases count
	totalTestCases := int32(len(testCases))

	// Verify passed test cases count
	if passedTestCases > totalTestCases {
		return nil, errors.NewAppError(nil, "Passed test cases greater than total test cases", http.StatusInternalServerError)
	}

	// Create the averaged submission
	count := len(submissionResponses)
	lastResp := submissionResponses[len(submissionResponses)-1]

	// If any test failed, use its details, otherwise use averages
	memory := int32(totalMemory / count)
	avgTimeRaw := totalTime / float64(count)
	avgTime := time.Duration(avgTimeRaw * float64(time.Second)).Truncate(time.Millisecond)

	avgSubmission := domain.Submission{
		Status:        sql.SubmissionStatus(lastResp.Status.Description),
		Memory:        memory,
		Time:          avgTime,
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

	// Save to database
	_, err = h.repo.CreateSubmission(ctx, dbSubmission)
	if err != nil {
		return nil, errors.HandleDatabaseError(err, "create submission")
	}

	// Prepare response
	language := judge0.LanguageIDToLanguage(int(lastLanguageID))

	response := domain.Submission{
		Id:              submissionId,
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

	request, err := validateCreateSubmissionRequest(r, claims.UserID)
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
		UserId:       userId,
		SubmissionId: idUUID,
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

// QueryParams holds processed query parameters
type SubmissionQueryParams struct {
	problemId int
	status    string
	order     string
	sort      string
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
	var problemId int32
	if id != "" {
		i, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return nil, err
		}
		if i < 1 {
			return nil, errors.NewAppError(nil, "page number is less than 1", http.StatusBadRequest)
		}
		problemId = int32(i)
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
		ProblemId: problemId,
		Status:    sql.SubmissionStatus(status),
		Sort:      sql.ProblemSort(sort),
		Order:     sql.SortDirection(order),
		Page:      pageInt,
		PerPage:   perPageInt,
	}, nil
}
