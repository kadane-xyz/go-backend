package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

type ProblemHint struct {
	Description string `json:"description"`
	Answer      []byte `json:"answer"`
}

type ProblemCode struct {
	Language string `json:"language"`
	Code     []byte `json:"code"`
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
	Code        ProblemRequestCode   `json:"code"`
	Hints       []ProblemRequestHint `json:"hints"`
	Points      int                  `json:"points"`
	Solution    string               `json:"solution"`
	TestCases   []TestCase           `json:"testCases"`
}

type Problem struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	Difficulty  string      `json:"difficulty"`
	Code        interface{} `json:"code"`
	Hints       interface{} `json:"hints"`
	Points      int         `json:"points"`
	Solution    interface{} `json:"solution,omitempty"`
	TestCases   interface{} `json:"testCases"`
	Starred     bool        `json:"starred"`
}

type ProblemResponse struct {
	Data sql.GetProblemRow `json:"data"`
}

type ProblemsResponse struct {
	Data []sql.GetProblemsRow `json:"data"`
}

type ProblemPaginationResponse struct {
	Data       []Problem  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Sort string

const (
	SortAlpha Sort = "alpha"
	SortIndex Sort = "index"
)

// GET: /problems
func (h *Handler) GetProblems(w http.ResponseWriter, r *http.Request) {
	titleSearch := strings.TrimSpace(r.URL.Query().Get("titleSearch"))
	difficulty := strings.TrimSpace(r.URL.Query().Get("difficulty"))
	sortType := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortType == "" {
		sortType = string(SortIndex)
	} else if sortType != string(SortAlpha) && sortType != string(SortIndex) {
		apierror.SendError(w, http.StatusBadRequest, "Invalid sort")
		return
	}

	order := strings.TrimSpace(r.URL.Query().Get("order"))
	if order == "" {
		order = "asc"
	} else if order != "asc" && order != "desc" {
		apierror.SendError(w, http.StatusBadRequest, "Invalid order")
		return
	}

	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.ParseInt(r.URL.Query().Get("perPage"), 10, 64)
	if err != nil || perPage < 1 {
		perPage = 10
	}

	if difficulty != "" {
		valid := (difficulty == string(sql.ProblemDifficultyEasy) ||
			difficulty == string(sql.ProblemDifficultyMedium) ||
			difficulty == string(sql.ProblemDifficultyHard))
		if !valid {
			apierror.SendError(w, http.StatusBadRequest, "Invalid difficulty")
			return
		}
	}

	problems, err := h.PostgresQueries.GetProblemsFilteredPaginated(r.Context(), sql.GetProblemsFilteredPaginatedParams{
		Title:         titleSearch,
		Difficulty:    difficulty,
		Sort:          sortType,
		SortDirection: order,
		PerPage:       int32(perPage),
		Page:          int32(page),
	})
	if err != nil {
		log.Printf("Error getting problems: %v", err)
		apierror.SendError(w, http.StatusInternalServerError, "Failed to get problems")
		return
	}

	totalCount := len(problems)
	if totalCount == 0 {
		response := ProblemPaginationResponse{
			Data:       []Problem{},
			Pagination: Pagination{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	lastPage := (int64(totalCount) + perPage - 1) / perPage

	if lastPage == 0 {
		lastPage = 1
	}

	fromIndex := (page - 1) * perPage
	toIndex := fromIndex + perPage

	if toIndex > int64(len(problems)) {
		toIndex = int64(len(problems))
	}

	if page > lastPage {
		apierror.SendError(w, http.StatusBadRequest, "Page out of bounds")
		return
	}

	paginatedProblems := problems[fromIndex:toIndex]

	responseData := []Problem{}
	for _, problem := range paginatedProblems {
		responseData = append(responseData, Problem{
			ID:          int(problem.ID),
			Title:       problem.Title,
			Description: problem.Description.String,
			Tags:        problem.Tags,
			Difficulty:  string(problem.Difficulty),
			Code:        problem.CodeJson,
			Hints:       problem.HintsJson,
			Points:      int(problem.Points),
			Solution:    problem.SolutionsJson,
			TestCases:   problem.TestCasesJson,
			Starred:     problem.Starred,
		})
	}

	// Return an empty array if no matches (status 200)
	response := ProblemPaginationResponse{
		Data: responseData,
		Pagination: Pagination{
			Page:      page,
			PerPage:   perPage,
			DataCount: int64(len(problems)),
			LastPage:  lastPage,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST: /problems
func (h *Handler) CreateProblem(w http.ResponseWriter, r *http.Request) {
	var request ProblemRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Check problem fields
	if request.Title == "" || request.Description == "" || len(request.Solution) == 0 {
		apierror.SendError(w, http.StatusBadRequest, "Title, description, and solution are required")
		return
	}

	if request.Code.Language == "" || request.Code.Code == "" {
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
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create problem")
		return
	}

	// 2. Create hints using the problem ID
	for _, hint := range request.Hints {
		err = h.PostgresQueries.CreateProblemHint(context.Background(), sql.CreateProblemHintParams{
			ProblemID:   pgtype.Int4{Int32: int32(problemID), Valid: true},
			Description: hint.Description,
			Answer:      hint.Answer,
		})
		if err != nil {
			apierror.SendError(w, http.StatusInternalServerError, "Failed to create hint")
			return
		}
	}

	// 3. Create codes using the problem ID
	err = h.PostgresQueries.CreateProblemCode(context.Background(), sql.CreateProblemCodeParams{
		ProblemID: pgtype.Int4{Int32: int32(problemID), Valid: true},
		Language:  sql.ProblemLanguage(request.Code.Language),
		Code:      request.Code.Code,
	})
	if err != nil {
		apierror.SendError(w, http.StatusInternalServerError, "Failed to create code")
		return
	}

	// 4. Create test cases using the problem ID
	for _, testCase := range request.TestCases {
		_, err = h.PostgresQueries.CreateProblemTestCase(context.Background(), sql.CreateProblemTestCaseParams{
			ProblemID:  pgtype.Int4{Int32: int32(problemID), Valid: true},
			Input:      testCase.Input,
			Output:     testCase.Output,
			Visibility: sql.Visibility(testCase.Visibility),
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
	idInt, err := strconv.Atoi(id)
	if err != nil {
		apierror.SendError(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	problem, err := h.PostgresQueries.GetProblem(context.Background(), int32(idInt))
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
