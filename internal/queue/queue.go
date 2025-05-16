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
	Username  string    `json:"username"`  // Original display name
	LowerName string    `json:"lowerName"` // Lowercase version for matching
	JoinTime  time.Time `json:"joinTime"`
	IsMod     bool      `json:"isMod"`
}

// QueueState represents the saved state of the queue
type QueueState struct {
	Enabled bool         `json:"enabled"`
	Users   []QueuedUser `json:"users"`
}

// Queue manages the user queue
type Queue struct {
	users      []QueuedUser
	enabled    bool
	mu         sync.RWMutex
	stateFile  string
	saveTicker *time.Ticker
	done       chan bool
}

// NewQueue creates a new queue manager
func NewQueue(stateFile string) *Queue {
	q := &Queue{
		users:     make([]QueuedUser, 0),
		enabled:   false,
		stateFile: stateFile,
		done:      make(chan bool),
	}

	// Try to load existing state
	if err := q.loadState(); err != nil {
		fmt.Printf("Warning: Could not load queue state: %v\n", err)
	}

	// Start periodic save
	q.startPeriodicSave()

	return q
}

// startPeriodicSave starts a goroutine to periodically save the queue state
func (q *Queue) startPeriodicSave() {
	q.saveTicker = time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-q.saveTicker.C:
				if err := q.saveState(); err != nil {
					fmt.Printf("Error saving queue state: %v\n", err)
				}
			case <-q.done:
				q.saveTicker.Stop()
				return
			}
		}
	}()
}

// saveState saves the current queue state to a file
func (q *Queue) saveState() error {
	q.mu.RLock()
	state := QueueState{
		Enabled: q.enabled,
		Users:   q.users,
	}
	q.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(q.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %v", err)
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal queue state: %v", err)
	}

	// Write to file
	if err := os.WriteFile(q.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write queue state: %v", err)
	}

	return nil
}

// loadState loads the queue state from a file
func (q *Queue) loadState() error {
	// Check if file exists
	if _, err := os.Stat(q.stateFile); os.IsNotExist(err) {
		return nil // No state file, start fresh
	}

	// Read file
	data, err := os.ReadFile(q.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read queue state: %v", err)
	}

	// Unmarshal JSON
	var state QueueState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal queue state: %v", err)
	}

	// Update queue
	q.mu.Lock()
	q.enabled = state.Enabled
	q.users = state.Users
	q.mu.Unlock()

	return nil
}

// Close stops the periodic save and saves the final state
func (q *Queue) Close() error {
	q.done <- true
	return q.saveState()
}

// Enable starts the queue system
func (q *Queue) Enable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = true
	q.users = make([]QueuedUser, 0) // Clear queue when enabling

	// Save state after enabling
	if err := q.saveState(); err != nil {
		fmt.Printf("Warning: Failed to save queue state after enabling: %v\n", err)
	}
}

// Disable stops the queue system and clears the queue
func (q *Queue) Disable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = false
	q.users = make([]QueuedUser, 0)

	// Save state after disabling
	if err := q.saveState(); err != nil {
		fmt.Printf("Warning: Failed to save queue state after disabling: %v\n", err)
	}
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
	q.users = make([]QueuedUser, 0)

	// Save state after clearing queue
	if err := q.saveState(); err != nil {
		fmt.Printf("Warning: Failed to save queue state after clearing queue: %v\n", err)
	}
	return count
}

// Add adds a user to the queue
func (q *Queue) Add(username string, isMod bool) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return fmt.Errorf("queue system is currently disabled")
	}

	lowerName := strings.ToLower(username)
	// Check if user is already in queue
	for _, user := range q.users {
		if user.LowerName == lowerName {
			return fmt.Errorf("user is already in queue")
		}
	}

	q.users = append(q.users, QueuedUser{
		Username:  username,  // Preserve original display name
		LowerName: lowerName, // Store lowercase for matching
		JoinTime:  time.Now(),
		IsMod:     isMod,
	})

	// Save state after adding user
	if err := q.saveState(); err != nil {
		fmt.Printf("Warning: Failed to save queue state after adding user: %v\n", err)
	}
	return nil
}

// Remove removes a user from the queue
func (q *Queue) Remove(username string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	lowerName := strings.ToLower(username)
	for i, user := range q.users {
		if user.LowerName == lowerName {
			// Remove user by slicing
			q.users = append(q.users[:i], q.users[i+1:]...)

			// Save state after removing user
			if err := q.saveState(); err != nil {
				fmt.Printf("Warning: Failed to save queue state after removing user: %v\n", err)
			}
			return true
		}
	}
	return false
}

// List returns the current queue
func (q *Queue) List() []QueuedUser {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Return a copy to prevent external modifications
	users := make([]QueuedUser, len(q.users))
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

	lowerName := strings.ToLower(username)
	for i, user := range q.users {
		if user.LowerName == lowerName {
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

	lowerName := strings.ToLower(username)
	// Check if user is already in queue
	for _, user := range q.users {
		if user.LowerName == lowerName {
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

	// Create new user
	newUser := QueuedUser{
		Username:  username,  // Preserve original display name
		LowerName: lowerName, // Store lowercase for matching
		JoinTime:  time.Now(),
		IsMod:     isMod,
	}

	// Insert at position (converting from 1-based to 0-based index)
	position--
	if position == len(q.users) {
		// Append to end
		q.users = append(q.users, newUser)
	} else {
		// Insert at position
		q.users = append(q.users[:position], append([]QueuedUser{newUser}, q.users[position:]...)...)
	}

	// Save state after adding user at position
	if err := q.saveState(); err != nil {
		fmt.Printf("Warning: Failed to save queue state after adding user at position: %v\n", err)
	}
	return nil
}

// Pop removes and returns the first user from the queue
func (q *Queue) Pop() (*QueuedUser, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.enabled {
		return nil, fmt.Errorf("queue system is currently disabled")
	}

	if len(q.users) == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	// Get first user and remove them
	user := q.users[0]
	q.users = q.users[1:]

	// Save state after popping user
	if err := q.saveState(); err != nil {
		fmt.Printf("Warning: Failed to save queue state after popping user: %v\n", err)
	}
	return &user, nil
}
