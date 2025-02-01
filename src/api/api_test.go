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

// newTestRequest creates a new HTTP request with the firebase token added to the context.
func newTestRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	return req.WithContext(context.WithValue(req.Context(), middleware.FirebaseTokenKey, firebaseToken))
}

// applyRouteParams adds url parameters using chi's RouteContext.
func applyRouteParams(req *http.Request, params map[string]string) *http.Request {
	if params == nil {
		return req
	}
	routeCtx := chi.NewRouteContext()
	for key, value := range params {
		routeCtx.URLParams.Add(key, value)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

// applyQueryParams adds query parameters to the request URL.
func applyQueryParams(req *http.Request, queryParams map[string]string) *http.Request {
	if queryParams == nil {
		return req
	}
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	return req
}

// executeTestRequest runs the handler with the given request and checks that the status code is as expected.
func executeTestRequest(t *testing.T, req *http.Request, expectedStatus int, handlerFunc http.HandlerFunc) {
	w := httptest.NewRecorder()
	handlerFunc(w, req)
	if w.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d, message: %s", expectedStatus, w.Code, extractErrorMessage(w.Body))
	}
}

// extractErrorMessage tries to decode a JSON error or falls back to the raw body.
func extractErrorMessage(body *bytes.Buffer) string {
	if body.Len() > 0 {
		var response struct {
			Error struct {
				StatusCode int    `json:"statusCode"`
				Message    string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(body).Decode(&response); err != nil {
			return body.String()
		}
		return response.Error.Message
	}
	return "no response body"
}
