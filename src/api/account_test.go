package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"kadane.xyz/go-backend/v2/src/middleware"
)

func TestGetAccount(t *testing.T) {
	baseReq, err := http.NewRequest("GET", "/accounts/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	baseCtx := context.WithValue(baseReq.Context(), middleware.FirebaseTokenKey, firebaseToken)

	testCases := []struct {
		name           string
		params         map[string]string
		input          string
		expectedStatus int
	}{
		{
			name:           "Get account missing id",
			input:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get account",
			input:          "123abc",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Get account with attributes",
			input:          "123abc",
			params:         map[string]string{"attributes": "true"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseCtx)
			routeCtx := chi.NewRouteContext()
			for key, value := range testcase.params {
				routeCtx.URLParams.Add(key, value)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			handler.GetAccount(w, req)

			if w.Code != testcase.expectedStatus {
				t.Errorf("Expected status %d, got %d", testcase.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAccounts(t *testing.T) {
	baseReq, err := http.NewRequest("GET", "/accounts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	baseCtx := context.WithValue(baseReq.Context(), middleware.FirebaseTokenKey, firebaseToken)

	testCases := []struct {
		name           string
		params         map[string]string
		expectedStatus int
	}{
		{
			name:           "Get accounts by username",
			params:         map[string]string{"usernames": "[\"johndoe\", \"janedoe\"]"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts by locations",
			params:         map[string]string{"locations": "[\"New York\", \"London\"]"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts sort by",
			params:         map[string]string{"sortBy": "createdAt"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts order by",
			params:         map[string]string{"orderBy": "desc"},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get accounts by username and locations",
			params: map[string]string{
				"usernames": "[\"johndoe\", \"janedoe\"]",
				"locations": "[\"New York\", \"London\"]",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get accounts by username locations sort and order",
			params: map[string]string{
				"usernames": "[\"johndoe\", \"janedoe\"]",
				"locations": "[\"New York\", \"London\"]",
				"sortBy":    "createdAt",
				"orderBy":   "desc",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Get accounts by username sort and order",
			params: map[string]string{
				"usernames": "[\"johndoe\", \"janedoe\"]",
				"sortBy":    "level",
				"orderBy":   "asc",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseCtx)
			routeCtx := chi.NewRouteContext()
			for key, value := range testcase.params {
				routeCtx.URLParams.Add(key, value)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()

			handler.GetAccounts(w, req)

			if w.Code != testcase.expectedStatus {
				t.Errorf("Expected status %d, got %d", testcase.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetAccountByUsername(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest("GET", "/accounts/username/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	newctx := context.WithValue(req.Context(), middleware.FirebaseTokenKey, firebaseToken)
	req = req.WithContext(newctx)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("username", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	w := httptest.NewRecorder()
	handler.GetAccountByUsername(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
