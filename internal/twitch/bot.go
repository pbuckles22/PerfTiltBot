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
	"github.com/pbuckles22/PBChatBot/internal/utils"
)

// formatTime formats a time in the channel's configured timezone and prints the correct timezone abbreviation
func (b *Bot) formatTime(t time.Time) string {
	return utils.FormatTimeForDisplay(t, b.cfg.Timezone)
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
	channelConfigPath := fmt.Sprintf("configs/channels/%s_config_secrets.yaml", channel)
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
	log.Printf("[Token] Startup: expires in %s", timeUntilExpiry.Round(time.Second))

	// Calculate initial check interval based on time until expiry
	checkInterval := calculateCheckInterval(timeUntilExpiry)

	log.Printf("[Token] First check in %s", checkInterval.Round(time.Second))

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

	// Start token refresh goroutine (only once, not on every connect)
	go b.refreshTokenLoop(ctx)

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

	return nil
}

// refreshTokenLoop periodically checks and refreshes the token
func (b *Bot) refreshTokenLoop(ctx context.Context) {
	// Calculate initial check interval based on time until expiry
	timeUntilExpiry := time.Until(b.authManager.ExpiresAt)
	checkInterval := calculateCheckInterval(timeUntilExpiry)
	nextCheckTime := time.Now().Add(checkInterval)

	log.Printf("[Token Refresh Loop] Starting refresh loop. First check at: %s (in %s)",
		formatTimeForLogs(nextCheckTime),
		checkInterval.Round(time.Second))

	// Ensure positive interval for initial ticker
	if checkInterval <= 0 {
		log.Printf("[Token Refresh Loop] WARNING: Initial calculated interval is %v, using 1 second instead", checkInterval)
		checkInterval = 1 * time.Second
	}

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
			// Calculate time until expiry
			timeUntilExpiry := time.Until(b.authManager.ExpiresAt)

			// Calculate next check interval and time
			checkInterval = calculateCheckInterval(timeUntilExpiry)

			// Only refresh if we're within minimum time of expiry
			if timeUntilExpiry <= minRefreshTime {
				log.Printf("[Token] Refreshing (expires in %s)", timeUntilExpiry.Round(time.Second))

				// Store the old expiry time for comparison
				oldExpiry := b.authManager.ExpiresAt

				if err := b.authManager.RefreshToken(); err != nil {
					log.Printf("Error refreshing token: %v", err)
					continue
				}
				b.client.SetIRCToken("oauth:" + b.authManager.AccessToken)

				// Calculate new check interval based on new token expiry
				timeUntilExpiry = time.Until(b.authManager.ExpiresAt)
				checkInterval = calculateCheckInterval(timeUntilExpiry)

				log.Printf("[Token] Refreshed: %s -> %s (next check in %s)",
					formatTimeForLogs(oldExpiry),
					formatTimeForLogs(b.authManager.ExpiresAt),
					checkInterval.Round(time.Second))

				// Reset ticker with new interval (ensure positive interval)
				if checkInterval <= 0 {
					log.Printf("[Token Refresh Loop] WARNING: Calculated interval is %v, using 1 second instead", checkInterval)
					checkInterval = 1 * time.Second
				}
				ticker.Reset(checkInterval)

				log.Printf("") // Blank line after refresh operation
			} else {
				log.Printf("[Token] Valid (expires in %s, next check in %s)",
					timeUntilExpiry.Round(time.Second),
					checkInterval.Round(time.Second))

				// Reset ticker with new interval (ensure positive interval)
				if checkInterval <= 0 {
					log.Printf("[Token Refresh Loop] WARNING: Calculated interval is %v, using 1 second instead", checkInterval)
					checkInterval = 1 * time.Second
				}
				ticker.Reset(checkInterval)

				log.Printf("") // Blank line after valid check
			}
		}
	}
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
