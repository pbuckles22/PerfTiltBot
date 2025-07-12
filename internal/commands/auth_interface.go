package commands

import "time"

// AuthManagerInterface defines the interface for authentication management
// This allows the commands package to use auth functionality without creating import cycles
type AuthManagerInterface interface {
	GetAccessToken() (string, error)
	RefreshToken() error
	IsTokenValid() bool
	GetExpiresAt() time.Time
}
