package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

// commandManager is a package-level variable that holds the command manager instance
var commandManager *CommandManager

// SetCommandManager sets the command manager instance for the handlers
func SetCommandManager(cm *CommandManager) {
	commandManager = cm
}

// handleHelp shows the list of available commands
func handleHelp(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	commands := commandManager.GetCommandList()
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
}

// handlePing checks if the bot is alive
func handlePing(message twitch.PrivateMessage, args []string) string {
	return "Pong! 🏓"
}

// handleStartQueue starts the queue system
func handleStartQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if queue.IsEnabled() {
		return "Queue system is already running!"
	}
	queue.Enable()
	return fmt.Sprintf("@%s has started the queue system!", message.User.Name)
}

// handleEndQueue ends the queue system
func handleEndQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is already disabled!"
	}
	queue.Disable()
	return fmt.Sprintf("@%s has ended the queue system!", message.User.Name)
}

// handleClearQueue clears all users from the queue
func handleClearQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}
	count := queue.Clear()
	return fmt.Sprintf("@%s cleared the queue! Removed %d user(s).", message.User.Name, count)
}

// handleJoin adds a user to the queue
func handleJoin(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}

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
	if len(args) == 0 {
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
	lastArg := args[len(args)-1]
	position, err := strconv.Atoi(lastArg)
	hasPosition := err == nil

	// Get all usernames (excluding position if present)
	usernames := args
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
}

// handleLeave removes a user from the queue
func handleLeave(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}

	var targetUser string
	if len(args) > 0 && isPrivileged(message) {
		// Mod/VIP is removing someone else
		targetUser = strings.TrimPrefix(args[0], "@")
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
}

// handleQueue shows the current queue
func handleQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
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
}

// handlePosition shows a user's position in the queue
func handlePosition(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}

	position := queue.Position(message.User.Name)
	if position == -1 {
		return fmt.Sprintf("@%s, you are not in the queue!", message.User.Name)
	}
	return fmt.Sprintf("@%s, you are at position %d in the queue!", message.User.Name, position)
}

// handlePop removes users from the front of the queue
func handlePop(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}

	// Parse count from message if provided
	count := 1 // Default to 1 if no count specified
	if len(args) > 0 {
		var err error
		count, err = strconv.Atoi(args[0])
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
}

// handleRemove removes specified users from the queue
func handleRemove(message twitch.PrivateMessage, args []string) string {
	if len(args) == 0 {
		return "Usage: !remove <username1> [username2] ... or !remove <position1> [position2] ..."
	}

	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}

	// Check if all arguments are numbers (positions)
	allPositions := true
	positions := make([]int, 0)
	for _, arg := range args {
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
		for _, username := range args {
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
}

// handleMove moves a user to a new position in the queue
func handleMove(message twitch.PrivateMessage, args []string) string {
	if len(args) < 2 {
		return "Usage: !move <username> <position>"
	}

	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}

	username := args[0]
	position, err := strconv.Atoi(args[1])
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
}

// handlePauseQueue pauses the queue system
func handlePauseQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}
	if queue.IsPaused() {
		return "Queue system is already paused!"
	}
	err := queue.Pause()
	if err != nil {
		return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
	}
	return fmt.Sprintf("@%s has paused the queue system!", message.User.Name)
}

// handleUnpauseQueue resumes the queue system
func handleUnpauseQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}
	if !queue.IsPaused() {
		return "Queue system is not paused!"
	}
	err := queue.Unpause()
	if err != nil {
		return fmt.Sprintf("@%s, %s", message.User.Name, err.Error())
	}
	return fmt.Sprintf("@%s has resumed the queue system!", message.User.Name)
}

// handleSaveQueue saves the current queue state
func handleSaveQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if err := queue.SaveState("/app/data/queue_state.json"); err != nil {
		return fmt.Sprintf("Failed to save queue state: %v", err)
	}
	return "Queue state has been saved successfully!"
}

// handleRestoreQueue restores the queue state from a saved file
func handleRestoreQueue(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if err := queue.LoadState("/app/data/queue_state.json"); err != nil {
		return fmt.Sprintf("Failed to restore queue state: %v", err)
	}
	return "Queue state has been restored successfully!"
}

// handleDeleteQueue deletes the saved queue state
func handleDeleteQueue(message twitch.PrivateMessage, args []string) string {
	if err := os.Remove("/app/data/queue_state.json"); err != nil {
		if os.IsNotExist(err) {
			return "No saved queue state exists to delete."
		}
		return fmt.Sprintf("Failed to delete saved queue state: %v", err)
	}
	return "Saved queue state has been deleted successfully!"
}

// handleKill shuts down the bot
func handleKill(message twitch.PrivateMessage, args []string) string {
	// Signal that we want to shut down
	commandManager.RequestShutdown()
	return fmt.Sprintf("@%s has initiated bot shutdown. Goodbye! 👋", message.User.Name)
}
