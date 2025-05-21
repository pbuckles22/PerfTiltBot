package commands

import (
	"fmt"
	"log"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PerfTiltBot/internal/config"
	twitchauth "github.com/pbuckles22/PerfTiltBot/internal/twitch"
)

// formatTimeET formats a time in the channel's configured timezone
func formatTimeET(t time.Time, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Printf("Error loading timezone %s: %v, falling back to America/New_York", timezone, err)
		loc, _ = time.LoadLocation("America/New_York")
	}
	return t.In(loc).Format("2006-01-02 15:04:05 ET")
}

// RegisterAuthCommand registers the auth command
func RegisterAuthCommand(cm *CommandManager, authManager *twitchauth.AuthManager) {
	cm.RegisterCommand(&Command{
		Name:        "auth",
		Description: "Shows token authentication information",
		ModOnly:     true, // Only moderators can use this command
		Handler: func(message twitchirc.PrivateMessage, args []string) string {
			// Get the channel's config
			cfg, err := config.Load("configs/secrets.yaml")
			if err != nil {
				log.Printf("Error loading config: %v", err)
				return "Error loading configuration"
			}

			// Check if the message is a whisper
			isWhisper := message.Channel == "jtv"

			// Format the response
			response := fmt.Sprintf(
				"Last refresh: %s | Expires: %s | Next check: %s",
				formatTimeET(authManager.GetLastRefreshTime(), cfg.Twitch.Timezone),
				formatTimeET(authManager.GetExpiresAt(), cfg.Twitch.Timezone),
				formatTimeET(time.Now().Add(5*time.Minute), cfg.Twitch.Timezone),
			)

			// If it's a whisper, prefix with /w
			if isWhisper {
				return fmt.Sprintf("/w %s %s", message.User.DisplayName, response)
			}

			return response
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
