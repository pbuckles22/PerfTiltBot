package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// QueuedUser represents a user in the queue
type QueuedUser struct {
	Username string
	JoinTime time.Time
	IsMod    bool
}

// Queue manages the user queue
type Queue struct {
	users   []QueuedUser
	enabled bool
	paused  bool
	mu      sync.RWMutex
}

// QueueState represents the serializable state of the queue
type QueueState struct {
	Enabled bool         `json:"enabled"`
	Paused  bool         `json:"paused"`
	Users   []QueuedUser `json:"users"`
}

// NewQueue creates a new queue manager
func NewQueue() *Queue {
	return &Queue{
		users:   make([]QueuedUser, 0),
		enabled: false,
		paused:  false,
	}
}

// Enable starts the queue system
func (q *Queue) Enable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = true
	q.paused = false
	q.users = make([]QueuedUser, 0) // Clear queue when enabling
}

// Disable stops the queue system and clears the queue
func (q *Queue) Disable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = false
	q.paused = false
	q.users = make([]QueuedUser, 0)
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
	q.users = make([]QueuedUser, 0)
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

	// Check if user is already in queue
	for _, user := range q.users {
		if user.Username == username {
			return fmt.Errorf("user is already in queue")
		}
	}

	q.users = append(q.users, QueuedUser{
		Username: username,
		JoinTime: time.Now(),
		IsMod:    isMod,
	})
	return nil
}

// Remove removes a user from the queue
func (q *Queue) Remove(username string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, user := range q.users {
		if user.Username == username {
			// Remove user by slicing
			q.users = append(q.users[:i], q.users[i+1:]...)
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

	for i, user := range q.users {
		if user.Username == username {
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
		if user.Username == username {
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
		Username: username,
		JoinTime: time.Now(),
		IsMod:    isMod,
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

	// Get first user
	user := q.users[0]

	// Remove first user
	q.users = q.users[1:]

	return &user, nil
}

// PopN removes and returns the first N users from the queue
func (q *Queue) PopN(count int) ([]QueuedUser, error) {
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
	users := make([]QueuedUser, count)
	copy(users, q.users[:count])

	// Remove first N users
	q.users = q.users[count:]

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
		if user.Username == username {
			// Remove the user from the queue
			q.users = append(q.users[:i], q.users[i+1:]...)
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
		if user.Username == username {
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
	q.users = append(q.users[:position], append([]QueuedUser{user}, q.users[position:]...)...)

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
		if user.Username == username {
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

	return nil
}

// SaveState saves the current queue state to a file
func (q *Queue) SaveState(filename string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	state := QueueState{
		Enabled: q.enabled,
		Paused:  q.paused,
		Users:   make([]QueuedUser, len(q.users)),
	}
	copy(state.Users, q.users)

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal queue state: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write queue state: %w", err)
	}

	return nil
}

// LoadState loads the queue state from a file
func (q *Queue) LoadState(filename string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no saved queue state found")
		}
		return fmt.Errorf("failed to read queue state: %w", err)
	}

	var state QueueState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal queue state: %w", err)
	}

	q.enabled = state.Enabled
	q.paused = state.Paused
	q.users = make([]QueuedUser, len(state.Users))
	copy(q.users, state.Users)

	return nil
}
