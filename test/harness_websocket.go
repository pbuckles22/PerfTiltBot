package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type WebSocketTestConfig struct {
	BotName        string `yaml:"bot_name"`
	BotTestChannel string `yaml:"bot_test_channel"`
	ClientID       string `yaml:"client_id"`
	ClientSecret   string `yaml:"client_secret"`
	OAuth          string `yaml:"oauth"`
	RefreshToken   string `yaml:"refresh_token"`
}

type TwitchChatMessage struct {
	Type string `json:"type"`
	Data struct {
		Timestamp string `json:"timestamp"`
		Message   string `json:"message"`
		User      struct {
			ID          string `json:"id"`
			Login       string `json:"login"`
			DisplayName string `json:"display_name"`
		} `json:"user"`
		Channel struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"channel"`
	} `json:"data"`
}

func loadWebSocketTestConfig(configPath string) (*WebSocketTestConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config WebSocketTestConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Validate required fields
	if config.BotName == "" || config.OAuth == "" || config.BotTestChannel == "" {
		return nil, fmt.Errorf("missing required fields in config")
	}

	return &config, nil
}

func main() {
	// Load test bot configuration
	configPath := "configs/bots/testbot/pbtestbot_auth_secrets.yaml"
	config, err := loadWebSocketTestConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded WebSocket test config for: %s\n", config.BotName)
	fmt.Printf("Testing in channel: %s\n", config.BotTestChannel)

	// Connect to Twitch Chat WebSocket (simpler approach)
	// This uses the same WebSocket that the web chat uses
	conn, _, err := websocket.DefaultDialer.Dial("wss://irc-ws.chat.twitch.tv:443", nil)
	if err != nil {
		fmt.Printf("Failed to connect to WebSocket: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Connected to Twitch Chat WebSocket\n")

	// Send CAP REQ for tags and commands
	capReq := "CAP REQ :twitch.tv/tags twitch.tv/commands"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(capReq)); err != nil {
		fmt.Printf("Failed to send CAP REQ: %v\n", err)
		os.Exit(1)
	}

	// Send PASS and NICK for authentication
	passCmd := fmt.Sprintf("PASS %s", config.OAuth)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(passCmd)); err != nil {
		fmt.Printf("Failed to send PASS: %v\n", err)
		os.Exit(1)
	}

	nickCmd := fmt.Sprintf("NICK %s", config.BotName)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(nickCmd)); err != nil {
		fmt.Printf("Failed to send NICK: %v\n", err)
		os.Exit(1)
	}

	// Join the channel
	joinCmd := fmt.Sprintf("JOIN #%s", config.BotTestChannel)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(joinCmd)); err != nil {
		fmt.Printf("Failed to send JOIN: %v\n", err)
		os.Exit(1)
	}

	// Send and test commands
	tests := []string{"!ping", "!help", "!uptime", "!startqueue", "!queue", "!endqueue"}

	for _, cmd := range tests {
		privmsgCmd := fmt.Sprintf("PRIVMSG #%s :%s", config.BotTestChannel, cmd)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(privmsgCmd)); err != nil {
			fmt.Printf("Failed to send PRIVMSG: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Sent command: %s\n", cmd)

		// Listen for 3 seconds for responses
		found := false
		listenStart := time.Now()
		for time.Since(listenStart) < 3*time.Second {
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					fmt.Printf("WebSocket read error: %v\n", err)
					break
				}
				// Timeout, continue to next iteration
				continue
			}
			messageStr := string(message)
			if strings.Contains(messageStr, "PRIVMSG") {
				fmt.Printf("[RESPONSE] %s\n", messageStr)
				found = true
				break // Exit immediately after receiving a response
			}
		}
		if !found {
			fmt.Printf("No response received for command: %s\n", cmd)
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Test run complete.")
}
