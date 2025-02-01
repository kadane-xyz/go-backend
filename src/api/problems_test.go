package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/src/middleware"
	"kadane.xyz/go-backend/v2/src/sql/sql"
)

func TestGetProblems(t *testing.T) {
	req, err := http.NewRequest("GET", "/problems", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	newctx := context.WithValue(req.Context(), middleware.FirebaseTokenKey, firebaseToken)
	req = req.WithContext(newctx)

	w := httptest.NewRecorder()
	handler.GetProblems(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCreateProblem(t *testing.T) {
	problem := Problem{
		Title:        "Test Problem",
		Description:  "This is a test problem",
		FunctionName: "testFunction",
		Tags:         []string{"test", "problem"},
		Difficulty:   sql.ProblemDifficultyEasy,
		Code:         map[string]string{"go": "func testFunction() { return 42 }"},
		Hints:        []ProblemRequestHint{{Description: "This is a hint", Answer: "42"}},
		Points:       100,
		Solution:     "func testFunction() { return 42 }",
		TestCases:    []TestCase{{Description: "Test case 1", Input: []TestCaseInput{{Value: "1", Type: "int"}, {Value: "2", Type: "int"}, {Value: "3", Type: "int"}}, Output: "6", Visibility: sql.VisibilityPublic}},
	}

	jsonBody, err := json.Marshal(problem)
	if err != nil {
		t.Fatalf("Failed to marshal problem: %v", err)
	}

	req, err := http.NewRequest("POST", "/problems", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	routeCtx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	newctx := context.WithValue(req.Context(), middleware.FirebaseTokenKey, firebaseToken)
	req = req.WithContext(newctx)

	w := httptest.NewRecorder()
	handler.CreateProblem(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}
