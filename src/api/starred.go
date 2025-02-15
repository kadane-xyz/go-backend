package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/judge0"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type Starred struct {
	ID      interface{} `json:"id"` // can be int32 or string
	Starred bool        `json:"starred"`
}

type StarredResponse struct {
	Data Starred `json:"data"`
}

type StarredsResponse struct {
	Data []Starred `json:"data"`
}

// starred problems
type StarProblemRequest struct {
	ProblemID int32 `json:"problemId"`
}

type StarredProblem struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Difficulty  string   `json:"difficulty"`
	Points      int      `json:"points"`
	Starred     bool     `json:"starred"`
}

type StarredProblemsResponse struct {
	Data []StarredProblem `json:"data"`
}

// starred solutions
type StarSolutionRequest struct {
	SolutionID int32 `json:"solutionId"`
}

type StarredSolution struct {
	Id        int64            `json:"id"`
	Username  string           `json:"username"`
	Title     string           `json:"title"`
	Date      pgtype.Timestamp `json:"date"`
	Tags      []string         `json:"tags"`
	Body      string           `json:"body"`
	Votes     int32            `json:"votes"`
	ProblemId int64            `json:"problemId"`
	Starred   bool             `json:"starred"`
}

type StarredSolutionsResponse struct {
	Data []StarredSolution `json:"data"`
}

// starred submissions
type StarSubmissionRequest struct {
	SubmissionID string `json:"submissionId"`
}

type StarredSubmission struct {
	Id            pgtype.UUID          `json:"id"`
	Token         string               `json:"token"`
	Stdout        string               `json:"stdout"`
	Time          string               `json:"time"`
	Memory        int32                `json:"memory"`
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

type StarredSubmissionsResponse struct {
	Data []StarredSubmission `json:"data"`
}

// GET

// GET: /starred/problems
func (h *Handler) GetStarredProblems(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredProblems, err := h.PostgresQueries.GetStarredProblems(r.Context(), userId)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to retrieve starred problems")
		return
	}

	if len(starredProblems) == 0 {
		EmptyDataArrayResponse(w)
		return
	}

	var responseData []StarredProblem
	for _, problem := range starredProblems {
		responseData = append(responseData, StarredProblem{
			ID:          int(problem.ID),
			Title:       problem.Title,
			Description: problem.Description.String,
			Tags:        problem.Tags,
			Difficulty:  string(problem.Difficulty),
			Points:      int(problem.Points),
			Starred:     problem.Starred,
		})
	}

	response := StarredProblemsResponse{
		Data: responseData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET: /starred/solutions
func (h *Handler) GetStarredSolutions(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredSolutions, err := h.PostgresQueries.GetStarredSolutions(r.Context(), userId)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to retrieve starred solutions")
		return
	}

	if len(starredSolutions) == 0 {
		EmptyDataArrayResponse(w)
		return
	}

	var responseData []StarredSolution
	for _, solution := range starredSolutions {
		responseData = append(responseData, StarredSolution{
			Id:        solution.ID,
			Username:  solution.Username,
			Title:     solution.Title,
			Date:      solution.CreatedAt,
			Tags:      solution.Tags,
			Body:      solution.Body,
			Votes:     solution.Votes.Int32,
			ProblemId: solution.ProblemID.Int64,
			Starred:   solution.Starred,
		})
	}

	response := StarredSolutionsResponse{
		Data: responseData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET: /starred/submissions
func (h *Handler) GetStarredSubmissions(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredSubmissions, err := h.PostgresQueries.GetStarredSubmissions(r.Context(), userId)
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to retrieve starred submissions")
		return
	}

	if len(starredSubmissions) == 0 {
		EmptyDataArrayResponse(w)
		return
	}

	var responseData []StarredSubmission
	for _, submission := range starredSubmissions {
		responseData = append(responseData, StarredSubmission{
			Id:             submission.ID,
			Stdout:         submission.Stdout.String,
			Time:           submission.Time.String,
			Memory:         submission.Memory.Int32,
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

	response := StarredSubmissionsResponse{
		Data: responseData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PUT

// PUT: /starred/problems
func (h *Handler) PutStarProblem(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	var problemRequest StarProblemRequest
	err = json.NewDecoder(r.Body).Decode(&problemRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if problemRequest.ProblemID == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	starred, err := h.PostgresQueries.PutStarredProblem(r.Context(), sql.PutStarredProblemParams{
		UserID:    userId,
		ProblemID: problemRequest.ProblemID,
	})
	if err != nil {
		//apierror.SendError(w, http.StatusInternalServerError, "Failed to star problem")
		apierror.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response StarredResponse
	response.Data.ID = problemRequest.ProblemID // Set ID to problem ID
	response.Data.Starred = starred

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PUT: /starred/solutions
func (h *Handler) PutStarSolution(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	var solutionRequest StarSolutionRequest
	err = json.NewDecoder(r.Body).Decode(&solutionRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if solutionRequest.SolutionID == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Invalid solution ID")
		return
	}

	starred, err := h.PostgresQueries.PutStarredSolution(r.Context(), sql.PutStarredSolutionParams{
		UserID:     userId,
		SolutionID: solutionRequest.SolutionID,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to star solution")
		return
	}

	var response StarredResponse
	response.Data.ID = solutionRequest.SolutionID // Set ID to solution ID
	response.Data.Starred = starred

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PUT: /starred/submissions
func (h *Handler) PutStarSubmission(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	var submissionRequest StarSubmissionRequest
	err = json.NewDecoder(r.Body).Decode(&submissionRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if submissionRequest.SubmissionID == "" {
		apierror.SendError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	idUUID, err := uuid.Parse(submissionRequest.SubmissionID)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	submissionID := pgtype.UUID{Bytes: idUUID, Valid: true}

	starred, err := h.PostgresQueries.PutStarredSubmission(r.Context(), sql.PutStarredSubmissionParams{
		UserID:       userId,
		SubmissionID: submissionID,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to star submission")
		return
	}

	var response StarredResponse
	response.Data.ID = submissionRequest.SubmissionID // Set ID to submission ID
	response.Data.Starred = starred

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
