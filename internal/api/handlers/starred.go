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
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type StarredHandler struct {
	repo *repository.StarredRepository
}

func NewStarredHandler(repo *repository.StarredRepository) *StarredHandler {
	return &StarredHandler{repo: repo}
}

// GET: /starred/problems
func (h *StarredHandler) GetStarredProblems(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	starredProblems, err := h.repo.GetStarredProblems(r.Context(), claims.UserID)
	if err != nil {
		return errors.HandleDatabaseError(err, "starred problems")
	}

	if len(starredProblems) == 0 {
		httputils.EmptyDataArrayResponse(w)
		return nil
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

	return nil
}

// GET: /starred/solutions
func (h *SolutionsHandler) GetStarredSolutions(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	starredSolutions, err := h.repo.GetStarredSolutions(r.Context(), claims.UserID)
	if err != nil {
		return errors.HandleDatabaseError(err, "starred solutions")
	}

	if len(starredSolutions) == 0 {
		httputils.EmptyDataArrayResponse(w)
		return nil
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
func (h *SubmissionHandler) GetStarredSubmissions(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	starredSubmissions, err := h.repo.GetStarredSubmissions(r.Context(), claims.UserID)
	if err != nil {
		return errors.HandleDatabaseError(err, "starred submissions")
	}

	if len(starredSubmissions) == 0 {
		httputils.EmptyDataArrayResponse(w)
		return nil
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

	return nil
}

// PUT

// PUT: /starred/problems
func (h *ProblemHandler) PutStarProblem(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	problemRequest, err := httputils.DecodeJSONRequest[domain.StarProblemRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if problemRequest.ProblemID == 0 {
		return errors.NewApiError(err, "invalid problem id", http.StatusBadRequest)
	}

	starred, err := h.repo.PutStarredProblem(r.Context(), sql.PutStarredProblemParams{
		UserID:    claims.UserID,
		ProblemID: problemRequest.ProblemID,
	})
	if err != nil {
		//SendError(w, http.StatusInternalServerError, "Failed to star problem")
		return errors.HandleDatabaseError(err, "starred problem")
	}

	var response domain.StarredResponse
	response.Data.ID = problemRequest.ProblemID // Set ID to problem ID
	response.Data.Starred = starred

	httputils.SendJSONResponse(w, http.StatusOK, response)

	return nil
}

// PUT: /starred/solutions
func (h *SolutionsHandler) PutStarSolution(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	solutionRequest, err := httputils.DecodeJSONRequest[domain.StarSolutionRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if solutionRequest.SolutionID == 0 {
		return errors.NewApiError(w, http.StatusBadRequest, "Invalid solution ID")
	}

	starred, err := h.repo.PutStarredSolution(r.Context(), sql.PutStarredSolutionParams{
		UserID:     claims.UserID,
		SolutionID: solutionRequest.SolutionID,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "starred solution")
	}

	var response domain.StarredResponse
	response.Data.ID = solutionRequest.SolutionID // Set ID to solution ID
	response.Data.Starred = starred

	httputils.SendJSONResponse(w, http.StatusOK, response)

	return nil
}

// PUT: /starred/submissions
func (h *SubmissionHandler) PutStarSubmission(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	submissionRequest, err := httputils.DecodeJSONRequest[domain.StarSubmissionRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if submissionRequest.SubmissionID == "" {
		return errors.NewApiError(err, "Invalid submission ID", http.StatusBadRequest)
	}

	idUUID, err := uuid.Parse(submissionRequest.SubmissionID)
	if err != nil {
		return errors.NewApiError(err, "Invalid submission ID", http.StatusBadRequest)
	}

	submissionID := pgtype.UUID{Bytes: idUUID, Valid: true}

	starred, err := h.repo.PutStarredSubmission(r.Context(), sql.PutStarredSubmissionParams{
		UserID:       claims.UserID,
		SubmissionID: submissionID,
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "starred solution")
	}

	var response domain.StarredResponse
	response.Data.ID = submissionRequest.SubmissionID // Set ID to submission ID
	response.Data.Starred = starred

	httputils.SendJSONResponse(w, http.StatusOK, response)

	return nil

}
