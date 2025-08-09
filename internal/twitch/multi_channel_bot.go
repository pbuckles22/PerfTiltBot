package twitch

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	channelstats "github.com/pbuckles22/PBChatBot/internal/channel"
	"github.com/pbuckles22/PBChatBot/internal/commands"
	"github.com/pbuckles22/PBChatBot/internal/config"
)

// ChannelBot represents a single channel's bot instance
type ChannelBot struct {
	channel        string
	client         *twitchirc.Client
	commandManager *commands.CommandManager
	channelStats   *channelstats.ChannelStats
	cfg            *config.Config
	startTime      time.Time
	connected      bool
	mu             sync.RWMutex
}

// MultiChannelBot manages multiple channel connections
type MultiChannelBot struct {
	authManager *AuthManager
	secretsPath string
	botUsername string
	channels    map[string]*ChannelBot
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewMultiChannelBot creates a new multi-channel bot instance
func NewMultiChannelBot(authManager *AuthManager, secretsPath string, botUsername string) *MultiChannelBot {
	ctx, cancel := context.WithCancel(context.Background())

	return &MultiChannelBot{
		authManager: authManager,
		secretsPath: secretsPath,
		botUsername: botUsername,
		channels:    make(map[string]*ChannelBot),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// AddChannel adds a new channel to the multi-channel bot
func (mcb *MultiChannelBot) AddChannel(channelName string) error {
	mcb.mu.Lock()
	defer mcb.mu.Unlock()

	// Check if channel already exists
	if _, exists := mcb.channels[channelName]; exists {
		return fmt.Errorf("channel %s already exists", channelName)
	}

	// Load the channel's config
	channelConfigPath := fmt.Sprintf("configs/channels/%s_config_secrets.yaml", channelName)
	cfg, err := config.Load(channelConfigPath)
	if err != nil {
		return fmt.Errorf("error loading config for channel %s: %w", channelName, err)
	}

	// Create command manager for this channel
	cm := commands.NewCommandManager(
		"!", // Hardcoded command prefix
		cfg.DataPath,
		channelName,
	)
	commands.RegisterBasicCommands(cm)
	commands.RegisterUptimeCommand(cm)
	commands.RegisterAuthCommand(cm, mcb.authManager)

	// Initialize channel stats
	channelStats := channelstats.NewChannelStats(cfg.DataPath)

	// Create channel bot instance
	channelBot := &ChannelBot{
		channel:        channelName,
		commandManager: cm,
		channelStats:   channelStats,
		cfg:            cfg,
		startTime:      time.Now(),
		connected:      false,
	}

	mcb.channels[channelName] = channelBot
	log.Printf("Added channel: %s", channelName)

	return nil
}

// ConnectToChannel connects to a specific channel
func (mcb *MultiChannelBot) ConnectToChannel(channelName string) error {
	mcb.mu.RLock()
	channelBot, exists := mcb.channels[channelName]
	mcb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	channelBot.mu.Lock()
	defer channelBot.mu.Unlock()

	if channelBot.connected {
		return fmt.Errorf("channel %s is already connected", channelName)
	}

	// Get initial access token
	token, err := mcb.authManager.GetAccessToken()
	if err != nil {
		return fmt.Errorf("error getting access token for channel %s: %w", channelName, err)
	}

	// Create Twitch client for this channel
	channelBot.client = twitchirc.NewClient(mcb.botUsername, "oauth:"+token)

	// Set up connection handler
	channelBot.client.OnConnect(func() {
		log.Printf("[%s] Successfully connected to Twitch IRC", channelName)
		log.Printf("[%s] Joining channel: %s", channelName, channelName)
		channelBot.client.Join(channelName)

		channelBot.mu.Lock()
		channelBot.connected = true
		channelBot.mu.Unlock()
	})

	// Set up message handler
	channelBot.client.OnPrivateMessage(func(message twitchirc.PrivateMessage) {
		// Record chatter stats
		channelBot.channelStats.RecordChatMessage(message.User.Name)

		// Check if token needs refresh
		if !mcb.authManager.IsTokenValid() {
			newToken, err := mcb.authManager.GetAccessToken()
			if err != nil {
				log.Printf("[%s] Error refreshing token: %v", channelName, err)
				return
			}
			channelBot.client.SetIRCToken("oauth:" + newToken)
		}

		// Handle commands
		if response, isCommand := channelBot.commandManager.HandleMessage(message); isCommand && response != "" {
			channelBot.client.Say(message.Channel, response)
		}
	})

	// Start connection in a goroutine
	mcb.wg.Add(1)
	go func() {
		defer mcb.wg.Done()

		for {
			select {
			case <-mcb.ctx.Done():
				return
			default:
				if err := channelBot.client.Connect(); err != nil {
					log.Printf("[%s] Error connecting to Twitch IRC: %v", channelName, err)
					log.Printf("[%s] Attempting to reconnect in 30 seconds...", channelName)
					time.Sleep(30 * time.Second)
					continue
				}
				return
			}
		}
	}()

	log.Printf("[%s] Connection initiated", channelName)
	return nil
}

// ConnectToAllChannels connects to all added channels
func (mcb *MultiChannelBot) ConnectToAllChannels() error {
	mcb.mu.RLock()
	channels := make([]string, 0, len(mcb.channels))
	for channelName := range mcb.channels {
		channels = append(channels, channelName)
	}
	mcb.mu.RUnlock()

	for _, channelName := range channels {
		if err := mcb.ConnectToChannel(channelName); err != nil {
			log.Printf("Error connecting to channel %s: %v", channelName, err)
			// Continue with other channels even if one fails
		}
	}

	return nil
}

// DisconnectFromChannel disconnects from a specific channel
func (mcb *MultiChannelBot) DisconnectFromChannel(channelName string) error {
	mcb.mu.RLock()
	channelBot, exists := mcb.channels[channelName]
	mcb.mu.RUnlock()

	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	channelBot.mu.Lock()
	defer channelBot.mu.Unlock()

	if !channelBot.connected {
		return fmt.Errorf("channel %s is not connected", channelName)
	}

	if channelBot.client != nil {
		channelBot.client.Disconnect()
	}

	channelBot.connected = false
	log.Printf("[%s] Disconnected", channelName)

	return nil
}

// DisconnectFromAllChannels disconnects from all channels
func (mcb *MultiChannelBot) DisconnectFromAllChannels() {
	mcb.mu.RLock()
	channels := make([]string, 0, len(mcb.channels))
	for channelName := range mcb.channels {
		channels = append(channels, channelName)
	}
	mcb.mu.RUnlock()

	for _, channelName := range channels {
		mcb.DisconnectFromChannel(channelName)
	}
}

// GetChannelStatus returns the connection status of a channel
func (mcb *MultiChannelBot) GetChannelStatus(channelName string) (bool, error) {
	mcb.mu.RLock()
	channelBot, exists := mcb.channels[channelName]
	mcb.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("channel %s not found", channelName)
	}

	channelBot.mu.RLock()
	defer channelBot.mu.RUnlock()

	return channelBot.connected, nil
}

// GetAllChannelStatuses returns the connection status of all channels
func (mcb *MultiChannelBot) GetAllChannelStatuses() map[string]bool {
	mcb.mu.RLock()
	defer mcb.mu.RUnlock()

	statuses := make(map[string]bool)
	for channelName, channelBot := range mcb.channels {
		channelBot.mu.RLock()
		statuses[channelName] = channelBot.connected
		channelBot.mu.RUnlock()
	}

	return statuses
}

// GetChannelCount returns the number of channels managed by this bot
func (mcb *MultiChannelBot) GetChannelCount() int {
	mcb.mu.RLock()
	defer mcb.mu.RUnlock()
	return len(mcb.channels)
}

// Shutdown gracefully shuts down the multi-channel bot
func (mcb *MultiChannelBot) Shutdown() {
	log.Println("Shutting down multi-channel bot...")

	// Cancel context to stop all goroutines
	mcb.cancel()

	// Disconnect from all channels
	mcb.DisconnectFromAllChannels()

	// Wait for all goroutines to finish
	mcb.wg.Wait()

	log.Println("Multi-channel bot shutdown complete")
}

// StartTokenRefresh starts the token refresh loop for all channels
func (mcb *MultiChannelBot) StartTokenRefresh() {
	mcb.wg.Add(1)
	go func() {
		defer mcb.wg.Done()
		mcb.refreshTokenLoop()
	}()
}

// refreshTokenLoop periodically checks and refreshes the token for all channels
func (mcb *MultiChannelBot) refreshTokenLoop() {
	// Calculate initial check interval based on time until expiry
	timeUntilExpiry := time.Until(mcb.authManager.ExpiresAt)
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
		case <-mcb.ctx.Done():
			log.Printf("[Token Refresh Loop] Context done, stopping refresh loop")
			return
		case <-ticker.C:
			// Calculate time until expiry
			timeUntilExpiry := time.Until(mcb.authManager.ExpiresAt)

			// Calculate next check interval and time
			checkInterval = calculateCheckInterval(timeUntilExpiry)

			// Only refresh if we're within minimum time of expiry
			if timeUntilExpiry <= minRefreshTime {
				log.Printf("[Token] Refreshing (expires in %s)", timeUntilExpiry.Round(time.Second))

				// Store the old expiry time for comparison
				oldExpiry := mcb.authManager.ExpiresAt

				if err := mcb.authManager.RefreshToken(); err != nil {
					log.Printf("Error refreshing token: %v", err)
					continue
				}

				// Update token for all connected channels
				mcb.mu.RLock()
				for channelName, channelBot := range mcb.channels {
					channelBot.mu.RLock()
					if channelBot.connected && channelBot.client != nil {
						channelBot.client.SetIRCToken("oauth:" + mcb.authManager.AccessToken)
						log.Printf("[%s] Token updated", channelName)
					}
					channelBot.mu.RUnlock()
				}
				mcb.mu.RUnlock()

				// Calculate new check interval based on new token expiry
				timeUntilExpiry = time.Until(mcb.authManager.ExpiresAt)
				checkInterval = calculateCheckInterval(timeUntilExpiry)

				log.Printf("[Token] Refreshed: %s -> %s (next check in %s)",
					formatTimeForLogs(oldExpiry),
					formatTimeForLogs(mcb.authManager.ExpiresAt),
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
