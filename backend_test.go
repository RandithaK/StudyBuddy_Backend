package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RandithaK/StudyBuddy_Backend/pkg/models"
	"github.com/RandithaK/StudyBuddy_Backend/pkg/server"
	"github.com/RandithaK/StudyBuddy_Backend/pkg/store"
	"github.com/google/uuid"
)

func TestHealthAndGetTasks(t *testing.T) {
	ctx := context.Background()
	s, _ := store.NewStore(ctx, "")
	server.SeedStore(s)

	r := server.SetupRouter(s)

	// Health
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestEmailVerification(t *testing.T) {
	ctx := context.Background()
	s, _ := store.NewStore(ctx, "")

	// 1. Create unverified user
	token := uuid.New().String()
	user := models.User{
		ID:                "test-verify-id",
		Name:              "Verify User",
		Email:             "verify@example.com",
		Password:          "hashed",
		IsVerified:        false,
		VerificationToken: token,
	}
	s.CreateUser(user)

	// 2. Verify email
	req := httptest.NewRequest(http.MethodGet, "/verify-email?token="+token, nil)
	rr := httptest.NewRecorder()
	r := server.SetupRouter(s)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// 3. Check user status
	updatedUser, err := s.GetUser(user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if !updatedUser.IsVerified {
		t.Fatal("user should be verified")
	}
	if updatedUser.VerificationToken != "" {
		t.Fatal("verification token should be cleared")
	}
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	s, _ := store.NewStore(ctx, "")
	server.SeedStore(s)
	r := server.SetupRouter(s)

	// 1. Login with seeded test user
	loginBody := `{"query":"mutation Login($input: LoginInput!){ login(input:$input){ token user{ id name email } } }","variables":{"input":{"email":"test@example.com","password":"password"}}}`
	req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for login, got %d", rr.Code)
	}
	// Extract token
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal login response: %v", err)
	}
	token := resp["data"].(map[string]any)["login"].(map[string]any)["token"].(string)

	// 2. Change password
	changeBody := `{"query":"mutation ChangePassword($input: ChangePasswordInput!){ changePassword(input:$input){ success message } }","variables":{"input":{"currentPassword":"password","newPassword":"validpass123"}}}`
	req = httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(changeBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for change password, got %d", rr.Code)
	}
	// Confirm change success
	if !strings.Contains(rr.Body.String(), "password updated") {
		t.Fatalf("unexpected change password response: %s", rr.Body.String())
	}

	// 3. Login with new password
	loginBody2 := `{"query":"mutation Login($input: LoginInput!){ login(input:$input){ token user{ id name email } } }","variables":{"input":{"email":"test@example.com","password":"validpass123"}}}`
	req = httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(loginBody2))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for login with new password, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "token") {
		t.Fatalf("expected token in login response with new password: %s", rr.Body.String())
	}
}
