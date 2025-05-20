<#
.SYNOPSIS
    PerfTiltBot Management Script

.DESCRIPTION
    This script manages PerfTiltBot instances for different Twitch channels.
    It handles building, starting, stopping, and monitoring bot instances.

.COMMANDS
    start <channel_name>
        Starts a bot instance for the specified channel.
        - Builds the Docker image if it doesn't exist
        - Copies channel-specific secrets file
        - Creates a named container with mounted volumes
        - Starts the bot with proper configuration

    stop-channel <channel_name>
        Stops and removes a specific channel's bot instance.
        - Gracefully stops the container
        - Removes the container
        - Preserves the data volume for future use

    stop-all
        Stops and removes all running bot instances.
        - Stops all perftiltbot containers
        - Removes all containers
        - Preserves all data volumes

    list
        Lists all running bot instances.
        - Shows channel names
        - Shows container names
        - Shows running status

    build
        Builds the Docker image for the bot.
        - Uses multi-stage build for smaller image size
        - Includes version tagging
        - Sets up proper environment

    list-channels <bot>
        Lists all channels using a specific bot.
        - Shows channel names
        - Shows config files

    update-bot <bot>
        Updates the shared bot configuration for a specific bot.
        - Opens the bot configuration file for editing
        - Validates the updated configuration
        - Saves the updated configuration

.EXAMPLES
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
    .\run_bot.ps1 list-channels perftiltbot

    # Update shared bot configuration
    .\run_bot.ps1 update-bot perftiltbot

.NOTES
    - Requires Docker Desktop to be running
    - Requires PowerShell 7.0 or higher
    - Channel-specific secrets files must exist in configs/
    - Each channel gets its own container and data volume
    - Bot configuration file (bot.yaml) must exist in configs/
#>

# Function to build the Docker image
function Build-Image {
    Write-Host "Building Docker image..."
    docker build -t perftiltbot .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Failed to build Docker image"
        exit 1
    }
    Write-Host "Docker image built successfully!"
}

# Function to validate configuration
function Test-Config {
    param (
        [string]$ConfigFile
    )

    $requiredFields = @("bot_name", "channel", "oauth", "client_id", "client_secret")
    $missingFields = @()

    foreach ($field in $requiredFields) {
        if (-not (Select-String -Path $ConfigFile -Pattern "^${field}:" -Quiet)) {
            $missingFields += $field
        }
    }

    if ($missingFields.Count -gt 0) {
        Write-Host "Error: Missing required fields in $ConfigFile:"
        $missingFields | ForEach-Object { Write-Host $_ }
        return $false
    }
    return $true
}

# Function to list channels using a specific bot
function Get-ChannelsByBot {
    param (
        [string]$BotName
    )

    Write-Host "Channels using bot: $BotName"
    Write-Host "----------------------------"
    
    $found = $false
    Get-ChildItem "configs/*_secrets.yaml" | ForEach-Object {
        $botName = (Select-String -Path $_.FullName -Pattern 'bot_name:' | ForEach-Object { $_.Line -replace '.*bot_name:\s*"([^"]+)".*', '$1' })
        if ($botName -eq $BotName) {
            $channel = $_.BaseName -replace '_secrets$', ''
            Write-Host "Channel: $channel"
            Write-Host "Config file: $($_.FullName)"
            Write-Host "----------------------------"
            $found = $true
        }
    }

    if (-not $found) {
        Write-Host "No channels found using bot: $BotName"
    }
}

# Function to update shared bot configuration
function Update-BotConfig {
    param (
        [string]$BotName
    )

    $botSecrets = "configs/${BotName}_secrets.yaml"
    $tempFile = "configs/temp_update.yaml"

    # Check if bot config exists
    if (-not (Test-Path $botSecrets)) {
        Write-Host "Error: Bot configuration not found: $botSecrets"
        return $false
    }

    # Create backup
    Copy-Item $botSecrets "${botSecrets}.bak"
    Write-Host "Created backup at ${botSecrets}.bak"

    # Create temporary file with current config
    Copy-Item $botSecrets $tempFile

    # Edit the temporary file
    if ($env:EDITOR) {
        & $env:EDITOR $tempFile
    } else {
        notepad $tempFile
    }

    # Validate the updated configuration
    if (Test-Config $tempFile) {
        # Update the bot config
        Move-Item $tempFile $botSecrets -Force
        Write-Host "Bot configuration updated successfully"
        
        # List affected channels
        Write-Host "Affected channels:"
        Get-ChannelsByBot $BotName
        return $true
    } else {
        Write-Host "Error: Invalid configuration. Changes not saved."
        Remove-Item $tempFile
        return $false
    }
}

