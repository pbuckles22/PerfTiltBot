package main

import (
	"fmt"
	"os"
	"time"

	twitch "github.com/gempir/go-twitch-irc/v4"
)

func main() {
	// TODO: Replace with your test bot credentials and channel
	testUsername := os.Getenv("BOT_TEST_USER") // or hardcode for now
	testOAuth := os.Getenv("BOT_TEST_OAUTH")   // format: oauth:xxxxxx
	testChannel := os.Getenv("BOT_TEST_CHANNEL")

	if testUsername == "" || testOAuth == "" || testChannel == "" {
		fmt.Println("Please set BOT_TEST_USER, BOT_TEST_OAUTH, and BOT_TEST_CHANNEL environment variables.")
		os.Exit(1)
	}

	client := twitch.NewClient(testUsername, testOAuth)
	responses := make(chan string, 10)

	client.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		if msg.Channel == testChannel {
			responses <- fmt.Sprintf("%s: %s", msg.User.Name, msg.Message)
		}
	})

	client.OnConnect(func() {
		fmt.Println("Connected to Twitch IRC as test user.")
		client.Join(testChannel)
	})

	go func() {
		err := client.Connect()
		if err != nil {
			fmt.Printf("IRC connection error: %v\n", err)
			os.Exit(1)
		}
	}()

	time.Sleep(3 * time.Second) // Wait for join

	tests := []struct {
		command string
		expect  string // substring to expect in response
	}{
		{"!ping", "Pong"},
		{"!uptime", "Bot has been running"},
		{"!help", "Available commands"},
	}

	for _, test := range tests {
		fmt.Printf("Testing %s... ", test.command)
		client.Say(testChannel, test.command)
		passed := false
		timeout := time.After(8 * time.Second)
		for !passed {
			select {
			case resp := <-responses:
				if test.expect == "" || (test.expect != "" && containsIgnoreCase(resp, test.expect)) {
					fmt.Println("PASS")
					passed = true
				}
			case <-timeout:
				fmt.Println("FAIL (timeout)")
				passed = true
			}
		}
	}

	fmt.Println("Basic health checks complete.")
	os.Exit(0)
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsIgnoreCase(s[1:], substr) || containsIgnoreCase(s, substr[1:]))
}
