package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting PerfTiltBot...")

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-sigChan

	log.Println("Shutting down gracefully...")
}
