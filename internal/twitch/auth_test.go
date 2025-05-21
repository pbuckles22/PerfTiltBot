package twitch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenRefresh(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type: application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		// Create mock response
		response := TokenResponse{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresIn:    3600, // 1 hour
			TokenType:    "bearer",
			Scope:        []string{"chat:read", "chat:edit"},
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a new auth manager with test credentials
	am := NewAuthManager(
		"test_client_id",
		"test_client_secret",
		"test_refresh_token",
	)

	// Override the token endpoint URL for testing
	originalTokenURL := "https://id.twitch.tv/oauth2/token"
	tokenURL = server.URL
	defer func() { tokenURL = originalTokenURL }()

	// Test initial state
	if am.AccessToken != "" {
		t.Error("Expected empty access token initially")
	}
	if !am.ExpiresAt.IsZero() {
		t.Error("Expected zero expiration time initially")
	}

	// Test token refresh
	err := am.RefreshToken()
	if err != nil {
		t.Errorf("Failed to refresh token: %v", err)
	}

	// Verify token was set
	if am.AccessToken != "mock_access_token" {
		t.Errorf("Expected access token 'mock_access_token', got '%s'", am.AccessToken)
	}

	// Verify refresh token was updated
	if am.RefreshTokenValue != "mock_refresh_token" {
		t.Errorf("Expected refresh token 'mock_refresh_token', got '%s'", am.RefreshTokenValue)
	}

	// Verify expiration time was set (should be roughly 1 hour from now)
	expectedExpiry := time.Now().Add(time.Hour)
	if am.ExpiresAt.Sub(expectedExpiry) > time.Minute {
		t.Errorf("Expected expiration time to be roughly 1 hour from now, got %v", am.ExpiresAt)
	}

	// Test token validity check
	if !am.IsTokenValid() {
		t.Error("Token should be valid after refresh")
	}

	// Test token expiration
	am.ExpiresAt = time.Now().Add(-1 * time.Hour) // Set expiration to 1 hour ago
	if am.IsTokenValid() {
		t.Error("Token should be invalid after expiration")
	}

	// Test token near expiration
	am.ExpiresAt = time.Now().Add(4 * time.Minute) // Set expiration to 4 minutes from now
	if am.IsTokenValid() {
		t.Error("Token should be considered invalid when within 5 minutes of expiration")
	}
}
