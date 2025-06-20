package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"kadane.xyz/go-backend/v2/internal/api/httputils"
	"kadane.xyz/go-backend/v2/internal/api/responses"
	"kadane.xyz/go-backend/v2/internal/database/repository"
	"kadane.xyz/go-backend/v2/internal/domain"
	"kadane.xyz/go-backend/v2/internal/errors"
	"kadane.xyz/go-backend/v2/internal/middleware"
)

type StarredHandler struct {
	repo repository.StarredRepository
}

func NewStarredHandler(repo repository.StarredRepository) *StarredHandler {
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

	httputils.SendJSONDataResponse(w, http.StatusOK, starredProblems)

	return nil
}

// GET: /starred/solutions
func (h *StarredHandler) GetStarredSolutions(w http.ResponseWriter, r *http.Request) error {
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

	httputils.SendJSONDataResponse(w, http.StatusOK, starredSolutions)

	return nil
}

// GET: /starred/submissions
func (h *StarredHandler) GetStarredSubmissions(w http.ResponseWriter, r *http.Request) error {
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

	httputils.SendJSONDataResponse(w, http.StatusOK, starredSubmissions)

	return nil
}

// PUT

// PUT: /starred/problems
func (h *StarredHandler) PutStarProblem(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	problemRequest, err := httputils.DecodeJSONRequest[domain.StarredRequest](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if problemRequest.ID == 0 {
		return errors.NewApiError(err, "invalid problem id", http.StatusBadRequest)
	}

	starred, err := h.repo.StarProblem(r.Context(), &domain.StarProblemParams{
		UserId:    claims.UserID,
		ProblemId: problemRequest.ID.(int32),
	})
	if err != nil {
		//SendError(w, http.StatusInternalServerError, "Failed to star problem")
		return errors.HandleDatabaseError(err, "starred problem")
	}

	response := responses.NewStarredResponse(problemRequest.ID, starred)

	httputils.SendJSONResponse(w, http.StatusOK, response)

	return nil
}

// PUT: /starred/solutions
func (h *StarredHandler) PutStarSolution(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	solutionRequest, err := httputils.DecodeJSONRequest[domain.StarredSolution](r)
	if err != nil {
		return errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if solutionRequest.ID == 0 {
		return errors.NewApiError(nil, "Invalid solution ID", http.StatusBadRequest)
	}

	starred, err := h.repo.StarSolution(r.Context(), &domain.StarSolutionParams{
		UserId:     claims.UserID,
		SolutionId: int32(solutionRequest.ID),
	})
	if err != nil {
		return errors.HandleDatabaseError(err, "starred solution")
	}

	response := responses.NewStarredResponse(solutionRequest.ID, starred)

	httputils.SendJSONResponse(w, http.StatusOK, response)

	return nil
}

func validateStarSubmissionRequest(r *http.Request, userId string) (*domain.StarSubmissionParams, error) {
	submissionRequest, err := httputils.DecodeJSONRequest[domain.StarredSubmission](r)
	if err != nil {
		return nil, errors.NewApiError(err, "validation", http.StatusBadRequest)
	}

	if !submissionRequest.ID.Valid {
		return nil, errors.NewApiError(err, "Invalid submission ID", http.StatusBadRequest)
	}

	idUUID, err := uuid.Parse(submissionRequest.ID.String())
	if err != nil {
		return nil, errors.NewApiError(err, "Invalid submission ID", http.StatusBadRequest)
	}

	return &domain.StarSubmissionParams{
		UserId:       userId,
		SubmissionId: idUUID,
	}, nil
}

// PUT: /starred/submissions
func (h *StarredHandler) PutStarSubmission(w http.ResponseWriter, r *http.Request) error {
	claims, err := middleware.GetClientClaims(r.Context())
	if err != nil {
		return err
	}

	params, err := validateStarSubmissionRequest(r, claims.UserID)
	if err != nil {
		return err
	}

	starred, err := h.repo.StarSubmission(r.Context(), params)
	if err != nil {
		return errors.HandleDatabaseError(err, "starred solution")
	}

	response := responses.NewStarredResponse(params.SubmissionId.String(), starred)

	httputils.SendJSONDataResponse(w, http.StatusOK, response)

	return nil

}
