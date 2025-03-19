package api

import (
	"net/http"
	"testing"
)

func TestGetAccount(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get account missing id",
			urlParams:      map[string]string{"id": ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get account",
			urlParams:      map[string]string{"id": "123abc"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get account with attributes",
			urlParams:      map[string]string{"id": "123abc"},
			queryParams:    map[string]string{"attributes": "true"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/accounts", nil)
			request = applyURLParams(request, testCase.urlParams)
			request = applyQueryParams(request, testCase.queryParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetAccount)
		})
	}
}

func TestGetAccounts(t *testing.T) {
	testCases := []TestingCase{
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

			request := newTestRequest(t, http.MethodGet, "/accounts", nil)
			request = applyQueryParams(request, testCase.queryParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetAccounts)
		})
	}
}

func TestGetAccountByUsername(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Get account by username",
			urlParams:      map[string]string{"username": "johndoe"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get account by username not found",
			urlParams:      map[string]string{"username": "notfound"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get account by username with attributes",
			urlParams:      map[string]string{"username": "janesmith"},
			queryParams:    map[string]string{"attributes": "true"},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodGet, "/accounts/username", nil)
			request = applyURLParams(request, testCase.urlParams)
			request = applyQueryParams(request, testCase.queryParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.GetAccountByUsername)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	testCases := []TestingCase{
		{
			name: "Create account with empty id",
			body: map[string]string{
				"id":       "",
				"username": "johndoe13",
				"email":    "johndoe13@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create account with empty username",
			body: map[string]string{
				"id":       "aabbcc213",
				"username": "",
				"email":    "johndoe13@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create account with empty email",
			body: map[string]string{
				"id":       "aabbcc213",
				"username": "johndoe13",
				"email":    "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create account",
			body: map[string]string{
				"id":       "aabbcc213",
				"username": "johndoe13",
				"email":    "johndoe13@example.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Create account with invalid email",
			body: map[string]string{
				"id":       "aabbcc213",
				"username": "johndoe13",
				"email":    "johndoe13",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPost, "/accounts", testCase.body)

			executeTestRequest(t, request, testCase.expectedStatus, handler.CreateAccount)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	testCases := []TestingCase{
		{
			name:      "Update account",
			urlParams: map[string]string{"id": "123abc"},
			body: map[string]string{
				"bio":          "Hello, I'm John Doe",
				"contactEmail": "john@example.com",
				"location":     "CN",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Update account",
			urlParams: map[string]string{"id": "123abc"},
			body: map[string]string{
				"bio":          "Hello, I'm John Doe",
				"contactEmail": "john@example.com",
				"location":     "CN",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequestWithBody(t, http.MethodPut, "/accounts", testCase.body)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.UpdateAccount)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	testCases := []TestingCase{
		{
			name:           "Delete account",
			urlParams:      map[string]string{"id": "789ghi"},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Delete account missing id",
			urlParams:      map[string]string{"id": ""},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := newTestRequest(t, http.MethodDelete, "/accounts", nil)
			request = applyURLParams(request, testCase.urlParams)

			executeTestRequest(t, request, testCase.expectedStatus, handler.DeleteAccount)
		})
	}
}
