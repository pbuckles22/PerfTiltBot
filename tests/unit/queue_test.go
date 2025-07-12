package unit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pbuckles22/PBChatBot/internal/queue"
)

func TestNewQueue(t *testing.T) {
	tempDir := t.TempDir()
	channel := "testchannel"

	q := queue.NewQueue(tempDir, channel)

	if q == nil {
		t.Fatal("NewQueue returned nil")
	}

	if q.IsEnabled() {
		t.Error("New queue should not be enabled by default")
	}

	if q.IsPaused() {
		t.Error("New queue should not be paused by default")
	}
}

func TestQueueEnableDisable(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")

	// Test Enable
	q.Enable()
	if !q.IsEnabled() {
		t.Error("Queue should be enabled after Enable()")
	}
	if q.IsPaused() {
		t.Error("Queue should not be paused after Enable()")
	}
	if q.Size() != 0 {
		t.Error("Queue should be empty after Enable()")
	}

	// Test Disable
	q.Disable()
	if q.IsEnabled() {
		t.Error("Queue should be disabled after Disable()")
	}
	if q.IsPaused() {
		t.Error("Queue should not be paused after Disable()")
	}
	if q.Size() != 0 {
		t.Error("Queue should be empty after Disable()")
	}
}

func TestQueueAdd(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")
	q.Enable()

	// Test adding user
	err := q.Add("testuser", false)
	if err != nil {
		t.Errorf("Failed to add user: %v", err)
	}

	if q.Size() != 1 {
		t.Errorf("Expected queue size 1, got %d", q.Size())
	}

	users := q.List()
	if len(users) != 1 || users[0] != "testuser" {
		t.Errorf("Expected user 'testuser', got %v", users)
	}

	// Test adding duplicate user
	err = q.Add("testuser", false)
	if err == nil {
		t.Error("Should not allow adding duplicate user")
	}
	if !strings.Contains(err.Error(), "already in queue") {
		t.Errorf("Expected 'already in queue' error, got: %v", err)
	}

	// Test adding user when disabled
	q.Disable()
	err = q.Add("anotheruser", false)
	if err == nil {
		t.Error("Should not allow adding user when disabled")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Errorf("Expected 'disabled' error, got: %v", err)
	}
}

func TestQueueRemove(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")
	q.Enable()

	// Add users
	q.Add("user1", false)
	q.Add("user2", false)
	q.Add("user3", false)

	// Test removing existing user
	removed := q.Remove("user2")
	if !removed {
		t.Error("Should successfully remove existing user")
	}

	if q.Size() != 2 {
		t.Errorf("Expected queue size 2, got %d", q.Size())
	}

	users := q.List()
	expected := []string{"user1", "user3"}
	if len(users) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, users)
	}

	// Test removing non-existent user
	removed = q.Remove("nonexistent")
	if removed {
		t.Error("Should not remove non-existent user")
	}

	// Test case-insensitive removal
	removed = q.Remove("USER1")
	if !removed {
		t.Error("Should remove user case-insensitively")
	}
}

func TestQueuePosition(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")
	q.Enable()

	// Add users
	q.Add("user1", false)
	q.Add("user2", false)
	q.Add("user3", false)

	// Test position of existing users
	if pos := q.Position("user1"); pos != 1 {
		t.Errorf("Expected position 1 for user1, got %d", pos)
	}

	if pos := q.Position("user2"); pos != 2 {
		t.Errorf("Expected position 2 for user2, got %d", pos)
	}

	if pos := q.Position("user3"); pos != 3 {
		t.Errorf("Expected position 3 for user3, got %d", pos)
	}

	// Test position of non-existent user
	if pos := q.Position("nonexistent"); pos != -1 {
		t.Errorf("Expected position -1 for non-existent user, got %d", pos)
	}

	// Test case-insensitive position
	if pos := q.Position("USER1"); pos != 1 {
		t.Errorf("Expected position 1 for USER1 (case-insensitive), got %d", pos)
	}
}

