package channel

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StreamSession represents a single streaming session
type StreamSession struct {
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Duration       time.Duration `json:"duration"`
	Game           string        `json:"game"`
	Title          string        `json:"title"`
	Viewers        int           `json:"viewers"`
	PeakViewers    int           `json:"peak_viewers"`
	AverageViewers float64       `json:"average_viewers"`
	ChatMessages   int           `json:"chat_messages"`
	UniqueChatters int           `json:"unique_chatters"`
}

// ChannelStats tracks overall channel statistics
type ChannelStats struct {
	mu sync.RWMutex

	// Current session
	CurrentSession *StreamSession `json:"current_session"`

	// Historical data
	Sessions []StreamSession `json:"sessions"`

	// Overall stats
	TotalStreamTime   time.Duration `json:"total_stream_time"`
	TotalSessions     int           `json:"total_sessions"`
	MaxViewers        int           `json:"max_viewers"`
	AverageViewers    float64       `json:"average_viewers"`
	TotalChatMessages int           `json:"total_chat_messages"`
	UniqueChatters    int           `json:"unique_chatters"`

	// File paths
	statsPath string
}

// NewChannelStats creates a new ChannelStats instance
func NewChannelStats(dataPath string) *ChannelStats {
	stats := &ChannelStats{
		statsPath: filepath.Join(dataPath, "channel_stats.json"),
	}

	// Load existing stats if available
	if err := stats.Load(); err != nil {
		log.Printf("Warning: Could not load existing channel stats: %v", err)
	}

	return stats
}

// StartSession starts tracking a new stream session
func (s *ChannelStats) StartSession(game, title string, viewers int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// End any existing session
	if s.CurrentSession != nil {
		s.endCurrentSession()
	}

	// Create new session
	s.CurrentSession = &StreamSession{
		StartTime:      time.Now(),
		Game:           game,
		Title:          title,
		Viewers:        viewers,
		PeakViewers:    viewers,
		AverageViewers: float64(viewers),
	}
}

// UpdateSession updates the current session with new data
func (s *ChannelStats) UpdateSession(game, title string, viewers int, chatMessages int, uniqueChatters int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentSession == nil {
		return
	}

	// Update session data
	s.CurrentSession.Game = game
	s.CurrentSession.Title = title
	s.CurrentSession.Viewers = viewers
	s.CurrentSession.ChatMessages = chatMessages
	s.CurrentSession.UniqueChatters = uniqueChatters

	// Update peak viewers
	if viewers > s.CurrentSession.PeakViewers {
		s.CurrentSession.PeakViewers = viewers
	}

	// Update average viewers
	duration := time.Since(s.CurrentSession.StartTime)
	s.CurrentSession.AverageViewers = (s.CurrentSession.AverageViewers*duration.Seconds() + float64(viewers)) / (duration.Seconds() + 1)
}

// EndSession ends the current stream session
func (s *ChannelStats) EndSession() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentSession == nil {
		return
	}

	s.endCurrentSession()
}

// endCurrentSession ends the current session and saves it to history
func (s *ChannelStats) endCurrentSession() {
	if s.CurrentSession == nil {
		return
	}

	// Set end time and calculate duration
	s.CurrentSession.EndTime = time.Now()
	s.CurrentSession.Duration = s.CurrentSession.EndTime.Sub(s.CurrentSession.StartTime)

	// Add to sessions history
	s.Sessions = append(s.Sessions, *s.CurrentSession)

	// Update overall stats
	s.TotalStreamTime += s.CurrentSession.Duration
	s.TotalSessions++
	s.TotalChatMessages += s.CurrentSession.ChatMessages
	s.UniqueChatters += s.CurrentSession.UniqueChatters

	if s.CurrentSession.PeakViewers > s.MaxViewers {
		s.MaxViewers = s.CurrentSession.PeakViewers
	}

	// Update average viewers
	totalViewerTime := 0.0
	for _, session := range s.Sessions {
		totalViewerTime += session.AverageViewers * session.Duration.Seconds()
	}
	s.AverageViewers = totalViewerTime / s.TotalStreamTime.Seconds()

	// Save stats
	if err := s.Save(); err != nil {
		log.Printf("Error saving channel stats: %v", err)
	}

	// Clear current session
	s.CurrentSession = nil
}

// GetStats returns a copy of the current stats
func (s *ChannelStats) GetStats() *ChannelStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a deep copy
	stats := &ChannelStats{
		CurrentSession:    s.CurrentSession,
		Sessions:          make([]StreamSession, len(s.Sessions)),
		TotalStreamTime:   s.TotalStreamTime,
		TotalSessions:     s.TotalSessions,
		MaxViewers:        s.MaxViewers,
		AverageViewers:    s.AverageViewers,
		TotalChatMessages: s.TotalChatMessages,
		UniqueChatters:    s.UniqueChatters,
		statsPath:         s.statsPath,
	}

	// Copy sessions
	copy(stats.Sessions, s.Sessions)

	return stats
}

// GetStatsForPeriod returns stats for a specific time period
func (s *ChannelStats) GetStatsForPeriod(start, end time.Time) *ChannelStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &ChannelStats{
		statsPath: s.statsPath,
	}

	// Filter sessions within the period
	for _, session := range s.Sessions {
		if session.StartTime.After(start) && session.EndTime.Before(end) {
			stats.Sessions = append(stats.Sessions, session)
			stats.TotalStreamTime += session.Duration
			stats.TotalSessions++
			stats.TotalChatMessages += session.ChatMessages
			stats.UniqueChatters += session.UniqueChatters

			if session.PeakViewers > stats.MaxViewers {
				stats.MaxViewers = session.PeakViewers
			}
		}
	}

	// Calculate average viewers
	if stats.TotalStreamTime > 0 {
		totalViewerTime := 0.0
		for _, session := range stats.Sessions {
			totalViewerTime += session.AverageViewers * session.Duration.Seconds()
		}
		stats.AverageViewers = totalViewerTime / stats.TotalStreamTime.Seconds()
	}

	return stats
}

// GetLastWeekStats returns stats for the last 7 days
func (s *ChannelStats) GetLastWeekStats() *ChannelStats {
	end := time.Now()
	start := end.AddDate(0, 0, -7)
	return s.GetStatsForPeriod(start, end)
}

// GetLastMonthStats returns stats for the last 30 days
func (s *ChannelStats) GetLastMonthStats() *ChannelStats {
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	return s.GetStatsForPeriod(start, end)
}

// Save saves the stats to disk
func (s *ChannelStats) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling stats: %w", err)
	}

	if err := os.WriteFile(s.statsPath, data, 0644); err != nil {
		return fmt.Errorf("error writing stats file: %w", err)
	}

	return nil
}

// Load loads the stats from disk
func (s *ChannelStats) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.statsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's okay
		}
		return fmt.Errorf("error reading stats file: %w", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("error unmarshaling stats: %w", err)
	}

	return nil
}
