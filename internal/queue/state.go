package queue

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// QueueState represents the serializable state of the queue
type QueueState struct {
	Enabled bool        `json:"enabled"`
	Users   []UserState `json:"users"`
}

// UserState represents a serializable user in the queue
type UserState struct {
	Username string    `json:"username"`
	JoinTime time.Time `json:"join_time"`
	IsMod    bool      `json:"is_mod"`
}

// SaveState saves the current queue state to a file
func (q *Queue) SaveState(filename string) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// Create state object
	state := QueueState{
		Enabled: q.enabled,
		Users:   make([]UserState, len(q.users)),
	}

	// Copy user data
	for i, user := range q.users {
		state.Users[i] = UserState{
			Username: user.Username,
			JoinTime: user.JoinTime,
			IsMod:    user.IsMod,
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filename, data, 0644)
}

// LoadState loads the queue state from a file
func (q *Queue) LoadState(filename string) error {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty queue
			return nil
		}
		return err
	}

	// Unmarshal JSON
	var state QueueState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// Update queue state
	q.mu.Lock()
	defer q.mu.Unlock()

	q.enabled = state.Enabled
	q.users = make([]QueuedUser, len(state.Users))

	// Copy user data
	for i, user := range state.Users {
		q.users[i] = QueuedUser{
			Username: user.Username,
			JoinTime: user.JoinTime,
			IsMod:    user.IsMod,
		}
	}

	return nil
}
