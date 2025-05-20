package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/pbuckles22/PerfTiltBot/internal/commands"
	"github.com/pbuckles22/PerfTiltBot/internal/config"
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

	// Create Twitch client
	client := twitch.NewClient(cfg.Twitch.BotUsername, cfg.Twitch.BotToken)

	// Channel to track successful connection
	connectionEstablished := make(chan bool)

	// Register handlers
	client.OnConnect(func() {
		log.Printf("Successfully connected to Twitch IRC")
		// Join the channel after connection is established
		log.Printf("Attempting to join channel: %s", cfg.Twitch.Channel)
		client.Join(cfg.Twitch.Channel)
		connectionEstablished <- true
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Printf("Message from %s: %s", message.User.Name, message.Message)

		// Handle commands
		if response, isCommand := cm.HandleMessage(message); isCommand && response != "" {
			client.Say(message.Channel, response)
		}
	})

	// Connect to Twitch IRC in a separate goroutine
	go func() {
		if err := client.Connect(); err != nil {
			log.Fatalf("Error connecting to Twitch IRC: %v", err)
		}
	}()

	// Wait for successful connection
	select {
	case <-connectionEstablished:
		log.Println("Bot is now running. Press Ctrl+C to exit.")
	case <-time.After(time.Second * 30):
		log.Fatal("Failed to establish connection within timeout")
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

	// Send a part message before disconnecting
	log.Printf("Leaving channel: %s", cfg.Twitch.Channel)
	client.Depart(cfg.Twitch.Channel)

	// Create a shutdown timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer shutdownCancel()

	// Create a channel to signal disconnect completion
	done := make(chan bool)
	go func() {
		client.Disconnect()
		done <- true
	}()

	// Wait for disconnect with timeout
	select {
	case <-done:
		log.Println("Successfully disconnected from Twitch")
	case <-shutdownCtx.Done():
		log.Println("Forced shutdown after timeout")
	}
}
