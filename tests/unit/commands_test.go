package unit

import (
	"strings"
	"testing"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PBChatBot/internal/commands"
)

// Mock message for testing
func createMockMessage(username, message string, isMod, isVIP, isBroadcaster bool) twitchirc.PrivateMessage {
	badges := make(map[string]int)
	if isMod {
		badges["moderator"] = 1
	}
	if isVIP {
		badges["vip"] = 1
	}
	if isBroadcaster {
		badges["broadcaster"] = 1
	}

	return twitchirc.PrivateMessage{
		User: twitchirc.User{
			Name:   username,
			Badges: badges,
		},
		Message: message,
		Channel: "testchannel",
	}
}

func TestHandlePing(t *testing.T) {
	msg := createMockMessage("testuser", "!ping", false, false, false)

	response := commands.HandlePing(msg, []string{})

	if response != "Pong! 🏓" {
		t.Errorf("Expected 'Pong! 🏓', got '%s'", response)
	}
}

func TestHandleStartQueue(t *testing.T) {
	// Test starting queue when disabled
	msg := createMockMessage("testuser", "!startqueue", false, false, false)

	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_start")
	commands.SetCommandManager(cm)

	response := commands.HandleStartQueue(msg, []string{})

	if !strings.Contains(response, "started the queue system") {
		t.Errorf("Expected 'started the queue system', got '%s'", response)
	}

	if !cm.GetQueue().IsEnabled() {
		t.Error("Queue should be enabled after start")
	}

	// Test starting queue when already enabled
	response = commands.HandleStartQueue(msg, []string{})

	if !strings.Contains(response, "already running") {
		t.Errorf("Expected 'already running', got '%s'", response)
	}
}

func TestHandleEndQueue(t *testing.T) {
	msg := createMockMessage("testuser", "!endqueue", false, false, false)

	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_end")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	response := commands.HandleEndQueue(msg, []string{})

	if !strings.Contains(response, "ended the queue system") {
		t.Errorf("Expected 'ended the queue system', got '%s'", response)
	}

	if cm.GetQueue().IsEnabled() {
		t.Error("Queue should be disabled after end")
	}

	// Test ending queue when already disabled
	response = commands.HandleEndQueue(msg, []string{})

	if !strings.Contains(response, "already disabled") {
		t.Errorf("Expected 'already disabled', got '%s'", response)
	}
}

func TestHandleJoin(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_join")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Test joining self
	msg := createMockMessage("testuser", "!join", false, false, false)
	response := commands.HandleJoin(msg, []string{})

	if !strings.Contains(response, "joined queue at position 1") {
		t.Errorf("Expected 'joined queue at position 1', got '%s'", response)
	}

	if cm.GetQueue().Size() != 1 {
		t.Error("Queue should have 1 user after join")
	}

	// Test joining when queue is disabled
	cm.GetQueue().Disable()
	response = commands.HandleJoin(msg, []string{})

	if !strings.Contains(response, "disabled") {
		t.Errorf("Expected 'disabled', got '%s'", response)
	}

	// Test joining with specific username (moderator)
	cm.GetQueue().Enable()
	modMsg := createMockMessage("moduser", "!join otheruser", true, false, false)
	response = commands.HandleJoin(modMsg, []string{"otheruser"})

	if !strings.Contains(response, "otheruser joined queue") {
		t.Errorf("Expected 'otheruser joined queue', got '%s'", response)
	}

	// Test joining with specific username (non-moderator) - should succeed for first user
	regularMsg := createMockMessage("regularuser", "!join anotheruser", false, false, false)
	response = commands.HandleJoin(regularMsg, []string{"anotheruser"})

	if !strings.Contains(response, "anotheruser joined queue") {
		t.Errorf("Expected 'anotheruser joined queue', got '%s'", response)
	}
}

