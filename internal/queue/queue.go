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
	users []QueuedUser
	mu    sync.RWMutex
}

// NewQueue creates a new queue manager
func NewQueue() *Queue {
	return &Queue{
		users: make([]QueuedUser, 0),
	}
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
