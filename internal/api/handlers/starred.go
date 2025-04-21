package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/database/sql"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/judge0"
)

type StarredHandler struct {
	repo *repository.StarredRepository
}

func NewStarredHandler(repo *repository.StarredRepository) *StarredHandler {
	return &StarredHandler{repo: repo}
}

// GET: /starred/problems
func (h *StarredHandler) GetStarredProblems(w http.ResponseWriter, r *http.Request) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredProblems, err := h.repo.GetStarredProblems(r.Context(), userId)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to retrieve starred problems")
		return
	}

	if len(starredProblems) == 0 {
		httputils.EmptyDataArrayResponse(w)
		return
	}

	var response []domain.StarredProblem
	for _, problem := range starredProblems {
		response = append(response, domain.StarredProblem{
			ID:          int(problem.ID),
			Title:       problem.Title,
			Description: problem.Description.String,
			Tags:        problem.Tags,
			Difficulty:  string(problem.Difficulty),
			Points:      int(problem.Points),
			Starred:     problem.Starred,
		})
	}

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}

// GET: /starred/solutions
func (h *SolutionsHandler) GetStarredSolutions(w http.ResponseWriter, r *http.Request) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredSolutions, err := h.repo.GetStarredSolutions(r.Context(), userId)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to retrieve starred solutions")
		return
	}

	if len(starredSolutions) == 0 {
		httputils.EmptyDataArrayResponse(w)
		return
	}

	var response []domain.StarredSolution
	for _, solution := range starredSolutions {
		response = append(response, domain.StarredSolution{
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

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}

// GET: /starred/submissions
func (h *SubmissionHandler) GetStarredSubmissions(w http.ResponseWriter, r *http.Request) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	starredSubmissions, err := h.repo.GetStarredSubmissions(r.Context(), userId)
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to retrieve starred submissions")
		return
	}

	if len(starredSubmissions) == 0 {
		httputils.EmptyDataArrayResponse(w)
		return
	}

	var response []domain.StarredSubmission
	for _, submission := range starredSubmissions {
		response = append(response, domain.StarredSubmission{
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

	httputils.SendJSONDataResponse(w, http.StatusOK, response)
}

// PUT

// PUT: /starred/problems
func (h *ProblemHandler) PutStarProblem(w http.ResponseWriter, r *http.Request) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	problemRequest, apiErr := httputils.DecodeJSONRequest[domain.StarProblemRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if problemRequest.ProblemID == 0 {
		errors.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	starred, err := h.repo.PutStarredProblem(r.Context(), sql.PutStarredProblemParams{
		UserID:    userId,
		ProblemID: problemRequest.ProblemID,
	})
	if err != nil {
		//SendError(w, http.StatusInternalServerError, "Failed to star problem")
		errors.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response domain.StarredResponse
	response.Data.ID = problemRequest.ProblemID // Set ID to problem ID
	response.Data.Starred = starred

	httputils.SendJSONResponse(w, http.StatusOK, response)
}

// PUT: /starred/solutions
func (h *SolutionsHandler) PutStarSolution(w http.ResponseWriter, r *http.Request) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	solutionRequest, apiErr := httputils.DecodeJSONRequest[domain.StarSolutionRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if solutionRequest.SolutionID == 0 {
		errors.SendError(w, http.StatusBadRequest, "Invalid solution ID")
		return
	}

	starred, err := h.repo.PutStarredSolution(r.Context(), sql.PutStarredSolutionParams{
		UserID:     userId,
		SolutionID: solutionRequest.SolutionID,
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to star solution")
		return
	}

	var response domain.StarredResponse
	response.Data.ID = solutionRequest.SolutionID // Set ID to solution ID
	response.Data.Starred = starred

	httputils.SendJSONResponse(w, http.StatusOK, response)
}

// PUT: /starred/submissions
func (h *SubmissionHandler) PutStarSubmission(w http.ResponseWriter, r *http.Request) {
	userId, err := httputils.GetClientUserID(w, r)
	if err != nil {
		return
	}

	submissionRequest, apiErr := httputils.DecodeJSONRequest[domain.StarSubmissionRequest](r)
	if apiErr != nil {
		errors.SendError(w, apiErr.StatusCode(), apiErr.Message())
		return
	}

	if submissionRequest.SubmissionID == "" {
		errors.SendError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	idUUID, err := uuid.Parse(submissionRequest.SubmissionID)
	if err != nil {
		errors.SendError(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	submissionID := pgtype.UUID{Bytes: idUUID, Valid: true}

	starred, err := h.repo.PutStarredSubmission(r.Context(), sql.PutStarredSubmissionParams{
		UserID:       userId,
		SubmissionID: submissionID,
	})
	if err != nil {
		errors.SendError(w, http.StatusInternalServerError, "Failed to star submission")
		return
	}

	var response domain.StarredResponse
	response.Data.ID = submissionRequest.SubmissionID // Set ID to submission ID
	response.Data.Starred = starred

	httputils.SendJSONResponse(w, http.StatusOK, response)
}
