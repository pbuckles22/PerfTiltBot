package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PBChatBot/internal/commands"
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

type ChannelConfig struct {
	BotName  string `yaml:"bot_name"`
	Channel  string `yaml:"channel"`
	DataPath string `yaml:"data_path"`
	Commands struct {
		Queue struct {
			MaxSize         int `yaml:"max_size"`
			DefaultPosition int `yaml:"default_position"`
			DefaultPopCount int `yaml:"default_pop_count"`
		} `yaml:"queue"`
		Cooldowns struct {
			Default   int `yaml:"default"`
			Moderator int `yaml:"moderator"`
			VIP       int `yaml:"vip"`
		} `yaml:"cooldowns"`
	} `yaml:"commands"`
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

func loadChannelConfig(path string) (*ChannelConfig, error) {
	config := &ChannelConfig{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading channel config file: %w", err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel config file: %w", err)
	}

	// Validate required fields
	if config.BotName == "" {
		return nil, fmt.Errorf("bot_name is required")
	}
	if config.Channel == "" {
		return nil, fmt.Errorf("channel is required")
	}

	return config, nil
}

func main() {
	log.Println("Starting PBChatBot...")

	// Get channel name from environment variable
	channelName := os.Getenv("CHANNEL_NAME")
	if channelName == "" {
		log.Fatal("CHANNEL_NAME environment variable is required")
	}

	// Get bot name from environment variable
	botName := os.Getenv("BOT_NAME")
	if botName == "" {
		log.Fatal("BOT_NAME environment variable is required")
	}

	// Load bot auth config
	botAuthConfig, err := loadBotAuthConfig(fmt.Sprintf("configs/bots/%s_auth_secrets.yaml", botName))
	if err != nil {
		log.Fatalf("Failed to load bot auth configuration: %v", err)
	}

	// Load channel config
	channelConfig, err := loadChannelConfig(fmt.Sprintf("configs/channels/%s_config_secrets.yaml", channelName))
	if err != nil {
		log.Fatalf("Failed to load channel configuration: %v", err)
	}

	// Verify bot names match
	if botAuthConfig.BotName != channelConfig.BotName {
		log.Fatalf("Bot name mismatch: auth config has %s, channel config has %s",
			botAuthConfig.BotName, channelConfig.BotName)
	}

	log.Printf("Loaded configuration for bot: %s, channel: %s",
		botAuthConfig.BotName, channelConfig.Channel)

	// Create auth manager
	authManager := twitch.NewAuthManager(
		botAuthConfig.ClientID,
		botAuthConfig.ClientSecret,
		botAuthConfig.RefreshToken,
		fmt.Sprintf("configs/bots/%s_auth_secrets.yaml", botName),
	)

	// Create command manager
	cm := commands.NewCommandManager(
		"!", // Hardcoded command prefix
		channelConfig.DataPath,
		channelConfig.Channel,
	)
	commands.RegisterBasicCommands(cm)
	commands.RegisterUptimeCommand(cm)
	commands.RegisterAuthCommand(cm, authManager)

	// Create bot instance
	bot := twitch.NewBot(
		channelConfig.Channel,
		authManager,
		fmt.Sprintf("configs/bots/%s_auth_secrets.yaml", botName),
		botAuthConfig.BotName,
	)

	// Register command handlers
	bot.RegisterCommandHandler(func(message twitchirc.PrivateMessage) string {
		if response, isCommand := cm.HandleMessage(message); isCommand && response != "" {
			return response
		}
		return ""
	})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Twitch
	if err := bot.Connect(ctx); err != nil {
		log.Fatalf("Error connecting to Twitch: %v", err)
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either shutdown signal or kill command
	go func() {
		<-sigChan
		cm.RequestShutdown()
	}()

	// Wait for shutdown request
	cm.WaitForShutdown()

	// Graceful shutdown
	log.Println("Shutting down gracefully...")
	cancel() // Cancel the context to stop token refresh loop
}
