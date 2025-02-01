package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
		urlParamId     string
		queryParams    map[string]string
		expectedStatus int
	}{
		{
			name:           "Get account missing id",
			urlParamId:     "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get account",
			urlParamId:     "123abc",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get account with attributes",
			urlParamId:     "123abc",
			queryParams:    map[string]string{"attributes": "true"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseCtx)

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", testcase.urlParamId)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			if testCase.queryParams != nil {
				q := req.URL.Query()
				for key, value := range testCase.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			handler.GetAccount(w, req)

			if w.Code != testcase.expectedStatus {
				var errMsg string
				if w.Body.Len() > 0 {
					var response struct {
						Error struct {
							StatusCode int    `json:"statusCode"`
							Message    string `json:"message"`
						} `json:"error"`
					}
					if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
						errMsg = w.Body.String() // Fallback to raw body if JSON parsing fails
					} else {
						errMsg = response.Error.Message
					}
				} else {
					errMsg = "no response body"
				}
				t.Errorf("Expected status %d, got %d, message: %s", testcase.expectedStatus, w.Code, errMsg)
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
				var errMsg string
				if w.Body.Len() > 0 {
					var response struct {
						Error struct {
							StatusCode int    `json:"statusCode"`
							Message    string `json:"message"`
						} `json:"error"`
					}
					if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
						errMsg = w.Body.String() // Fallback to raw body if JSON parsing fails
					} else {
						errMsg = response.Error.Message
					}
				} else {
					errMsg = "no response body"
				}
				t.Errorf("Expected status %d, got %d, message: %s", testcase.expectedStatus, w.Code, errMsg)
			}
		})
	}
}

func TestGetAccountByUsername(t *testing.T) {
	baseReq, err := http.NewRequest("GET", "/accounts/username/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	baseCtx := context.WithValue(baseReq.Context(), middleware.FirebaseTokenKey, firebaseToken)

	testCases := []struct {
		name             string
		urlParamUsername string
		queryParams      map[string]string
		expectedStatus   int
	}{
		{
			name:             "Get account by username",
			urlParamUsername: "johndoe",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Get account by username not found",
			urlParamUsername: "notfound",
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "Get account by username with attributes",
			urlParamUsername: "janesmith",
			queryParams:      map[string]string{"attributes": "true"},
			expectedStatus:   http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseCtx)

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("username", testcase.urlParamUsername)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			if testCase.queryParams != nil {
				q := req.URL.Query()
				for key, value := range testCase.queryParams {
					q.Add(key, value)
				}
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			handler.GetAccountByUsername(w, req)

			if w.Code != testcase.expectedStatus {
				var errMsg string
				if w.Body.Len() > 0 {
					var response struct {
						Error struct {
							StatusCode int    `json:"statusCode"`
							Message    string `json:"message"`
						} `json:"error"`
					}
					if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
						errMsg = w.Body.String() // Fallback to raw body if JSON parsing fails
					} else {
						errMsg = response.Error.Message
					}
				} else {
					errMsg = "no response body"
				}
				t.Errorf("Expected status %d, got %d, message: %s", testcase.expectedStatus, w.Code, errMsg)
			}
		})
	}
}

func TestCreateAccount(t *testing.T) {
	baseReq, err := http.NewRequest("POST", "/accounts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	baseCtx := context.WithValue(baseReq.Context(), middleware.FirebaseTokenKey, firebaseToken)

	testCases := []struct {
		name           string
		input          CreateAccountRequest
		expectedStatus int
	}{
		{
			name: "Create account with empty id",
			input: CreateAccountRequest{
				ID:       "",
				Username: "johndoe13",
				Email:    "johndoe13@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create account with empty username",
			input: CreateAccountRequest{
				ID:       "aabbcc213",
				Username: "",
				Email:    "johndoe13@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create account with empty email",
			input: CreateAccountRequest{
				ID:       "aabbcc213",
				Username: "johndoe13",
				Email:    "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create account",
			input: CreateAccountRequest{
				ID:       "aabbcc213",
				Username: "johndoe13",
				Email:    "johndoe13@example.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Create account with invalid email",
			input: CreateAccountRequest{
				ID:       "aabbcc213",
				Username: "johndoe13",
				Email:    "johndoe13",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest("POST", "/accounts", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req = req.WithContext(baseCtx)
			req.Header.Set("Content-Type", "application/json")

			body, err := json.Marshal(testcase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			req.Body = io.NopCloser(bytes.NewBuffer(body))

			routeCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			handler.CreateAccount(w, req)

			if w.Code != testcase.expectedStatus {
				var errMsg string
				if w.Body.Len() > 0 {
					var response struct {
						Error struct {
							StatusCode int    `json:"statusCode"`
							Message    string `json:"message"`
						} `json:"error"`
					}
					if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
						errMsg = w.Body.String() // Fallback to raw body if JSON parsing fails
					} else {
						errMsg = response.Error.Message
					}
				} else {
					errMsg = "no response body"
				}
				t.Errorf("Expected status %d, got %d, message: %s", testcase.expectedStatus, w.Code, errMsg)
			}
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	baseReq, err := http.NewRequest("PUT", "/accounts/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	baseCtx := context.WithValue(baseReq.Context(), middleware.FirebaseTokenKey, firebaseToken)

	testCases := []struct {
		name           string
		urlParamId     string
		input          AccountAttributes
		expectedStatus int
	}{
		{
			name:       "Update account",
			urlParamId: "123abc",
			input: AccountAttributes{
				ID:           "123abc",
				Bio:          "Hello, I'm John Doe",
				ContactEmail: "john@example.com",
				Location:     "CN",
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
			routeCtx.URLParams.Add("id", testcase.urlParamId)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			body, err := json.Marshal(testcase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}
			req.Body = io.NopCloser(bytes.NewBuffer(body))

			w := httptest.NewRecorder()
			handler.UpdateAccount(w, req)

			if w.Code != testcase.expectedStatus {
				var errMsg string
				if w.Body.Len() > 0 {
					var response struct {
						Error struct {
							StatusCode int    `json:"statusCode"`
							Message    string `json:"message"`
						} `json:"error"`
					}
					if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
						errMsg = w.Body.String() // Fallback to raw body if JSON parsing fails
					} else {
						errMsg = response.Error.Message
					}
				} else {
					errMsg = "no response body"
				}
				t.Errorf("Expected status %d, got %d, message: %s", testcase.expectedStatus, w.Code, errMsg)
			}
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	baseReq, err := http.NewRequest("DELETE", "/accounts/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	baseCtx := context.WithValue(baseReq.Context(), middleware.FirebaseTokenKey, firebaseToken)

	testCases := []struct {
		name           string
		urlParamId     string
		expectedStatus int
	}{
		{
			name:           "Delete account",
			urlParamId:     "123abc",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Delete account missing id",
			urlParamId:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseCtx)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", testcase.urlParamId)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

			w := httptest.NewRecorder()
			handler.DeleteAccount(w, req)

			if w.Code != testcase.expectedStatus {
				var errMsg string
				if w.Body.Len() > 0 {
					var response struct {
						Error struct {
							StatusCode int    `json:"statusCode"`
							Message    string `json:"message"`
						} `json:"error"`
					}
					if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
						errMsg = w.Body.String() // Fallback to raw body if JSON parsing fails
					} else {
						errMsg = response.Error.Message
					}
				} else {
					errMsg = "no response body"
				}
				t.Errorf("Expected status %d, got %d, message: %s", testcase.expectedStatus, w.Code, errMsg)
			}
		})
	}
}
