package twitch

import (
	"time"

	"github.com/pbuckles22/PBChatBot/internal/utils"
)

// Constants for token refresh
const (
	checkInterval  = 1 * time.Minute // Check every minute
	minRefreshTime = 5 * time.Minute // Refresh when 5 minutes left
)

// calculateCheckInterval determines when to check next based on when we need to refresh
// We want to refresh when there are 5 minutes left, so we start checking when there are 10 minutes left
func calculateCheckInterval(timeUntilExpiry time.Duration) time.Duration {
	// If less than minimum time, refresh token immediately
	if timeUntilExpiry <= minRefreshTime {
		return 1 * time.Second // Return 1 second instead of 0 to avoid ticker panic
	}

	// Ensure we don't return negative or zero intervals
	if timeUntilExpiry <= 0 {
		return 1 * time.Second
	}

	// Calculate when we need to start checking (10 minutes before expiry)
	// This gives us time to check and refresh before the 5-minute mark
	timeUntilFirstCheck := timeUntilExpiry - 10*time.Minute

	// If we're already within 10 minutes of expiry, check every minute
	if timeUntilFirstCheck <= 0 {
		return 1 * time.Minute
	}

	// Otherwise, wait until we're 10 minutes away from expiry
	return timeUntilFirstCheck
}

// formatTimeForLogs formats time for debug logs in PST
func formatTimeForLogs(t time.Time) string {
	return utils.FormatTimeForLogs(t)
}
