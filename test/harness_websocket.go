package main

import (
	"fmt"
	"io/ioutil"
	"net"
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

// verifyQueueState sends a queue command and verifies the expected state
func verifyQueueState(conn *websocket.Conn, channel string, expectedState string, timeout time.Duration) (bool, error) {
	return sendCommandAndWait(conn, channel, "!queue", expectedState, timeout)
}

// checkBackupFiles checks if backup files exist for debugging
func checkBackupFiles(channel string) {
	fmt.Printf("[DEBUG] Checking for backup files for channel: %s\n", channel)

	// Check for backup file
	backupFile := fmt.Sprintf("data/queue_backup_%s.json", channel)
	if _, err := os.Stat(backupFile); err == nil {
		fmt.Printf("[DEBUG] ✓ Backup file exists: %s\n", backupFile)
	} else {
		fmt.Printf("[DEBUG] ✗ Backup file missing: %s (error: %v)\n", backupFile, err)
	}

	// Check for auto-save file
	autoSaveFile := fmt.Sprintf("data/queue_state_%s.json", channel)
	if _, err := os.Stat(autoSaveFile); err == nil {
		fmt.Printf("[DEBUG] ✓ Auto-save file exists: %s\n", autoSaveFile)
	} else {
		fmt.Printf("[DEBUG] ✗ Auto-save file missing: %s (error: %v)\n", autoSaveFile, err)
	}
}

// connectToTwitch establishes a WebSocket connection to Twitch with retry logic
func connectToTwitch(config *WebSocketTestConfig) (*websocket.Conn, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("[RECONNECT] Attempt %d/%d to connect to Twitch\n", attempt+1, maxRetries)
			time.Sleep(5 * time.Second) // Wait before retry
		}

		conn, _, err := websocket.DefaultDialer.Dial("wss://irc-ws.chat.twitch.tv:443", nil)
		if err != nil {
			lastErr = err
			fmt.Printf("[ERROR] Failed to connect (attempt %d): %v\n", attempt+1, err)
			continue
		}

		// Send CAP REQ for tags and commands
		capReq := "CAP REQ :twitch.tv/tags twitch.tv/commands"
		if err := conn.WriteMessage(websocket.TextMessage, []byte(capReq)); err != nil {
			conn.Close()
			lastErr = err
			fmt.Printf("[ERROR] Failed to send CAP REQ (attempt %d): %v\n", attempt+1, err)
			continue
		}

		// Send PASS and NICK for authentication
		passCmd := fmt.Sprintf("PASS %s", config.OAuth)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(passCmd)); err != nil {
			conn.Close()
			lastErr = err
			fmt.Printf("[ERROR] Failed to send PASS (attempt %d): %v\n", attempt+1, err)
			continue
		}

		nickCmd := fmt.Sprintf("NICK %s", config.BotName)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(nickCmd)); err != nil {
			conn.Close()
			lastErr = err
			fmt.Printf("[ERROR] Failed to send NICK (attempt %d): %v\n", attempt+1, err)
			continue
		}

		// Join the channel
		joinCmd := fmt.Sprintf("JOIN #%s", config.BotTestChannel)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(joinCmd)); err != nil {
			conn.Close()
			lastErr = err
			fmt.Printf("[ERROR] Failed to send JOIN (attempt %d): %v\n", attempt+1, err)
			continue
		}

		fmt.Printf("✓ Connected to Twitch Chat WebSocket (attempt %d)\n", attempt+1)
		return conn, nil
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, lastErr)
}

