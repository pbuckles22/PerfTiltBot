package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

// RegisterBasicCommands adds the default commands to the command manager.
// This includes core functionality like help, ping, and queue management.
func RegisterBasicCommands(cm *CommandManager) {
	// Help command - Shows available commands based on queue state
	// When queue is disabled: Shows only basic commands
	// When queue is enabled: Shows both basic and queue commands
	cm.RegisterCommand(Command{
		Name:        "help",
		Description: "Shows the list of available commands",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			commands := cm.GetCommandList()
			var commandList []string

			// Define base commands that are always shown regardless of queue state
			baseCommands := map[string]bool{
				"help":       true,
				"ping":       true,
				"startqueue": true,
			}

			// Define queue commands that are only shown when queue is enabled
			queueCommands := map[string]bool{
				"help":       true,
				"queue":      true,
				"join":       true,
				"leave":      true,
				"position":   true,
				"endqueue":   true,
				"clearqueue": true,
			}

			// Build the list of commands to display
			for _, cmd := range commands {
				// Add aliases to the description if any exist
				desc := cmd.Description
				if len(cmd.Aliases) > 0 {
					desc = fmt.Sprintf("%s (aliases: %s)", desc, strings.Join(cmd.Aliases, ", !"))
				}
				cmdInfo := fmt.Sprintf("%s: %s", cmd.Name, desc)

				// Only show base commands when queue is disabled
				if baseCommands[cmd.Name] && !queue.IsEnabled() {
					commandList = append(commandList, cmdInfo)
					// Show queue commands only when queue is enabled
				} else if queueCommands[cmd.Name] && queue.IsEnabled() {
					commandList = append(commandList, cmdInfo)
				}
			}

			if len(commandList) == 0 {
				return "No commands available."
			}

			return fmt.Sprintf("Available commands: %s", strings.Join(commandList, " | "))
		},
	})

	// Ping command - Simple bot health check
	cm.RegisterCommand(Command{
		Name:        "ping",
		Description: "Check if the bot is running",
		Handler: func(message twitch.PrivateMessage) string {
			return "Pong! 🏓"
		},
	})

	// Start Queue command - Enables the queue system (mod only)
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

	// End Queue command - Disables and clears the queue (mod only)
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

	// Queue command - Shows current queue status
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

			// Build numbered list of users in queue
			var userList []string
			for i, user := range users {
				userList = append(userList, fmt.Sprintf("%d. %s", i+1, user.Username))
			}

			return fmt.Sprintf("Current queue (%d): %s", len(users), strings.Join(userList, ", "))
		},
	})

	// Join command - Adds a user to the queue
	// Regular users can only add themselves
	// Mods/VIPs can add others and specify positions
	cm.RegisterCommand(Command{
		Name:        "join",
		Description: "Join the queue. Mods/VIPs can add others with !join [username] or !join [username] [position]",
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			parts := strings.Fields(message.Message)

			// Handle regular users (can only add themselves)
			if !isPrivileged(message) {
				err := queue.Add(message.User.Name, false)
				if err != nil {
					return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
				}
				position := queue.Position(message.User.Name)
				return fmt.Sprintf("@%s has joined the queue at position %d!", message.User.Name, position)
			}

			// Handle privileged users (mods/VIPs)
			if len(parts) < 2 {
				// No target specified, add themselves
				err := queue.Add(message.User.Name, true)
				if err != nil {
					return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
				}
				position := queue.Position(message.User.Name)
				return fmt.Sprintf("@%s has joined the queue at position %d!", message.User.Name, position)
			}

			// Get target user to add
			targetUser := strings.TrimPrefix(parts[1], "@")
			isMod := message.User.Badges["moderator"] > 0 || message.User.Badges["broadcaster"] > 0

			// Check if position was specified
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

	// Leave command - Removes a user from the queue
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

	// Clear command - Removes all users from the queue (mod only)
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

	// Position command - Shows a user's position in queue
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

	// Kill command - Safely shuts down the bot (mod only)
	cm.RegisterCommand(Command{
		Name:        "kill",
		Description: "Safely shut down the bot (Mods/VIPs only)",
		ModOnly:     false,
		Handler: func(message twitch.PrivateMessage) string {
			// Check if user is privileged (Mod, VIP, or Broadcaster)
			if !isPrivileged(message) {
				return "This command can only be used by moderators and VIPs."
			}

			// Signal that we want to shut down
			cm.RequestShutdown()
			return fmt.Sprintf("@%s has initiated bot shutdown. Goodbye! 👋", message.User.Name)
		},
	})

	// Pop command - Removes the first user from the queue (mod only)
	cm.RegisterCommand(Command{
		Name:        "pop",
		Description: "Remove the first user from the queue (Mods/VIPs only)",
		ModOnly:     false,
		Handler: func(message twitch.PrivateMessage) string {
			// Check if user is privileged (Mod, VIP, or Broadcaster)
			if !isPrivileged(message) {
				return "This command can only be used by moderators and VIPs."
			}

			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			user, err := queue.Pop()
			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}

			return fmt.Sprintf("@%s has been removed from the front of the queue!", user.Username)
		},
	})

	// Remove command - Removes a specified user from the queue (Mods/VIPs only)
	cm.RegisterCommand(Command{
		Name:        "remove",
		Aliases:     []string{"rm"},
		Description: "Remove a specified user from the queue (Mods/VIPs only)",
		ModOnly:     false,
		Handler: func(message twitch.PrivateMessage) string {
			// Check if user is privileged (Mod, VIP, or Broadcaster)
			if !isPrivileged(message) {
				return "This command can only be used by moderators and VIPs."
			}

			args := strings.Fields(message.Message)
			if len(args) < 2 {
				return "Usage: !remove <username>"
			}

			username := args[1]
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			removed, err := queue.RemoveUser(username)
			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}

			if removed {
				return fmt.Sprintf("@%s has been removed from the queue!", username)
			} else {
				return fmt.Sprintf("@%s is not in the queue.", username)
			}
		},
	})
}
