package commands

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PerfTiltBot/internal/queue"
)

// Command represents a chat command that can be executed by users.
// Each command has a name, optional aliases, description, and a handler function
// that processes the command when triggered.
type Command struct {
	// Primary name of the command (e.g., "join", "leave")
	Name string
	// Alternative names for the command (e.g., ["j"] for join)
	Aliases []string
	// Human-readable description of what the command does
	Description string
	// Function that executes when the command is triggered
	// Takes a Twitch message as input and returns a response string
	Handler func(message twitch.PrivateMessage, args []string) string
	// If true, only moderators can use this command
	ModOnly bool
	// If true, only privileged users (mods, VIPs, broadcasters) can use this command
	IsPrivileged bool
	// Cooldown configuration for the command
	Cooldown CooldownConfig
}

// CommandManager handles the registration and execution of all chat commands.
// It maintains a thread-safe registry of commands and manages the queue system.
type CommandManager struct {
	// Map of command names/aliases to their Command objects
	// Keys are lowercase to ensure case-insensitive matching
	commands map[string]*Command
	// Character that must prefix all commands (e.g., "!")
	prefix string
	// Queue system for managing user entries
	queue *queue.Queue
	// Mutex for thread-safe access to the commands map
	mu sync.RWMutex
	// Channel to signal shutdown request
	shutdownCh chan struct{}
	// Cooldown manager for handling command cooldowns
	cooldown *CooldownManager
}

// NewCommandManager creates a new command manager
func NewCommandManager(prefix string) *CommandManager {
	cm := &CommandManager{
		commands:   make(map[string]*Command),
		prefix:     prefix,
		queue:      queue.NewQueue(),
		shutdownCh: make(chan struct{}),
		cooldown:   NewCooldownManager(),
	}
	SetCommandManager(cm)
	return cm
}

// RequestShutdown signals that the bot should shut down.
// This is typically called by the kill command.
func (cm *CommandManager) RequestShutdown() {
	close(cm.shutdownCh)
}

// WaitForShutdown blocks until a shutdown is requested.
// This should be called in the main loop to handle graceful shutdown.
func (cm *CommandManager) WaitForShutdown() {
	<-cm.shutdownCh
}

// RegisterCommand adds a new command to the manager's registry.
// Both the main command name and all aliases are registered in lowercase
// to ensure case-insensitive matching when processing messages.
func (cm *CommandManager) RegisterCommand(cmd *Command) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Register the main command name (converted to lowercase)
	cm.commands[strings.ToLower(cmd.Name)] = cmd

	// Register all aliases (also converted to lowercase)
	for _, alias := range cmd.Aliases {
		cm.commands[strings.ToLower(alias)] = cmd
	}

	// Set default cooldown if not specified
	if cmd.Cooldown == (CooldownConfig{}) {
		cmd.Cooldown = DefaultCooldownConfig()
	}
	cm.cooldown.SetCooldown(cmd.Name, cmd.Cooldown)
}

// isPrivileged checks if a user has moderator, broadcaster, or VIP privileges.
// These privileges may grant access to restricted commands or special features.
func isPrivileged(message twitch.PrivateMessage) bool {
	return message.User.Badges["moderator"] > 0 ||
		message.User.Badges["broadcaster"] > 0 ||
		message.User.Badges["vip"] > 0
}

// HandleMessage processes incoming chat messages and executes commands if present.
// Returns a tuple containing:
// - response: The message to send back to chat (empty if no response needed)
// - isCommand: True if the message was a command attempt (even if invalid)
func (cm *CommandManager) HandleMessage(message twitch.PrivateMessage) (response string, isCommand bool) {
	// Check if the message starts with the command prefix
	if !strings.HasPrefix(message.Message, cm.prefix) {
		return "", false
	}

	// Remove the prefix and split into command and arguments
	parts := strings.Fields(strings.TrimPrefix(message.Message, cm.prefix))
	if len(parts) == 0 {
		return "", false
	}

	// Look up the command in our registry (case-insensitive)
	commandName := strings.ToLower(parts[0])

	cm.mu.RLock()
	command, exists := cm.commands[commandName]
	cm.mu.RUnlock()

	if !exists {
		// Message started with prefix but command wasn't found
		return "", true
	}

	// Check if this is a mod-only command
	if command.ModOnly && message.User.Badges["moderator"] == 0 && message.User.Badges["broadcaster"] == 0 {
		return "This command can only be used by moderators.", true
	}

	// Check if this is a privileged command
	if command.IsPrivileged && !isPrivileged(message) {
		return "This command can only be used by moderators and VIPs.", true
	}

	// Check cooldown
	if remaining := cm.cooldown.CheckCooldown(command.Name, message); remaining > 0 {
		// Only show cooldown message if we haven't shown it for this cooldown period
		if cm.cooldown.ShouldShowCooldownMessage(command.Name, message) {
			// Update the last message time
			cm.cooldown.UpdateLastMessageTime(command.Name, message)
			// Send cooldown message
			return fmt.Sprintf("@%s, this command is on cooldown. Please wait %s.", message.User.Name, FormatCooldown(remaining)), true
		}
		// Don't show message, but still indicate this was a command attempt
		return "", true
	}

	// Execute the command's handler and return its response
	return command.Handler(message, parts[1:]), true
}

// GetCommandList returns a deduplicated list of all registered commands.
// Commands with aliases are only returned once, using their primary name.
func (cm *CommandManager) GetCommandList() []Command {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Use a map to deduplicate commands with aliases
	uniqueCommands := make(map[string]Command)
	for _, cmd := range cm.commands {
		uniqueCommands[cmd.Name] = *cmd
	}

	// Convert map to slice for return
	commands := make([]Command, 0, len(uniqueCommands))
	for _, cmd := range uniqueCommands {
		commands = append(commands, cmd)
	}
	return commands
}

// GetQueue returns the queue manager instance.
// This allows commands to interact with the queue system.
func (cm *CommandManager) GetQueue() *queue.Queue {
	return cm.queue
}
