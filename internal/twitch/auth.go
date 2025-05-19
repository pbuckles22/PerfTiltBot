package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenResponse represents the response from Twitch's OAuth token endpoint
type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	TokenType    string   `json:"token_type"`
	Scope        []string `json:"scope"`
}

// AuthManager handles Twitch OAuth token management
type AuthManager struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
	AccessToken  string
	ExpiresAt    time.Time
}

// NewAuthManager creates a new Twitch authentication manager
func NewAuthManager(clientID, clientSecret, refreshToken string) *AuthManager {
	return &AuthManager{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}
}

// RefreshToken refreshes the OAuth token using the refresh token
func (am *AuthManager) RefreshToken() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", am.RefreshToken)
	data.Set("client_id", am.ClientID)
	data.Set("client_secret", am.ClientSecret)

	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	am.AccessToken = tokenResp.AccessToken
	am.RefreshToken = tokenResp.RefreshToken
	am.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// GetAccessToken returns the current access token, refreshing if necessary
func (am *AuthManager) GetAccessToken() (string, error) {
	// If token is expired or will expire in the next 5 minutes, refresh it
	if time.Until(am.ExpiresAt) < 5*time.Minute {
		if err := am.RefreshToken(); err != nil {
			return "", fmt.Errorf("error refreshing token: %w", err)
		}
	}
	return am.AccessToken, nil
}

// IsTokenValid checks if the current token is valid
func (am *AuthManager) IsTokenValid() bool {
	return time.Until(am.ExpiresAt) > 5*time.Minute
}
