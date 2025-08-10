package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/maybemaby/oapibase/api/auth"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func bootstrapManager() *auth.JwtManager {
	return &auth.JwtManager{
		AccessTokenSecret:    []byte("very-long-access-secret"),
		AccessTokenLifetime:  time.Minute * 5,
		RefreshTokenSecret:   []byte("very-long-refresh-secret"),
		RefreshTokenLifetime: time.Hour,
	}
}

func TestRequireAccessTokenPass(t *testing.T) {
	manager := bootstrapManager()

	handler := auth.RequireAccessToken(manager)(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	validAccessToken, _ := manager.EncodeAccessToken(auth.SessionData{
		UserId: 1,
		Role:   "user",
	})

	req.Header.Set("Authorization", "Bearer "+validAccessToken)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireAccessToken401(t *testing.T) {
	manager := bootstrapManager()

	handler := auth.RequireAccessToken(manager)(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireAccessNoBearer401(t *testing.T) {
	manager := bootstrapManager()

	handler := auth.RequireAccessToken(manager)(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	validAccessToken, _ := manager.EncodeAccessToken(auth.SessionData{
		UserId: 1,
		Role:   "user",
	})

	req.Header.Set("Authorization", validAccessToken)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRefreshHandler(t *testing.T) {
	manager := bootstrapManager()

	handler := auth.RefreshTokenHandler(manager)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)

	validRefreshToken, _ := manager.EncodeRefreshToken(auth.SessionData{
		UserId: 1,
		Role:   "user",
	})

	// If the test runs too fast, the same token might get generated
	time.Sleep(1 * time.Second)

	req.Header.Set("Authorization", "Bearer "+validRefreshToken)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response auth.RefreshTokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.AccessToken == "" || response.RefreshToken == "" {
		t.Error("Expected non-empty access and refresh tokens in response")
	}

	if response.RefreshToken == validRefreshToken {
		t.Errorf("Expected new refresh token, got the same as input %s", validRefreshToken)
	}
}
