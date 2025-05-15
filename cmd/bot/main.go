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

	log.Printf("Loaded configuration for bot: %s, channel: %s", config.Twitch.BotUsername, config.Twitch.Channel)

	// Create Twitch client
	client := twitch.NewClient(config.Twitch.BotUsername, config.Twitch.BotToken)

	// Channel to track successful connection
	connectionEstablished := make(chan bool)

	// Register handlers
	client.OnConnect(func() {
		log.Printf("Successfully connected to Twitch IRC")
		// Join the channel after connection is established
		log.Printf("Attempting to join channel: %s", config.Twitch.Channel)
		client.Join(config.Twitch.Channel)
		connectionEstablished <- true
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Printf("Message from %s: %s", message.User.Name, message.Message)
	})

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Connect in a goroutine
	log.Printf("Attempting to connect to Twitch IRC...")
	go func() {
		if err := client.Connect(); err != nil {
			log.Printf("Connection error: %v", err)
			connectionEstablished <- false
		}
	}()

	// Wait for either successful connection or timeout
	select {
	case success := <-connectionEstablished:
		if !success {
			log.Fatal("Failed to establish connection to Twitch")
		}
		log.Printf("Connection and channel join successful")
	case <-time.After(time.Second * 30): // Increased timeout to 30 seconds
		log.Fatal("Timeout while establishing connection to Twitch")
	}

	log.Printf("Bot is fully running in channel: %s", config.Twitch.Channel)

	client.Say(config.Twitch.Channel, "Hello Bitches! W pbuck")

	fmt.Println("Press CTRL-C to exit.")

	// Wait for shutdown signal
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down gracefully...")

	// Send a part message before disconnecting
	log.Printf("Leaving channel: %s", config.Twitch.Channel)
	client.Depart(config.Twitch.Channel)

	// Create a shutdown timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10) // Increased timeout to 10 seconds
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