func TestHandleLeave(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Add user to queue
	cm.GetQueue().Add("testuser", false)

	// Test leaving self
	msg := createMockMessage("testuser", "!leave", false, false, false)
	response := commands.HandleLeave(msg, []string{})

	if !strings.Contains(response, "left queue") {
		t.Errorf("Expected 'left queue', got '%s'", response)
	}

	// Wait a moment for auto-save to complete
	time.Sleep(100 * time.Millisecond)

	if cm.GetQueue().Size() != 0 {
		t.Error("Queue should be empty after leave")
	}

	// Test leaving when not in queue
	response = commands.HandleLeave(msg, []string{})

	if !strings.Contains(response, "not in the queue") {
		t.Errorf("Expected 'not in the queue', got '%s'", response)
	}

	// Test leaving when queue is disabled
	cm.GetQueue().Disable()
	response = commands.HandleLeave(msg, []string{})

	if !strings.Contains(response, "disabled") {
		t.Errorf("Expected 'disabled', got '%s'", response)
	}
}

func TestHandleQueue(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_queue")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Test empty queue
	msg := createMockMessage("testuser", "!queue", false, false, false)
	response := commands.HandleQueue(msg, []string{})

	if !strings.Contains(response, "empty") {
		t.Errorf("Expected 'empty', got '%s'", response)
	}

	// Add users and test
	cm.GetQueue().Add("user1", false)
	cm.GetQueue().Add("user2", false)

	response = commands.HandleQueue(msg, []string{})

	if !strings.Contains(response, "Queue: user1, user2 (2 total)") {
		t.Errorf("Expected 'Queue: user1, user2 (2 total)', got '%s'", response)
	}

	// Test when queue is disabled
	cm.GetQueue().Disable()
	response = commands.HandleQueue(msg, []string{})

	if !strings.Contains(response, "disabled") {
		t.Errorf("Expected 'disabled', got '%s'", response)
	}
}

func TestHandlePosition(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_position")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Add users
	cm.GetQueue().Add("user1", false)
	cm.GetQueue().Add("user2", false)
	cm.GetQueue().Add("user3", false)

	// Test position of self
	msg := createMockMessage("user2", "!position", false, false, false)
	response := commands.HandlePosition(msg, []string{})

	if !strings.Contains(response, "user2 is at position 2") {
		t.Errorf("Expected 'user2 is at position 2', got '%s'", response)
	}

	// Test position of specific user
	response = commands.HandlePosition(msg, []string{"user1"})

	if !strings.Contains(response, "user1 is at position 1") {
		t.Errorf("Expected 'user1 is at position 1', got '%s'", response)
	}

	// Test position of non-existent user
	response = commands.HandlePosition(msg, []string{"nonexistent"})

	if !strings.Contains(response, "not in the queue") {
		t.Errorf("Expected 'not in the queue', got '%s'", response)
	}

	// Test position by number
	response = commands.HandlePosition(msg, []string{"2"})

	if !strings.Contains(response, "User at position 2 is user2") {
		t.Errorf("Expected 'User at position 2 is user2', got '%s'", response)
	}

	// Test invalid position number
	response = commands.HandlePosition(msg, []string{"999"})

	if !strings.Contains(response, "Invalid position") {
		t.Errorf("Expected 'Invalid position', got '%s'", response)
	}
}

