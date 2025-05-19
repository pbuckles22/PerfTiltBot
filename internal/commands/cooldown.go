package commands

import (
	"fmt"
	"sync"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

// UserType represents different types of users in the system
type UserType string

const (
	UserTypeRegular     UserType = "regular"
	UserTypeVIP         UserType = "vip"
	UserTypeMod         UserType = "mod"
	UserTypeBroadcaster UserType = "broadcaster"
)

// CooldownConfig represents the cooldown settings for a command
type CooldownConfig struct {
	// Cooldown durations for different user types
	Regular     time.Duration
	VIP         time.Duration
	Mod         time.Duration
	Broadcaster time.Duration
}

// DefaultCooldownConfig returns a default cooldown configuration
func DefaultCooldownConfig() CooldownConfig {
	return CooldownConfig{
		Regular:     30 * time.Second,
		VIP:         15 * time.Second,
		Mod:         5 * time.Second,
		Broadcaster: 0, // No cooldown for broadcaster
	}
}

// CooldownManager handles command cooldowns for different user types
type CooldownManager struct {
	// Map of command names to their cooldown configurations
	configs map[string]CooldownConfig
	// Map of command names to user last usage times
	lastUsage map[string]map[string]time.Time
	// Map of command names to user last cooldown message times
	lastMessage map[string]map[string]time.Time
	mu          sync.RWMutex
}

// NewCooldownManager creates a new cooldown manager
func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		configs:     make(map[string]CooldownConfig),
		lastUsage:   make(map[string]map[string]time.Time),
		lastMessage: make(map[string]map[string]time.Time),
	}
}

// SetCooldown sets the cooldown configuration for a command
func (cm *CooldownManager) SetCooldown(commandName string, config CooldownConfig) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.configs[commandName] = config
	if _, exists := cm.lastUsage[commandName]; !exists {
		cm.lastUsage[commandName] = make(map[string]time.Time)
	}
	if _, exists := cm.lastMessage[commandName]; !exists {
		cm.lastMessage[commandName] = make(map[string]time.Time)
	}
}

// GetUserType determines the user type based on their badges
func GetUserType(message twitch.PrivateMessage) UserType {
	if message.User.Badges["broadcaster"] > 0 {
		return UserTypeBroadcaster
	}
	if message.User.Badges["moderator"] > 0 {
		return UserTypeMod
	}
	if message.User.Badges["vip"] > 0 {
		return UserTypeVIP
	}
	return UserTypeRegular
}

// CheckCooldown checks if a command is on cooldown for a user
// Returns remaining cooldown duration if on cooldown, 0 if not
func (cm *CooldownManager) CheckCooldown(commandName string, message twitch.PrivateMessage) time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Get cooldown config for command
	config, exists := cm.configs[commandName]
	if !exists {
		return 0 // No cooldown if not configured
	}

	// Get user type
	userType := GetUserType(message)

	// Get cooldown duration for user type
	var cooldown time.Duration
	switch userType {
	case UserTypeBroadcaster:
		cooldown = config.Broadcaster
	case UserTypeMod:
		cooldown = config.Mod
	case UserTypeVIP:
		cooldown = config.VIP
	default:
		cooldown = config.Regular
	}

	// No cooldown if duration is 0
	if cooldown == 0 {
		return 0
	}

	// Get last usage time for this command and user
	lastUsage, exists := cm.lastUsage[commandName][message.User.Name]
	if !exists {
		return 0 // No previous usage
	}

	// Calculate remaining cooldown
	remaining := cooldown - time.Since(lastUsage)
	if remaining <= 0 {
		return 0 // Cooldown expired
	}

	return remaining
}

// ShouldShowCooldownMessage checks if we should show the cooldown message to the user
func (cm *CooldownManager) ShouldShowCooldownMessage(commandName string, message twitch.PrivateMessage) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Get last message time for this command and user
	lastMessage, exists := cm.lastMessage[commandName][message.User.Name]
	if !exists {
		return true // No previous message
	}

	// Get cooldown config for command
	config, exists := cm.configs[commandName]
	if !exists {
		return true // No cooldown config, show message
	}

	// Get user type
	userType := GetUserType(message)

	// Get cooldown duration for user type
	var cooldown time.Duration
	switch userType {
	case UserTypeBroadcaster:
		cooldown = config.Broadcaster
	case UserTypeMod:
		cooldown = config.Mod
	case UserTypeVIP:
		cooldown = config.VIP
	default:
		cooldown = config.Regular
	}

	// If cooldown has expired, show message
	return time.Since(lastMessage) >= cooldown
}

// UpdateLastUsage updates the last usage time for a command and user
func (cm *CooldownManager) UpdateLastUsage(commandName string, message twitch.PrivateMessage) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.lastUsage[commandName]; !exists {
		cm.lastUsage[commandName] = make(map[string]time.Time)
	}
	cm.lastUsage[commandName][message.User.Name] = time.Now()
}

// UpdateLastMessageTime updates the last time we showed a cooldown message to a user
func (cm *CooldownManager) UpdateLastMessageTime(commandName string, message twitch.PrivateMessage) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.lastMessage[commandName]; !exists {
		cm.lastMessage[commandName] = make(map[string]time.Time)
	}
	cm.lastMessage[commandName][message.User.Name] = time.Now()
}

// FormatCooldown formats a cooldown duration into a human-readable string
func FormatCooldown(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
