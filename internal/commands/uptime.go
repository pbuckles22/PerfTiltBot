package commands

import (
	"fmt"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

// RegisterUptimeCommand registers the uptime command
func RegisterUptimeCommand(cm *CommandManager) {
	cm.RegisterCommand(&Command{
		Name:        "uptime",
		Aliases:     []string{"up"},
		Description: "Shows how long the bot has been running",
		Handler: func(message twitch.PrivateMessage, args []string) string {
			uptime := time.Since(cm.GetBotStartTime())
			hours := int(uptime.Hours())
			minutes := int(uptime.Minutes()) % 60
			seconds := int(uptime.Seconds()) % 60

			if hours > 0 {
				return fmt.Sprintf("Bot has been running for %d hours, %d minutes, and %d seconds", hours, minutes, seconds)
			} else if minutes > 0 {
				return fmt.Sprintf("Bot has been running for %d minutes and %d seconds", minutes, seconds)
			} else {
				return fmt.Sprintf("Bot has been running for %d seconds", seconds)
			}
		},
	})
}
