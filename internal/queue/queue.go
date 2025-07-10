package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// QueuedUser represents a user in the queue
type QueuedUser struct {
	Username string
	JoinTime time.Time
	IsMod    bool
}

// QueueState represents the persistent state of the queue
type QueueState struct {
	Channel     string   `json:"channel"`      // Channel name this queue belongs to
	Queue       []string `json:"queue"`        // List of usernames in queue
	LastUpdated int64    `json:"last_updated"` // Unix timestamp of last update
}

// Queue represents a queue of users
type Queue struct {
	users    []string
	mu       sync.RWMutex
	dataPath string
	channel  string
	enabled  bool
	paused   bool
}

// NewQueue creates a new queue manager
func NewQueue(dataPath string, channel string) *Queue {
	q := &Queue{
		users:    make([]string, 0),
		dataPath: dataPath,
		channel:  channel,
		enabled:  false,
		paused:   false,
	}
	q.LoadState()
	return q
}

// Enable starts the queue system
func (q *Queue) Enable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = true
	q.paused = false
	q.users = make([]string, 0) // Clear queue when enabling
	q.autoSave()                // Auto-save after enabling
}

// Disable stops the queue system and clears the queue
func (q *Queue) Disable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = false
	q.paused = false
	q.users = make([]string, 0)
	q.autoSave() // Auto-save after disabling
}

// Pause pauses the queue system (no new additions allowed)
func (q *Queue) Pause() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return fmt.Errorf("queue system is currently disabled")
	}

	if q.paused {
		return fmt.Errorf("queue system is already paused")
	}

	q.paused = true
	q.autoSave() // Auto-save after pausing
	return nil
}

// Unpause resumes the queue system
func (q *Queue) Unpause() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return fmt.Errorf("queue system is currently disabled")
	}

	if !q.paused {
		return fmt.Errorf("queue system is not paused")
	}

	q.paused = false
	q.autoSave() // Auto-save after unpausing
	return nil
}

// IsPaused returns whether the queue system is paused
func (q *Queue) IsPaused() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.paused
}

// IsEnabled returns whether the queue system is enabled
func (q *Queue) IsEnabled() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.enabled
}

// Clear removes all users from the queue
func (q *Queue) Clear() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := len(q.users)
	q.users = make([]string, 0)
	q.autoSave() // Auto-save after clearing
	return count
}

// Add adds a user to the queue
func (q *Queue) Add(username string, isMod bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return fmt.Errorf("queue system is currently disabled")
	}

	if q.paused && !isMod {
		return fmt.Errorf("queue system is currently paused")
	}

	// Check if user is already in queue (case-insensitive check)
	for _, user := range q.users {
		if strings.EqualFold(user, username) {
			return fmt.Errorf("user is already in queue")
		}
	}

	// Store the username with its exact capitalization
	q.users = append(q.users, username)
	q.autoSave() // Auto-save after adding user
	return nil
}

// Remove removes a user from the queue
func (q *Queue) Remove(username string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, user := range q.users {
		if strings.EqualFold(user, username) {
			// Remove user by slicing
			q.users = append(q.users[:i], q.users[i+1:]...)
			q.autoSave() // Auto-save after removing user
			return true
		}
	}
	return false
}

// List returns the current queue
func (q *Queue) List() []string {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Return a copy to prevent external modifications
	users := make([]string, len(q.users))
	copy(users, q.users)
	return users
}

// Size returns the current queue size
func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.users)
}

// Position returns the position of a user in the queue (1-based)
func (q *Queue) Position(username string) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for i, user := range q.users {
		if strings.EqualFold(user, username) {
			return i + 1
		}
	}
	return -1
}

// AddAtPosition adds a user to the queue at the specified position (1-based)
func (q *Queue) AddAtPosition(username string, position int, isMod bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return fmt.Errorf("queue system is currently disabled")
	}

	if q.paused && !isMod {
		return fmt.Errorf("queue system is currently paused")
	}

	// Check if user is already in queue
	for _, user := range q.users {
		if strings.EqualFold(user, username) {
			return fmt.Errorf("user is already in queue")
		}
	}

	// Validate position
	if position < 1 {
		position = 1
	}
	if position > len(q.users)+1 {
		position = len(q.users) + 1
	}

	// Store the username with its exact capitalization
	newUser := username

	// Insert at position (converting from 1-based to 0-based index)
	position--
	if position == len(q.users) {
		// Append to end
		q.users = append(q.users, newUser)
	} else {
		// Insert at position
		q.users = append(q.users[:position], append([]string{newUser}, q.users[position:]...)...)
	}
	q.autoSave() // Auto-save after adding user at position
	return nil
}

