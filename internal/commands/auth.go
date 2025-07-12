package commands

import (
	"fmt"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PBChatBot/internal/utils"
)

// formatTimeET formats a time in the channel's configured timezone
func formatTimeET(t time.Time, timezone string) string {
	return utils.FormatTimeForDisplay(t, timezone)
}

// RegisterAuthCommand registers the auth command
func RegisterAuthCommand(cm *CommandManager, authManager AuthManagerInterface) {
	cm.RegisterCommand(&Command{
		Name:        "auth",
		Description: "Refreshes the bot's authentication token",
		ModOnly:     true, // Only moderators can use this command
		Handler: func(message twitchirc.PrivateMessage, args []string) string {
			// Only allow channel owner to use this command
			if message.User.Name != message.Channel {
				return "This command can only be used by the channel owner."
			}

			// Refresh the token
			if err := authManager.RefreshToken(); err != nil {
				return fmt.Sprintf("Error refreshing token: %v", err)
			}

			return "Token refreshed successfully!"
		},
	})
}

// calculateNextCheckTime determines when the next token validity check will occur
func calculateNextCheckTime(expiresAt time.Time) time.Time {
	timeUntilExpiry := time.Until(expiresAt)

	// Use the same intervals as in the bot's refreshTokenLoop
	switch {
	case timeUntilExpiry <= 5*time.Minute:
		return time.Now().Add(0) // Will trigger immediate refresh
	case timeUntilExpiry <= 10*time.Minute:
		return time.Now().Add(3 * time.Minute)
	case timeUntilExpiry <= 20*time.Minute:
		return time.Now().Add(5 * time.Minute)
	case timeUntilExpiry <= 30*time.Minute:
		return time.Now().Add(7 * time.Minute)
	case timeUntilExpiry <= time.Hour:
		return time.Now().Add(10 * time.Minute)
	default:
		return time.Now().Add(30 * time.Minute)
	}
}
