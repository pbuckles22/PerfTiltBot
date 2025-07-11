package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	twitch "github.com/gempir/go-twitch-irc/v4"
	"gopkg.in/yaml.v3"
)

type TestBotConfig struct {
	BotName        string `yaml:"bot_name"`
	BotTestChannel string `yaml:"bot_test_channel"`
	ClientID       string `yaml:"client_id"`
	ClientSecret   string `yaml:"client_secret"`
	OAuth          string `yaml:"oauth"`
	RefreshToken   string `yaml:"refresh_token"`
}

func loadTestBotConfig(configPath string) (*TestBotConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config TestBotConfig
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

func mainIRC() {
	// Load test bot configuration
	configPath := "configs/bots/testbot/pbtestbot_auth_secrets.yaml"
	config, err := loadTestBotConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded test bot config for: %s\n", config.BotName)
	fmt.Printf("Testing in channel: %s\n", config.BotTestChannel)

	// TODO: Add token refresh logic here if needed
	// For now, we'll use the OAuth token as-is

	client := twitch.NewClient(config.BotName, config.OAuth)
	responses := make(chan string, 1000) // Much larger buffer
	connectedUsername := make(chan string, 1)

	// Track all messages received for debugging
	var allMessages []string
	var messageMutex sync.Mutex

	// Simple message counter for debugging
	var messageCount int32

	client.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		atomic.AddInt32(&messageCount, 1)
		fmt.Printf("[DEBUG] PrivateMessage #%d from %s in channel '%s': %s\n",
			atomic.LoadInt32(&messageCount), msg.User.Name, msg.Channel, msg.Message)

		// Capture ALL messages from the target channel
		if msg.Channel == config.BotTestChannel {
			messageStr := fmt.Sprintf("%s: %s", msg.User.Name, msg.Message)
			fmt.Printf("[DEBUG] Adding to responses: %s\n", messageStr)

			messageMutex.Lock()
			allMessages = append(allMessages, messageStr)
			messageMutex.Unlock()

			// Send to channel (non-blocking)
			select {
			case responses <- messageStr:
				fmt.Printf("[DEBUG] Successfully sent to responses channel\n")
			default:
				fmt.Printf("[ERROR] Responses channel is full! Dropping message: %s\n", messageStr)
			}
		}
	})

	// Also listen for other message types that might contain bot responses
	client.OnClearChatMessage(func(msg twitch.ClearChatMessage) {
		fmt.Printf("[DEBUG] ClearChatMessage: %s\n", msg.Message)
	})

	client.OnWhisperMessage(func(msg twitch.WhisperMessage) {
		fmt.Printf("[DEBUG] WhisperMessage from %s: %s\n", msg.User.Name, msg.Message)
		responses <- fmt.Sprintf("%s: %s", msg.User.Name, msg.Message)
	})

	// Listen for all message types to debug connection issues
	client.OnUserNoticeMessage(func(msg twitch.UserNoticeMessage) {
		fmt.Printf("[DEBUG] UserNoticeMessage: %s\n", msg.Message)
	})

	client.OnUserStateMessage(func(msg twitch.UserStateMessage) {
		fmt.Printf("[DEBUG] UserStateMessage in channel %s: %s\n", msg.Channel, msg.User.Name)
		if msg.Channel == config.BotTestChannel {
			connectedUsername <- msg.User.Name
		}
	})

	// Add more message type handlers to catch everything
	client.OnGlobalUserStateMessage(func(msg twitch.GlobalUserStateMessage) {
		fmt.Printf("[DEBUG] GlobalUserStateMessage: %s\n", msg.User.Name)
	})

	client.OnRoomStateMessage(func(msg twitch.RoomStateMessage) {
		fmt.Printf("[DEBUG] RoomStateMessage in channel %s\n", msg.Channel)
	})

	client.OnNamesMessage(func(msg twitch.NamesMessage) {
		fmt.Printf("[DEBUG] NamesMessage in %s: %d users\n", msg.Channel, len(msg.Users))
	})

	client.OnPingMessage(func(msg twitch.PingMessage) {
		fmt.Printf("[DEBUG] PingMessage: %s\n", msg.Message)
	})

	client.OnPongMessage(func(msg twitch.PongMessage) {
		fmt.Printf("[DEBUG] PongMessage: %s\n", msg.Message)
	})

	client.OnUnsetMessage(func(msg twitch.RawMessage) {
		fmt.Printf("[DEBUG] UnsetMessage: %s\n", msg.Raw)
		// Check if this is a PRIVMSG from the bot
		if strings.Contains(msg.Raw, "PRIVMSG") && strings.Contains(msg.Raw, "perftiltbot") {
			fmt.Printf("[DEBUG] Found bot response in raw message: %s\n", msg.Raw)
			// Try to extract the message content
			parts := strings.Split(msg.Raw, " :")
			if len(parts) >= 2 {
				messageContent := parts[len(parts)-1]
				fmt.Printf("[DEBUG] Extracted message: %s\n", messageContent)
				responses <- fmt.Sprintf("perftiltbot: %s", messageContent)
			}
		}
		// Log ALL PRIVMSG messages to see what's coming through
		if strings.Contains(msg.Raw, "PRIVMSG") {
			fmt.Printf("[DEBUG] ALL PRIVMSG: %s\n", msg.Raw)
		}
	})

	client.OnConnect(func() {
		fmt.Printf("Connected to Twitch IRC as test user: %s\n", config.BotName)
		client.Join(config.BotTestChannel)
	})

	// Connection health monitoring would go here if available

	// Add more detailed message logging
	client.OnClearChatMessage(func(msg twitch.ClearChatMessage) {
		fmt.Printf("[DEBUG] ClearChatMessage: %s\n", msg.Message)
	})

	client.OnWhisperMessage(func(msg twitch.WhisperMessage) {
		fmt.Printf("[DEBUG] WhisperMessage from %s: %s\n", msg.User.Name, msg.Message)
	})

	client.OnUserNoticeMessage(func(msg twitch.UserNoticeMessage) {
		fmt.Printf("[DEBUG] UserNoticeMessage: %s\n", msg.Message)
	})

	client.OnUserStateMessage(func(msg twitch.UserStateMessage) {
		fmt.Printf("[DEBUG] UserStateMessage in channel %s: %s\n", msg.Channel, msg.User.Name)
		if msg.Channel == config.BotTestChannel {
			connectedUsername <- msg.User.Name
		}
	})

	// Add more message type handlers to catch everything
	client.OnGlobalUserStateMessage(func(msg twitch.GlobalUserStateMessage) {
		fmt.Printf("[DEBUG] GlobalUserStateMessage: %s\n", msg.User.Name)
	})

	client.OnRoomStateMessage(func(msg twitch.RoomStateMessage) {
		fmt.Printf("[DEBUG] RoomStateMessage in channel %s\n", msg.Channel)
	})

	client.OnNamesMessage(func(msg twitch.NamesMessage) {
		fmt.Printf("[DEBUG] NamesMessage in %s: %d users\n", msg.Channel, len(msg.Users))
	})

	client.OnPingMessage(func(msg twitch.PingMessage) {
		fmt.Printf("[DEBUG] PingMessage: %s\n", msg.Message)
	})

	client.OnPongMessage(func(msg twitch.PongMessage) {
		fmt.Printf("[DEBUG] PongMessage: %s\n", msg.Message)
	})

	client.OnUnsetMessage(func(msg twitch.RawMessage) {
		fmt.Printf("[DEBUG] UnsetMessage: %s\n", msg.Raw)
		// Check if this is a PRIVMSG from the bot
		if strings.Contains(msg.Raw, "PRIVMSG") && strings.Contains(msg.Raw, "perftiltbot") {
			fmt.Printf("[DEBUG] Found bot response in raw message: %s\n", msg.Raw)
			// Try to extract the message content
			parts := strings.Split(msg.Raw, " :")
			if len(parts) >= 2 {
				messageContent := parts[len(parts)-1]
				fmt.Printf("[DEBUG] Extracted message: %s\n", messageContent)
				responses <- fmt.Sprintf("perftiltbot: %s", messageContent)
			}
		}
	})

	go func() {
		err := client.Connect()
		if err != nil {
			fmt.Printf("IRC connection error: %v\n", err)
			os.Exit(1)
		}
	}()

	time.Sleep(3 * time.Second) // Wait for join

	startTime := time.Now()

	// Verify the connected username matches the configured bot name (case-insensitive)
	select {
	case actualUsername := <-connectedUsername:
		fmt.Printf("[DEBUG] actualUsername: >%s<, config.BotName: >%s<\n", actualUsername, config.BotName)
		if !strings.EqualFold(actualUsername, config.BotName) {
			fmt.Printf("SECURITY WARNING: Connected as '%s' but config expects '%s'\n", actualUsername, config.BotName)
			fmt.Printf("OAuth token belongs to different account. Exiting for security.\n")
			os.Exit(1)
		}
		fmt.Printf("✓ Verified connected as: %s\n", actualUsername)
	case <-time.After(5 * time.Second):
		fmt.Printf("Warning: Could not verify connected username (timeout)\n")
	}

	// Simple listening test - just listen for any messages for 10 seconds
	fmt.Printf("\n=== LISTENING TEST ===\n")
	fmt.Printf("Listening for any messages for 10 seconds...\n")
	listenStart := time.Now()
	for time.Since(listenStart) < 10*time.Second {
		time.Sleep(1 * time.Second)
		messageMutex.Lock()
		currentCount := len(allMessages)
		messageMutex.Unlock()
		fmt.Printf("[DEBUG] After %ds: %d messages received\n", int(time.Since(listenStart).Seconds()), currentCount)
	}
	fmt.Printf("Listening test complete. Total messages: %d\n", len(allMessages))

	tests := []struct {
		command string
		expect  string // substring to expect in response
	}{
		{"!ping", "Pong"},
		{"!help", "Available commands"},
		{"!uptime", "running"},
		{"!startqueue", "started the queue system"},
		{"!queue", "Queue"},
		{"!endqueue", "ended the queue system"},
	}

	for _, test := range tests {
		fmt.Printf("Testing %s... ", test.command)
		fmt.Printf("[DEBUG] Starting test at %s\n", time.Now().Format("15:04:05.000"))

		// Get current message count before sending command
		messageMutex.Lock()
		messageCountBefore := len(allMessages)
		messageMutex.Unlock()

		fmt.Printf("[DEBUG] Sending command at %s\n", time.Now().Format("15:04:05.000"))
		client.Say(config.BotTestChannel, test.command)
		commandSentTime := time.Now()

		// Wait for response with multiple attempts
		passed := false
		for attempt := 0; attempt < 6; attempt++ { // Try up to 6 times (3 seconds total)
			time.Sleep(500 * time.Millisecond)

			// Check if we received any new messages
			messageMutex.Lock()
			messageCountAfter := len(allMessages)
			newMessages := allMessages[messageCountBefore:]
			messageMutex.Unlock()

			fmt.Printf("[DEBUG] Attempt %d: Messages before: %d, after: %d, new: %d\n",
				attempt+1, messageCountBefore, messageCountAfter, len(newMessages))

			// Check each new message
			for _, msg := range newMessages {
				responseTime := time.Since(commandSentTime)
				fmt.Printf("[DEBUG] Checking message: %s\n", msg)
				if test.expect == "" || containsIgnoreCaseAndTrim(msg, test.expect) {
					fmt.Printf("PASS (%.3fs) - Found matching response on attempt %d\n", responseTime.Seconds(), attempt+1)
					passed = true
					break
				}
			}

			if passed {
				break
			}
		}

		if !passed {
			fmt.Printf("FAIL - No matching response found after 6 attempts. Expected: '%s'\n", test.expect)
			messageMutex.Lock()
			allMessagesCopy := make([]string, len(allMessages))
			copy(allMessagesCopy, allMessages)
			messageMutex.Unlock()
			fmt.Printf("[DEBUG] All messages captured so far (%d total):\n", len(allMessagesCopy))
			for i, msg := range allMessagesCopy {
				fmt.Printf("  %d: %s\n", i+1, msg)
			}
		}

		// Minimal delay between commands to avoid rate limiting
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Basic health checks complete.")
	fmt.Printf("Total test time: %s\n", time.Since(startTime).Round(time.Millisecond))
	fmt.Println("\n=== MANUAL TESTING INSTRUCTIONS ===")
	fmt.Println("Due to Twitch IRC rate limiting, automated testing of multiple commands")
	fmt.Println("may not work reliably. Please test the following commands manually:")
	fmt.Println("  - !uptime")
	fmt.Println("  - !help")
	fmt.Println("  - Any other commands you want to verify")
	fmt.Println("\nThe test bot is now connected and ready for manual testing.")
	os.Exit(0)
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsIgnoreCase(s[1:], substr) || containsIgnoreCase(s, substr[1:]))
}

func containsIgnoreCaseAndTrim(s, substr string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	substr = strings.ToLower(strings.TrimSpace(substr))
	return strings.Contains(s, substr)
}
