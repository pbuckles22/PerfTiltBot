package queue

import (
	"fmt"
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
	mu      sync.RWMutex
}

// NewQueue creates a new queue manager
func NewQueue() *Queue {
	return &Queue{
		users:   make([]QueuedUser, 0),
		enabled: false,
	}
}

// Enable starts the queue system
func (q *Queue) Enable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = true
	q.users = make([]QueuedUser, 0) // Clear queue when enabling
}

// Disable stops the queue system and clears the queue
func (q *Queue) Disable() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enabled = false
	q.users = make([]QueuedUser, 0)
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
