package channel

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// StreamSession represents a single streaming session
type StreamSession struct {
	StartTime      time.Time      `json:"start_time"`
	EndTime        time.Time      `json:"end_time"`
	Duration       time.Duration  `json:"duration"`
	Game           string         `json:"game"`
	Title          string         `json:"title"`
	Viewers        int            `json:"viewers"`
	PeakViewers    int            `json:"peak_viewers"`
	AverageViewers float64        `json:"average_viewers"`
	ChatMessages   int            `json:"chat_messages"`
	UniqueChatters int            `json:"unique_chatters"`
	ChatterCounts  map[string]int `json:"chatter_counts"` // username -> message count
	SessionID      string         `json:"session_id"`     // Unique identifier for the session
}

// ChannelStats tracks overall channel statistics
type ChannelStats struct {
	mu sync.RWMutex

	// Current session
	CurrentSession *StreamSession `json:"current_session"`

	// Historical data
	Sessions []StreamSession `json:"sessions"`

	// Overall stats
	TotalStreamTime   time.Duration  `json:"total_stream_time"`
	TotalSessions     int            `json:"total_sessions"`
	MaxViewers        int            `json:"max_viewers"`
	AverageViewers    float64        `json:"average_viewers"`
	TotalChatMessages int            `json:"total_chat_messages"`
	UniqueChatters    int            `json:"unique_chatters"`
	ChatterTotals     map[string]int `json:"chatter_totals"`   // username -> total messages
	LastSessionEnd    time.Time      `json:"last_session_end"` // When the last session ended

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

	// Check if we can resume the previous session
	if s.canResumePreviousSession(game, title) {
		// Resume the previous session
		s.CurrentSession = &StreamSession{
			StartTime:      s.Sessions[len(s.Sessions)-1].StartTime, // Keep original start time
			Game:           game,
			Title:          title,
			Viewers:        viewers,
			PeakViewers:    s.Sessions[len(s.Sessions)-1].PeakViewers,
			AverageViewers: s.Sessions[len(s.Sessions)-1].AverageViewers,
			ChatMessages:   s.Sessions[len(s.Sessions)-1].ChatMessages,
			UniqueChatters: s.Sessions[len(s.Sessions)-1].UniqueChatters,
			ChatterCounts:  s.Sessions[len(s.Sessions)-1].ChatterCounts,
			SessionID:      s.Sessions[len(s.Sessions)-1].SessionID,
		}
		// Remove the previous session from history since we're resuming it
		s.Sessions = s.Sessions[:len(s.Sessions)-1]
		return
	}

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
		ChatterCounts:  make(map[string]int),
		SessionID:      generateSessionID(),
	}
}

// canResumePreviousSession checks if we can resume the previous session
func (s *ChannelStats) canResumePreviousSession(game, title string) bool {
	if len(s.Sessions) == 0 {
		return false
	}

	lastSession := s.Sessions[len(s.Sessions)-1]
	timeSinceEnd := time.Since(s.LastSessionEnd)

	// Can resume if:
	// 1. Less than 30 minutes since last session ended
	// 2. Same game and title
	// 3. Last session wasn't too long ago (e.g., within last 24 hours)
	return timeSinceEnd < 30*time.Minute &&
		lastSession.Game == game &&
		lastSession.Title == title &&
		time.Since(lastSession.StartTime) < 24*time.Hour
}

// generateSessionID creates a unique session identifier
func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
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

// RecordChatMessage records a chat message from a user
func (s *ChannelStats) RecordChatMessage(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentSession == nil {
		return
	}

	// Update session chatter counts
	s.CurrentSession.ChatMessages++
	s.CurrentSession.ChatterCounts[username]++
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

	if s.ChatterTotals == nil {
		s.ChatterTotals = make(map[string]int)
	}
	for user, count := range s.CurrentSession.ChatterCounts {
		s.ChatterTotals[user] += count
	}

	// Update unique chatters
	unique := make(map[string]struct{})
	for _, session := range s.Sessions {
		for user := range session.ChatterCounts {
			unique[user] = struct{}{}
		}
	}
	s.UniqueChatters = len(unique)

	if s.CurrentSession.PeakViewers > s.MaxViewers {
		s.MaxViewers = s.CurrentSession.PeakViewers
	}

	// Update average viewers
	totalViewerTime := 0.0
	for _, session := range s.Sessions {
		totalViewerTime += session.AverageViewers * session.Duration.Seconds()
	}
	s.AverageViewers = totalViewerTime / s.TotalStreamTime.Seconds()

	// Save the end time of this session
	s.LastSessionEnd = s.CurrentSession.EndTime

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
		ChatterTotals:     make(map[string]int),
		statsPath:         s.statsPath,
	}

	// Copy sessions
	copy(stats.Sessions, s.Sessions)

	// Copy chatter totals
	for user, count := range s.ChatterTotals {
		stats.ChatterTotals[user] = count
	}

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

// GetTopChatters returns the top N chatters by message count
func (s *ChannelStats) GetTopChatters(n int) []struct {
	User  string
	Count int
} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type pair struct {
		User  string
		Count int
	}
	var chatters []pair
	for user, count := range s.ChatterTotals {
		chatters = append(chatters, pair{user, count})
	}
	// Sort descending
	sort.Slice(chatters, func(i, j int) bool { return chatters[i].Count > chatters[j].Count })
	if n > len(chatters) {
		n = len(chatters)
	}
	result := make([]struct {
		User  string
		Count int
	}, n)
	for i := 0; i < n; i++ {
		result[i] = struct {
			User  string
			Count int
		}{chatters[i].User, chatters[i].Count}
	}
	return result
}
