package twitch

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PerfTiltBot/internal/config"
)

// formatTime formats a time in the channel's configured timezone
func (b *Bot) formatTime(t time.Time) string {
	loc, err := time.LoadLocation(b.cfg.Twitch.Timezone)
	if err != nil {
		log.Printf("Error loading timezone %s: %v, falling back to America/New_York", b.cfg.Twitch.Timezone, err)
		loc, _ = time.LoadLocation("America/New_York")
	}
	return t.In(loc).Format("2006-01-02 15:04:05 ET")
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
}

// NewBot creates a new Twitch bot instance
func NewBot(channel string, authManager *AuthManager, secretsPath string, botUsername string) *Bot {
	// Load the channel's config
	cfg, err := config.Load(secretsPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		cfg = &config.Config{}
		cfg.Twitch.Timezone = "America/New_York" // Default timezone if config fails to load
	}

	return &Bot{
		channel:     channel,
		authManager: authManager,
		secretsPath: secretsPath,
		botUsername: botUsername,
		startTime:   time.Now(),
		cfg:         cfg,
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

	// Start connection in a goroutine
	go func() {
		if err := b.client.Connect(); err != nil {
			log.Printf("Error connecting to Twitch IRC: %v", err)
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

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			timeUntilExpiry := time.Until(b.authManager.ExpiresAt)
			log.Printf("[Token Check] Checking token validity. Current expiry: %s (expires in %s)",
				b.formatTime(b.authManager.ExpiresAt),
				timeUntilExpiry.Round(time.Second))

			// Only refresh if we're within 5 minutes of expiry
			if timeUntilExpiry <= 5*time.Minute {
				log.Printf("[Token Refresh] Token expires in %s, refreshing...", timeUntilExpiry.Round(time.Second))
				newToken, err := b.authManager.GetAccessToken()
				if err != nil {
					log.Printf("Error refreshing token: %v", err)
					continue
				}
				b.client.SetIRCToken("oauth:" + newToken)
				log.Printf("[Token Refresh] Token refreshed. New expiry: %s",
					b.formatTime(b.authManager.ExpiresAt))

				// Update check interval for next check
				timeUntilExpiry = time.Until(b.authManager.ExpiresAt)
				checkInterval = calculateCheckInterval(timeUntilExpiry)
				ticker.Reset(checkInterval)
			} else {
				log.Printf("[Token Check] Token is still valid (expires in %s)", timeUntilExpiry.Round(time.Second))
			}
		}
	}
}

// calculateCheckInterval determines how often to check token validity
// based on the remaining time until expiry
func calculateCheckInterval(timeUntilExpiry time.Duration) time.Duration {
	// If less than 5 minutes, refresh token
	if timeUntilExpiry <= 5*time.Minute {
		return 0 // Will trigger immediate refresh
	}

	// If less than 10 minutes, check every 3 minutes
	if timeUntilExpiry <= 10*time.Minute {
		return 3 * time.Minute
	}

	// If less than 20 minutes, check every 5 minutes
	if timeUntilExpiry <= 20*time.Minute {
		return 5 * time.Minute
	}

	// If less than 30 minutes, check every 7 minutes
	if timeUntilExpiry <= 30*time.Minute {
		return 7 * time.Minute
	}

	// If less than 1 hour, check every 10 minutes
	if timeUntilExpiry <= time.Hour {
		return 10 * time.Minute
	}

	// Otherwise, check every 30 minutes
	return 30 * time.Minute
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
