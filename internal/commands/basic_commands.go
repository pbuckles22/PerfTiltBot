package commands

import (
	"fmt"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

// RegisterBasicCommands adds the default commands to the command manager
func RegisterBasicCommands(cm *CommandManager) {
	// Help command
	cm.RegisterCommand(Command{
		Name:        "help",
		Description: "Shows the list of available commands",
		Handler: func(message twitch.PrivateMessage) string {
			commands := cm.GetCommandList()
			var cmdList []string
			for _, cmd := range commands {
				cmdList = append(cmdList, fmt.Sprintf("%s: %s", cmd.Name, cmd.Description))
			}
			return fmt.Sprintf("Available commands: %s", strings.Join(cmdList, " | "))
		},
	})

	// Ping command
	cm.RegisterCommand(Command{
		Name:        "ping",
		Description: "Check if the bot is running",
		Handler: func(message twitch.PrivateMessage) string {
			return "Pong! 🏓"
		},
	})

	// Queue command
	cm.RegisterCommand(Command{
		Name:        "queue",
		Description: "Shows the current queue status",
		Handler: func(message twitch.PrivateMessage) string {
			return "Queue system coming soon!"
		},
	})

	// Join command
	cm.RegisterCommand(Command{
		Name:        "join",
		Description: "Join the queue",
		Handler: func(message twitch.PrivateMessage) string {
			return fmt.Sprintf("@%s, queue joining will be implemented soon!", message.User.Name)
		},
	})

	// Leave command
	cm.RegisterCommand(Command{
		Name:        "leave",
		Description: "Leave the queue",
		Handler: func(message twitch.PrivateMessage) string {
			return fmt.Sprintf("@%s, queue leaving will be implemented soon!", message.User.Name)
		},
	})
}