func TestHandlePop(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_pop")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Add users
	cm.GetQueue().Add("user1", false)
	cm.GetQueue().Add("user2", false)
	cm.GetQueue().Add("user3", false)

	// Test popping single user (default)
	msg := createMockMessage("moduser", "!pop", true, false, false)
	response := commands.HandlePop(msg, []string{})

	if !strings.Contains(response, "Popped: user1") {
		t.Errorf("Expected 'Popped: user1', got '%s'", response)
	}

	if cm.GetQueue().Size() != 2 {
		t.Error("Queue should have 2 users after pop")
	}

	// Test popping multiple users
	response = commands.HandlePop(msg, []string{"2"})

	if !strings.Contains(response, "Popped: user2, user3") {
		t.Errorf("Expected 'Popped: user2, user3', got '%s'", response)
	}

	if cm.GetQueue().Size() != 0 {
		t.Error("Queue should be empty after pop 2")
	}

	// Test popping from empty queue
	response = commands.HandlePop(msg, []string{})

	if !strings.Contains(response, "empty") {
		t.Errorf("Expected 'empty', got '%s'", response)
	}

	// Test invalid pop count
	response = commands.HandlePop(msg, []string{"invalid"})

	if !strings.Contains(response, "Invalid number") {
		t.Errorf("Expected 'Invalid number', got '%s'", response)
	}

	// Test negative pop count
	response = commands.HandlePop(msg, []string{"-1"})

	if !strings.Contains(response, "Invalid number") {
		t.Errorf("Expected 'Invalid number', got '%s'", response)
	}
}

func TestHandleRemove(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_remove")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Add users
	cm.GetQueue().Add("user1", false)
	cm.GetQueue().Add("user2", false)
	cm.GetQueue().Add("user3", false)

	// Test removing by username
	msg := createMockMessage("moduser", "!remove user2", true, false, false)
	response := commands.HandleRemove(msg, []string{"user2"})

	if !strings.Contains(response, "removed from queue") {
		t.Errorf("Expected 'removed from queue', got '%s'", response)
	}

	if cm.GetQueue().Size() != 2 {
		t.Error("Queue should have 2 users after remove")
	}

	// Test removing by position
	response = commands.HandleRemove(msg, []string{"1"})

	if !strings.Contains(response, "removed from queue") {
		t.Errorf("Expected 'removed from queue', got '%s'", response)
	}

	// Test removing non-existent user
	response = commands.HandleRemove(msg, []string{"nonexistent"})

	if !strings.Contains(response, "not in the queue") {
		t.Errorf("Expected 'not in the queue', got '%s'", response)
	}

	// Test removing from invalid position
	response = commands.HandleRemove(msg, []string{"999"})

	if !strings.Contains(response, "Invalid position") {
		t.Errorf("Expected 'Invalid position', got '%s'", response)
	}

	// Test missing argument
	response = commands.HandleRemove(msg, []string{})

	if !strings.Contains(response, "Usage:") {
		t.Errorf("Expected usage message, got '%s'", response)
	}
}

func TestHandleMove(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_move")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Add users
	cm.GetQueue().Add("user1", false)
	cm.GetQueue().Add("user2", false)
	cm.GetQueue().Add("user3", false)

	// Test moving by username
	msg := createMockMessage("moduser", "!move user2 3", true, false, false)
	response := commands.HandleMove(msg, []string{"user2", "3"})

	if !strings.Contains(response, "moved to position 3") {
		t.Errorf("Expected 'moved to position 3', got '%s'", response)
	}

	users := cm.GetQueue().List()
	expected := []string{"user1", "user3", "user2"}
	if len(users) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, users)
	}

	// Test moving by position
	response = commands.HandleMove(msg, []string{"1", "2"})

	if !strings.Contains(response, "moved to position 2") {
		t.Errorf("Expected 'moved to position 2', got '%s'", response)
	}

	// Test moving non-existent user
	response = commands.HandleMove(msg, []string{"nonexistent", "1"})

	if !strings.Contains(response, "not in the queue") {
		t.Errorf("Expected 'not in the queue', got '%s'", response)
	}

	// Test invalid target position
	response = commands.HandleMove(msg, []string{"user1", "invalid"})

	if !strings.Contains(response, "Invalid target position") {
		t.Errorf("Expected 'Invalid target position', got '%s'", response)
	}

	// Test missing arguments
	response = commands.HandleMove(msg, []string{"user1"})

	if !strings.Contains(response, "Usage:") {
		t.Errorf("Expected usage message, got '%s'", response)
	}
}

