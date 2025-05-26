package twitch

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	channelstats "github.com/pbuckles22/PBChatBot/internal/channel"
	"github.com/pbuckles22/PBChatBot/internal/config"
)

// Constants for token refresh
const (
	tokenRefreshPercentage = 25               // Check at 25% of remaining time
	minRefreshTime         = 15 * time.Minute // Minimum time before expiry to refresh
)

// formatTime formats a time in the channel's configured timezone and prints the correct timezone abbreviation
func (b *Bot) formatTime(t time.Time) string {
	loc, err := time.LoadLocation("America/New_York") // Default timezone
	if err != nil {
		log.Printf("Error loading timezone: %v, falling back to America/New_York", err)
	}
	tzTime := t.In(loc)
	return tzTime.Format("2006-01-02 15:04:05 MST")
}

// Bot represents a Twitch chat bot
type Bot struct {
	channel         string
	authManager     *AuthManager
	client          *twitch.Client
	commandHandlers []func(twitch.PrivateMessage) string
	secretsPath     string
	botUsername     string
	startTime       time.Time
	cfg             *config.Config
	channelStats    *channelstats.ChannelStats
}

// NewBot creates a new Twitch bot instance
func NewBot(channel string, authManager *AuthManager, secretsPath string, botUsername string) *Bot {
	// Load the channel's config
	channelConfigPath := fmt.Sprintf("configs/%s_config_secrets.yaml", channel)
	cfg, err := config.Load(channelConfigPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		cfg = &config.Config{
			Channel:  channel,
			DataPath: fmt.Sprintf("/app/data/%s", channel),
		}
	}

	// Initialize channel stats using the same data path as the queue
	channelStats := channelstats.NewChannelStats(cfg.DataPath)

	return &Bot{
		channel:      channel,
		authManager:  authManager,
		secretsPath:  secretsPath,
		botUsername:  botUsername,
		startTime:    time.Now(),
		cfg:          cfg,
		channelStats: channelStats,
	}
}

// Connect establishes a connection to Twitch IRC
func (b *Bot) Connect(ctx context.Context) error {
	// Get initial access token, refreshing only if needed
	token, err := b.authManager.GetAccessToken()
	if err != nil {
		return fmt.Errorf("error getting initial access token: %w", err)
	}

	// Log token validity and expiry at startup
	timeUntilExpiry := time.Until(b.authManager.ExpiresAt)
	log.Printf("[Startup Token Check] Token expiry: %s (expires in %s)",
		b.formatTime(b.authManager.ExpiresAt),
		timeUntilExpiry.Round(time.Second))

	// Calculate and show the complete check schedule
	remainingTime := timeUntilExpiry
	currentTime := time.Now()
	checkNumber := 1

	log.Printf("[Token Check Schedule]")
	for remainingTime > minRefreshTime {
		checkInterval := remainingTime * tokenRefreshPercentage / 100
		nextCheckTime := currentTime.Add(checkInterval)
		timeAfterCheck := remainingTime - checkInterval
		log.Printf("  Check %d: %s (in %s, %s remaining after check)",
			checkNumber,
			b.formatTime(nextCheckTime),
			checkInterval.Round(time.Second),
			timeAfterCheck.Round(time.Second))

		remainingTime -= checkInterval
		currentTime = nextCheckTime
		checkNumber++
	}
	log.Printf("  Final check: %s (at %s remaining)",
		b.formatTime(b.authManager.ExpiresAt.Add(-minRefreshTime)),
		minRefreshTime)

	// Create Twitch client with bot username and new token
	b.client = twitch.NewClient(b.botUsername, "oauth:"+token)

	// Set up connection handler
	b.client.OnConnect(func() {
		log.Printf("Successfully connected to Twitch IRC")
		log.Printf("Joining channel: %s", b.channel)
		b.client.Join(b.channel)
	})

	// Set up message handler
	b.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// Record chatter stats
		b.channelStats.RecordChatMessage(message.User.Name)
		// Check if token needs refresh
		if !b.authManager.IsTokenValid() {
			newToken, err := b.authManager.GetAccessToken()
			if err != nil {
				log.Printf("Error refreshing token: %v", err)
				return
			}
			b.client.SetIRCToken("oauth:" + newToken)
		}

		// Handle commands
		for _, handler := range b.commandHandlers {
			if response := handler(message); response != "" {
				// Check if response is a whisper command
				if strings.HasPrefix(response, "/w ") {
					// Extract the whisper command parts
					parts := strings.SplitN(response, " ", 3)
					if len(parts) == 3 {
						b.client.Say(message.Channel, fmt.Sprintf("/w %s %s", parts[1], parts[2]))
					}
				} else {
					b.client.Say(message.Channel, response)
				}
				break
			}
		}
	})

	// Start connection in a goroutine with reconnection logic
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := b.client.Connect(); err != nil {
					log.Printf("Error connecting to Twitch IRC: %v", err)
					log.Printf("Attempting to reconnect in 30 seconds...")
					time.Sleep(30 * time.Second)
					continue
				}
				return
			}
		}
	}()

	// Start token refresh goroutine
	go b.refreshTokenLoop(ctx)

	return nil
}

