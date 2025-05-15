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
			queue := cm.GetQueue()
			users := queue.List()

			if len(users) == 0 {
				return "The queue is currently empty."
			}

			var userList []string
			for i, user := range users {
				userList = append(userList, fmt.Sprintf("%d. %s", i+1, user.Username))
			}

			return fmt.Sprintf("Current queue (%d): %s", len(users), strings.Join(userList, ", "))
		},
	})

	// Join command
	cm.RegisterCommand(Command{
		Name:        "join",
		Description: "Join the queue",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			isMod := message.User.Badges["moderator"] > 0 || message.User.Badges["broadcaster"] > 0
			err := queue.Add(message.User.Name, isMod)

			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}

			position := queue.Position(message.User.Name)
			return fmt.Sprintf("@%s has joined the queue at position %d!", message.User.Name, position)
		},
	})

	// Leave command
	cm.RegisterCommand(Command{
		Name:        "leave",
		Description: "Leave the queue",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			if queue.Remove(message.User.Name) {
				return fmt.Sprintf("@%s has left the queue!", message.User.Name)
			}
			return fmt.Sprintf("@%s, you are not in the queue!", message.User.Name)
		},
	})

	// Clear command (mod only)
	cm.RegisterCommand(Command{
		Name:        "clearqueue",
		Description: "Clear the entire queue (Mods only)",
		ModOnly:     true,
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			count := queue.Clear()
			return fmt.Sprintf("@%s cleared the queue! Removed %d user(s).", message.User.Name, count)
		},
	})

	// Position command
	cm.RegisterCommand(Command{
		Name:        "position",
		Description: "Check your position in the queue",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			position := queue.Position(message.User.Name)

			if position == -1 {
				return fmt.Sprintf("@%s, you are not in the queue!", message.User.Name)
			}

			return fmt.Sprintf("@%s, you are at position %d in the queue!", message.User.Name, position)
		},
	})
}
