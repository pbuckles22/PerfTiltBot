package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pbuckles22/PBChatBot/internal/utils"
	"gopkg.in/yaml.v3"
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
	ClientID          string
	ClientSecret      string
	RefreshTokenValue string
	AccessToken       string
	ExpiresAt         time.Time
	SecretsPath       string
	lastRefreshTime   time.Time
	etLocation        *time.Location
}

// tokenURL is the endpoint for token operations
var tokenURL = "https://id.twitch.tv/oauth2/token"

// NewAuthManager creates a new Twitch authentication manager
func NewAuthManager(clientID, clientSecret, refreshToken, secretsPath string) *AuthManager {
	loc := utils.GetLogLocation()

	return &AuthManager{
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		RefreshTokenValue: refreshToken,
		SecretsPath:       secretsPath,
		lastRefreshTime:   time.Now().In(loc),
		etLocation:        loc,
	}
}

// RefreshToken refreshes the OAuth token using the refresh token
func (am *AuthManager) RefreshToken() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", am.RefreshTokenValue)
	data.Set("client_id", am.ClientID)
	data.Set("client_secret", am.ClientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
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
	am.RefreshTokenValue = tokenResp.RefreshToken
	am.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).In(am.etLocation)

	// Persist the new refresh token to the secrets file
	if err := am.persistRefreshToken(); err != nil {
		return fmt.Errorf("error persisting refresh token: %w", err)
	}

	am.lastRefreshTime = time.Now().In(am.etLocation)

	return nil
}

// persistRefreshToken saves the new refresh token to the secrets file
func (am *AuthManager) persistRefreshToken() error {
	// Read the current secrets file
	data, err := os.ReadFile(am.SecretsPath)
	if err != nil {
		return fmt.Errorf("error reading secrets file: %w", err)
	}

	// Parse the YAML
	var secrets map[string]interface{}
	if err := yaml.Unmarshal(data, &secrets); err != nil {
		return fmt.Errorf("error parsing secrets file: %w", err)
	}

	// Update the refresh token
	if twitch, ok := secrets["twitch"].(map[string]interface{}); ok {
		twitch["refresh_token"] = am.RefreshTokenValue
		secrets["twitch"] = twitch
	}

	// Write back to file
	newData, err := yaml.Marshal(secrets)
	if err != nil {
		return fmt.Errorf("error marshaling secrets: %w", err)
	}

	if err := os.WriteFile(am.SecretsPath, newData, 0644); err != nil {
		return fmt.Errorf("error writing secrets file: %w", err)
	}

	return nil
}

// GetAccessToken returns the current access token, refreshing if necessary
func (am *AuthManager) GetAccessToken() (string, error) {
	if !am.IsTokenValid() {
		log.Printf("[Auth] Refreshing token...")
		if err := am.RefreshToken(); err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		am.lastRefreshTime = time.Now().In(am.etLocation)
		log.Printf("[Auth] Token refreshed successfully")
		log.Printf("") // Blank line after auth refresh
	}
	return am.AccessToken, nil
}

// IsTokenValid checks if the current token is valid
func (am *AuthManager) IsTokenValid() bool {
	timeUntilExpiry := time.Until(am.ExpiresAt)
	return timeUntilExpiry > 1*time.Minute
}

// GetLastRefreshTime returns when the token was last refreshed
func (am *AuthManager) GetLastRefreshTime() time.Time {
	return am.lastRefreshTime
}

// GetExpiresAt returns the time when the current token expires
func (am *AuthManager) GetExpiresAt() time.Time {
	return am.ExpiresAt
}
