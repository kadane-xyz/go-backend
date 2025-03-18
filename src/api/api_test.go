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
	"kadane.xyz/go-backend/v2/src/apierror"
	"kadane.xyz/go-backend/v2/src/middleware"
)

type TestingCase struct {
	name           string            // name of the test case
	queryParams    map[string]string // query params of the request ie. filtering params, pagination params, etc.
	urlParams      map[string]string // url params of the request ie. id, slug, etc.
	body           any               // body of the request ie. json body for POST requests
	expectedStatus int               // expected status code of the response
	//expectedOutput any               // expected output of the response
}

// newTestRequestWithBody creates a new HTTP request with the given method, url, and body.
func newTestRequestWithBody(t *testing.T, method, url string, body any) *http.Request {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}
	return newTestRequest(t, method, url, bytes.NewBuffer(jsonBody))
}

// newTestRequest creates a new HTTP request with the firebase token added to the context.
func newTestRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	return req.WithContext(context.WithValue(req.Context(), middleware.ClientTokenKey, clientToken))
}

// applyRouteParams adds url parameters using chi's RouteContext.
func applyURLParams(req *http.Request, params map[string]string) *http.Request {
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
	var response apierror.APIError
	err := json.NewDecoder(body).Decode(&response)
	if err != nil {
		return body.String()
	}

	return response.Data.Error.Message
}
