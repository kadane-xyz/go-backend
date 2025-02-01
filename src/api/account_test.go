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
	req, err := http.NewRequest("GET", "/accounts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	newctx := context.WithValue(req.Context(), middleware.FirebaseTokenKey, firebaseToken)
	req = req.WithContext(newctx)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", firebaseToken.UserID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	w := httptest.NewRecorder()
	handler.GetAccount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	t.Logf("response: %s", w.Body.String())
}

func TestGetAccountByUsername(t *testing.T) {
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
