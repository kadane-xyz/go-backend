package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
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
	Language   string `json:"language"`
	SourceCode string `json:"sourceCode"`
	Stdin      string `json:"stdin"`
	ProblemID  string `json:"problemId"`
	//Wait       bool   `json:"wait"`
}

type SubmissionResponse struct {
	Data *judge0.SubmissionResponse `json:"data"`
}

type SubmissionResult struct {
	Token         string         `json:"token"`
	Stdout        string         `json:"stdout"`
	Time          string         `json:"time"`
	Memory        int            `json:"memory"`
	Stderr        string         `json:"stderr"`
	CompileOutput string         `json:"compileOutput"`
	Message       string         `json:"message"`
	Status        StatusResponse `json:"status"`
	Language      LanguageInfo   `json:"language"`
	// Our custom fields
	AccountID      string    `json:"accountId"`
	SubmittedCode  string    `json:"submittedCode"`
	SubmittedStdin string    `json:"submittedStdin"`
	ProblemID      string    `json:"problemId"`
	CreatedAt      time.Time `json:"createdAt"`
}

type SubmissionResultResponse struct {
	Data *SubmissionResult `json:"data"`
}

type SubmissionsResponse struct {
	Data []SubmissionResult `json:"data"`
}

type StatusResponse struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

type LanguageInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *Handler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for comment creation")
		return
	}

	var submissionRequest Submission
	err := json.NewDecoder(r.Body).Decode(&submissionRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid submission data format")
		return
	}

	problemId := submissionRequest.ProblemID
	if problemId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing problem ID")
		return
	}

	idUUID, err := uuid.Parse(problemId)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	languageID := judge0.LanguageToLanguageID(submissionRequest.Language)

	// get expected output
	expectedOutput, err := h.PostgresQueries.GetProblemSolution(r.Context(), pgtype.UUID{Bytes: idUUID, Valid: true})
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem solution")
		return
	}

	submission := judge0.Submission{
		LanguageID:     languageID,
		SourceCode:     submissionRequest.SourceCode,
		Stdin:          submissionRequest.Stdin,
		ExpectedOutput: base64.StdEncoding.EncodeToString(expectedOutput),
	}

	submissionResponse, err := h.Judge0Client.CreateSubmissionAndWait(submission)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	submissionSql := sql.CreateSubmissionParams{
		Token: submissionResponse.Token,
		Stdout: pgtype.Text{
			String: submissionResponse.Stdout,
			Valid:  submissionResponse.Stdout != "",
		},
		Time: pgtype.Text{
			String: submissionResponse.Time,
			Valid:  submissionResponse.Time != "",
		},
		MemoryUsed: pgtype.Int4{
			Int32: int32(submissionResponse.Memory),
			Valid: true,
		},
		Stderr: pgtype.Text{
			String: submissionResponse.Stderr,
			Valid:  submissionResponse.Stderr != "",
		},
		CompileOutput: pgtype.Text{
			String: submissionResponse.CompileOutput,
			Valid:  submissionResponse.CompileOutput != "",
		},
		Message: pgtype.Text{
			String: submissionResponse.Message,
			Valid:  submissionResponse.Message != "",
		},
		Status:            sql.SubmissionStatus(submissionResponse.Status.Description),
		StatusID:          int32(submissionResponse.Status.ID),
		StatusDescription: submissionResponse.Status.Description,
		LanguageID:        int32(languageID),
		LanguageName:      submissionRequest.Language,
		AccountID:         userId,
		SubmittedCode:     submissionRequest.SourceCode,
		SubmittedStdin:    submissionRequest.Stdin,
		ProblemID:         pgtype.UUID{Bytes: idUUID, Valid: true},
	}

	// create submission in db
	result, err := h.PostgresQueries.CreateSubmission(r.Context(), submissionSql)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	response := SubmissionResultResponse{
		Data: &SubmissionResult{
			Token:         result.Token,
			Stdout:        result.Stdout.String,
			Time:          result.Time.String,
			Memory:        int(result.MemoryUsed.Int32),
			Stderr:        result.Stderr.String,
			CompileOutput: result.CompileOutput.String,
			Message:       result.Message.String,
			Status: StatusResponse{
				ID:          int(result.StatusID),
				Description: result.StatusDescription,
			},
			Language: LanguageInfo{
				ID:   int(result.LanguageID),
				Name: result.LanguageName,
			},
			// Our custom fields
			AccountID:      userId,
			SubmittedCode:  submissionRequest.SourceCode,
			SubmittedStdin: submissionRequest.Stdin,
			ProblemID:      problemId,
			CreatedAt:      result.CreatedAt.Time,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for comment creation")
		return
	}

	token := chi.URLParam(r, "token")
	if token == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing token")
		return
	}

	result, err := h.PostgresQueries.GetSubmissionByToken(r.Context(), token)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get submission")
		return
	}

	problemId := uuid.UUID(result.ProblemID.Bytes).String()

	response := SubmissionResultResponse{
		Data: &SubmissionResult{
			Token:         result.Token,
			Stdout:        result.Stdout.String,
			Time:          result.Time.String,
			Memory:        int(result.MemoryUsed.Int32),
			Stderr:        result.Stderr.String,
			CompileOutput: result.CompileOutput.String,
			Message:       result.Message.String,
			Status: StatusResponse{
				ID:          int(result.StatusID),
				Description: result.StatusDescription,
			},
			Language: LanguageInfo{
				ID:   int(result.LanguageID),
				Name: result.LanguageName,
			},
			// Our custom fields
			AccountID:      userId,
			SubmittedCode:  result.SubmittedCode,
			SubmittedStdin: result.SubmittedStdin,
			ProblemID:      problemId,
			CreatedAt:      result.CreatedAt.Time,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetSubmissionsByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing username")
		return
	}

	var problemUUID pgtype.UUID
	problemId := r.URL.Query().Get("problemId")
	if problemId != "" {
		idUUID, err := uuid.Parse(problemId)
		if err != nil {
			apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
			return
		}
		problemUUID = pgtype.UUID{Bytes: idUUID, Valid: true}
	}

	accountId, err := h.PostgresQueries.GetAccountIDByUsername(r.Context(), username)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get account ID")
		return
	}

	submissions, err := h.PostgresQueries.GetSubmissionsByUsername(r.Context(), sql.GetSubmissionsByUsernameParams{
		AccountID: accountId,
		Column2:   problemUUID,
	})
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get submissions")
		return
	}

	submissionResults := make([]SubmissionResult, 0)
	for _, submission := range submissions {
		problemId := uuid.UUID(submission.ProblemID.Bytes).String()
		submissionResults = append(submissionResults, SubmissionResult{
			Token:         submission.Token,
			Stdout:        submission.Stdout.String,
			Time:          submission.Time.String,
			Memory:        int(submission.MemoryUsed.Int32),
			Stderr:        submission.Stderr.String,
			CompileOutput: submission.CompileOutput.String,
			Message:       submission.Message.String,
			Status: StatusResponse{
				ID:          int(submission.StatusID),
				Description: submission.StatusDescription,
			},
			Language: LanguageInfo{
				ID:   int(submission.LanguageID),
				Name: submission.LanguageName,
			},
			AccountID:      submission.AccountID,
			SubmittedCode:  submission.SubmittedCode,
			SubmittedStdin: submission.SubmittedStdin,
			ProblemID:      problemId,
			CreatedAt:      submission.CreatedAt.Time,
		})
	}

	response := SubmissionsResponse{
		Data: submissionResults,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