// refreshTokenLoop periodically checks and refreshes the token
func (b *Bot) refreshTokenLoop(ctx context.Context) {
	// Calculate initial check interval based on time until expiry
	timeUntilExpiry := time.Until(b.authManager.ExpiresAt)
	checkInterval := calculateCheckInterval(timeUntilExpiry)
	nextCheckTime := time.Now().Add(checkInterval)

	log.Printf("[Token Refresh Loop] Starting refresh loop. First check at: %s (in %s)",
		b.formatTime(nextCheckTime),
		checkInterval.Round(time.Second))

	ticker := time.NewTicker(checkInterval)
	defer func() {
		ticker.Stop()
		log.Printf("[Token Refresh Loop] Ticker stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Token Refresh Loop] Context done, stopping refresh loop")
			return
		case <-ticker.C:
			log.Printf("[Token Refresh Loop] Ticker fired at %s", b.formatTime(time.Now()))

			// Get current time and calculate time until expiry
			now := time.Now()
			timeUntilExpiry := time.Until(b.authManager.ExpiresAt)

			// Calculate next check interval and time
			checkInterval = calculateCheckInterval(timeUntilExpiry)
			nextCheckTime = now.Add(checkInterval)

			log.Printf("[Token Check] Checking token validity. Current expiry: %s (expires in %s). Next check at: %s",
				b.formatTime(b.authManager.ExpiresAt),
				timeUntilExpiry.Round(time.Second),
				b.formatTime(nextCheckTime))

			// Only refresh if we're within minimum time of expiry
			if timeUntilExpiry <= minRefreshTime {
				log.Printf("[Token Refresh] Token expires in %s, refreshing...", timeUntilExpiry.Round(time.Second))
				newToken, err := b.authManager.GetAccessToken()
				if err != nil {
					log.Printf("Error refreshing token: %v", err)
					continue
				}
				b.client.SetIRCToken("oauth:" + newToken)

				// Calculate new check interval based on new token expiry
				timeUntilExpiry = time.Until(b.authManager.ExpiresAt)
				checkInterval = calculateCheckInterval(timeUntilExpiry)
				nextCheckTime = now.Add(checkInterval)

				log.Printf("[Token Refresh] Token refreshed. New expiry: %s. Starting new check cycle:",
					b.formatTime(b.authManager.ExpiresAt))
				log.Printf("  - Next check at: %s (in %s)",
					b.formatTime(nextCheckTime),
					checkInterval.Round(time.Second))

				// Reset ticker with new interval
				ticker.Reset(checkInterval)
				log.Printf("[Token Refresh Loop] Ticker reset for next check at: %s", b.formatTime(nextCheckTime))

				// Log the complete new check schedule
				remainingTime := timeUntilExpiry
				currentTime := now
				checkNumber := 1

				log.Printf("[New Token Check Schedule]")
				for remainingTime > minRefreshTime {
					checkInterval := remainingTime * tokenRefreshPercentage / 100
					nextCheckTime := currentTime.Add(checkInterval)
					timeAfterCheck := remainingTime - checkInterval
					log.Printf("  Check %d: %s (in %s, %s remaining after check)",
						checkNumber,
						b.formatTime(nextCheckTime),
						checkInterval.Round(time.Second),
						timeAfterCheck.Round(time.Second))

					remainingTime -= checkInterval
					currentTime = nextCheckTime
					checkNumber++
				}
				log.Printf("  Final check: %s (at %s remaining)",
					b.formatTime(b.authManager.ExpiresAt.Add(-minRefreshTime)),
					minRefreshTime)
			} else {
				log.Printf("[Token Check] Token is still valid (expires in %s). Next check at: %s",
					timeUntilExpiry.Round(time.Second),
					b.formatTime(nextCheckTime))
			}
		}
	}
}

// calculateCheckInterval determines how often to check token validity
// based on the remaining time until expiry
func calculateCheckInterval(timeUntilExpiry time.Duration) time.Duration {
	// If less than minimum time, refresh token immediately
	if timeUntilExpiry <= minRefreshTime {
		return 0 // Will trigger immediate refresh
	}

	// Calculate next check time as percentage of the current remaining time
	// For example, with 4 hours and 25%:
	// First check: 4h - (4h * 0.25) = 3h remaining
	// Next check: 3h - (3h * 0.25) = 2h15m remaining
	// Next check: 2h15m - (2h15m * 0.25) = 1h41m15s remaining
	// And so on until we hit minimum time
	return timeUntilExpiry * tokenRefreshPercentage / 100
}

// RegisterCommandHandler adds a new command handler
func (b *Bot) RegisterCommandHandler(handler func(twitch.PrivateMessage) string) {
	b.commandHandlers = append(b.commandHandlers, handler)
}

// IsCommand checks if a message is a command
func (b *Bot) IsCommand(message string) bool {
	return strings.HasPrefix(message, "!")
}

// GetCommandName extracts the command name from a message
func (b *Bot) GetCommandName(message string) string {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimPrefix(parts[0], "!")
}

// GetCommandArgs extracts the command arguments from a message
func (b *Bot) GetCommandArgs(message string) []string {
	parts := strings.Fields(message)
	if len(parts) <= 1 {
		return nil
	}
	return parts[1:]
}
