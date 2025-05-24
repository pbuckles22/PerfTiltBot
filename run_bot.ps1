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

    restart-all
        Stops and restarts all running bot instances with the latest Docker image.

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

    # Restart all bots
    .\run_bot.ps1 restart-all

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
    docker build -t pbchatbot .
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
        Write-Host "Error: Missing required fields in ${ConfigFile}:"
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
    Get-ChildItem "configs/*_config_secrets.yaml" | ForEach-Object {
        $botName = (Select-String -Path $_.FullName -Pattern 'bot_name:' | ForEach-Object { $_.Line -replace '.*bot_name:\s*"([^"]+)".*', '$1' })
        if ($botName -eq $BotName) {
            $channel = $_.BaseName -replace '_config_secrets$', ''
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

    $botAuth = "configs/${BotName}_auth_secrets.yaml"
    $tempFile = "configs/temp_update.yaml"

    # Check if bot config exists
    if (-not (Test-Path $botAuth)) {
        Write-Host "Error: Bot authentication file not found: $botAuth"
        return $false
    }

    # Create backup
    Copy-Item $botAuth "${botAuth}.bak"
    Write-Host "Created backup at ${botAuth}.bak"

    # Create temporary file with current config
    Copy-Item $botAuth $tempFile

    # Edit the temporary file
    if ($env:EDITOR) {
        & $env:EDITOR $tempFile
    } else {
        notepad $tempFile
    }

    # Validate the updated configuration
    if (Test-Config $tempFile) {
        # Update the bot config
        Move-Item $tempFile $botAuth -Force
        Write-Host "Bot authentication updated successfully"
        
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

    $CHANNEL_CONFIG = Join-Path "configs" "${CHANNEL}_config_secrets.yaml"
    $BOT_CONFIG = Join-Path "configs" "bot.yaml"
    $TEMP_SECRETS = Join-Path "configs" "temp_secrets.yaml"
    $FINAL_SECRETS = Join-Path "configs" "secrets.yaml"

    # Check if channel config exists
    if (-not (Test-Path $CHANNEL_CONFIG)) {
        Write-Host "Error: Channel configuration file not found: $CHANNEL_CONFIG"
        Write-Host "Please create a channel configuration file at: $CHANNEL_CONFIG"
        exit 1
    }

    # Check if bot config exists
    if (-not (Test-Path $BOT_CONFIG)) {
        Write-Host "Error: Bot configuration file not found: $BOT_CONFIG"
        Write-Host "Please create a bot configuration file at: $BOT_CONFIG"
        exit 1
    }

    # Extract bot name from channel config, preserving case
    $BOT_NAME = (Select-String -Path $CHANNEL_CONFIG -Pattern 'bot_name:' | ForEach-Object { $_.Line -replace '.*bot_name:\s*"([^"]+)".*', '$1' })
    if (-not $BOT_NAME) {
        Write-Host "Error: bot_name not found in $CHANNEL_CONFIG"
        exit 1
    }

    # Extract channel name from config, preserving case
    $CHANNEL_NAME = (Select-String -Path $CHANNEL_CONFIG -Pattern 'channel:' | ForEach-Object { $_.Line -replace '.*channel:\s*"([^"]+)".*', '$1' })
    if (-not $CHANNEL_NAME) {
        Write-Host "Error: channel not found in $CHANNEL_CONFIG"
        exit 1
    }

    # Create container name using exact case from configs
    $CONTAINER_NAME = "${BOT_NAME}-${CHANNEL_NAME}"

    # Check if bot auth exists
    $BOT_AUTH = Join-Path "configs" "${BOT_NAME}_auth_secrets.yaml"
    if (Test-Path $BOT_AUTH) {
        Write-Host "Found bot authentication for $BOT_NAME"
        # Merge bot auth with channel config
        Write-Host "Merging configurations..."
        # First copy bot auth as base
        Copy-Item $BOT_AUTH $TEMP_SECRETS -Force
        # Then merge channel-specific overrides
        yq eval-all "select(fileIndex == 0) * select(fileIndex == 1)" $TEMP_SECRETS $CHANNEL_CONFIG > $FINAL_SECRETS
        Remove-Item $TEMP_SECRETS
        # Explicitly set the channel field to ensure it is present
        yq eval ".channel = `"$CHANNEL_NAME`"" -i $FINAL_SECRETS
        # Restructure the YAML to match the expected format in the bot code
        yq eval '.twitch = {"bot_token": .oauth, "client_id": .client_id, "client_secret": .client_secret, "refresh_token": .refresh_token, "bot_username": .bot_name, "channel": .channel, "data_path": .data_path}' -i $FINAL_SECRETS
    } else {
        Write-Host "Error: Bot authentication file not found: $BOT_AUTH"
        exit 1
    }

    # Validate the final configuration
    if (-not (Test-Config $FINAL_SECRETS)) {
        Write-Host "Error: Invalid configuration after merging"
        exit 1
    }

    # Check if container is already running or exists
    $existingContainer = docker ps -a -q -f "name=$CONTAINER_NAME"
    if ($existingContainer) {
        Write-Host "Container $CONTAINER_NAME already exists. Stopping and removing it..."
        docker stop $CONTAINER_NAME
        docker rm $CONTAINER_NAME
    }

    # Run the container
    Write-Host "Starting bot for channel: $CHANNEL_NAME"
    docker run -d `
        --name $CONTAINER_NAME `
        -v "${PWD}/configs/secrets.yaml:/app/configs/secrets.yaml" `
        -v "${PWD}/configs/bot.yaml:/app/configs/bot.yaml" `
        -v "${BOT_NAME}-${CHANNEL_NAME}-data:/app/data" `
        pbchatbot

    Write-Host "Bot started successfully!"
    Write-Host "Container name: $CONTAINER_NAME"
    Write-Host "To view logs: docker logs $CONTAINER_NAME"
    Write-Host "To stop: docker stop $CONTAINER_NAME"
}

# Function to list all running bot instances
function List-Bots {
    Write-Host "`nRunning PerfTiltBot instances:"
    Write-Host "----------------------------"
    $containers = docker ps --format "{{.Names}}"
    if ($containers) {
        foreach ($container in $containers) {
            # Extract bot name and channel from container name
            if ($container -match "(.+)-(.+)") {
                $botName = $matches[1]
                $channel = $matches[2]
                Write-Host "Bot: $botName"
                Write-Host "Channel: $channel"
                Write-Host "Container: $container"
                Write-Host "Status: Running"
                Write-Host "----------------------------"
            }
        }
    } else {
        Write-Host "No running bot instances found"
    }
}

# Function to stop all bot instances
function Stop-All-Bots {
    Write-Host "Stopping all bot instances..."
    $containers = docker ps -a --format "{{.Names}}"
    if ($containers) {
        foreach ($container in $containers) {
            if ($container -match "(.+)-(.+)") {
                Write-Host "Stopping container: $container"
                docker stop $container
                docker rm $container
            }
        }
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
    
    # Get the bot name from the channel config
    $CHANNEL_CONFIG = Join-Path "configs" "${CHANNEL}_config_secrets.yaml"
    if (-not (Test-Path $CHANNEL_CONFIG)) {
        Write-Host "Error: Channel configuration file not found: $CHANNEL_CONFIG"
        exit 1
    }

    # Extract bot name from channel config, preserving case
    $BOT_NAME = (Select-String -Path $CHANNEL_CONFIG -Pattern 'bot_name:' | ForEach-Object { $_.Line -replace '.*bot_name:\s*"([^"]+)".*', '$1' })
    if (-not $BOT_NAME) {
        Write-Host "Error: bot_name not found in $CHANNEL_CONFIG"
        exit 1
    }

    $CONTAINER_NAME = "${BOT_NAME}-${CHANNEL}"
    Write-Host "Stopping bot for channel: $CHANNEL"
    
    $container = docker ps -a -q -f "name=$CONTAINER_NAME"
    if ($container) {
        docker stop $CONTAINER_NAME
        docker rm $CONTAINER_NAME
        Write-Host "Bot stopped and removed for channel: $CHANNEL"
    } else {
        Write-Host "No running bot instance found for channel: $CHANNEL"
    }
}

# Function to restart all bots
function Restart-AllBots {
    Write-Host "Starting bot restart process..."
    $configs = Get-ChildItem -Path "configs" -Filter "*_config_secrets.yaml"
    
    foreach ($config in $configs) {
        $channel = $config.BaseName -replace "_config_secrets$", ""
        Write-Host "`nProcessing channel: $channel"
        
        # Stop the specific channel
        Write-Host "Stopping bot for channel: $channel"
        Stop-Channel-Bot -CHANNEL $channel
        
        # Start the channel
        Write-Host "Starting bot for channel: $channel"
        Start-Bot -CHANNEL $channel
        
        # Wait a moment to ensure the bot is up before moving to the next
        Start-Sleep -Seconds 2
    }
    
    Write-Host "`nAll bots have been restarted successfully!"
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
    Write-Host "  .\run_bot.ps1 restart-all            - Stop and restart all bots with latest image"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\run_bot.ps1 start pbuckles"
    Write-Host "  .\run_bot.ps1 stop-channel pbuckles"
    Write-Host "  .\run_bot.ps1 build"
    Write-Host "  .\run_bot.ps1 list-channels perftiltbot"
    Write-Host "  .\run_bot.ps1 update-bot perftiltbot"
    Write-Host "  .\run_bot.ps1 restart-all"
    Write-Host ""
    Write-Host "Shortcut:"
    Write-Host "  .\run_bot.ps1 <channel_name>         - Same as 'start <channel_name>'"
    exit 1
}

$command = $args[0]
$channel = $args[1]

# If only one argument is provided and it's not a known command, treat it as a channel name
if ($args.Count -eq 1 -and $command -notin @("start", "stop-channel", "build", "list", "stop-all", "list-channels", "update-bot", "restart-all")) {
    # Check if image exists, build if it doesn't
    $imageExists = docker images -q pbchatbot
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
        $imageExists = docker images -q pbchatbot
        if (-not $imageExists) {
            Build-Image
        }
        Start-Bot -CHANNEL $channel
    }
    "stop-channel" {
        if ($args.Count -lt 2) {
            Write-Host "Error: Channel name required for stop-channel command"
            Write-Host "Usage: .\run_bot.ps1 stop-channel <channel_name>"
            exit 1
        }
        Stop-Channel-Bot -CHANNEL $channel
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
        Get-ChannelsByBot -BotName $channel
    }
    "update-bot" {
        if ($args.Count -lt 2) {
            Write-Host "Error: Bot name required for update-bot command"
            Write-Host "Usage: .\run_bot.ps1 update-bot <bot_name>"
            exit 1
        }
        Update-BotConfig -BotName $channel
    }
    "restart-all" {
        Restart-AllBots
    }
    default {
        Write-Host "Error: Unknown command '$command'"
        Write-Host "Run .\run_bot.ps1 without arguments to see usage"
        exit 1
    }
} 