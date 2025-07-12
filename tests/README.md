# PerfTiltBot Tests

This directory contains comprehensive tests for the PerfTiltBot Twitch chat bot.

## Test Structure

```
tests/
├── unit/           # Unit tests for individual components
│   ├── queue_test.go      # Queue package tests
│   └── commands_test.go   # Command handler tests
├── integration/    # Integration tests (future)
├── websocket/      # WebSocket tests (future)
├── run_tests.go    # Test runner script
└── README.md       # This file
```

## Test Categories

### Unit Tests (`tests/unit/`)
- **queue_test.go**: Tests for the queue package functionality
  - Queue creation and state management
  - Adding/removing users
  - Position tracking
  - Pause/unpause functionality
  - State persistence
- **commands_test.go**: Tests for command handlers
  - All command responses
  - Permission checking
  - Error handling
  - Message formatting

### Integration Tests (`tests/integration/`)
- Tests that verify multiple components work together
- End-to-end command processing
- Configuration loading and validation
- Database/file system interactions

### WebSocket Tests (`tests/websocket/`)
- Real Twitch IRC connection tests
- Message sending/receiving
- Connection stability
- Rate limiting behavior

## Running Tests

### Using the Test Runner

The easiest way to run tests is using the test runner script:

```bash
# Run all tests
go run tests/run_tests.go

# Run only unit tests
go run tests/run_tests.go -type unit

# Run with verbose output
go run tests/run_tests.go -type unit -v

# Run specific test categories
go run tests/run_tests.go -type integration
go run tests/run_tests.go -type websocket
```

### Using Go Test Directly

You can also run tests using the standard Go test command:

```bash
# Run all tests in the tests directory
go test ./tests/...

# Run only unit tests
go test ./tests/unit/...

# Run with verbose output
go test -v ./tests/unit/...

# Run a specific test file
go test ./tests/unit/queue_test.go

# Run a specific test function
go test -run TestHandlePing ./tests/unit/commands_test.go
```

### Test Options

- `-v`: Verbose output showing each test case
- `-count=1`: Disable test caching
- `-timeout=30s`: Set test timeout
- `-race`: Enable race condition detection

## Writing Tests

### Unit Test Guidelines

1. **Isolation**: Each test should be independent and not rely on other tests
2. **Cleanup**: Use `t.TempDir()` for temporary files
3. **Naming**: Use descriptive test names that explain what is being tested
4. **Coverage**: Test both success and failure cases
5. **Mocking**: Use mocks for external dependencies

### Example Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Setup
    tempDir := t.TempDir()
    q := queue.NewQueue(tempDir, "testchannel")
    
    // Test case 1: Success scenario
    t.Run("success_case", func(t *testing.T) {
        err := q.Add("user1", false)
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
    })
    
    // Test case 2: Error scenario
    t.Run("error_case", func(t *testing.T) {
        err := q.Add("user1", false) // Duplicate
        if err == nil {
            t.Error("Expected error for duplicate user")
        }
    })
}
```

### Integration Test Guidelines

1. **Real Dependencies**: Use actual file system and network connections
2. **Test Data**: Create realistic test data
3. **Cleanup**: Ensure proper cleanup of test resources
4. **Timeouts**: Set appropriate timeouts for network operations

### WebSocket Test Guidelines

1. **Rate Limiting**: Respect Twitch's rate limits
2. **Connection Management**: Handle connection failures gracefully
3. **Authentication**: Use test credentials when possible
4. **Isolation**: Each test should use a separate connection

## Test Configuration

### Environment Variables

Some tests may require environment variables:

```bash
# For WebSocket tests
export TWITCH_TEST_USERNAME="your_test_username"
export TWITCH_TEST_OAUTH="your_test_oauth_token"
export TWITCH_TEST_CHANNEL="your_test_channel"

# For integration tests
export TEST_DATA_DIR="/path/to/test/data"
```

### Test Data

Test data should be stored in `tests/data/` and referenced by tests as needed. This includes:
- Configuration files
- Sample queue states
- Mock Twitch messages

## Continuous Integration

Tests are automatically run in CI/CD pipelines:

1. **Unit Tests**: Run on every commit
2. **Integration Tests**: Run on pull requests
3. **WebSocket Tests**: Run on main branch only (due to rate limits)

## Coverage

To generate test coverage reports:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./tests/...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out
```

## Troubleshooting

### Common Issues

1. **Import Path Errors**: Ensure you're running tests from the project root
2. **Permission Errors**: Check file permissions for test directories
3. **Network Timeouts**: Increase timeout values for slow connections
4. **Rate Limiting**: Add delays between WebSocket tests

### Debug Mode

Enable debug output for tests:

```bash
# Set debug environment variable
export TEST_DEBUG=1

# Run tests with debug output
go test -v ./tests/unit/...
```

## Contributing

When adding new tests:

1. Follow the existing test structure and naming conventions
2. Add tests for both success and failure cases
3. Update this README if adding new test categories
4. Ensure tests pass in CI/CD environment
5. Add appropriate documentation for complex test scenarios 