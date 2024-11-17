package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/judge0"
)

type Submission struct {
	Language   string `json:"language"`
	SourceCode string `json:"sourceCode"`
	Stdin      string `json:"stdin"`
	ProblemID  int    `json:"problemId"`
	Wait       bool   `json:"wait"`
}

type SubmissionResponse struct {
	Data *judge0.SubmissionResponse `json:"data"`
}

type SubmissionResultResponse struct {
	Data *judge0.SubmissionResult `json:"data"`
}

func (h *Handler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	/*userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for comment creation")
		return
	}*/

	var submissionRequest Submission
	err := json.NewDecoder(r.Body).Decode(&submissionRequest)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid submission data format")
		return
	}

	languageID := judge0.LanguageToLanguageID(submissionRequest.Language)

	submission := judge0.Submission{
		LanguageID: languageID,
		SourceCode: submissionRequest.SourceCode,
		Stdin:      submissionRequest.Stdin,
	}

	submissionResponse, err := h.Judge0Client.CreateSubmission(submission)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create submission")
		return
	}

	var response SubmissionResponse
	response.Data = submissionResponse

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	// Get userid from middleware context
	/*userId := r.Context().Value(middleware.FirebaseTokenKey).(middleware.FirebaseTokenInfo).UserID
	if userId == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing user ID for comment creation")
		return
	}*/

	token := chi.URLParam(r, "token")
	if token == "" {
		apierror.SendError(w, http.StatusBadRequest, "Missing token")
		return
	}

	/*problemId, err := strconv.Atoi(r.URL.Query().Get("problemId"))
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}*/

	result, err := h.Judge0Client.GetSubmission(token)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get submission")
		return
	}

	/*hash := sha256.Sum256([]byte(result.Stdout))

	expectedOutputHash, err := h.PostgresQueries.GetProblemSolutionExpectedOutputHash(r.Context(), pgtype.Int8{Int64: int64(problemId), Valid: true})
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem solution expected output hash")
		return
	}

	var response string
	if bytes.Equal(hash[:], expectedOutputHash) {
		response = "correct answer"
	} else {
		response = "wrong answer"
	}

	log.Printf("Response: %s", response)

	w.Header().Set("Content-Type", "application/json")
	if response == "correct answer" {
		json.NewEncoder(w).Encode(SubmissionResultResponse{Data: response})
	} else {
		apierror.SendError(w, http.StatusBadRequest, "Wrong answer")
	}*/

	response := SubmissionResultResponse{Data: result}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
