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

// GetCommandManager returns the command manager instance
func GetCommandManager() *CommandManager {
	return commandManager
}

// handleHelp shows the list of available commands
func handleHelp(message twitch.PrivateMessage, args []string) string {
	commands := commandManager.GetCommandList()
	var commandList []string

	// Build the list of commands to display based on user permissions
	for _, cmd := range commands {
		// Check if user has permission to use this command
		if cmd.ModOnly && !isPrivileged(message) {
			continue // Skip mod-only commands for non-privileged users
		}
		if cmd.IsPrivileged && !isPrivileged(message) {
			continue // Skip privileged commands for regular users
		}

		// Build command info with name and aliases
		cmdInfo := fmt.Sprintf("!%s", cmd.Name)
		if len(cmd.Aliases) > 0 {
			aliases := make([]string, len(cmd.Aliases))
			for i, alias := range cmd.Aliases {
				aliases[i] = fmt.Sprintf("!%s", alias)
			}
			cmdInfo = fmt.Sprintf("%s (%s)", cmdInfo, strings.Join(aliases, ", "))
		}

		// Add description
		cmdInfo = fmt.Sprintf("%s: %s", cmdInfo, cmd.Description)

		// Add permission info
		if cmd.ModOnly {
			cmdInfo = fmt.Sprintf("%s [Mod Only]", cmdInfo)
		} else if cmd.IsPrivileged {
			cmdInfo = fmt.Sprintf("%s [Mod/VIP]", cmdInfo)
		}

		commandList = append(commandList, cmdInfo)
	}

	if len(commandList) == 0 {
		return "No commands available."
	}

	// Group commands by category
	var baseCommands []string
	var queueCommands []string

	for _, cmd := range commandList {
		// Base commands that are always available
		if strings.Contains(cmd, "help") || strings.Contains(cmd, "ping") || strings.Contains(cmd, "uptime") {
			baseCommands = append(baseCommands, cmd)
		} else {
			queueCommands = append(queueCommands, cmd)
		}
	}

	// Build the response
	var response strings.Builder
	response.WriteString("Available commands:\n")

	if len(baseCommands) > 0 {
		response.WriteString("Base Commands:\n")
		for _, cmd := range baseCommands {
			response.WriteString(fmt.Sprintf("• %s\n", cmd))
		}
	}

	if len(queueCommands) > 0 {
		response.WriteString("\nQueue Commands:\n")
		for _, cmd := range queueCommands {
			response.WriteString(fmt.Sprintf("• %s\n", cmd))
		}
	}

	return response.String()
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
	return fmt.Sprintf("Queue cleared (%d users removed)", count)
}

// handleJoin handles the !join command
func handleJoin(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is currently disabled."
	}

	// If no arguments provided, add the command user
	if len(args) == 0 {
		err := cm.GetQueue().Add(message.User.Name, isPrivileged(message))
		if err != nil {
			return fmt.Sprintf("Error joining queue: %v", err)
		}
		pos := cm.GetQueue().Position(message.User.Name)
		total := cm.GetQueue().Size()
		return fmt.Sprintf("%s joined queue at position %d (%d total)", message.User.Name, pos, total)
	}

	// If arguments provided and user is privileged, add all specified users
	if isPrivileged(message) {
		var responses []string
		for _, username := range args {
			// Use the exact username provided in the command
			err := cm.GetQueue().Add(username, true)
			if err != nil {
				responses = append(responses, fmt.Sprintf("Error adding %s: %v", username, err))
			} else {
				pos := cm.GetQueue().Position(username)
				total := cm.GetQueue().Size()
				responses = append(responses, fmt.Sprintf("%s joined queue at position %d (%d total)", username, pos, total))
			}
		}
		return strings.Join(responses, " ")
	}

	// If not privileged, only add the first user with exact case
	err := cm.GetQueue().Add(args[0], false)
	if err != nil {
		return fmt.Sprintf("Error joining queue: %v", err)
	}
	pos := cm.GetQueue().Position(args[0])
	total := cm.GetQueue().Size()
	return fmt.Sprintf("%s joined queue at position %d (%d total)", args[0], pos, total)
}

