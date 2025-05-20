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

.NOTES
    - Requires Docker Desktop to be running
    - Requires PowerShell 7.0 or higher
    - Channel-specific secrets files must exist in configs/
    - Each channel gets its own container and data volume
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

# Function to start a bot for a specific channel
function Start-Bot {
    param (
        [string]$CHANNEL
    )

    $SECRETS_FILE = "configs/${CHANNEL}_secrets.yaml"
    $CONTAINER_NAME = "perftiltbot-${CHANNEL}"

    # Check if secrets file exists
    if (-not (Test-Path $SECRETS_FILE)) {
        Write-Host "Error: Secrets file not found: $SECRETS_FILE"
        Write-Host "Please create a secrets file at: $SECRETS_FILE"
        exit 1
    }

    # Copy channel-specific secrets to secrets.yaml
    Write-Host "Setting up configuration for channel: $CHANNEL"
    Copy-Item $SECRETS_FILE "configs/secrets.yaml" -Force

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
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\run_bot.ps1 start pbuckles"
    Write-Host "  .\run_bot.ps1 stop-channel pbuckles"
    Write-Host "  .\run_bot.ps1 build"
    Write-Host ""
    Write-Host "Shortcut:"
    Write-Host "  .\run_bot.ps1 <channel_name>         - Same as 'start <channel_name>'"
    exit 1
}

$command = $args[0]

# If only one argument is provided and it's not a known command, treat it as a channel name
if ($args.Count -eq 1 -and $command -notin @("start", "stop-channel", "build", "list", "stop-all")) {
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
    default {
        Write-Host "Error: Unknown command '$command'"
        Write-Host "Run .\run_bot.ps1 without arguments to see usage"
        exit 1
    }
} 