package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pbuckles22/PBChatBot/internal/twitch"
	"gopkg.in/yaml.v3"
)

type BotAuthConfig struct {
	BotName      string `yaml:"bot_name"`
	OAuth        string `yaml:"oauth"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RefreshToken string `yaml:"refresh_token"`
}

func loadBotAuthConfig(path string) (*BotAuthConfig, error) {
	config := &BotAuthConfig{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading bot auth file: %w", err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing bot auth file: %w", err)
	}

	// Validate required fields
	if config.BotName == "" {
		return nil, fmt.Errorf("bot_name is required")
	}
	if config.OAuth == "" {
		return nil, fmt.Errorf("oauth token is required")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("client_id is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required")
	}
	if config.RefreshToken == "" {
		return nil, fmt.Errorf("refresh_token is required")
	}

	return config, nil
}

func main() {
	log.Println("Starting PBChatBot Multi-Channel...")

	// Get bot name from environment variable
	botName := os.Getenv("BOT_NAME")
	if botName == "" {
		log.Fatal("BOT_NAME environment variable is required")
	}

	// Get channel list from environment variable (comma-separated)
	channelList := os.Getenv("CHANNEL_NAMES")
	if channelList == "" {
		log.Fatal("CHANNEL_NAMES environment variable is required (comma-separated list)")
	}

	// Parse channel names
	channels := strings.Split(channelList, ",")
	for i, channel := range channels {
		channels[i] = strings.TrimSpace(channel)
	}

	log.Printf("Bot: %s, Channels: %v", botName, channels)

	// Load bot auth config
	botAuthConfig, err := loadBotAuthConfig(fmt.Sprintf("configs/bots/%s_auth_secrets.yaml", botName))
	if err != nil {
		log.Fatalf("Failed to load bot auth configuration: %v", err)
	}

	// Create auth manager
	authManager := twitch.NewAuthManager(
		botAuthConfig.ClientID,
		botAuthConfig.ClientSecret,
		botAuthConfig.RefreshToken,
		fmt.Sprintf("configs/bots/%s_auth_secrets.yaml", botName),
	)

	// Create multi-channel bot instance
	multiBot := twitch.NewMultiChannelBot(
		authManager,
		fmt.Sprintf("configs/bots/%s_auth_secrets.yaml", botName),
		botAuthConfig.BotName,
	)

	// Add all channels to the multi-channel bot
	for _, channelName := range channels {
		if err := multiBot.AddChannel(channelName); err != nil {
			log.Printf("Error adding channel %s: %v", channelName, err)
			continue
		}
		log.Printf("Successfully added channel: %s", channelName)
	}

	// Start token refresh loop
	multiBot.StartTokenRefresh()

	// Connect to all channels
	if err := multiBot.ConnectToAllChannels(); err != nil {
		log.Printf("Error connecting to channels: %v", err)
	}

	// Log initial status
	statuses := multiBot.GetAllChannelStatuses()
	log.Printf("Initial connection statuses: %v", statuses)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		multiBot.Shutdown()
	}()

	// Keep the main thread alive
	select {}
}
