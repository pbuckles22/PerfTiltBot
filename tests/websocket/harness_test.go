package websocket

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"
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

// sendCommandWithRetry sends a command with retry logic for connection resilience
func sendCommandWithRetry(conn *websocket.Conn, channel string, command string, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("[RETRY] Attempt %d/%d for command: %s\n", attempt+1, maxRetries, command)
			time.Sleep(2 * time.Second) // Wait before retry
		}

		privmsgCmd := fmt.Sprintf("PRIVMSG #%s :%s", channel, command)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(privmsgCmd)); err != nil {
			lastErr = err
			fmt.Printf("[ERROR] Failed to send command (attempt %d): %v\n", attempt+1, err)
			continue
		}
		return nil // Success
	}
	return fmt.Errorf("failed to send command after %d attempts: %v", maxRetries, lastErr)
}

// waitForResponse waits for a specific response pattern with timeout
func waitForResponse(conn *websocket.Conn, expectedPattern string, timeout time.Duration) (bool, string, error) {
	start := time.Now()
	lastReadTime := time.Now()

	for time.Since(start) < timeout {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		// Use panic recovery to catch the "repeated read on failed websocket connection" panic
		var message []byte
		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					if strings.Contains(fmt.Sprintf("%v", r), "repeated read on failed") {
						err = fmt.Errorf("websocket failed state: %v", r)
					} else {
						// Re-panic for other panics
						panic(r)
					}
				}
			}()
			_, message, err = conn.ReadMessage()
		}()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				fmt.Printf("[ERROR] WebSocket connection closed: %v\n", err)
				return false, "", err
			}
			if time.Since(lastReadTime) > 10*time.Second {
				fmt.Printf("[ERROR] No successful reads for 10 seconds, connection may be dead\n")
				return false, "", err
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if strings.Contains(err.Error(), "repeated read on failed") || strings.Contains(err.Error(), "websocket failed state") {
				fmt.Printf("[ERROR] Connection is in failed state: %v\n", err)
				return false, "", fmt.Errorf("websocket failed state")
			}
			fmt.Printf("[WARNING] Read error (continuing): %v\n", err)
			continue
		}
		lastReadTime = time.Now()
		messageStr := string(message)
		fmt.Printf("[DEBUG] Raw message: %s\n", messageStr)
		if strings.Contains(messageStr, "PRIVMSG") {
			fmt.Printf("[RESPONSE] %s\n", messageStr)
			if strings.Contains(strings.ToLower(messageStr), strings.ToLower(expectedPattern)) {
				return true, messageStr, nil
			}
		}
	}
	fmt.Printf("[TIMEOUT] Expected pattern '%s' not found within %v\n", expectedPattern, timeout)
	return false, "", nil
}

// checkConnectionHealth performs a quick health check on the WebSocket connection
func checkConnectionHealth(conn *websocket.Conn) bool {
	// Don't try to read from the connection as it might be in a failed state
	// Instead, just check if we can write to it
	err := conn.WriteMessage(websocket.TextMessage, []byte("PING :tmi.twitch.tv"))
	if err != nil {
		fmt.Printf("[HEALTH] Connection write failed: %v\n", err)
		return false
	}
	return true
}

// sendCommandAndWait sends a command and waits for a specific response
func sendCommandAndWait(conn *websocket.Conn, channel string, command string, expectedResponse string, timeout time.Duration) (bool, error) {
	if !checkConnectionHealth(conn) {
		return false, fmt.Errorf("connection health check failed before sending command")
	}
	if err := sendCommandWithRetry(conn, channel, command, 3); err != nil {
		return false, fmt.Errorf("failed to send command: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	found, _, err := waitForResponse(conn, expectedResponse, timeout)
	if err != nil {
		return false, err
	}
	if !found {
		return false, fmt.Errorf("expected response '%s' not found for command '%s'", expectedResponse, command)
	}
	return true, nil
}

// runTestWithReconnect runs a test with automatic reconnection if the connection fails
func runTestWithReconnect(conn **websocket.Conn, config *WebSocketTestConfig, test struct {
	command     string
	expect      string
	description string
}, timeout time.Duration) (bool, error) {
	success, err := sendCommandAndWait(*conn, config.BotTestChannel, test.command, test.expect, timeout)
	if err == nil {
		return success, nil
	}
	isConnectionError := strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "websocket") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "health check failed") ||
		strings.Contains(err.Error(), "websocket failed state")
	if isConnectionError {
		fmt.Printf("[RECONNECT] Connection issue detected (%s), attempting to reconnect...\n", err.Error())
		(*conn).Close()
		newConn, reconnectErr := connectToTwitch(config)
		if reconnectErr != nil {
			return false, fmt.Errorf("failed to reconnect: %v", reconnectErr)
		}
		if clearErr := clearQueueAndWait(newConn, config.BotTestChannel); clearErr != nil {
			return false, fmt.Errorf("failed to clear queue after reconnect: %v", clearErr)
		}
		*conn = newConn
		fmt.Printf("[RECONNECT] Retrying test after reconnection...\n")
		return sendCommandAndWait(*conn, config.BotTestChannel, test.command, test.expect, timeout)
	}
	return success, err
}

