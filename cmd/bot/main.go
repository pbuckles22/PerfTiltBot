package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	twitchirc "github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PerfTiltBot/internal/commands"
	"github.com/pbuckles22/PerfTiltBot/internal/config"
	"github.com/pbuckles22/PerfTiltBot/internal/twitch"
	"gopkg.in/yaml.v3"
)

type SecretsConfig struct {
	Twitch struct {
		BotToken     string `yaml:"bot_token"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		BotUsername  string `yaml:"bot_username"`
		Channel      string `yaml:"channel"`
		DataPath     string `yaml:"data_path"`
		RefreshToken string `yaml:"refresh_token"`
	} `yaml:"twitch"`
}

type BotConfig struct {
	Bot struct {
		CommandPrefix string `yaml:"command_prefix"`
		Cooldowns     struct {
			Default   int `yaml:"default"`
			Moderator int `yaml:"moderator"`
		} `yaml:"cooldowns"`
		Permissions struct {
			ModeratorOnly   []string `yaml:"moderator_only"`
			BroadcasterOnly []string `yaml:"broadcaster_only"`
		} `yaml:"permissions"`
	} `yaml:"bot"`
}

func loadSecretsConfig(path string) (*SecretsConfig, error) {
	config := &SecretsConfig{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading secrets file: %w", err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing secrets file: %w", err)
	}

	// Validate required fields
	if config.Twitch.BotToken == "" {
		return nil, fmt.Errorf("twitch bot token is required")
	}
	if config.Twitch.BotUsername == "" {
		return nil, fmt.Errorf("twitch bot username is required")
	}
	if config.Twitch.Channel == "" {
		return nil, fmt.Errorf("twitch channel is required")
	}

	return config, nil
}

func loadBotConfig(path string) (*BotConfig, error) {
	config := &BotConfig{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading bot config file: %w", err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing bot config file: %w", err)
	}

	// Set default command prefix if not specified
	if config.Bot.CommandPrefix == "" {
		config.Bot.CommandPrefix = "!"
	}

	return config, nil
}

func main() {
	log.Println("Starting PerfTiltBot...")

	// Load configurations
	cfg, err := config.Load("configs/secrets.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	botConfig, err := loadBotConfig("configs/bot.yaml")
	if err != nil {
		log.Fatalf("Failed to load bot configuration: %v", err)
	}

	log.Printf("Loaded configuration for bot: %s, channel: %s", cfg.Twitch.BotUsername, cfg.Twitch.Channel)

	// Create command manager
	cm := commands.NewCommandManager(
		botConfig.Bot.CommandPrefix,
		cfg.Twitch.DataPath,
		cfg.Twitch.Channel,
	)
	commands.RegisterBasicCommands(cm)

	// Create auth manager
	authManager := twitch.NewAuthManager(
		cfg.Twitch.ClientID,
		cfg.Twitch.ClientSecret,
		cfg.Twitch.RefreshToken,
		"configs/secrets.yaml",
	)

	// Create bot instance
	bot := twitch.NewBot(cfg.Twitch.Channel, authManager, "configs/secrets.yaml", cfg.Twitch.BotUsername)

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
