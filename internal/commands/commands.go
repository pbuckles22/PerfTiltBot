package commands

import (
	"strings"
	"sync"

	"github.com/gempir/go-twitch-irc/v4"
)

// Command represents a chat command
type Command struct {
	Name        string
	Description string
	Handler     func(message twitch.PrivateMessage) string
}

// CommandManager handles all chat commands
type CommandManager struct {
	commands map[string]Command
	prefix   string
	mu       sync.RWMutex
}

// NewCommandManager creates a new command manager
func NewCommandManager(prefix string) *CommandManager {
	return &CommandManager{
		commands: make(map[string]Command),
		prefix:   prefix,
	}
}

// RegisterCommand adds a new command to the manager
func (cm *CommandManager) RegisterCommand(cmd Command) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.commands[strings.ToLower(cmd.Name)] = cmd
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