// clearQueueAndWait clears the queue and waits for confirmation
func clearQueueAndWait(conn *websocket.Conn, channel string) error {
	fmt.Printf("[SETUP] Clearing queue for clean test state...\n")

	// End queue system if running
	if err := sendCommandWithRetry(conn, channel, "!endqueue", 3); err != nil {
		return fmt.Errorf("failed to end queue: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Start queue system fresh
	if err := sendCommandWithRetry(conn, channel, "!startqueue", 3); err != nil {
		return fmt.Errorf("failed to start queue: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Clear any existing users
	if err := sendCommandWithRetry(conn, channel, "!clearqueue", 3); err != nil {
		return fmt.Errorf("failed to clear queue: %v", err)
	}
	time.Sleep(1 * time.Second)

	fmt.Printf("[SETUP] Queue cleared and ready for testing\n")
	return nil
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

	// Connect to Twitch with retry logic
	conn, err := connectToTwitch(config)
	if err != nil {
		fmt.Printf("Failed to connect to Twitch: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Wait a bit for connection to stabilize
	time.Sleep(3 * time.Second)

	// Clear queue for clean test state
	if err := clearQueueAndWait(conn, config.BotTestChannel); err != nil {
		fmt.Printf("Failed to clear queue: %v\n", err)
		os.Exit(1)
	}

	// Check initial backup file state
	fmt.Printf("\n=== INITIAL BACKUP FILE STATE ===\n")
	checkBackupFiles(config.BotTestChannel)

	// Track test results
	totalTests := 0
	passed := 0
	failed := 0
	skipped := 0

	// Test Group 1: Basic connectivity and info tests
	fmt.Printf("\n=== TEST GROUP 1: BASIC CONNECTIVITY ===\n")
	basicTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!ping", "Pong", "Basic bot connectivity"},
		{"!help", "Available commands", "Command listing"},
		{"!uptime", "running", "Bot uptime"},
	}

	for _, test := range basicTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 2: Queue system lifecycle (with proper state verification)
	fmt.Printf("\n=== TEST GROUP 2: QUEUE SYSTEM LIFECYCLE ===\n")
	queueLifecycleTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!queue", "currently empty", "Empty queue verification"},
		{"!join", "joined queue", "Self-join"},
		{"!queue", "pbtestbot", "Queue state after join"},
		{"!position", "position 1", "Self position check"},
	}

	for _, test := range queueLifecycleTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 3: Basic queue operations (with state verification)
	fmt.Printf("\n=== TEST GROUP 3: BASIC QUEUE OPERATIONS ===\n")
	basicQueueTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!join testuser1", "joined queue", "Add single user"},
		{"!join testuser2", "joined queue", "Add second user"},
		{"!queue", "testuser1", "Queue state with multiple users"},
		{"!move testuser1 5", "moved to position", "Move user by name"},
		{"!queue", "testuser1", "Queue state after move"},
	}

	for _, test := range basicQueueTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 4: Multi-user operations
	fmt.Printf("\n=== TEST GROUP 4: MULTI-USER OPERATIONS ===\n")
	multiUserTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!join multi1 multi2 multi3", "joined queue", "Multi-user join"},
		{"!queue", "multi1", "Queue state after multi-join"},
		{"!pop 1", "Popped:", "Pop single user"},
		{"!queue", "testuser2", "Queue state after pop"},
		{"!pop 2", "Popped:", "Pop multiple users"},
		{"!queue", "testuser1", "Queue state after multi-pop"},
	}

	for _, test := range multiUserTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 5: Remove operations
	fmt.Printf("\n=== TEST GROUP 5: REMOVE OPERATIONS ===\n")
	removeTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!remove testuser1", "removed from queue", "Remove user by name"},
		{"!queue", "multi3", "Queue state after remove"},
		{"!remove 1", "removed from queue", "Remove user by position"},
		{"!queue", "pbtestbot", "Queue state after position remove"},
		{"!leave pbtestbot", "left queue", "Leave self"},
		{"!queue", "currently empty", "Queue state after leave"},
	}

	for _, test := range removeTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 6: Edge cases and error conditions
	fmt.Printf("\n=== TEST GROUP 6: EDGE CASES AND ERRORS ===\n")
	edgeCaseTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!join edgeuser", "joined the queue", "Add user for edge case testing"},
		{"!move edgeuser 1", "moved to position", "Move to same position (no-op)"},
		{"!queue", "edgeuser", "Queue state after no-op move"},
		{"!pop", "Popped from queue", "Pop with no arguments (default 1)"},
		{"!queue", "currently empty", "Queue state after default pop"},
		{"!join testuser", "joined the queue", "Add user for invalid pop test"},
		{"!pop 0", "Invalid number", "Pop with invalid argument (0)"},
		{"!pop -1", "Invalid number", "Pop with invalid argument (negative)"},
		{"!pop abc", "Invalid number", "Pop with invalid argument (non-numeric)"},
		{"!move nonexistent 1", "not in the queue", "Move non-existent user"},
		{"!move 999 1", "Invalid from position", "Move from invalid position"},
		{"!move testuser abc", "Invalid target position", "Move to invalid position"},
		{"!remove nonexistent", "not in the queue", "Remove non-existent user"},
		{"!remove 999", "Invalid position", "Remove from invalid position"},
	}

	for _, test := range edgeCaseTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 7: Clear queue operations
	fmt.Printf("\n=== TEST GROUP 7: CLEAR QUEUE OPERATIONS ===\n")
	clearTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!clearqueue", "cleared the queue", "Clear queue"},
		{"!queue", "currently empty", "Queue state after clear"},
	}

	for _, test := range clearTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 8: Manual backup/restore system (ISOLATED)
	fmt.Printf("\n=== TEST GROUP 8: MANUAL BACKUP/RESTORE SYSTEM ===\n")
	fmt.Printf("This group tests the manual backup system in isolation...\n")
	time.Sleep(3 * time.Second) // Extra delay before backup tests

	manualBackupTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!join finaluser", "joined the queue", "Add user for backup testing"},
		{"!savequeue", "Queue state has been saved", "Manual backup"},
		{"!queue", "finaluser", "Queue state after manual backup"},
		{"!leave finaluser", "left the queue", "Remove user after backup"},
		{"!queue", "currently empty", "Queue state after leave"},
		{"!restorequeue", "Queue state has been restored", "Manual restore (loads from backup file)"},
		{"!queue", "finaluser", "Queue state after manual restore"},
	}

	for _, test := range manualBackupTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 9: Auto-save/restore system (ISOLATED)
	fmt.Printf("\n=== TEST GROUP 9: AUTO-SAVE/RESTORE SYSTEM ===\n")
	fmt.Printf("This group tests the auto-save system in isolation...\n")
	time.Sleep(3 * time.Second) // Extra delay before auto-save tests

	autoSaveTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!join crashuser", "joined the queue", "Add user for auto-save testing"},
		{"!queue", "crashuser", "Queue state before auto-restore"},
		{"!restoreauto", "Auto-save state has been restored", "Auto-restore (loads from auto-save file)"},
		{"!queue", "crashuser", "Queue state after auto-restore"},
	}

	for _, test := range autoSaveTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 10: Restore comparison (ISOLATED)
	fmt.Printf("\n=== TEST GROUP 10: RESTORE COMPARISON ===\n")
	fmt.Printf("This group demonstrates the difference between restore commands...\n")
	time.Sleep(3 * time.Second) // Extra delay before comparison tests

	restoreComparisonTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!join testuser1", "joined the queue", "Add user for restore comparison"},
		{"!join testuser2", "joined the queue", "Add second user for restore comparison"},
		{"!savequeue", "Queue state has been saved", "Create manual backup with 2 users"},
		{"!queue", "testuser1", "Queue state after manual backup (should have testuser1, testuser2)"},
		{"!join testuser3", "joined the queue", "Add third user (auto-saved)"},
		{"!leave testuser1", "left the queue", "Remove first user (auto-saved)"},
		{"!queue", "testuser2", "Queue state before restore comparison (should have testuser2, testuser3)"},
		{"!restorequeue", "Queue state has been restored", "Manual restore (should have testuser1, testuser2 from backup file)"},
		{"!queue", "testuser1", "Queue state after manual restore (from backup file)"},
		{"!restoreauto", "Auto-save state has been restored", "Auto-restore (should have testuser2, testuser3 from auto-save file)"},
		{"!queue", "testuser2", "Queue state after auto-restore (from auto-save file)"},
	}

	for _, test := range restoreComparisonTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	// Test Group 11: Queue control operations
	fmt.Printf("\n=== TEST GROUP 11: QUEUE CONTROL OPERATIONS ===\n")
	queueControlTests := []struct {
		command     string
		expect      string
		description string
	}{
		{"!pausequeue", "Queue is now paused", "Pause queue"},
		{"!unpausequeue", "Queue is now open again", "Unpause queue"},
		{"!endqueue", "ended the queue system", "End queue system"},
	}

	for _, test := range queueControlTests {
		totalTests++
		fmt.Printf("\n[TEST %d] Testing: %s (%s)\n", totalTests, test.command, test.description)

		if success, err := runTestWithReconnect(&conn, config, test, 5*time.Second); err != nil {
			fmt.Printf("✗ FAIL: %v\n", err)
			failed++
		} else if success {
			fmt.Printf("✓ PASS: %s\n", test.command)
			passed++
		}
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("\n=== TEST SUMMARY ===\n")
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Skipped: %d\n", skipped)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passed)/float64(totalTests)*100)

	if failed > 0 {
		fmt.Printf("\n⚠️  Some tests failed. This may be due to:\n")
		fmt.Printf("   - WebSocket connection instability\n")
		fmt.Printf("   - Bot rate limiting\n")
		fmt.Printf("   - Network issues\n")
		fmt.Printf("   - Asynchronous message processing delays\n")
	}

	fmt.Printf("\n=== BACKUP SYSTEM EXPLANATION ===\n")
	fmt.Printf("The bot uses two separate save/restore systems:\n")
	fmt.Printf("1. Auto-save: Automatically saves after every operation to queue_state_<channel>.json\n")
	fmt.Printf("2. Manual backup: Created by !savequeue command to queue_backup_<channel>.json\n")
	fmt.Printf("3. !restorequeue: Loads from manual backup file (queue_backup_<channel>.json)\n")
	fmt.Printf("4. !restoreauto: Loads from auto-save file (queue_state_<channel>.json)\n")
	fmt.Printf("5. !endqueue: Clears the auto-save file to prevent restoring an ended queue\n")
	fmt.Printf("\nThis separation prevents manual backups from being overwritten by auto-saves.\n")
	fmt.Printf("\nTest groups are separated to handle asynchronous message processing properly.\n")

	fmt.Println("\nTest run complete.")
}