// handleLeave handles the !leave command
func handleLeave(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is currently disabled."
	}

	username := message.User.Name
	if len(args) > 0 && isPrivileged(message) {
		username = args[0]
	}

	// Get the current queue to find the exact case of the username
	users := cm.GetQueue().List()
	var exactUsername string
	for _, user := range users {
		if strings.EqualFold(user, username) {
			exactUsername = user
			break
		}
	}

	if exactUsername == "" {
		return fmt.Sprintf("%s is not in the queue!", username)
	}

	if cm.GetQueue().Remove(exactUsername) {
		return fmt.Sprintf("%s left queue", exactUsername)
	}
	return fmt.Sprintf("%s is not in the queue!", username)
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
		userList = append(userList, fmt.Sprintf("%d) %s", i+1, user))
	}

	return fmt.Sprintf("Queue: %s (%d total)", strings.Join(users, ", "), len(users))
}

// handlePosition shows a user's position in the queue
func handlePosition(message twitch.PrivateMessage, args []string) string {
	queue := commandManager.GetQueue()
	if !queue.IsEnabled() {
		return "Queue system is currently disabled."
	}

	// If no arguments, show position of command user
	if len(args) == 0 {
		position := queue.Position(message.User.Name)
		if position == -1 {
			return fmt.Sprintf("@%s, you are not in the queue!", message.User.Name)
		}
		return fmt.Sprintf("%s is at position %d", message.User.Name, position)
	}

	// Try to parse argument as a position number
	position, err := strconv.Atoi(args[0])
	if err == nil {
		// If it's a valid number, get the user at that position
		users := queue.List()
		if position < 1 || position > len(users) {
			return fmt.Sprintf("Invalid position. Queue has %d users.", len(users))
		}
		username := users[position-1]
		return fmt.Sprintf("User at position %d is %s", position, username)
	}

	// If not a number, treat as username
	username := args[0]
	position = queue.Position(username)
	if position == -1 {
		return fmt.Sprintf("%s is not in the queue!", username)
	}
	return fmt.Sprintf("%s is at position %d", username, position)
}

// handlePop handles the !pop command
func handlePop(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is currently disabled."
	}

	count := 1
	if len(args) > 0 {
		var err error
		count, err = strconv.Atoi(args[0])
		if err != nil || count < 1 {
			return "Invalid number of users to pop. Please specify a positive number."
		}
	}

	users, err := cm.GetQueue().PopN(count)
	if err != nil {
		return fmt.Sprintf("Error popping users: %v", err)
	}

	if len(users) == 0 {
		return "Queue is empty."
	}

	// Format the response
	var response strings.Builder
	response.WriteString("Popped: ")
	for i, user := range users {
		if i > 0 {
			response.WriteString(", ")
		}
		response.WriteString(user)
	}

	return response.String()
}

// handleRemove handles the !remove command
func handleRemove(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is currently disabled."
	}

	if len(args) < 1 {
		return "Usage: !remove <username> or !remove <position>"
	}

	// Try to parse the argument as a position number
	position, err := strconv.Atoi(args[0])
	if err == nil {
		// If it's a valid number, get the user at that position
		users := cm.GetQueue().List()
		if position < 1 || position > len(users) {
			return fmt.Sprintf("Invalid position. Queue has %d users.", len(users))
		}
		username := users[position-1]
		if cm.GetQueue().Remove(username) {
			return fmt.Sprintf("%s (position %d) removed from queue", username, position)
		}
		return fmt.Sprintf("Error removing user at position %d", position)
	}

	// If not a number, treat as username
	username := args[0]
	// Get the current queue to find the exact case of the username
	users := cm.GetQueue().List()
	var exactUsername string
	for _, user := range users {
		if strings.EqualFold(user, username) {
			exactUsername = user
			break
		}
	}

	if exactUsername == "" {
		return fmt.Sprintf("%s is not in the queue!", username)
	}

	if cm.GetQueue().Remove(exactUsername) {
		return fmt.Sprintf("%s removed from queue", exactUsername)
	}
	return fmt.Sprintf("Error removing %s from the queue.", username)
}

