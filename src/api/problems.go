package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"encoding/base64"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type ProblemHint struct {
	Description []byte `json:"description"`
	Answer      []byte `json:"answer"`
}

type ProblemCode struct {
	Language string `json:"language"`
	Code     []byte `json:"code"`
}

type ProblemTestCase struct {
	Description    string `json:"description"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expectedOutput"`
}

type ProblemRequestHint struct {
	Description string `json:"description"`
	Answer      string `json:"answer"`
}

type ProblemRequestCode struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type ProblemRequest struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Tags        []string             `json:"tags"`
	Difficulty  string               `json:"difficulty"`
	Code        []ProblemRequestCode `json:"code"`
	Hints       []ProblemRequestHint `json:"hints"`
	Points      int                  `json:"points"`
	Solution    string               `json:"solution"`
	TestCases   []ProblemTestCase    `json:"testCases"`
}

type Problem struct {
	ID          pgtype.UUID       `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Difficulty  string            `json:"difficulty"`
	Code        []ProblemCode     `json:"code"`
	Hints       []ProblemHint     `json:"hints"`
	Points      int               `json:"points"`
	Solution    string            `json:"solution,omitempty"`
	TestCases   []ProblemTestCase `json:"testCases"`
}

type ProblemResponse struct {
	Data sql.GetProblemRow `json:"data"`
}

type ProblemsResponse struct {
	Data []sql.GetProblemsRow `json:"data"`
}

// GET: /problems
func (h *Handler) GetProblems(w http.ResponseWriter, r *http.Request) {
	problems, err := h.PostgresQueries.GetProblems(r.Context())
	if err != nil {
		log.Printf("Error getting problems: %v", err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problems")
		return
	}

	// Map the SQL response to our API response
	response := ProblemsResponse{
		Data: problems,
	}

	sendJSONResponse(w, http.StatusOK, response)
}

// Helper function to send JSON response
func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// POST: /problems
func (h *Handler) CreateProblem(w http.ResponseWriter, r *http.Request) {
	var request ProblemRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	log.Println(request)

	// Check problem fields
	if request.Title == "" || request.Description == "" || len(request.Solution) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Title, description, and solution are required")
		return
	}

	if len(request.Code) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "At least one code is required")
		return
	}

	if request.Points < 0 {
		apierror.SendError(w, http.StatusBadRequest, "Points must be greater than 0")
		return
	}

	if len(request.Solution) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Solution is required")
		return
	}

	// 1. Create problem first
	problemID, err := h.PostgresQueries.CreateProblem(context.Background(), sql.CreateProblemParams{
		Title:       request.Title,
		Description: pgtype.Text{String: request.Description, Valid: true},
		Points:      int32(request.Points),
		Tags:        request.Tags,
		Difficulty:  sql.ProblemDifficulty(request.Difficulty),
	})
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create problem")
		return
	}

	// 2. Create hints using the problem ID
	for _, hint := range request.Hints {
		descBytes, err := base64.StdEncoding.DecodeString(hint.Description)
		if err != nil {
			apierror.SendError(w, http.StatusBadRequest, "Invalid base64 in hint description")
			return
		}

		answerBytes, err := base64.StdEncoding.DecodeString(hint.Answer)
		if err != nil {
			apierror.SendError(w, http.StatusBadRequest, "Invalid base64 in hint answer")
			return
		}

		err = h.PostgresQueries.CreateProblemHint(context.Background(), sql.CreateProblemHintParams{
			ProblemID:   problemID,
			Description: descBytes,
			Answer:      answerBytes,
		})
		if err != nil {
			log.Println(err)
			apierror.SendError(w, http.StatusInternalServerError, "Failed to create hint")
			return
		}
	}

	// 3. Create codes using the problem ID
	for _, code := range request.Code {
		codeBytes, err := base64.StdEncoding.DecodeString(code.Code)
		if err != nil {
			apierror.SendError(w, http.StatusBadRequest, "Invalid base64 in code")
			return
		}

		err = h.PostgresQueries.CreateProblemCode(context.Background(), sql.CreateProblemCodeParams{
			ProblemID: problemID,
			Language:  sql.ProblemLanguage(code.Language),
			Code:      codeBytes,
		})
		if err != nil {
			log.Println(err)
			apierror.SendError(w, http.StatusInternalServerError, "Failed to create code")
			return
		}
	}

	// 4. Create solution using the problem ID
	solutionBytes, err := base64.StdEncoding.DecodeString(request.Solution)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid base64 in solution")
		return
	}

	_, err = h.PostgresQueries.CreateProblemSolution(context.Background(), sql.CreateProblemSolutionParams{
		ProblemID:      problemID,
		ExpectedOutput: solutionBytes,
	})
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create solution")
		return
	}

	// 5. Create test cases using the problem ID
	for _, testCase := range request.TestCases {
		_, err = h.PostgresQueries.CreateProblemTestCase(context.Background(), sql.CreateProblemTestCaseParams{
			ProblemID:      problemID,
			Description:    testCase.Description,
			Input:          testCase.Input,
			ExpectedOutput: testCase.ExpectedOutput,
		})
		if err != nil {
			log.Println(err)
			apierror.SendError(w, http.StatusInternalServerError, "Failed to create test case")
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

// GET: /problems/:id
func (h *Handler) GetProblem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	idUUID, err := uuid.Parse(id)
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	problem, err := h.PostgresQueries.GetProblem(context.Background(), pgtype.UUID{Bytes: idUUID, Valid: true})
	if err != nil {
		log.Println(err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problem")
		return
	}

	response := ProblemResponse{
		Data: problem,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
