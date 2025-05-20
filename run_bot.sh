#!/bin/bash

# PerfTiltBot Management Script
#
# This script manages PerfTiltBot instances for different Twitch channels.
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
#   ./run_bot.sh start pbuckles
#   ./run_bot.sh stop-channel pbuckles
#   ./run_bot.sh list
#   ./run_bot.sh stop-all
#   ./run_bot.sh build
#   ./run_bot.sh list-channels perftiltbot
#   ./run_bot.sh update-bot perftiltbot
#
# Shortcut:
#   ./run_bot.sh <channel_name>  - Same as 'start <channel_name>'
#
# Notes:
#   - Requires Docker to be running
#   - Channel-specific secrets files must exist in configs/
#   - Each channel gets its own container and data volume

# Function to build the Docker image
build_image() {
    echo "Building Docker image..."
    docker build -t perftiltbot .
    if [ $? -ne 0 ]; then
        echo "Error: Failed to build Docker image"
        exit 1
    fi
    echo "Docker image built successfully!"
}

# Function to validate configuration
validate_config() {
    local config_file=$1
    local required_fields=("bot_name" "channel" "oauth" "client_id" "client_secret")
    local missing_fields=()

    for field in "${required_fields[@]}"; do
        if ! grep -q "^${field}:" "$config_file"; then
            missing_fields+=("$field")
        fi
    done

    if [ ${#missing_fields[@]} -gt 0 ]; then
        echo "Error: Missing required fields in $config_file:"
        printf '%s\n' "${missing_fields[@]}"
        return 1
    fi
    return 0
}

# Function to list channels using a specific bot
list_channels_by_bot() {
    local BOT_NAME=$1
    local found=0

    echo "Channels using bot: $BOT_NAME"
    echo "----------------------------"
    
    for file in configs/*_secrets.yaml; do
        if [ -f "$file" ]; then
            local bot_name=$(grep "bot_name:" "$file" | cut -d'"' -f2)
            if [ "$bot_name" = "$BOT_NAME" ]; then
                local channel=$(basename "$file" _secrets.yaml)
                echo "Channel: $channel"
                echo "Config file: $file"
                echo "----------------------------"
                found=1
            fi
        fi
    done

    if [ $found -eq 0 ]; then
        echo "No channels found using bot: $BOT_NAME"
    fi
}

# Function to update shared bot configuration
update_bot_config() {
    local BOT_NAME=$1
    local BOT_SECRETS="configs/${BOT_NAME}_secrets.yaml"
    local TEMP_FILE="configs/temp_update.yaml"

    # Check if bot config exists
    if [ ! -f "$BOT_SECRETS" ]; then
        echo "Error: Bot configuration not found: $BOT_SECRETS"
        return 1
    fi

    # Create backup
    cp "$BOT_SECRETS" "${BOT_SECRETS}.bak"
    echo "Created backup at ${BOT_SECRETS}.bak"

    # Create temporary file with current config
    cp "$BOT_SECRETS" "$TEMP_FILE"

    # Edit the temporary file
    if [ -n "$EDITOR" ]; then
        $EDITOR "$TEMP_FILE"
    else
        vi "$TEMP_FILE"
    fi

    # Validate the updated configuration
    if validate_config "$TEMP_FILE"; then
        # Update the bot config
        mv "$TEMP_FILE" "$BOT_SECRETS"
        echo "Bot configuration updated successfully"
        
        # List affected channels
        echo "Affected channels:"
        list_channels_by_bot "$BOT_NAME"
    else
        echo "Error: Invalid configuration. Changes not saved."
        rm "$TEMP_FILE"
        return 1
    fi
}

# Function to start a bot for a specific channel
start_bot() {
    local CHANNEL=$1
    local SECRETS_FILE="configs/${CHANNEL}_secrets.yaml"
    local CONTAINER_NAME="perftiltbot-${CHANNEL}"
    local BOT_CONFIG="configs/bot.yaml"
    local TEMP_SECRETS="configs/temp_secrets.yaml"

    # Check if secrets file exists
    if [ ! -f "$SECRETS_FILE" ]; then
        echo "Error: Secrets file not found: $SECRETS_FILE"
        echo "Please create a secrets file at: $SECRETS_FILE"
        exit 1
    fi

    # Check if bot config exists
    if [ ! -f "$BOT_CONFIG" ]; then
        echo "Error: Bot configuration file not found: $BOT_CONFIG"
        echo "Please create a bot configuration file at: $BOT_CONFIG"
        exit 1
    fi

    # Extract bot name from channel secrets
    local BOT_NAME=$(grep "bot_name:" "$SECRETS_FILE" | cut -d'"' -f2)
    if [ -z "$BOT_NAME" ]; then
        echo "Error: bot_name not found in $SECRETS_FILE"
        exit 1
    fi

    # Check if bot-specific config exists
    local BOT_SECRETS="configs/${BOT_NAME}_secrets.yaml"
    if [ -f "$BOT_SECRETS" ]; then
        echo "Found bot-specific configuration for $BOT_NAME"
        # Merge bot secrets with channel secrets
        echo "Merging configurations..."
        # First copy bot secrets as base
        cp "$BOT_SECRETS" "$TEMP_SECRETS"
        # Then merge channel-specific overrides
        yq eval-all 'select(fileIndex == 0) * select(fileIndex == 1)' "$TEMP_SECRETS" "$SECRETS_FILE" > "configs/secrets.yaml"
        rm "$TEMP_SECRETS"
    else
        echo "No bot-specific configuration found, using channel configuration directly"
        cp "$SECRETS_FILE" "configs/secrets.yaml"
    fi

    # Validate the final configuration
    if ! validate_config "configs/secrets.yaml"; then
        echo "Error: Invalid configuration after merging"
        exit 1
    fi

    # Check if container is already running
    if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
        echo "Container $CONTAINER_NAME is already running"
        echo "Stopping and removing existing container..."
        docker stop "$CONTAINER_NAME"
        docker rm "$CONTAINER_NAME"
    fi

    # Run the container
    echo "Starting bot for channel: $CHANNEL"
    docker run -d \
        --name "$CONTAINER_NAME" \
        -v "${PWD}/configs/secrets.yaml:/app/configs/secrets.yaml" \
        -v "${PWD}/configs/bot.yaml:/app/configs/bot.yaml" \
        -v "perftiltbot-${CHANNEL}-data:/app/data" \
        perftiltbot

    echo "Bot started successfully!"
    echo "Container name: $CONTAINER_NAME"
    echo "To view logs: docker logs $CONTAINER_NAME"
    echo "To stop: docker stop $CONTAINER_NAME"
}

# Function to list all running bot instances
list_bots() {
    echo -e "\nRunning PerfTiltBot instances:"
    echo "----------------------------"
    local containers=$(docker ps --format "{{.Names}}" | grep "perftiltbot-")
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r container; do
            local channel=${container#perftiltbot-}
            echo "Channel: $channel"
            echo "Container: $container"
            echo "Status: Running"
            echo "----------------------------"
        done
    else
        echo "No running bot instances found"
    fi
}

# Function to stop all bot instances
stop_all_bots() {
    echo "Stopping all PerfTiltBot instances..."
    local containers=$(docker ps -q -f "name=perftiltbot-")
    if [ -n "$containers" ]; then
        docker stop $containers
        docker rm $containers
        echo "All bot instances stopped and removed"
    else
        echo "No running bot instances found"
    fi
}

# Function to stop a specific channel's bot instance
stop_channel_bot() {
    local CHANNEL=$1
    local CONTAINER_NAME="perftiltbot-${CHANNEL}"
    echo "Stopping bot for channel: $CHANNEL"
    
    if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
        docker stop "$CONTAINER_NAME"
        docker rm "$CONTAINER_NAME"
        echo "Bot stopped and removed for channel: $CHANNEL"
    else
        echo "No running bot instance found for channel: $CHANNEL"
    fi
}

# Main script logic
if [ $# -eq 0 ]; then
    echo "Usage:"
    echo "  ./run_bot.sh start <channel_name>    - Start bot for a channel"
    echo "  ./run_bot.sh stop-channel <channel>  - Stop bot for a specific channel"
    echo "  ./run_bot.sh build                   - Build Docker image"
    echo "  ./run_bot.sh list                    - List running bot instances"
    echo "  ./run_bot.sh stop-all               - Stop all bot instances"
    echo "  ./run_bot.sh list-channels <bot>    - List channels using a specific bot"
    echo "  ./run_bot.sh update-bot <bot>       - Update shared bot configuration"
    echo ""
    echo "Examples:"
    echo "  ./run_bot.sh start pbuckles"
    echo "  ./run_bot.sh stop-channel pbuckles"
    echo "  ./run_bot.sh build"
    echo "  ./run_bot.sh list-channels perftiltbot"
    echo "  ./run_bot.sh update-bot perftiltbot"
    echo ""
    echo "Shortcut:"
    echo "  ./run_bot.sh <channel_name>         - Same as 'start <channel_name>'"
    exit 1
fi

# If only one argument is provided and it's not a known command, treat it as a channel name
if [ $# -eq 1 ] && [[ ! "$1" =~ ^(start|stop-channel|build|list|stop-all|list-channels|update-bot)$ ]]; then
    # Check if image exists, build if it doesn't
    if [ -z "$(docker images -q perftiltbot)" ]; then
        build_image
    fi
    start_bot "$1"
    exit 0
fi

# Handle commands
case "$1" in
    "start")
        if [ $# -lt 2 ]; then
            echo "Error: Channel name required for start command"
            echo "Usage: ./run_bot.sh start <channel_name>"
            exit 1
        fi
        # Check if image exists, build if it doesn't
        if [ -z "$(docker images -q perftiltbot)" ]; then
            build_image
        fi
        start_bot "$2"
        ;;
    "stop-channel")
        if [ $# -lt 2 ]; then
            echo "Error: Channel name required for stop-channel command"
            echo "Usage: ./run_bot.sh stop-channel <channel_name>"
            exit 1
        fi
        stop_channel_bot "$2"
        ;;
    "build")
        build_image
        ;;
    "list")
        list_bots
        ;;
    "stop-all")
        stop_all_bots
        ;;
    "list-channels")
        if [ $# -lt 2 ]; then
            echo "Error: Bot name required for list-channels command"
            echo "Usage: ./run_bot.sh list-channels <bot_name>"
            exit 1
        fi
        list_channels_by_bot "$2"
        ;;
    "update-bot")
        if [ $# -lt 2 ]; then
            echo "Error: Bot name required for update-bot command"
            echo "Usage: ./run_bot.sh update-bot <bot_name>"
            exit 1
        fi
        update_bot_config "$2"
        ;;
    *)
        echo "Error: Unknown command '$1'"
        echo "Run ./run_bot.sh without arguments to see usage"
        exit 1
        ;;
esac 