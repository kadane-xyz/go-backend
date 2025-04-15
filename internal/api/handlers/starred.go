package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

type StarredHandler struct {
	starredRepo *repository.StarredRepository
}

func NewStarredHandler(starredRepo *repository.StarredRepository) *StarredHandler {
	return &StarredHandler{starredRepo: starredRepo}
}

// GET: /starred/problems
func (h *StarredHandler) GetStarredProblems(w http.ResponseWriter, r *http.Request) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredProblems, err := h.PostgresQueries.GetStarredProblems(r.Context(), userId)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve starred problems")
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

	SendJSONResponse(w, http.StatusOK, response)
}

// GET: /starred/solutions
func (h *Handler) GetStarredSolutions(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredSolutions, err := h.PostgresQueries.GetStarredSolutions(r.Context(), userId)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve starred solutions")
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

	SendJSONResponse(w, http.StatusOK, response)
}

// GET: /starred/submissions
func (h *Handler) GetStarredSubmissions(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredSubmissions, err := h.PostgresQueries.GetStarredSubmissions(r.Context(), userId)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to retrieve starred submissions")
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

	SendJSONResponse(w, http.StatusOK, response)
}

// PUT

// PUT: /starred/problems
func (h *Handler) PutStarProblem(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	problemRequest, apiErr := DecodeJSONRequest[StarProblemRequest](r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if problemRequest.ProblemID == 0 {
		SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	starred, err := h.PostgresQueries.PutStarredProblem(r.Context(), sql.PutStarredProblemParams{
		UserID:    userId,
		ProblemID: problemRequest.ProblemID,
	})
	if err != nil {
		//SendError(w, http.StatusInternalServerError, "Failed to star problem")
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response StarredResponse
	response.Data.ID = problemRequest.ProblemID // Set ID to problem ID
	response.Data.Starred = starred

	SendJSONResponse(w, http.StatusOK, response)
}

// PUT: /starred/solutions
func (h *Handler) PutStarSolution(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	solutionRequest, apiErr := DecodeJSONRequest[StarSolutionRequest](r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if solutionRequest.SolutionID == 0 {
		SendError(w, http.StatusBadRequest, "Invalid solution ID")
		return
	}

	starred, err := h.PostgresQueries.PutStarredSolution(r.Context(), sql.PutStarredSolutionParams{
		UserID:     userId,
		SolutionID: solutionRequest.SolutionID,
	})
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to star solution")
		return
	}

	var response StarredResponse
	response.Data.ID = solutionRequest.SolutionID // Set ID to solution ID
	response.Data.Starred = starred

	SendJSONResponse(w, http.StatusOK, response)
}

// PUT: /starred/submissions
func (h *Handler) PutStarSubmission(w http.ResponseWriter, r *http.Request) {
	userId, err := GetClientUserID(w, r)
	if err != nil {
		return
	}

	submissionRequest, apiErr := DecodeJSONRequest[StarSubmissionRequest](r)
	if apiErr != nil {
		SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if submissionRequest.SubmissionID == "" {
		SendError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	idUUID, err := uuid.Parse(submissionRequest.SubmissionID)
	if err != nil {
		SendError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	submissionID := pgtype.UUID{Bytes: idUUID, Valid: true}

	starred, err := h.PostgresQueries.PutStarredSubmission(r.Context(), sql.PutStarredSubmissionParams{
		UserID:       userId,
		SubmissionID: submissionID,
	})
	if err != nil {
		SendError(w, http.StatusInternalServerError, "Failed to star submission")
		return
	}

	var response StarredResponse
	response.Data.ID = submissionRequest.SubmissionID // Set ID to submission ID
	response.Data.Starred = starred

	SendJSONResponse(w, http.StatusOK, response)
}
