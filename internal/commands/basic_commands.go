package commands

import (
	"fmt"
	"strconv"
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
			queue := cm.GetQueue()
			commands := cm.GetCommandList()
			var baseList []string
			var queueList []string

			// Define base commands that are always shown
			baseCommands := map[string]bool{
				"help":       true,
				"ping":       true,
				"startqueue": true,
				"endqueue":   true,
			}

			// Define queue-specific commands
			queueCommands := map[string]bool{
				"queue":      true,
				"join":       true,
				"leave":      true,
				"clearqueue": true,
				"position":   true,
			}

			for _, cmd := range commands {
				desc := cmd.Description
				if len(cmd.Aliases) > 0 {
					desc = fmt.Sprintf("%s (aliases: %s)", desc, strings.Join(cmd.Aliases, ", "))
				}
				cmdInfo := fmt.Sprintf("%s: %s", cmd.Name, desc)

				if baseCommands[cmd.Name] {
					baseList = append(baseList, cmdInfo)
				} else if queueCommands[cmd.Name] {
					queueList = append(queueList, cmdInfo)
				}
			}

			result := fmt.Sprintf("Base commands: %s", strings.Join(baseList, " | "))
			if queue.IsEnabled() {
				result += fmt.Sprintf("\nQueue commands: %s", strings.Join(queueList, " | "))
			}
			return result
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

	// Start Queue command (mod only)
	cm.RegisterCommand(Command{
		Name:        "startqueue",
		Aliases:     []string{"sq"},
		Description: "Start the queue system (Mods only)",
		ModOnly:     true,
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			if queue.IsEnabled() {
				return "Queue system is already running!"
			}
			queue.Enable()
			return fmt.Sprintf("@%s has started the queue system!", message.User.Name)
		},
	})

	// End Queue command (mod only)
	cm.RegisterCommand(Command{
		Name:        "endqueue",
		Aliases:     []string{"eq"},
		Description: "End the queue system and clear the queue (Mods only)",
		ModOnly:     true,
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			if !queue.IsEnabled() {
				return "Queue system is already disabled!"
			}
			queue.Disable()
			return fmt.Sprintf("@%s has ended the queue system!", message.User.Name)
		},
	})

	// Queue command
	cm.RegisterCommand(Command{
		Name:        "queue",
		Aliases:     []string{"q"},
		Description: "Shows the current queue status",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

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
		Description: "Join the queue. Mods/VIPs can add others with !join [username] or !join [username] [position]",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			parts := strings.Fields(message.Message)

			// Handle regular users
			if !isPrivileged(message) {
				err := queue.Add(message.User.Name, false)
				if err != nil {
					return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
				}
				position := queue.Position(message.User.Name)
				return fmt.Sprintf("@%s has joined the queue at position %d!", message.User.Name, position)
			}

			// Handle privileged users
			if len(parts) < 2 {
				err := queue.Add(message.User.Name, true)
				if err != nil {
					return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
				}
				position := queue.Position(message.User.Name)
				return fmt.Sprintf("@%s has joined the queue at position %d!", message.User.Name, position)
			}

			// Get target user
			targetUser := strings.TrimPrefix(parts[1], "@")
			isMod := message.User.Badges["moderator"] > 0 || message.User.Badges["broadcaster"] > 0

			// Check for position parameter
			if len(parts) > 2 {
				position, err := strconv.Atoi(parts[2])
				if err != nil {
					return fmt.Sprintf("@%s, invalid position number provided.", message.User.Name)
				}

				err = queue.AddAtPosition(targetUser, position, isMod)
				if err != nil {
					return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
				}
				return fmt.Sprintf("@%s added %s to the queue at position %d!", message.User.Name, targetUser, position)
			}

			// No position specified, add to end
			err := queue.Add(targetUser, isMod)
			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}
			position := queue.Position(targetUser)
			return fmt.Sprintf("@%s added %s to the queue at position %d!", message.User.Name, targetUser, position)
		},
	})

	// Leave command
	cm.RegisterCommand(Command{
		Name:        "leave",
		Description: "Leave the queue. Mods/VIPs can remove others with !leave [username]",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			parts := strings.Fields(message.Message)

			var targetUser string
			if len(parts) > 1 && isPrivileged(message) {
				// Mod/VIP is removing someone else
				targetUser = strings.TrimPrefix(parts[1], "@")
			} else {
				// Regular user leaving themselves
				targetUser = message.User.Name
			}

			if queue.Remove(targetUser) {
				if targetUser != message.User.Name {
					return fmt.Sprintf("@%s has removed %s from the queue!", message.User.Name, targetUser)
				}
				return fmt.Sprintf("@%s has left the queue!", targetUser)
			}
			return fmt.Sprintf("@%s, %s is not in the queue!", message.User.Name, targetUser)
		},
	})

	// Clear command (mod only)
	cm.RegisterCommand(Command{
		Name:        "clearqueue",
		Aliases:     []string{"cq"},
		Description: "Clear the entire queue (Mods only)",
		ModOnly:     true,
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			count := queue.Clear()
			return fmt.Sprintf("@%s cleared the queue! Removed %d user(s).", message.User.Name, count)
		},
	})

	// Position command
	cm.RegisterCommand(Command{
		Name:        "position",
		Aliases:     []string{"pos"},
		Description: "Check your position in the queue",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			position := queue.Position(message.User.Name)
			if position == -1 {
				return fmt.Sprintf("@%s, you are not in the queue!", message.User.Name)
			}
			return fmt.Sprintf("@%s, you are at position %d in the queue!", message.User.Name, position)
		},
	})
}
