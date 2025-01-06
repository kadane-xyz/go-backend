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
	"kadane.xyz/go-backend/v2/src/middleware"
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
	ProblemID      int       `json:"problemId"`
	CreatedAt      time.Time `json:"createdAt"`
	Starred        bool      `json:"starred"`
}

type SubmissionRequest struct {
	Language   string `json:"language"`
	SourceCode string `json:"sourceCode"`
	ProblemID  int    `json:"problemId"`
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
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for comment creation")
		return
	}

	var submissionRequest SubmissionRequest
	err := json.NewDecoder(r.Body).Decode(&submissionRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid submission data format")
		return
	}

	problemId := submissionRequest.ProblemID
	if problemId == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Missing problem ID")
		return
	}

	languageID := judge0.LanguageToLanguageID(submissionRequest.Language)

	submissionId := uuid.New() // unique id for the batch submission to use for db reference

	// get expected output from all test cases
	testCases, err := h.PostgresQueries.GetProblemTestCases(r.Context(), pgtype.Int4{Int32: int32(problemId), Valid: true})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem solution")
		return
	}

	// create submissions for each test case to be used in judge0 batch submission
	submissions := []judge0.Submission{}

	for _, testCase := range testCases {
		submissions = append(submissions, judge0.Submission{
			LanguageID:     languageID,
			SourceCode:     submissionRequest.SourceCode,
			Stdin:          testCase.Input,
			ExpectedOutput: testCase.Output,
		})
	}

	// create judge0 batch submission
	submissionResponses, err := h.Judge0Client.CreateSubmissionBatchAndWait(submissions)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

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
	// store language id and name for db
	lastLanguageID := lastResp.Language.ID
	lastLanguageName := lastResp.Language.Name

	if failedSubmission != nil {
		avgSubmission = *failedSubmission
	}

	dbSubmission := sql.CreateSubmissionParams{
		ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
		AccountID:     userId,
		ProblemID:     int32(problemId),
		SubmittedCode: submissionRequest.SourceCode,
		Status:        avgSubmission.Status,
		Stdout:        pgtype.Text{String: avgSubmission.Stdout, Valid: true},
		Time:          pgtype.Text{String: avgSubmission.Time, Valid: true},
		Memory:        pgtype.Int4{Int32: int32(avgSubmission.Memory), Valid: true},
		Stderr:        pgtype.Text{String: avgSubmission.Stderr, Valid: true},
		CompileOutput: pgtype.Text{String: avgSubmission.CompileOutput, Valid: true},
		Message:       pgtype.Text{String: avgSubmission.Message, Valid: true},
		LanguageID:    int32(lastLanguageID),
		LanguageName:  lastLanguageName,
	}

	// create submission in db
	_, err = h.PostgresQueries.CreateSubmission(r.Context(), dbSubmission)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	response := SubmissionResponse{
		Data: Submission{
			Id:             submissionId.String(),
			Stdout:         avgSubmission.Stdout,
			Time:           avgSubmission.Time,
			Memory:         int(avgSubmission.Memory),
			Stderr:         avgSubmission.Stderr,
			CompileOutput:  avgSubmission.CompileOutput,
			Message:        avgSubmission.Message,
			Status:         avgSubmission.Status,
			Language:       judge0.LanguageIDToLanguage(int(lastLanguageID)),
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
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for comment creation")
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
			ProblemID:      int(result.ProblemID),
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
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for comment creation")
		return
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing username")
		return
	}

	id := r.URL.Query().Get("problemId")
	problemId, err := strconv.Atoi(id)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
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
			ProblemID:      int(submission.ProblemID),
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
