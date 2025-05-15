package commands

import (
	"strings"
	"sync"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PerfTiltBot/internal/queue"
)

// Command represents a chat command
type Command struct {
	Name        string
	Description string
	Handler     func(message twitch.PrivateMessage) string
	ModOnly     bool
}

// CommandManager handles all chat commands
type CommandManager struct {
	commands map[string]Command
	prefix   string
	queue    *queue.Queue
	mu       sync.RWMutex
}

// NewCommandManager creates a new command manager
func NewCommandManager(prefix string) *CommandManager {
	return &CommandManager{
		commands: make(map[string]Command),
		prefix:   prefix,
		queue:    queue.NewQueue(),
	}
}

// RegisterCommand adds a new command to the manager
func (cm *CommandManager) RegisterCommand(cmd Command) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.commands[strings.ToLower(cmd.Name)] = cmd
}

// isModerator checks if the user is a moderator or broadcaster
func isModerator(message twitch.PrivateMessage) bool {
	return message.User.Badges["moderator"] > 0 || message.User.Badges["broadcaster"] > 0
}

// HandleMessage processes incoming chat messages and executes commands
func (cm *CommandManager) HandleMessage(message twitch.PrivateMessage) (response string, isCommand bool) {
	if !strings.HasPrefix(message.Message, cm.prefix) {
		return "", false
	}

	// Split the message into command and arguments
	parts := strings.Fields(strings.TrimPrefix(message.Message, cm.prefix))
	if len(parts) == 0 {
		return "", false
	}

	commandName := strings.ToLower(parts[0])

	cm.mu.RLock()
	command, exists := cm.commands[commandName]
	cm.mu.RUnlock()

	if !exists {
		return "", true
	}

	// Check if command is mod-only
	if command.ModOnly && !isModerator(message) {
		return "This command can only be used by moderators.", true
	}

	return command.Handler(message), true
}

// GetCommandList returns a list of all registered commands
func (cm *CommandManager) GetCommandList() []Command {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	commands := make([]Command, 0, len(cm.commands))
	for _, cmd := range cm.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// GetQueue returns the queue manager
func (cm *CommandManager) GetQueue() *queue.Queue {
	return cm.queue
}
