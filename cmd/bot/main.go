package main

import (
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

	// Create Twitch client with connection timeout
	client := twitch.NewClient(config.Twitch.BotUsername, config.Twitch.BotToken)
	client.SetConnectionTimeout(time.Second * 10)

	// Register handlers
	client.OnConnect(func() {
		log.Printf("Connected to Twitch IRC")
		client.Join(config.Twitch.Channel)
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		log.Printf("Message from %s: %s", message.User.Name, message.Message)
	})

	client.OnDisconnect(func() {
		log.Printf("Disconnected from Twitch IRC")
	})

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Connect in a goroutine so we can handle shutdown
	connected := make(chan bool)
	var connectErr error
	go func() {
		log.Printf("Attempting to connect to Twitch IRC...")
		if err := client.Connect(); err != nil {
			connectErr = err
			connected <- false
			return
		}
		connected <- true
	}()

	// Wait for either connection or timeout
	select {
	case success := <-connected:
		if !success {
			log.Fatalf("Failed to connect to Twitch: %v", connectErr)
		}
	case <-time.After(time.Second * 15):
		log.Fatal("Timeout while connecting to Twitch")
	}

	log.Printf("Bot is running in channel: %s", config.Twitch.Channel)
	fmt.Println("Press CTRL-C to exit.")

	// Wait for shutdown signal
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down gracefully...")

	// Send a part message before disconnecting
	client.Depart(config.Twitch.Channel)

	// Create a shutdown timeout
	done := make(chan bool)
	go func() {
		client.Disconnect()
		done <- true
	}()

	// Wait for disconnect with timeout
	select {
	case <-done:
		log.Println("Successfully disconnected from Twitch")
	case <-time.After(time.Second * 5):
		log.Println("Forced shutdown after timeout")
	}
}