// handleMove handles the !move command
func handleMove(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is currently disabled."
	}

	if len(args) < 2 {
		return "Usage: !move <username/position> <position>"
	}

	// Get the current queue
	users := cm.GetQueue().List()
	var exactUsername string

	// Try to parse first argument as a position number
	fromPosition, err := strconv.Atoi(args[0])
	if err == nil {
		// If it's a valid number, get the user at that position
		if fromPosition < 1 || fromPosition > len(users) {
			return fmt.Sprintf("Invalid from position. Queue has %d users.", len(users))
		}
		exactUsername = users[fromPosition-1]
	} else {
		// If not a number, treat as username
		username := args[0]
		// Find the exact case of the username
		for _, user := range users {
			if strings.EqualFold(user, username) {
				exactUsername = user
				break
			}
		}
	}

	if exactUsername == "" {
		return fmt.Sprintf("%s is not in the queue!", args[0])
	}

	// Parse the target position
	toPosition, err := strconv.Atoi(args[1])
	if err != nil {
		return "Invalid target position. Please provide a number."
	}

	err = cm.GetQueue().MoveUser(exactUsername, toPosition)
	if err != nil {
		return fmt.Sprintf("Error moving user: %v", err)
	}

	return fmt.Sprintf("%s moved to position %d", exactUsername, toPosition)
}

// handlePause pauses the queue system
func handlePause(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is not enabled"
	}

	if err := cm.GetQueue().Pause(); err != nil {
		return fmt.Sprintf("Error pausing queue: %v", err)
	}
	return "Queue is now paused. No new entries can be added until the queue is unpaused."
}

// handleUnpause handles the !unpause command
func handleUnpause(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is not enabled"
	}

	if err := cm.GetQueue().Unpause(); err != nil {
		return fmt.Sprintf("Error unpausing queue: %v", err)
	}
	return "Queue is now open again."
}

// handleSaveState handles the !save command
func handleSaveState(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	queue := cm.GetQueue()

	// Get current queue state before saving
	users := queue.List()

	if err := queue.SaveBackup(); err != nil {
		return fmt.Sprintf("Error saving queue state: %v", err)
	}

	// Provide more detailed feedback
	if len(users) == 0 {
		return "Queue state has been saved (empty queue)"
	}
	return fmt.Sprintf("Queue state has been saved with %d user(s)", len(users))
}

// handleLoadState handles the !load command
func handleLoadState(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	queue := cm.GetQueue()

	// If queue is disabled, enable it first
	wasDisabled := !queue.IsEnabled()
	if wasDisabled {
		queue.Enable()
	}

	// Try to restore the saved queue state from backup
	if err := queue.LoadBackup(); err != nil {
		if wasDisabled {
			return "Queue system has been started!"
		}
		// Provide more specific error message
		if os.IsNotExist(err) {
			return "No backup file found. Use !savequeue to create a backup first."
		}
		return fmt.Sprintf("Error loading queue state: %v", err)
	}

	users := queue.List()
	if wasDisabled {
		return fmt.Sprintf("Queue system has been started and restored with %d user(s)!", len(users))
	}
	return fmt.Sprintf("Queue state has been restored with %d user(s)!", len(users))
}

// handleRestoreAuto handles the !restoreauto command (for testing crash recovery)
func handleRestoreAuto(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	queue := cm.GetQueue()

	// If queue is disabled, enable it first
	wasDisabled := !queue.IsEnabled()
	if wasDisabled {
		queue.Enable()
	}

	// Try to restore from the auto-save file (simulating crash recovery)
	if err := queue.LoadState(); err != nil {
		if wasDisabled {
			return "Queue system has been started!"
		}
		return fmt.Sprintf("Error loading auto-save state: %v", err)
	}

	users := queue.List()
	if wasDisabled {
		return fmt.Sprintf("Queue system has been started and auto-restored with %d user(s)!", len(users))
	}
	return fmt.Sprintf("Auto-save state has been restored with %d user(s)!", len(users))
}

// handleKill handles the !kill command
func handleKill(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	cm.RequestShutdown()
	return "Bot shutdown initiated. Goodbye! 👋"
}

// handleRestart handles the !restart command
func handleRestart(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	cm.RequestShutdown()
	return "Bot restart initiated. See you soon! 🔄"
}

// handleEnable handles the !enable command
func handleEnable(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	cm.GetQueue().Enable()
	return "Queue system has been enabled!"
}

// handleDisable handles the !disable command
func handleDisable(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	cm.GetQueue().Disable()
	return "Queue system has been disabled!"
}

// handleClear handles the !clear command
func handleClear(message twitch.PrivateMessage, args []string) string {
	cm := GetCommandManager()
	if !cm.GetQueue().IsEnabled() {
		return "Queue system is currently disabled."
	}

	count := cm.GetQueue().Clear()
	return fmt.Sprintf("Queue cleared! Removed %d user(s).", count)
}
