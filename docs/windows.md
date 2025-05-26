# Windows Setup Guide

This guide covers setting up and running PBChatBot on Windows.

## Prerequisites

1. **PowerShell 7.0 or higher**
   ```powershell
   # Install using Scoop
   scoop install pwsh
   ```

2. **Docker Desktop for Windows**
   - Download from [Docker's website](https://www.docker.com/products/docker-desktop)
   - Ensure WSL 2 backend is enabled
   - Start Docker Desktop before running the bot

3. **Scoop Package Manager** (recommended)
   ```powershell
   # Install Scoop
   Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
   irm get.scoop.sh | iex
   ```

## Installation

1. **Clone the repository**
   ```powershell
   git clone https://github.com/pbuckles22/PBChatBot.git
   cd PBChatBot
   ```

2. **Install dependencies using Scoop**
   ```powershell
   scoop install git go docker
   ```

## Configuration

1. **Create bot authentication file**
   - Create `configs/<bot_name>_auth_secrets.yaml`
   - See `configs/examples/README.md` for format

2. **Create channel configuration**
   - Create `configs/<channel>_config_secrets.yaml`
   - See `configs/examples/README.md` for format

## Running the Bot

The bot can be managed using the PowerShell script `run_bot.ps1`:

```powershell
# Start a bot for a channel
.\run_bot.ps1 start pbuckles

# Stop a specific channel's bot
.\run_bot.ps1 stop-channel pbuckles

# List all running bots
.\run_bot.ps1 list

# Stop all bots
.\run_bot.ps1 stop-all

# Build the Docker image
.\run_bot.ps1 build

# List channels using a specific bot
.\run_bot.ps1 list-channels pbchatbot

# Update shared bot configuration
.\run_bot.ps1 update-bot pbchatbot

# Restart all bots
.\run_bot.ps1 restart-all
```

## Windows-Specific Features

1. **PowerShell Integration**
   - Uses native PowerShell commands for better Windows integration
   - Supports Windows path formats
   - Uses Windows-style line endings

2. **Editor Integration**
   - Uses `notepad` as default editor if `$env:EDITOR` is not set
   - Supports Windows text editors

3. **Docker Desktop Integration**
   - Uses Docker Desktop's Windows integration
   - Supports WSL 2 backend
   - Uses Windows-style volume mounts

## Troubleshooting

1. **Docker Issues**
   - Ensure Docker Desktop is running
   - Check WSL 2 backend is enabled
   - Verify Docker service is running

2. **PowerShell Issues**
   - Run PowerShell as Administrator if needed
   - Check execution policy: `Get-ExecutionPolicy`
   - Set execution policy if needed: `Set-ExecutionPolicy RemoteSigned -Scope CurrentUser`

3. **Path Issues**
   - Use PowerShell's `Join-Path` for path construction
   - Use `$PWD` for current directory
   - Use forward slashes in paths

## Scoop Integration

The bot can be installed and managed using Scoop:

```powershell
# Add the bucket (if not already added)
scoop bucket add pbchatbot https://github.com/pbuckles22/PBChatBot

# Install the bot
scoop install pbchatbot

# Update the bot
scoop update pbchatbot

# Uninstall the bot
scoop uninstall pbchatbot
```

## Development on Windows

1. **IDE Setup**
   - VS Code with Go extension
   - PowerShell extension for script editing
   - Docker extension for container management

2. **Testing**
   ```powershell
   # Run tests
   go test ./...

   # Run specific test
   go test ./internal/commands
   ```

3. **Building**
   ```powershell
   # Build for Windows
   go build -o bot.exe cmd/bot/main.go

   # Build for Linux (Docker)
   .\run_bot.ps1 build
   ``` 