# Function to start a bot for a specific channel
function Start-Bot {
    param (
        [string]$CHANNEL
    )

    $SECRETS_FILE = "configs/${CHANNEL}_secrets.yaml"
    $CONTAINER_NAME = "perftiltbot-${CHANNEL}"
    $BOT_CONFIG = "configs/bot.yaml"
    $TEMP_SECRETS = "configs/temp_secrets.yaml"

    # Check if secrets file exists
    if (-not (Test-Path $SECRETS_FILE)) {
        Write-Host "Error: Secrets file not found: $SECRETS_FILE"
        Write-Host "Please create a secrets file at: $SECRETS_FILE"
        exit 1
    }

    # Check if bot config exists
    if (-not (Test-Path $BOT_CONFIG)) {
        Write-Host "Error: Bot configuration file not found: $BOT_CONFIG"
        Write-Host "Please create a bot configuration file at: $BOT_CONFIG"
        exit 1
    }

    # Extract bot name from channel secrets
    $BOT_NAME = (Select-String -Path $SECRETS_FILE -Pattern 'bot_name:' | ForEach-Object { $_.Line -replace '.*bot_name:\s*"([^"]+)".*', '$1' })
    if (-not $BOT_NAME) {
        Write-Host "Error: bot_name not found in $SECRETS_FILE"
        exit 1
    }

    # Check if bot-specific config exists
    $BOT_SECRETS = "configs/${BOT_NAME}_secrets.yaml"
    if (Test-Path $BOT_SECRETS) {
        Write-Host "Found bot-specific configuration for $BOT_NAME"
        # Merge bot secrets with channel secrets
        Write-Host "Merging configurations..."
        # First copy bot secrets as base
        Copy-Item $BOT_SECRETS $TEMP_SECRETS -Force
        # Then merge channel-specific overrides
        yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' $TEMP_SECRETS $SECRETS_FILE > "configs/secrets.yaml"
        Remove-Item $TEMP_SECRETS
    } else {
        Write-Host "No bot-specific configuration found, using channel configuration directly"
        Copy-Item $SECRETS_FILE "configs/secrets.yaml" -Force
    }

    # Validate the final configuration
    if (-not (Test-Config "configs/secrets.yaml")) {
        Write-Host "Error: Invalid configuration after merging"
        exit 1
    }

    # Check if container is already running
    $runningContainer = docker ps -q -f "name=$CONTAINER_NAME"
    if ($runningContainer) {
        Write-Host "Container $CONTAINER_NAME is already running"
        Write-Host "Stopping and removing existing container..."
        docker stop $CONTAINER_NAME
        docker rm $CONTAINER_NAME
    }

    # Run the container
    Write-Host "Starting bot for channel: $CHANNEL"
    docker run -d `
        --name $CONTAINER_NAME `
        -v "${PWD}/configs/secrets.yaml:/app/configs/secrets.yaml" `
        -v "${PWD}/configs/bot.yaml:/app/configs/bot.yaml" `
        -v "perftiltbot-${CHANNEL}-data:/app/data" `
        perftiltbot

    Write-Host "Bot started successfully!"
    Write-Host "Container name: $CONTAINER_NAME"
    Write-Host "To view logs: docker logs $CONTAINER_NAME"
    Write-Host "To stop: docker stop $CONTAINER_NAME"
}

# Function to list all running bot instances
function List-Bots {
    Write-Host "`nRunning PerfTiltBot instances:"
    Write-Host "----------------------------"
    $containers = docker ps --format "{{.Names}}" | Where-Object { $_ -like "perftiltbot-*" }
    if ($containers) {
        foreach ($container in $containers) {
            $channel = $container -replace "perftiltbot-", ""
            Write-Host "Channel: $channel"
            Write-Host "Container: $container"
            Write-Host "Status: Running"
            Write-Host "----------------------------"
        }
    } else {
        Write-Host "No running bot instances found"
    }
}

