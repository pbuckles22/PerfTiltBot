# PBChatBot Management Script
#
# This script manages PBChatBot instances for different Twitch channels.
# It handles building, starting, stopping, and monitoring bot instances.
#
# Commands:
#   start <channel_name>    - Start bot for a channel
#   stop-channel <channel>  - Stop bot for a specific channel
#   stop-all               - Stop all bot instances
#   list                   - List running bot instances
#   build                  - Build Docker image
#   list-channels <bot>    - List channels using a specific bot
#   update-bot <bot>        - Update shared bot configuration
#
# Examples:
#   .\run_bot.ps1 start pbuckles
#   .\run_bot.ps1 stop-channel pbuckles
#   .\run_bot.ps1 list
#   .\run_bot.ps1 stop-all
#   .\run_bot.ps1 build
#   .\run_bot.ps1 list-channels pbchatbot
#   .\run_bot.ps1 update-bot pbchatbot
#
# Shortcut:
#   .\run_bot.ps1 <channel_name>  - Same as 'start <channel_name>'
#
# Notes:
#   - Requires Docker Desktop to be running
#   - Channel-specific secrets files must exist in configs/
#   - Each channel gets its own container and data volume
#   - Uses Windows-style paths and PowerShell commands

# Function to build the Docker image
function Build-Image {
    Write-Host "Building Docker image..."
    docker build -t pbchatbot .
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error: Failed to build Docker image" -ForegroundColor Red
        exit 1
    }
    Write-Host "Docker image built successfully!" -ForegroundColor Green
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
        Write-Host "Error: Missing required fields in ${ConfigFile}:" -ForegroundColor Red
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
    
    $found = $false
    Write-Host "Channels using bot: $BotName"
    Write-Host "----------------------------"
    
    Get-ChildItem -Path "configs\*_config_secrets.yaml" | ForEach-Object {
        $botName = (Select-String -Path $_.FullName -Pattern 'bot_name:' | Select-Object -First 1).Line -replace '.*bot_name:\s*"([^"]+)".*', '$1'
        if ($botName -eq $BotName) {
            $channel = $_.BaseName -replace '_config_secrets$', ''
            Write-Host "Channel: $channel"
            Write-Host "Config file: $($_.FullName)"
            Write-Host "----------------------------"
            $found = $true
        }
    }
    
    if (-not $found) {
        Write-Host "No channels found using bot: $BotName" -ForegroundColor Yellow
    }
}

# Function to update shared bot configuration
function Update-BotConfig {
    param (
        [string]$BotName
    )
    
    $botSecrets = "configs\${BotName}_secrets.yaml"
    $tempFile = "configs\temp_update.yaml"
    
    if (-not (Test-Path $botSecrets)) {
        Write-Host "Error: Bot configuration not found: $botSecrets" -ForegroundColor Red
        return
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
        Move-Item -Force $tempFile $botSecrets
        Write-Host "Bot configuration updated successfully" -ForegroundColor Green
        
        # List affected channels
        Write-Host "Affected channels:"
        Get-ChannelsByBot $BotName
    } else {
        Write-Host "Error: Invalid configuration. Changes not saved." -ForegroundColor Red
        Remove-Item $tempFile
    }
}

