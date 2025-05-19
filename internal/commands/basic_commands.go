package commands

import (
	"fmt"
	"os"
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
		Aliases:     []string{"h"},
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
		Aliases:     []string{"j"},
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

			isMod := message.User.Badges["moderator"] > 0 || message.User.Badges["broadcaster"] > 0

			// Check if the last argument is a number (position)
			lastArg := parts[len(parts)-1]
			position, err := strconv.Atoi(lastArg)
			hasPosition := err == nil

			// Get all usernames (excluding position if present)
			usernames := parts[1:]
			if hasPosition {
				usernames = usernames[:len(usernames)-1]
			}

			// Process each username
			var addedUsers []string
			var alreadyInQueue []string
			for _, username := range usernames {
				username = strings.TrimPrefix(username, "@")
				if hasPosition {
					// If position is beyond queue length, add to end
					if position > len(queue.List()) {
						err = queue.Add(username, isMod)
						if err != nil {
							if err.Error() == "user is already in queue" {
								alreadyInQueue = append(alreadyInQueue, username)
								continue
							}
							return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
						}
						addedUsers = append(addedUsers, username)
					} else {
						err = queue.AddAtPosition(username, position, isMod)
						if err != nil {
							if err.Error() == "user is already in queue" {
								alreadyInQueue = append(alreadyInQueue, username)
								continue
							}
							return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
						}
						addedUsers = append(addedUsers, username)
						position++ // Increment position for next user
					}
				} else {
					err = queue.Add(username, isMod)
					if err != nil {
						if err.Error() == "user is already in queue" {
							alreadyInQueue = append(alreadyInQueue, username)
							continue
						}
						return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
					}
					addedUsers = append(addedUsers, username)
				}
			}

			// Build response message
			var responseParts []string
			if len(addedUsers) > 0 {
				if len(addedUsers) == 1 {
					pos := queue.Position(addedUsers[0])
					if hasPosition && position > len(queue.List()) {
						responseParts = append(responseParts, fmt.Sprintf("%s has been added to the end of the queue", addedUsers[0]))
					} else {
						responseParts = append(responseParts, fmt.Sprintf("added %s to the queue at position %d", addedUsers[0], pos))
					}
				} else {
					if hasPosition {
						// If we specified a position, show each user's position
						positions := make([]string, len(addedUsers))
						for i, user := range addedUsers {
							pos := queue.Position(user)
							if pos == len(queue.List()) {
								positions[i] = fmt.Sprintf("%s (end of queue)", user)
							} else {
								positions[i] = fmt.Sprintf("%s (position %d)", user, pos)
							}
						}
						responseParts = append(responseParts, fmt.Sprintf("added %s to the queue", strings.Join(positions, ", ")))
					} else {
						responseParts = append(responseParts, fmt.Sprintf("added %s to the queue", strings.Join(addedUsers, ", ")))
					}
				}
			}
			if len(alreadyInQueue) > 0 {
				if len(alreadyInQueue) == 1 {
					responseParts = append(responseParts, fmt.Sprintf("%s is already in the queue", alreadyInQueue[0]))
				} else {
					responseParts = append(responseParts, fmt.Sprintf("%s are already in the queue", strings.Join(alreadyInQueue, ", ")))
				}
			}

			if len(responseParts) == 0 {
				return fmt.Sprintf("@%s, no users were added to the queue.", message.User.Name)
			}

			return fmt.Sprintf("@%s %s!", message.User.Name, strings.Join(responseParts, " and "))
		},
	})

	// Leave command - Removes a user from the queue
	cm.RegisterCommand(Command{
		Name:        "leave",
		Aliases:     []string{"l"},
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

	// Pop command - Removes users from the front of the queue (mod only)
	cm.RegisterCommand(Command{
		Name:        "pop",
		Description: "Remove users from the front of the queue (Mods/VIPs only). Usage: !pop [count]",
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

			// Parse count from message if provided
			count := 1 // Default to 1 if no count specified
			parts := strings.Fields(message.Message)
			if len(parts) > 1 {
				var err error
				count, err = strconv.Atoi(parts[1])
				if err != nil || count < 1 {
					return "Invalid count provided. Please specify a positive number."
				}
			}

			// Pop N users from the queue
			users, err := queue.PopN(count)
			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}

			// Build response message
			if len(users) == 1 {
				return fmt.Sprintf("@%s has been removed from the front of the queue!", users[0].Username)
			} else {
				usernames := make([]string, len(users))
				for i, user := range users {
					usernames[i] = user.Username
				}
				return fmt.Sprintf("Removed %d users from the front of the queue: %s", len(users), strings.Join(usernames, ", "))
			}
		},
	})

	// Remove command - Removes a specified user from the queue (Mods/VIPs only)
	cm.RegisterCommand(Command{
		Name:        "remove",
		Aliases:     []string{"rm"},
		Description: "Remove specified users or positions from the queue (Mods/VIPs only). Usage: !remove [username1] [username2] ... or !remove [position1] [position2] ...",
		ModOnly:     false,
		Handler: func(message twitch.PrivateMessage) string {
			// Check if user is privileged (Mod, VIP, or Broadcaster)
			if !isPrivileged(message) {
				return "This command can only be used by moderators and VIPs."
			}

			args := strings.Fields(message.Message)
			if len(args) < 2 {
				return "Usage: !remove <username1> [username2] ... or !remove <position1> [position2] ..."
			}

			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			// Check if all arguments are numbers (positions)
			allPositions := true
			positions := make([]int, 0)
			for _, arg := range args[1:] {
				pos, err := strconv.Atoi(arg)
				if err != nil {
					allPositions = false
					break
				}
				positions = append(positions, pos)
			}

			var removedUsers []string
			var notFoundUsers []string
			var invalidPositions []int

			if allPositions {
				// Handle position-based removal
				for _, pos := range positions {
					if pos < 1 || pos > len(queue.List()) {
						invalidPositions = append(invalidPositions, pos)
						continue
					}
					user := queue.List()[pos-1]
					if queue.Remove(user.Username) {
						removedUsers = append(removedUsers, user.Username)
					}
				}
			} else {
				// Handle username-based removal
				for _, username := range args[1:] {
					username = strings.TrimPrefix(username, "@")
					removed, err := queue.RemoveUser(username)
					if err != nil {
						return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
					}
					if removed {
						removedUsers = append(removedUsers, username)
					} else {
						notFoundUsers = append(notFoundUsers, username)
					}
				}
			}

			// Build response message
			var responseParts []string

			if len(removedUsers) > 0 {
				if len(removedUsers) == 1 {
					responseParts = append(responseParts, fmt.Sprintf("removed %s from the queue", removedUsers[0]))
				} else {
					responseParts = append(responseParts, fmt.Sprintf("removed %s from the queue", strings.Join(removedUsers, ", ")))
				}
			}

			if len(notFoundUsers) > 0 {
				if len(notFoundUsers) == 1 {
					responseParts = append(responseParts, fmt.Sprintf("%s is not in the queue", notFoundUsers[0]))
				} else {
					responseParts = append(responseParts, fmt.Sprintf("%s are not in the queue", strings.Join(notFoundUsers, ", ")))
				}
			}

			if len(invalidPositions) > 0 {
				if len(invalidPositions) == 1 {
					responseParts = append(responseParts, fmt.Sprintf("position %d is invalid", invalidPositions[0]))
				} else {
					posStrs := make([]string, len(invalidPositions))
					for i, pos := range invalidPositions {
						posStrs[i] = fmt.Sprintf("%d", pos)
					}
					responseParts = append(responseParts, fmt.Sprintf("positions %s are invalid", strings.Join(posStrs, ", ")))
				}
			}

			if len(responseParts) == 0 {
				return fmt.Sprintf("@%s, no users were removed from the queue.", message.User.Name)
			}

			return fmt.Sprintf("@%s %s!", message.User.Name, strings.Join(responseParts, " and "))
		},
	})

	// Move command - Moves a specified user to a new position in the queue (Mods/VIPs only)
	cm.RegisterCommand(Command{
		Name:        "move",
		Aliases:     []string{"mv"},
		Description: "Move a specified user to a new position in the queue (Mods/VIPs only). If position is beyond queue length, user will be moved to the end.",
		ModOnly:     false,
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()

			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}

			// Check if user is privileged (Mod, VIP, or Broadcaster)
			if !isPrivileged(message) {
				return "This command can only be used by moderators and VIPs."
			}

			args := strings.Fields(message.Message)
			if len(args) < 3 {
				return "Usage: !move <username> <position>"
			}

			username := args[1]
			position, err := strconv.Atoi(args[2])
			if err != nil {
				return "Invalid position number provided."
			}

			// If position is beyond queue length, move to end
			if position > len(queue.List()) {
				err = queue.MoveToEnd(username)
				if err != nil {
					return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
				}
				return fmt.Sprintf("@%s has been moved to the end of the queue!", username)
			}

			err = queue.MoveUser(username, position)
			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}

			return fmt.Sprintf("@%s has been moved to position %d in the queue!", username, position)
		},
	})

	// Pause Queue command - Pauses the queue system (mod/VIP only)
	cm.RegisterCommand(Command{
		Name:         "pausequeue",
		Aliases:      []string{"pq"},
		Description:  "Pause the queue system (no new additions allowed) (Mods/VIPs only)",
		ModOnly:      false,
		IsPrivileged: true,
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}
			if queue.IsPaused() {
				return "Queue system is already paused!"
			}
			err := queue.Pause()
			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}
			return fmt.Sprintf("@%s has paused the queue system!", message.User.Name)
		},
	})

	// Unpause Queue command - Resumes the queue system (mod/VIP only)
	cm.RegisterCommand(Command{
		Name:         "unpausequeue",
		Aliases:      []string{"uq"},
		Description:  "Resume the queue system (Mods/VIPs only)",
		ModOnly:      false,
		IsPrivileged: true,
		Handler: func(message twitch.PrivateMessage) string {
			queue := cm.GetQueue()
			if !queue.IsEnabled() {
				return "The queue system is currently disabled."
			}
			if !queue.IsPaused() {
				return "Queue system is not paused!"
			}
			err := queue.Unpause()
			if err != nil {
				return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
			}
			return fmt.Sprintf("@%s has resumed the queue system!", message.User.Name)
		},
	})

	// Add savequeue command
	cm.RegisterCommand(Command{
		Name:        "savequeue",
		Aliases:     []string{"svq"},
		Description: "Save the current queue state to a file (Mods/VIPs only)",
		Handler: func(message twitch.PrivateMessage) string {
			if !isPrivileged(message) {
				return "Only moderators and VIPs can save the queue state."
			}

			queue := cm.GetQueue()
			if err := queue.SaveState("queue_state.json"); err != nil {
				return fmt.Sprintf("Failed to save queue state: %v", err)
			}

			return "Queue state has been saved successfully!"
		},
		ModOnly:      false,
		IsPrivileged: true,
	})

	// Add restorequeue command
	cm.RegisterCommand(Command{
		Name:        "restorequeue",
		Aliases:     []string{"rq"},
		Description: "Restore the queue state from a saved file (Mods/VIPs only)",
		Handler: func(message twitch.PrivateMessage) string {
			if !isPrivileged(message) {
				return "Only moderators and VIPs can restore the queue state."
			}

			queue := cm.GetQueue()
			if err := queue.LoadState("queue_state.json"); err != nil {
				return fmt.Sprintf("Failed to restore queue state: %v", err)
			}

			return "Queue state has been restored successfully!"
		},
		ModOnly:      false,
		IsPrivileged: true,
	})

	// Add deletequeue command
	cm.RegisterCommand(Command{
		Name:        "deletequeue",
		Aliases:     []string{"dq"},
		Description: "Delete the saved queue state file (Mods/VIPs only)",
		Handler: func(message twitch.PrivateMessage) string {
			if !isPrivileged(message) {
				return "Only moderators and VIPs can delete the saved queue state."
			}

			if err := os.Remove("queue_state.json"); err != nil {
				if os.IsNotExist(err) {
					return "No saved queue state exists to delete."
				}
				return fmt.Sprintf("Failed to delete saved queue state: %v", err)
			}

			return "Saved queue state has been deleted successfully!"
		},
		ModOnly:      false,
		IsPrivileged: true,
	})
}
