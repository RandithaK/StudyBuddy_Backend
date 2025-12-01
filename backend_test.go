package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthAndGetTasks(t *testing.T) {
	ctx := context.Background()
	store, _ := NewStore(ctx, "")
	seedStore(store)
	cfg := ServerConfig{Addr: "8081", JWTSecret: "test-secret"}

	r := setupRouter(store, cfg)

	// Health
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Get tasks
	req = httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for tasks, got %d", rr.Code)
	}
}