// Pop removes and returns the first user from the queue
func (q *Queue) Pop() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return "", fmt.Errorf("queue system is currently disabled")
	}

	if len(q.users) == 0 {
		return "", fmt.Errorf("queue is empty")
	}

	// Get first user
	user := q.users[0]

	// Remove first user
	q.users = q.users[1:]
	q.autoSave() // Auto-save after popping user

	return user, nil
}

// PopN removes and returns the first N users from the queue
func (q *Queue) PopN(count int) ([]string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return nil, fmt.Errorf("queue system is currently disabled")
	}

	if len(q.users) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	// Ensure count doesn't exceed queue size
	if count > len(q.users) {
		count = len(q.users)
	}

	// Get first N users
	users := make([]string, count)
	copy(users, q.users[:count])

	// Remove first N users
	q.users = q.users[count:]
	q.autoSave() // Auto-save after popping users

	return users, nil
}

// RemoveUser removes a specified user from the queue
func (q *Queue) RemoveUser(username string) (bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return false, fmt.Errorf("queue system is currently disabled")
	}

	for i, user := range q.users {
		if user == username {
			// Remove the user from the queue
			q.users = append(q.users[:i], q.users[i+1:]...)
			q.autoSave() // Auto-save after removing user
			return true, nil
		}
	}

	return false, nil
}

// MoveUser moves a user to a new position in the queue (1-based)
func (q *Queue) MoveUser(username string, position int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return fmt.Errorf("queue system is currently disabled")
	}

	// Find user's current position
	currentPos := -1
	for i, user := range q.users {
		if user == username {
			currentPos = i
			break
		}
	}

	if currentPos == -1 {
		return fmt.Errorf("user not found in queue")
	}

	// Validate position
	if position < 1 {
		position = 1
	}
	if position > len(q.users) {
		position = len(q.users)
	}

	// Convert to 0-based index
	position--

	// If same position, no need to move
	if currentPos == position {
		return nil
	}

	// Get user
	user := q.users[currentPos]

	// Remove from current position
	q.users = append(q.users[:currentPos], q.users[currentPos+1:]...)

	// Insert at new position
	q.users = append(q.users[:position], append([]string{user}, q.users[position:]...)...)
	q.autoSave() // Auto-save after moving user

	return nil
}

// MoveToEnd moves a user to the end of the queue
func (q *Queue) MoveToEnd(username string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return fmt.Errorf("queue system is currently disabled")
	}

	// Find user's current position
	currentPos := -1
	for i, user := range q.users {
		if user == username {
			currentPos = i
			break
		}
	}

	if currentPos == -1 {
		return fmt.Errorf("user not found in queue")
	}

	// If already at end, no need to move
	if currentPos == len(q.users)-1 {
		return nil
	}

	// Get user
	user := q.users[currentPos]

	// Remove from current position
	q.users = append(q.users[:currentPos], q.users[currentPos+1:]...)

	// Add to end
	q.users = append(q.users, user)
	q.autoSave() // Auto-save after moving user to end

	return nil
}

// autoSave automatically saves the queue state after modifications
// This method should be called after any queue modification operation
func (q *Queue) autoSave() {
	// Use a goroutine to avoid blocking the main operation
	go func() {
		if err := q.SaveState(); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Auto-save failed: %v\n", err)
		}
	}()
}

// SaveState saves the current queue state to a file
func (q *Queue) SaveState() error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Ensure the data directory exists
	if err := os.MkdirAll(q.dataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	state := QueueState{
		Channel:     q.channel,
		Queue:       q.users,
		LastUpdated: time.Now().Unix(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal queue state: %w", err)
	}

	// Use channel-specific filename
	filename := filepath.Join(q.dataPath, fmt.Sprintf("queue_state_%s.json", q.channel))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write queue state: %w", err)
	}

	return nil
}

// LoadState loads the queue state from a file
func (q *Queue) LoadState() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Use channel-specific filename
	filename := filepath.Join(q.dataPath, fmt.Sprintf("queue_state_%s.json", q.channel))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, start with empty queue
			q.users = make([]string, 0)
			return nil
		}
		return fmt.Errorf("failed to read queue state: %w", err)
	}

	var state QueueState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal queue state: %w", err)
	}

	// Verify the channel matches
	if state.Channel != q.channel {
		return fmt.Errorf("queue state channel mismatch: expected %s, got %s", q.channel, state.Channel)
	}

	q.users = state.Queue
	return nil
}

// GetDataPath returns the data path for this queue
func (q *Queue) GetDataPath() string {
	return q.dataPath
}
