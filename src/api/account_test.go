package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetAccount(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/accounts/", nil)

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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"id": testCase.urlParamId})
			req = applyQueryParams(req, testCase.queryParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetAccount)
		})
	}
}

func TestGetAccounts(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/accounts", nil)

	testCases := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
		{
			name:           "Get accounts by username",
			queryParams:    map[string]string{"usernames": "[\"johndoe\", \"janedoe\"]"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts by locations",
			queryParams:    map[string]string{"locations": "[\"New York\", \"London\"]"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts sort by",
			queryParams:    map[string]string{"sortBy": "createdAt"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts order by",
			queryParams:    map[string]string{"orderBy": "desc"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts by username and locations",
			queryParams:    map[string]string{"usernames": "[\"johndoe\", \"janedoe\"]", "locations": "[\"New York\", \"London\"]"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts by username locations sort and order",
			queryParams:    map[string]string{"usernames": "[\"johndoe\", \"janedoe\"]", "locations": "[\"New York\", \"London\"]", "sortBy": "createdAt", "orderBy": "desc"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get accounts by username sort and order",
			queryParams:    map[string]string{"usernames": "[\"johndoe\", \"janedoe\"]", "sortBy": "level", "orderBy": "asc"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyQueryParams(req, testCase.queryParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetAccounts)
		})
	}
}

func TestGetAccountByUsername(t *testing.T) {
	baseReq := newTestRequest(t, http.MethodGet, "/accounts/username/", nil)

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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := baseReq.Clone(baseReq.Context())
			req = applyRouteParams(req, map[string]string{"username": testCase.urlParamUsername})
			req = applyQueryParams(req, testCase.queryParams)

			executeTestRequest(t, req, testCase.expectedStatus, handler.GetAccountByUsername)
		})
	}
}

func TestCreateAccount(t *testing.T) {
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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := newTestRequest(t, http.MethodPost, "/accounts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req = applyRouteParams(req, nil)

			executeTestRequest(t, req, testCase.expectedStatus, handler.CreateAccount)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(testCase.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			req := newTestRequest(t, http.MethodPut, "/accounts", bytes.NewBuffer(body))
			req = applyRouteParams(req, map[string]string{"id": testCase.urlParamId})

			executeTestRequest(t, req, testCase.expectedStatus, handler.UpdateAccount)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	testCases := []struct {
		name           string
		urlParamId     string
		expectedStatus int
	}{
		{
			name:           "Delete account",
			urlParamId:     "789ghi",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Delete account missing id",
			urlParamId:     "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := newTestRequest(t, http.MethodDelete, "/accounts", nil)
			req = applyRouteParams(req, map[string]string{"id": testCase.urlParamId})

			executeTestRequest(t, req, testCase.expectedStatus, handler.DeleteAccount)
		})
	}
}