# Function to start a bot for a specific channel
function Start-Bot {
    param (
        [string]$Channel
    )
    
    $channelConfig = "configs\${Channel}_config_secrets.yaml"
    
    if (-not (Test-Path $channelConfig)) {
        Write-Host "Error: Channel configuration file not found: ${channelConfig}" -ForegroundColor Red
        Write-Host "Please create a channel configuration file at: ${channelConfig}"
        exit 1
    }
    
    # Extract bot name from channel config, preserving case
    $botName = (Select-String -Path $channelConfig -Pattern 'bot_name:' | Select-Object -First 1).Line -replace '.*bot_name:\s*"([^"]+)".*', '$1'
    if (-not $botName) {
        Write-Host "Error: bot_name not found in ${channelConfig}" -ForegroundColor Red
        exit 1
    }
    
    # Extract channel name from config, preserving case
    $channelName = (Select-String -Path $channelConfig -Pattern 'channel:' | Select-Object -First 1).Line -replace '.*channel:\s*"([^"]+)".*', '$1'
    if (-not $channelName) {
        Write-Host "Error: channel not found in ${channelConfig}" -ForegroundColor Red
        exit 1
    }
    
    # Create container name using exact case from configs
    $containerName = "${botName}-${channelName}"
    
    # Check if bot auth exists
    $botAuth = "configs\${botName}_auth_secrets.yaml"
    if (-not (Test-Path $botAuth)) {
        Write-Host "Error: Bot authentication file not found: ${botAuth}" -ForegroundColor Red
        exit 1
    }
    
    # Check if container is already running or exists
    if (docker ps -a -q -f name=$containerName) {
        Write-Host "Container ${containerName} already exists. Stopping and removing it..."
        docker stop $containerName
        docker rm $containerName
    }
    
    # Run the container
    Write-Host "Starting bot for channel: ${channelName}"
    
    # Convert Windows paths to Docker-compatible format and ensure they exist
    $botAuthPath = (Resolve-Path $botAuth).Path.Replace('\', '/')
    $channelConfigPath = (Resolve-Path $channelConfig).Path.Replace('\', '/')
    
    # Create volume mounts with proper syntax
    $volumeMounts = @(
        "-v", "${botAuthPath}:/app/configs/${botName}_auth_secrets.yaml",
        "-v", "${channelConfigPath}:/app/configs/${channelName}_config_secrets.yaml:ro",
        "-v", "${botName}-${channelName}-data:/app/data"
    )
    
    # Build the Docker command as an array to avoid PowerShell parsing issues
    $dockerArgs = @(
        "run",
        "-d",
        "--name", $containerName,
        "-e", "CHANNEL_NAME=${Channel}",
        "-e", "BOT_NAME=${botName}"
    )
    
    # Add volume mounts
    $dockerArgs += $volumeMounts
    
    # Add the image name
    $dockerArgs += "pbchatbot"
    
    # Execute the Docker command
    try {
        Write-Host "Running Docker command with arguments:"
        Write-Host ($dockerArgs -join ' ')
        docker $dockerArgs
        Write-Host "Bot started successfully!" -ForegroundColor Green
        Write-Host "Container name: ${containerName}"
        Write-Host "To view logs: docker logs ${containerName}"
        Write-Host "To stop: docker stop ${containerName}"
    }
    catch {
        Write-Host "Error starting bot: $_" -ForegroundColor Red
        exit 1
    }
}

# Function to list all running bot instances
function Get-Bots {
    Write-Host "`nRunning bot instances:"
    Write-Host "----------------------------"
    $containers = docker ps --format "{{.Names}}"
    if ($containers) {
        $containers | ForEach-Object {
            if ($_ -match '(.+)-(.+)') {
                $botName = $matches[1]
                $channel = $matches[2]
                Write-Host "Bot: $botName"
                Write-Host "Channel: $channel"
                Write-Host "Container: $_"
                Write-Host "Status: Running"
                Write-Host "----------------------------"
            }
        }
    } else {
        Write-Host "No running bot instances found"
    }
}

# Function to stop all bot instances
function Stop-AllBots {
    Write-Host "Stopping all bot instances..."
    $containers = docker ps -a --format "{{.Names}}"
    if ($containers) {
        $containers | ForEach-Object {
            if ($_ -match '(.+)-(.+)') {
                Write-Host "Stopping container: $_"
                docker stop $_
                docker rm $_
            }
        }
        Write-Host "All bot instances stopped and removed" -ForegroundColor Green
    } else {
        Write-Host "No running bot instances found"
    }
}

# Function to stop a specific channel's bot instance
function Stop-ChannelBot {
    param (
        [string]$Channel
    )
    
    $channelConfig = "configs\${Channel}_config_secrets.yaml"
    
    if (-not (Test-Path $channelConfig)) {
        Write-Host "Error: Channel configuration file not found: $channelConfig" -ForegroundColor Red
        exit 1
    }
    
    # Extract bot name from channel config, preserving case
    $botName = (Select-String -Path $channelConfig -Pattern 'bot_name:' | Select-Object -First 1).Line -replace '.*bot_name:\s*"([^"]+)".*', '$1'
    if (-not $botName) {
        Write-Host "Error: bot_name not found in $channelConfig" -ForegroundColor Red
        exit 1
    }
    
    $containerName = "${botName}-${Channel}"
    Write-Host "Stopping bot for channel: $Channel"
    
    if (docker ps -a -q -f name=$containerName) {
        docker stop $containerName
        docker rm $containerName
        Write-Host "Bot stopped and removed for channel: $Channel" -ForegroundColor Green
    } else {
        Write-Host "No running bot instance found for channel: $Channel"
    }
}

# Function to restart all bots
function Restart-AllBots {
    Write-Host "Starting bot restart process..."
    
    Get-ChildItem -Path "configs\*_config_secrets.yaml" | ForEach-Object {
        $channel = $_.BaseName -replace '_config_secrets$', ''
        Write-Host "`nProcessing channel: $channel"
        
        # Stop the specific channel
        Write-Host "Stopping bot for channel: $channel"
        Stop-ChannelBot $channel
        
        # Start the channel
        Write-Host "Starting bot for channel: $channel"
        Start-Bot $channel
        
        # Wait a moment to ensure the bot is up before moving to the next
        Start-Sleep -Seconds 2
    }
    
    Write-Host "`nAll bots have been restarted successfully!" -ForegroundColor Green
}

# Main script logic
$command = $args[0]
$channel = $args[1]

switch ($command) {
    "start" {
        if (-not $channel) {
            Write-Host "Error: Channel name required" -ForegroundColor Red
            exit 1
        }
        Start-Bot $channel
    }
    "stop-channel" {
        if (-not $channel) {
            Write-Host "Error: Channel name required" -ForegroundColor Red
            exit 1
        }
        Stop-ChannelBot $channel
    }
    "stop-all" {
        Stop-AllBots
    }
    "list" {
        Get-Bots
    }
    "build" {
        Build-Image
    }
    "list-channels" {
        if (-not $channel) {
            Write-Host "Error: Bot name required" -ForegroundColor Red
            exit 1
        }
        Get-ChannelsByBot $channel
    }
    "update-bot" {
        if (-not $channel) {
            Write-Host "Error: Bot name required" -ForegroundColor Red
            exit 1
        }
        Update-BotConfig $channel
    }
    "restart-all" {
        Restart-AllBots
    }
    default {
        if ($command -and -not $channel) {
            # If only one argument is provided, treat it as a channel name
            Start-Bot $command
        } else {
            Write-Host "Usage: .\run_bot.ps1 [command] [channel]"
            Write-Host "Commands:"
            Write-Host "  start <channel>     - Start a bot instance"
            Write-Host "  stop-channel <channel> - Stop a specific channel's bot"
            Write-Host "  stop-all           - Stop all bot instances"
            Write-Host "  list               - List all running bot instances"
            Write-Host "  build              - Build the Docker image"
            Write-Host "  list-channels <bot> - List all channels using a specific bot"
            Write-Host "  update-bot <bot>   - Update shared bot configuration"
            Write-Host "  restart-all        - Stop and restart all bots with latest image"
            Write-Host ""
            Write-Host "Shortcut: .\run_bot.ps1 <channel> (same as start)"
        }
    }
} 