func TestHandleClearQueue(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_clear")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Add users
	cm.GetQueue().Add("user1", false)
	cm.GetQueue().Add("user2", false)
	cm.GetQueue().Add("user3", false)

	// Test clearing queue
	msg := createMockMessage("moduser", "!clearqueue", true, false, false)
	response := commands.HandleClearQueue(msg, []string{})

	if !strings.Contains(response, "Queue cleared (3 users removed)") {
		t.Errorf("Expected 'Queue cleared (3 users removed)', got '%s'", response)
	}

	if cm.GetQueue().Size() != 0 {
		t.Error("Queue should be empty after clear")
	}

	// Test clearing empty queue
	response = commands.HandleClearQueue(msg, []string{})

	if !strings.Contains(response, "Queue cleared (0 users removed)") {
		t.Errorf("Expected 'Queue cleared (0 users removed)', got '%s'", response)
	}

	// Test clearing when disabled
	cm.GetQueue().Disable()
	response = commands.HandleClearQueue(msg, []string{})

	if !strings.Contains(response, "disabled") {
		t.Errorf("Expected 'disabled', got '%s'", response)
	}
}

func TestHandlePauseUnpause(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_pause")
	commands.SetCommandManager(cm)
	cm.GetQueue().Enable()

	// Test pause
	msg := createMockMessage("moduser", "!pausequeue", true, false, false)
	response := commands.HandlePause(msg, []string{})

	if !strings.Contains(response, "paused") {
		t.Errorf("Expected 'paused', got '%s'", response)
	}

	if !cm.GetQueue().IsPaused() {
		t.Error("Queue should be paused")
	}

	// Test pause when already paused
	response = commands.HandlePause(msg, []string{})

	if !strings.Contains(response, "already paused") {
		t.Errorf("Expected 'already paused', got '%s'", response)
	}

	// Test unpause
	response = commands.HandleUnpause(msg, []string{})

	if !strings.Contains(response, "open again") {
		t.Errorf("Expected 'open again', got '%s'", response)
	}

	if cm.GetQueue().IsPaused() {
		t.Error("Queue should not be paused")
	}

	// Test unpause when not paused
	response = commands.HandleUnpause(msg, []string{})

	if !strings.Contains(response, "not paused") {
		t.Errorf("Expected 'not paused', got '%s'", response)
	}
}

func TestHandleHelp(t *testing.T) {
	// Reset command manager for test
	commands.SetCommandManager(nil)
	tempDir := t.TempDir()
	cm := commands.NewCommandManager("!", tempDir, "testchannel_help")
	commands.SetCommandManager(cm)

	// Register some commands
	cm.RegisterCommand(&commands.Command{
		Name:        "help",
		Description: "Show help",
		Handler:     commands.HandleHelp,
	})
	cm.RegisterCommand(&commands.Command{
		Name:        "ping",
		Description: "Ping the bot",
		Handler:     commands.HandlePing,
	})
	cm.RegisterCommand(&commands.Command{
		Name:        "join",
		Description: "Join queue",
		Handler:     commands.HandleJoin,
	})

	msg := createMockMessage("testuser", "!help", false, false, false)
	response := commands.HandleHelp(msg, []string{})

	if !strings.Contains(response, "Available commands:") {
		t.Errorf("Expected 'Available commands:', got '%s'", response)
	}

	if !strings.Contains(response, "Base Commands:") {
		t.Errorf("Expected 'Base Commands:', got '%s'", response)
	}

	if !strings.Contains(response, "Queue Commands:") {
		t.Errorf("Expected 'Queue Commands:', got '%s'", response)
	}

	if !strings.Contains(response, "help") {
		t.Errorf("Expected 'help' in response, got '%s'", response)
	}

	if !strings.Contains(response, "ping") {
		t.Errorf("Expected 'ping' in response, got '%s'", response)
	}

	if !strings.Contains(response, "join") {
		t.Errorf("Expected 'join' in response, got '%s'", response)
	}
}
