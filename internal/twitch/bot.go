package twitch

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

// Bot represents a Twitch chat bot
type Bot struct {
	channel         string
	authManager     *AuthManager
	client          *twitch.Client
	commandHandlers []func(twitch.PrivateMessage) string
	secretsPath     string
	botUsername     string
}

// NewBot creates a new Twitch bot instance
func NewBot(channel string, authManager *AuthManager, secretsPath string, botUsername string) *Bot {
	return &Bot{
		channel:     channel,
		authManager: authManager,
		secretsPath: secretsPath,
		botUsername: botUsername,
	}
}

// Connect establishes a connection to Twitch IRC
func (b *Bot) Connect(ctx context.Context) error {
	// Force initial token refresh
	log.Printf("[Token Refresh] Performing initial token refresh...")
	if err := b.authManager.RefreshToken(); err != nil {
		return fmt.Errorf("error performing initial token refresh: %w", err)
	}
	log.Printf("[Token Refresh] Initial token refresh successful. Expires at: %s", b.authManager.ExpiresAt.Format(time.RFC3339))

	// Get initial access token
	token, err := b.authManager.GetAccessToken()
	if err != nil {
		return fmt.Errorf("error getting initial access token: %w", err)
	}

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
			b.client.SetIRCToken(newToken)
		}

		// Handle commands
		for _, handler := range b.commandHandlers {
			if response := handler(message); response != "" {
				b.client.Say(message.Channel, response)
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
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			timeUntilExpiry := time.Until(b.authManager.ExpiresAt)
			log.Printf("[Token Check] Checking token validity. Current expiry: %s (expires in %s)",
				b.authManager.ExpiresAt.Format(time.RFC3339),
				timeUntilExpiry.Round(time.Second))
			if !b.authManager.IsTokenValid() {
				log.Printf("[Token Refresh] Attempting to refresh Twitch access token...")
				newToken, err := b.authManager.GetAccessToken()
				if err != nil {
					log.Printf("Error refreshing token: %v", err)
					continue
				}
				b.client.SetIRCToken(newToken)
				log.Printf("[Token Refresh] Token refreshed. New expiry: %s", b.authManager.ExpiresAt.Format(time.RFC3339))
			} else {
				log.Printf("[Token Check] Token is still valid")
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
