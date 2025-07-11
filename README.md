# PerfTiltBot

A Twitch chat bot for queue management with automatic token refresh and timezone support.

## Documentation

- [Windows Setup Guide](docs/windows.md) - Windows-specific setup and usage instructions
- [AWS Deployment Guide](docs/aws.md) - AWS deployment and management instructions
- [Configuration Examples](configs/examples/README.md) - Example configuration files and setup

## Prerequisites

1. **Go 1.23 or higher**
   ```bash
   # Install Go
   go version
   ```

2. **Docker**
   - Docker Desktop (Windows/macOS)
   - Docker Engine (Linux)

## Features

- **Queue Management**: Join, leave, move, and manage user queues
- **Auto-Save**: Queue state automatically saved after every modification
- **Timezone Support**: Configurable timezone for user messages, consistent PST logging
- **Token Auto-Refresh**: Automatic OAuth token refresh with configurable intervals
- **Permission System**: Different command access for regular users, VIPs, moderators, and broadcasters
- **Cooldown Management**: Configurable cooldowns per user type
- **Multi-Channel Support**: Run multiple channels with different configurations

## Recent Updates

### v2.0.0
- **Auto-Save Queue State**: Queue is automatically saved after every modification (add, remove, move, pop, etc.)
- **Timezone Configuration**: 
  - Debug logs always in PST for consistency
  - User-facing messages use configurable timezone (defaults to EST)
- **Improved Token Refresh**: Fixed panic issues and improved logging
- **Cleaner Logs**: Simplified debug output with better formatting

## Quick Start

1. **Build the Docker image**:
   ```bash
   ./run_bot.sh build
   ```

2. **Start a bot for a channel**:
   ```bash
   ./run_bot.sh start <channel_name>
   ```

3. **View running bots**:
   ```bash
   ./run_bot.sh list
   ```

## Configuration

### Channel Configuration
Create a channel config file: `configs/channels/<channel>_config_secrets.yaml`

```yaml
bot_name: "mybot"
channel: "PerfectTilt"
data_path: "/app/data"
timezone: "America/New_York"  # Optional: defaults to EST

commands:
  queue:
    max_size: 100
    default_position: 1
    default_pop_count: 1
  cooldowns:
    default: 5
    moderator: 2
    vip: 3
```

### Bot Authentication
Create a bot auth file: `configs/bots/<bot_name>_auth_secrets.yaml`

**How to get your client_id, client_secret, access token, and refresh token:**
See [docs/twitch_bot_oauth.md](docs/twitch_bot_oauth.md) for a step-by-step guide.

```yaml
bot_name: "mybot"
oauth: "oauth:your_bot_oauth_token"
client_id: "your_bot_client_id"
client_secret: "your_bot_client_secret"
refresh_token: "your_refresh_token"
```

## Commands

See [Commands Documentation](docs/commands.md) for a complete list of available commands.

### Basic Commands
- `!help` - Show available commands
- `!join` - Join the queue
- `!queue` - Show current queue
- `!position` - Show your position

### Moderator Commands
- `!startqueue` - Start queue system
- `!pop` - Remove users from queue
- `!move` - Move users in queue
- `!savequeue` - Save queue state

## Development

### Building
```bash
# Build Docker image
./run_bot.sh build

# Run locally
go run cmd/bot/main.go
```

### Testing
```bash
# Run tests
go test ./...
```

## Docker Commands

```bash
# Start bot for channel
./run_bot.sh start <channel>

# Stop specific channel
./run_bot.sh stop-channel <channel>

# Stop all bots
./run_bot.sh stop-all

# List running bots
./run_bot.sh list

# Restart all bots
./run_bot.sh restart-all
```

## Architecture

- **Queue System**: Persistent queue with auto-save functionality
- **Token Management**: Automatic OAuth refresh with configurable intervals
- **Command System**: Modular command handlers with permission checks
- **Timezone Support**: Separate timezone handling for logs vs user display
- **Configuration**: YAML-based configuration per channel

## Security

- OAuth tokens are automatically refreshed
- Channel-specific data isolation
- Permission-based command access
- Secure configuration file handling

## Security Notes

- The bot's Docker image is regularly rebuilt with the latest Alpine and Go security patches.
- If a CVE is present in Alpine or Go, it may take a few days for the fix to appear in the official Alpine repositories.
- To minimize risk, always rebuild your image after a security fix is published.
- For critical CVEs, you may use Alpine edge/testing repositories, but this is not recommended for production due to stability concerns.
- See the Changelog for recent security-related updates.

## License

See [LICENSE](LICENSE) file for details.

## Authentication & OAuth

For step-by-step instructions on obtaining a Twitch OAuth access token and refresh token for your bot, see [docs/twitch_bot_oauth.md](docs/twitch_bot_oauth.md). 