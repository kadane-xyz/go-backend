package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
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
	AccountID      string    `json:"accountId"`
	SubmittedCode  string    `json:"submittedCode"`
	SubmittedStdin string    `json:"submittedStdin"`
	ProblemID      int32     `json:"problemId"`
	CreatedAt      time.Time `json:"createdAt"`
	Starred        bool      `json:"starred"`
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

// POST: /submissions
func (h *Handler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	var submissionRequest SubmissionRequest
	err = json.NewDecoder(r.Body).Decode(&submissionRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid submission data format")
		return
	}

	problemId := submissionRequest.ProblemID
	if problemId == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Missing problem ID")
		return
	}

	if submissionRequest.Language == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing language")
		return
	}

	// add check for language

	if submissionRequest.SourceCode == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing source code")
		return
	}

	problem, err := h.PostgresQueries.GetProblem(r.Context(), sql.GetProblemParams{
		ProblemID: int32(problemId),
		UserID:    userId,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	if problem.FunctionName == "" {
		apierror.SendError(w, http.StatusBadRequest, "Function name is missing from problem")
		return
	}

	// get expected output from all test cases
	problemTestCases, err := h.PostgresQueries.GetProblemTestCases(r.Context(), sql.GetProblemTestCasesParams{
		ProblemID:  problemId,
		Visibility: sql.VisibilityPrivate,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem solution")
		return
	}

	if len(problemTestCases) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "No test cases found")
		return
	}

	// create submissions for each test case to be used in judge0 batch submission
	submissions := []judge0.Submission{}

	var privateTestCases []TestCase
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

		privateTestCases = append(privateTestCases, TestCase{
			Input:  testCaseInput,
			Output: testCase.Output,
		})
	}

	if len(privateTestCases) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "No test cases found")
		return
	}

	for _, testCase := range privateTestCases {
		submissionRun := TemplateCreate(TemplateInput{
			Language:     submissionRequest.Language,
			SourceCode:   submissionRequest.SourceCode,
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
		apierror.SendError(w, http.StatusBadRequest, "Failed to create submissions")
		return
	}

	// create judge0 batch submission
	submissionResponses, err := h.Judge0Client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	if len(submissionResponses) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Failed to create submissions")
		return
	}

	// Calculate averages from all submission responses
	var totalMemory int
	var totalTime float64
	var failedSubmission *Submission

	// First pass: check for any failures and collect totals
	for _, resp := range submissionResponses {
		language := judge0.LanguageIDToLanguage(int(resp.Language.ID))
		if resp.Status.Description != "Accepted" {
			failedSubmission = &Submission{
				Status:        sql.SubmissionStatus(resp.Status.Description),
				Memory:        resp.Memory,
				Time:          resp.Time,
				Stdout:        resp.Stdout,
				Stderr:        resp.Stderr,
				CompileOutput: resp.CompileOutput,
				Message:       resp.Message,
				Language:      language,
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
	// store language id and name for db
	lastLanguageID := int32(lastResp.Language.ID)
	lastLanguageName := lastResp.Language.Name

	if failedSubmission != nil {
		avgSubmission = *failedSubmission
	}

	dbSubmission := sql.CreateSubmissionParams{
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		AccountID:     userId,
		ProblemID:     problemId,
		SubmittedCode: submissionRequest.SourceCode,
		Status:        avgSubmission.Status,
		Stdout:        avgSubmission.Stdout,
		Time:          avgSubmission.Time,
		Memory:        int32(avgSubmission.Memory),
		Stderr:        avgSubmission.Stderr,
		CompileOutput: avgSubmission.CompileOutput,
		Message:       avgSubmission.Message,
		LanguageID:    lastLanguageID,
		LanguageName:  lastLanguageName,
	}

	// create submission in db
	_, err = h.PostgresQueries.CreateSubmission(r.Context(), dbSubmission)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	language := judge0.LanguageIDToLanguage(int(lastLanguageID))
	submissionId := uuid.New() // unique id for the batch submission to use for db reference

	response := SubmissionResponse{
		Data: Submission{
			Id:             submissionId.String(),
			Stdout:         avgSubmission.Stdout,
			Time:           avgSubmission.Time,
			Memory:         avgSubmission.Memory,
			Stderr:         avgSubmission.Stderr,
			CompileOutput:  avgSubmission.CompileOutput,
			Message:        avgSubmission.Message,
			Status:         avgSubmission.Status,
			Language:       language,
			AccountID:      userId,
			SubmittedCode:  submissionRequest.SourceCode,
			SubmittedStdin: "",
			ProblemID:      problemId,
			CreatedAt:      time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET: /submissions/:submissionId
func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing submission ID")
		return
	}

	idUUID, err := uuid.Parse(token)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid submission ID")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET: /submissions/username/:username
func (h *Handler) GetSubmissionsByUsername(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing username")
		return
	}

	id := r.URL.Query().Get("problemId")
	var problemId int
	if id != "" {
		problemId, err = strconv.Atoi(id)
		if err != nil {
			apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
			return
		}
	}

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
			apierror.SendError(w, http.StatusBadRequest, "Invalid status parameter")
			return
		}
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "DESC"
	} else if order == "asc" {
		order = "ASC"
	} else if order == "desc" {
		order = "DESC"
	}

	// runtime, memory, createdAt
	sort := r.URL.Query().Get("sort")
	if sort == "runtime" {
		sort = "time"
	} else if sort == "memory" {
		sort = "memory"
	} else if sort == "created" {
		sort = "created_at"
	}

	submissions, err := h.PostgresQueries.GetSubmissionsByUsername(r.Context(), sql.GetSubmissionsByUsernameParams{
		Username:      username,
		ProblemID:     int32(problemId),
		Sort:          sort,
		SortDirection: order,
		Status:        sql.SubmissionStatus(status),
		UserID:        userId,
	})
	if err != nil {
		EmptyDataArrayResponse(w) // { data: [] }
		return
	}

	submissionResults := make([]Submission, 0)
	for _, submission := range submissions {
		submissionId := uuid.UUID(submission.ID.Bytes)
		submissionResults = append(submissionResults, Submission{
			Id:             submissionId.String(),
			Stdout:         submission.Stdout.String,
			Time:           submission.Time.String,
			Memory:         int(submission.Memory.Int32),
			Stderr:         submission.Stderr.String,
			CompileOutput:  submission.CompileOutput.String,
			Message:        submission.Message.String,
			Status:         submission.Status,
			Language:       judge0.LanguageIDToLanguage(int(submission.LanguageID)),
			AccountID:      submission.AccountID,
			SubmittedCode:  submission.SubmittedCode,
			SubmittedStdin: submission.SubmittedStdin.String,
			ProblemID:      submission.ProblemID,
			CreatedAt:      submission.CreatedAt.Time,
			Starred:        submission.Starred,
		})
	}

	response := SubmissionsResponse{
		Data: submissionResults,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