# Function to stop all bot instances
function Stop-All-Bots {
    Write-Host "Stopping all PerfTiltBot instances..."
    $containers = docker ps -q -f "name=perftiltbot-*"
    if ($containers) {
        docker stop $containers
        docker rm $containers
        Write-Host "All bot instances stopped and removed"
    } else {
        Write-Host "No running bot instances found"
    }
}

# Function to stop a specific channel's bot instance
function Stop-Channel-Bot {
    param (
        [string]$CHANNEL
    )
    
    $CONTAINER_NAME = "perftiltbot-${CHANNEL}"
    Write-Host "Stopping bot for channel: $CHANNEL"
    
    $container = docker ps -q -f "name=$CONTAINER_NAME"
    if ($container) {
        docker stop $CONTAINER_NAME
        docker rm $CONTAINER_NAME
        Write-Host "Bot stopped and removed for channel: $CHANNEL"
    } else {
        Write-Host "No running bot instance found for channel: $CHANNEL"
    }
}

# Main script logic
if ($args.Count -eq 0) {
    Write-Host "Usage:"
    Write-Host "  .\run_bot.ps1 start <channel_name>    - Start bot for a channel"
    Write-Host "  .\run_bot.ps1 stop-channel <channel>  - Stop bot for a specific channel"
    Write-Host "  .\run_bot.ps1 build                   - Build Docker image"
    Write-Host "  .\run_bot.ps1 list                    - List running bot instances"
    Write-Host "  .\run_bot.ps1 stop-all               - Stop all bot instances"
    Write-Host "  .\run_bot.ps1 list-channels <bot>    - List channels using a specific bot"
    Write-Host "  .\run_bot.ps1 update-bot <bot>       - Update shared bot configuration"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\run_bot.ps1 start pbuckles"
    Write-Host "  .\run_bot.ps1 stop-channel pbuckles"
    Write-Host "  .\run_bot.ps1 build"
    Write-Host "  .\run_bot.ps1 list-channels perftiltbot"
    Write-Host "  .\run_bot.ps1 update-bot perftiltbot"
    Write-Host ""
    Write-Host "Shortcut:"
    Write-Host "  .\run_bot.ps1 <channel_name>         - Same as 'start <channel_name>'"
    exit 1
}

$command = $args[0]

# If only one argument is provided and it's not a known command, treat it as a channel name
if ($args.Count -eq 1 -and $command -notin @("start", "stop-channel", "build", "list", "stop-all", "list-channels", "update-bot")) {
    # Check if image exists, build if it doesn't
    $imageExists = docker images -q perftiltbot
    if (-not $imageExists) {
        Build-Image
    }
    Start-Bot -CHANNEL $command
    exit 0
}

switch ($command) {
    "start" {
        if ($args.Count -lt 2) {
            Write-Host "Error: Channel name required for start command"
            Write-Host "Usage: .\run_bot.ps1 start <channel_name>"
            exit 1
        }
        # Check if image exists, build if it doesn't
        $imageExists = docker images -q perftiltbot
        if (-not $imageExists) {
            Build-Image
        }
        Start-Bot -CHANNEL $args[1]
    }
    "stop-channel" {
        if ($args.Count -lt 2) {
            Write-Host "Error: Channel name required for stop-channel command"
            Write-Host "Usage: .\run_bot.ps1 stop-channel <channel_name>"
            exit 1
        }
        Stop-Channel-Bot -CHANNEL $args[1]
    }
    "build" {
        Build-Image
    }
    "list" {
        List-Bots
    }
    "stop-all" {
        Stop-All-Bots
    }
    "list-channels" {
        if ($args.Count -lt 2) {
            Write-Host "Error: Bot name required for list-channels command"
            Write-Host "Usage: .\run_bot.ps1 list-channels <bot_name>"
            exit 1
        }
        Get-ChannelsByBot -BotName $args[1]
    }
    "update-bot" {
        if ($args.Count -lt 2) {
            Write-Host "Error: Bot name required for update-bot command"
            Write-Host "Usage: .\run_bot.ps1 update-bot <bot_name>"
            exit 1
        }
        Update-BotConfig -BotName $args[1]
    }
    default {
        Write-Host "Error: Unknown command '$command'"
        Write-Host "Run .\run_bot.ps1 without arguments to see usage"
        exit 1
    }
} 