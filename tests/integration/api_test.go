package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
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

	req = applyURLParams(req, tc.urlParams)

	req.URL.RawQuery = tc.queryParams.Encode()

	return req
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
