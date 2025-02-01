package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kadane.xyz/go-backend/v2/src/middleware"
)

func TestGetFriends(t *testing.T) {
	req, err := http.NewRequest("GET", "/friends", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	newctx := context.WithValue(req.Context(), middleware.FirebaseTokenKey, firebaseToken)
	req = req.WithContext(newctx)

	w := httptest.NewRecorder()
	handler.GetFriends(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	t.Logf("response: %s", w.Body.String())
}
