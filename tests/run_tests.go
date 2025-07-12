package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var testType = flag.String("type", "all", "Test type: all, unit, integration, websocket")
	var verbose = flag.Bool("v", false, "Verbose output")
	flag.Parse()

	// Get the project root directory
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Printf("Error finding project root: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Running tests from: %s\n", projectRoot)
	fmt.Printf("Test type: %s\n", *testType)

	switch strings.ToLower(*testType) {
	case "all":
		runAllTests(projectRoot, *verbose)
	case "unit":
		runUnitTests(projectRoot, *verbose)
	case "integration":
		runIntegrationTests(projectRoot, *verbose)
	case "websocket":
		runWebsocketTests(projectRoot, *verbose)
	default:
		fmt.Printf("Unknown test type: %s\n", *testType)
		fmt.Println("Available types: all, unit, integration, websocket")
		os.Exit(1)
	}
}

func findProjectRoot() (string, error) {
	// Start from current directory and look for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod file")
		}
		dir = parent
	}
}

func runAllTests(projectRoot string, verbose bool) {
	fmt.Println("\n=== Running All Tests ===")

	// Run unit tests
	fmt.Println("\n--- Unit Tests ---")
	runUnitTests(projectRoot, verbose)

	// Run integration tests (if they exist)
	fmt.Println("\n--- Integration Tests ---")
	runIntegrationTests(projectRoot, verbose)

	// Run websocket tests (if they exist)
	fmt.Println("\n--- WebSocket Tests ---")
	runWebsocketTests(projectRoot, verbose)
}

func runUnitTests(projectRoot string, verbose bool) {
	args := []string{"test", "./tests/unit/..."}
	if verbose {
		args = append(args, "-v")
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running: go %s\n", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		fmt.Printf("Unit tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Unit tests passed!")
}

func runIntegrationTests(projectRoot string, verbose bool) {
	// Check if integration tests exist
	integrationDir := filepath.Join(projectRoot, "tests", "integration")
	if _, err := os.Stat(integrationDir); os.IsNotExist(err) {
		fmt.Println("No integration tests found")
		return
	}

	args := []string{"test", "./tests/integration/..."}
	if verbose {
		args = append(args, "-v")
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running: go %s\n", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		fmt.Printf("Integration tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Integration tests passed!")
}

func runWebsocketTests(projectRoot string, verbose bool) {
	// Check if websocket tests exist
	websocketDir := filepath.Join(projectRoot, "tests", "websocket")
	if _, err := os.Stat(websocketDir); os.IsNotExist(err) {
		fmt.Println("No websocket tests found")
		return
	}

	args := []string{"test", "./tests/websocket/..."}
	if verbose {
		args = append(args, "-v")
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Running: go %s\n", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		fmt.Printf("WebSocket tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("WebSocket tests passed!")
}