func verifyQueueState(conn *websocket.Conn, channel string, expectedState string, timeout time.Duration) (bool, error) {
	return sendCommandAndWait(conn, channel, "!queue", expectedState, timeout)
}

func checkBackupFiles(channel string) {
	backupFiles := []string{
		fmt.Sprintf("data/%s_queue_state.json.backup", channel),
		fmt.Sprintf("data/%s_queue_state.json.backup2", channel),
	}

	for _, file := range backupFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("[BACKUP] Found backup file: %s\n", file)
		}
	}
}

func connectToTwitch(config *WebSocketTestConfig) (*websocket.Conn, error) {
	// Connect to Twitch IRC
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial("wss://irc-ws.chat.twitch.tv:443", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Twitch: %v", err)
	}

	// Send authentication
	authCmd := fmt.Sprintf("PASS %s", config.OAuth)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(authCmd)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send PASS: %v", err)
	}

	nickCmd := fmt.Sprintf("NICK %s", config.BotName)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(nickCmd)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send NICK: %v", err)
	}

	// Join the test channel
	joinCmd := fmt.Sprintf("JOIN #%s", config.BotTestChannel)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(joinCmd)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send JOIN: %v", err)
	}

	// Wait a moment for connection to establish
	time.Sleep(2 * time.Second)

	// Check if connection is healthy
	if !checkConnectionHealth(conn) {
		conn.Close()
		return nil, fmt.Errorf("connection health check failed after setup")
	}

	fmt.Printf("[CONNECT] Successfully connected to Twitch IRC\n")
	return conn, nil
}

func clearQueueAndWait(conn *websocket.Conn, channel string) error {
	// Try to clear the queue if it exists
	sendCommandWithRetry(conn, channel, "!clearqueue", 1)
	time.Sleep(1 * time.Second)
	return nil
}

// TestWebSocketCommands runs a comprehensive test of all bot commands via WebSocket
func TestWebSocketCommands(t *testing.T) {
	// Skip this test if no config file is available
	configPath := "configs/test_websocket_config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Skipping WebSocket test - no config file found at " + configPath)
	}

	config, err := loadWebSocketTestConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Twitch
	conn, err := connectToTwitch(config)
	if err != nil {
		t.Fatalf("Failed to connect to Twitch: %v", err)
	}
	defer conn.Close()

	// Clear any existing queue state
	if err := clearQueueAndWait(conn, config.BotTestChannel); err != nil {
		t.Fatalf("Failed to clear queue: %v", err)
	}

	// Define test cases
	tests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!ping", "Pong!", "Ping command"},
		{"!startqueue", "started the queue system", "Start queue"},
		{"!join", "joined queue at position 1", "Join queue"},
		{"!queue", "Queue: testuser", "Check queue"},
		{"!position", "testuser is at position 1", "Check position"},
		{"!endqueue", "ended the queue system", "End queue"},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			success, err := runTestWithReconnect(&conn, config, test, 10*time.Second)
			if err != nil {
				t.Errorf("Test failed: %v", err)
				return
			}
			if !success {
				t.Errorf("Expected response '%s' not found for command '%s'", test.expect, test.command)
			}
		})

		// Add delay between tests to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	// Check for backup files
	checkBackupFiles(config.BotTestChannel)
}

// TestWebSocketConnection tests basic connection functionality
func TestWebSocketConnection(t *testing.T) {
	// Skip this test if no config file is available
	configPath := "configs/test_websocket_config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Skipping WebSocket connection test - no config file found at " + configPath)
	}

	config, err := loadWebSocketTestConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test connection
	conn, err := connectToTwitch(config)
	if err != nil {
		t.Fatalf("Failed to connect to Twitch: %v", err)
	}
	defer conn.Close()

	// Test basic ping
	if !checkConnectionHealth(conn) {
		t.Error("Connection health check failed")
	}

	// Test sending a simple command
	err = sendCommandWithRetry(conn, config.BotTestChannel, "!ping", 3)
	if err != nil {
		t.Errorf("Failed to send ping command: %v", err)
	}
}

// TestWebSocketReconnection tests reconnection logic
func TestWebSocketReconnection(t *testing.T) {
	// Skip this test if no config file is available
	configPath := "configs/test_websocket_config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("Skipping WebSocket reconnection test - no config file found at " + configPath)
	}

	config, err := loadWebSocketTestConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Connect
	conn, err := connectToTwitch(config)
	if err != nil {
		t.Fatalf("Failed to connect to Twitch: %v", err)
	}

	// Close connection to simulate failure
	conn.Close()

	// Try to reconnect
	newConn, err := connectToTwitch(config)
	if err != nil {
		t.Fatalf("Failed to reconnect: %v", err)
	}
	defer newConn.Close()

	// Verify new connection works
	if !checkConnectionHealth(newConn) {
		t.Error("Reconnected connection health check failed")
	}
}