func TestQueuePop(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")
	q.Enable()

	// Add users
	q.Add("user1", false)
	q.Add("user2", false)
	q.Add("user3", false)

	// Test popping single user
	user, err := q.Pop()
	if err != nil {
		t.Errorf("Failed to pop user: %v", err)
	}
	if user != "user1" {
		t.Errorf("Expected popped user 'user1', got '%s'", user)
	}

	if q.Size() != 2 {
		t.Errorf("Expected queue size 2 after pop, got %d", q.Size())
	}

	// Test popping multiple users
	users, err := q.PopN(2)
	if err != nil {
		t.Errorf("Failed to pop multiple users: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0] != "user2" || users[1] != "user3" {
		t.Errorf("Expected users ['user2', 'user3'], got %v", users)
	}

	if q.Size() != 0 {
		t.Errorf("Expected empty queue, got size %d", q.Size())
	}

	// Test popping from empty queue
	_, err = q.Pop()
	if err == nil {
		t.Error("Should not be able to pop from empty queue")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected 'empty' error, got: %v", err)
	}
}

func TestQueueMoveUser(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")
	q.Enable()

	// Add users
	q.Add("user1", false)
	q.Add("user2", false)
	q.Add("user3", false)
	q.Add("user4", false)

	// Test moving user to different position
	err := q.MoveUser("user2", 4)
	if err != nil {
		t.Errorf("Failed to move user: %v", err)
	}

	users := q.List()
	expected := []string{"user1", "user3", "user4", "user2"}
	if len(users) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, users)
	}

	// Test moving to same position (should be no-op)
	err = q.MoveUser("user1", 1)
	if err != nil {
		t.Errorf("Moving to same position should not error: %v", err)
	}

	// Test moving non-existent user
	err = q.MoveUser("nonexistent", 2)
	if err == nil {
		t.Error("Should not be able to move non-existent user")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestQueuePauseUnpause(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")
	q.Enable()

	// Test pause
	err := q.Pause()
	if err != nil {
		t.Errorf("Failed to pause queue: %v", err)
	}
	if !q.IsPaused() {
		t.Error("Queue should be paused")
	}

	// Test pause when already paused
	err = q.Pause()
	if err == nil {
		t.Error("Should not be able to pause already paused queue")
	}

	// Test adding user when paused (should fail for non-mod)
	err = q.Add("user1", false)
	if err == nil {
		t.Error("Should not be able to add user when paused (non-mod)")
	}

	// Test adding user when paused (should succeed for mod)
	err = q.Add("user1", true)
	if err != nil {
		t.Errorf("Mod should be able to add user when paused: %v", err)
	}

	// Test unpause
	err = q.Unpause()
	if err != nil {
		t.Errorf("Failed to unpause queue: %v", err)
	}
	if q.IsPaused() {
		t.Error("Queue should not be paused")
	}

	// Test unpause when not paused
	err = q.Unpause()
	if err == nil {
		t.Error("Should not be able to unpause non-paused queue")
	}

	// Test adding user after unpause
	err = q.Add("user2", false)
	if err != nil {
		t.Errorf("Should be able to add user after unpause: %v", err)
	}
}

func TestQueueClear(t *testing.T) {
	tempDir := t.TempDir()
	q := queue.NewQueue(tempDir, "testchannel")
	q.Enable()

	// Add users
	q.Add("user1", false)
	q.Add("user2", false)
	q.Add("user3", false)

	// Test clear
	count := q.Clear()
	if count != 3 {
		t.Errorf("Expected to clear 3 users, got %d", count)
	}

	if q.Size() != 0 {
		t.Error("Queue should be empty after clear")
	}

	// Test clear on empty queue
	count = q.Clear()
	if count != 0 {
		t.Errorf("Expected to clear 0 users, got %d", count)
	}
}

func TestQueueStatePersistence(t *testing.T) {
	tempDir := t.TempDir()
	channel := "testchannel"

	// Create queue and add users
	q := queue.NewQueue(tempDir, channel)
	q.Enable()
	q.Add("user1", false)
	q.Add("user2", false)
	q.Add("user3", false)

	// Wait a moment for auto-save goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Verify state file was created
	stateFile := filepath.Join(tempDir, "queue_state_"+channel+".json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("Queue state file should be created")
	}

	// Create new queue instance (simulating restart)
	q2 := queue.NewQueue(tempDir, channel)

	// Queue should be disabled by default after restart
	if q2.IsEnabled() {
		t.Error("Queue should be disabled after restart")
	}

	// Enable the queue to load state
	q2.Enable()

	// Verify users were loaded
	if q2.Size() != 3 {
		t.Errorf("Expected 3 users after restart, got %d", q2.Size())
	}

	users := q2.List()
	expected := []string{"user1", "user2", "user3"}
	if len(users) != len(expected) {
		t.Errorf("Expected %v after restart, got %v", expected, users)
	}
}
