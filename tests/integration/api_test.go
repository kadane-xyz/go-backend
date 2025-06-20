package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"kadane.xyz/go-backend/v2/internal/middleware"
)

type TestingCase struct {
	name           string            // name of the test case
	method         string            // method of request
	url            string            // request url
	queryParams    url.Values        // query params of the request ie. filtering params, pagination params, etc.
	urlParams      map[string]string // url params of the request ie. id, slug, etc.
	body           any               // body of the request ie. json body for POST requests
	expectedStatus int               // expected status code of the response
	//expectedOutput any               // expected output of the response
}

// newTestRequestWithBody creates a new HTTP request with the given method, url, and body.
func (tc *TestingCase) buildRequest(ctx context.Context, t *testing.T) *http.Request {
	t.Helper()

	var body io.Reader

	if tc.body != nil {
		jsonBody, err := json.Marshal(tc.body)
		if err != nil {
			t.Fatalf("failed to marshall body: %v", err)
		}

		body = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, tc.method, tc.url, body)
	if err != nil {
		t.Fatalf("failed to created request: %v", err)
	}

	if tc.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req = applyURLPathValues(req, tc.urlParams)

	req.URL.RawQuery = tc.queryParams.Encode()

	return req.WithContext(context.WithValue(req.Context(), middleware.ClientTokenKey, clientToken))
}

func applyURLPathValues(req *http.Request, values map[string]string) *http.Request {
	if values == nil {
		return req
	}

	for key, value := range values {
		req.SetPathValue(key, value)
	}

	rawPath := req.URL.Path
	modified := false

	// Build the URL using the processed parameters
	for key, value := range values {
		placeholder := "{" + key + "}"
		if strings.Contains(rawPath, placeholder) {
			rawPath = strings.ReplaceAll(rawPath, placeholder, value)
			modified = true
		}
	}

	if modified {
		newURL := *req.URL
		newURL.Path = rawPath
		req.URL = &newURL
	}

	return req
}

// executeTestRequest runs the handler with the given request and checks that the status code is as expected.
func (tc *TestingCase) executeTestRequest(t *testing.T, req *http.Request) *http.Response {
	t.Helper()

	if TestingServer == nil {
		t.Fatalf("TestingServer is not initialized")
	}

	rec := httptest.NewRecorder()

	TestingServer.ServeHTTP(rec, req)

	if tc.expectedStatus != 0 && rec.Code != tc.expectedStatus {
		t.Errorf("handler returned unexpected status code: expected %d but got %d", tc.expectedStatus, rec.Code)

		return nil
	}

	return rec.Result()
}

func (tc *TestingCase) HandleTestCaseRequest(ctx context.Context, t *testing.T) *http.Response {
	t.Helper()

	return tc.executeTestRequest(t, tc.buildRequest(ctx, t))
}
