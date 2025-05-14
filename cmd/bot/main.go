package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Twitch struct {
		BotToken     string `yaml:"bot_token"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		BotUsername  string `yaml:"bot_username"`
		Channel      string `yaml:"channel"`
	} `yaml:"twitch"`
	APIs struct {
		ExampleAPIKey string `yaml:"example_api_key"`
	} `yaml:"apis"`
}

func loadConfig(path string) (*Config, error) {
	config := &Config{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
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

func main() {
	log.Println("Starting PerfTiltBot...")

	// Load configuration
	config, err := loadConfig("configs/secrets.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Bot configured for channel: %s", config.Twitch.Channel)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-sigChan

	log.Println("Shutting down gracefully...")
}
