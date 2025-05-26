#!/bin/bash

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
#   ./run_bot.sh start pbuckles
#   ./run_bot.sh stop-channel pbuckles
#   ./run_bot.sh list
#   ./run_bot.sh stop-all
#   ./run_bot.sh build
#   ./run_bot.sh list-channels pbchatbot
#   ./run_bot.sh update-bot pbchatbot
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
    docker build -t pbchatbot .
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
    
    for file in configs/*_config_secrets.yaml; do
        if [ -f "$file" ]; then
            local bot_name=$(grep "bot_name:" "$file" | sed -E 's/.*bot_name:[[:space:]]*"([^"]+)".*/\1/')
            if [ "$bot_name" = "$BOT_NAME" ]; then
                local channel=$(basename "$file" _config_secrets.yaml)
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
    local CHANNEL_CONFIG="configs/${CHANNEL}_config_secrets.yaml"

    # Check if channel config exists
    if [ ! -f "$CHANNEL_CONFIG" ]; then
        echo "Error: Channel configuration file not found: $CHANNEL_CONFIG"
        echo "Please create a channel configuration file at: $CHANNEL_CONFIG"
        exit 1
    fi

    # Extract bot name from channel config, preserving case
    local BOT_NAME=$(grep "bot_name:" "$CHANNEL_CONFIG" | sed -E 's/.*bot_name:[[:space:]]*"([^"]+)".*/\1/')
    if [ -z "$BOT_NAME" ]; then
        echo "Error: bot_name not found in $CHANNEL_CONFIG"
        exit 1
    fi

    # Extract channel name from config, preserving case
    local CHANNEL_NAME=$(grep "channel:" "$CHANNEL_CONFIG" | sed -E 's/.*channel:[[:space:]]*"([^"]+)".*/\1/')
    if [ -z "$CHANNEL_NAME" ]; then
        echo "Error: channel not found in $CHANNEL_CONFIG"
        exit 1
    fi

    # Create container name using exact case from configs
    local CONTAINER_NAME="${BOT_NAME}-${CHANNEL_NAME}"

    # Check if bot auth exists
    local BOT_AUTH="configs/${BOT_NAME}_auth_secrets.yaml"
    if [ ! -f "$BOT_AUTH" ]; then
        echo "Error: Bot authentication file not found: $BOT_AUTH"
        exit 1
    fi

    # Check if container is already running or exists
    if [ "$(docker ps -a -q -f name=$CONTAINER_NAME)" ]; then
        echo "Container $CONTAINER_NAME already exists. Stopping and removing it..."
        docker stop "$CONTAINER_NAME"
        docker rm "$CONTAINER_NAME"
    fi

    # Run the container
    echo "Starting bot for channel: $CHANNEL_NAME"
    docker run -d \
        --name "$CONTAINER_NAME" \
        -e "CHANNEL_NAME=$CHANNEL" \
        -v "$(pwd)/$BOT_AUTH:/app/configs/bot_auth.yaml" \
        -v "$(pwd)/$CHANNEL_CONFIG:/app/configs/${CHANNEL}_config_secrets.yaml" \
        -v "${BOT_NAME}-${CHANNEL_NAME}-data:/app/data" \
        pbchatbot

    echo "Bot started successfully!"
    echo "Container name: $CONTAINER_NAME"
    echo "To view logs: docker logs $CONTAINER_NAME"
    echo "To stop: docker stop $CONTAINER_NAME"
}

# Function to list all running bot instances
list_bots() {
    echo -e "\nRunning bot instances:"
    echo "----------------------------"
    local containers=$(docker ps --format "{{.Names}}")
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r container; do
            # Extract bot name and channel from container name
            if [[ $container =~ (.+)-(.+) ]]; then
                local bot_name="${BASH_REMATCH[1]}"
                local channel="${BASH_REMATCH[2]}"
                echo "Bot: $bot_name"
                echo "Channel: $channel"
                echo "Container: $container"
                echo "Status: Running"
                echo "----------------------------"
            fi
        done
    else
        echo "No running bot instances found"
    fi
}

# Function to stop all bot instances
stop_all_bots() {
    echo "Stopping all bot instances..."
    local containers=$(docker ps -a --format "{{.Names}}")
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r container; do
            if [[ $container =~ (.+)-(.+) ]]; then
                echo "Stopping container: $container"
                docker stop "$container"
                docker rm "$container"
            fi
        done
        echo "All bot instances stopped and removed"
    else
        echo "No running bot instances found"
    fi
}

# Function to stop a specific channel's bot instance
stop_channel_bot() {
    local CHANNEL=$1
    local CHANNEL_CONFIG="configs/${CHANNEL}_config_secrets.yaml"

    # Check if channel config exists
    if [ ! -f "$CHANNEL_CONFIG" ]; then
        echo "Error: Channel configuration file not found: $CHANNEL_CONFIG"
        exit 1
    fi

    # Extract bot name from channel config, preserving case
    local BOT_NAME=$(grep "bot_name:" "$CHANNEL_CONFIG" | sed -E 's/.*bot_name:[[:space:]]*"([^"]+)".*/\1/')
    if [ -z "$BOT_NAME" ]; then
        echo "Error: bot_name not found in $CHANNEL_CONFIG"
        exit 1
    fi

    local CONTAINER_NAME="${BOT_NAME}-${CHANNEL}"
    echo "Stopping bot for channel: $CHANNEL"
    
    if [ "$(docker ps -a -q -f name=$CONTAINER_NAME)" ]; then
        docker stop "$CONTAINER_NAME"
        docker rm "$CONTAINER_NAME"
        echo "Bot stopped and removed for channel: $CHANNEL"
    else
        echo "No running bot instance found for channel: $CHANNEL"
    fi
}

# Function to restart all bots
restart_all_bots() {
    echo "Starting bot restart process..."
    
    for config in configs/*_config_secrets.yaml; do
        if [ -f "$config" ]; then
            channel=$(basename "$config" _config_secrets.yaml)
            echo -e "\nProcessing channel: $channel"
            
            # Stop the specific channel
            echo "Stopping bot for channel: $channel"
            stop_channel_bot "$channel"
            
            # Start the channel
            echo "Starting bot for channel: $channel"
            start_bot "$channel"
            
            # Wait a moment to ensure the bot is up before moving to the next
            sleep 2
        fi
    done
    
    echo -e "\nAll bots have been restarted successfully!"
}

# Main script logic
command=$1
channel=$2

case "$command" in
    "start")
        if [ -z "$channel" ]; then
            echo "Error: Channel name required"
            exit 1
        fi
        start_bot "$channel"
        ;;
    "stop-channel")
        if [ -z "$channel" ]; then
            echo "Error: Channel name required"
            exit 1
        fi
        stop_channel_bot "$channel"
        ;;
    "stop-all")
        stop_all_bots
        ;;
    "list")
        list_bots
        ;;
    "build")
        build_image
        ;;
    "list-channels")
        if [ -z "$channel" ]; then
            echo "Error: Bot name required"
            exit 1
        fi
        list_channels_by_bot "$channel"
        ;;
    "update-bot")
        if [ -z "$channel" ]; then
            echo "Error: Bot name required"
            exit 1
        fi
        update_bot_config "$channel"
        ;;
    "restart-all")
        restart_all_bots
        ;;
    *)
        if [ -n "$command" ] && [ -z "$channel" ]; then
            # If only one argument is provided, treat it as a channel name
            start_bot "$command"
        else
            echo "Usage: ./run_bot.sh [command] [channel]"
            echo "Commands:"
            echo "  start <channel>     - Start a bot instance"
            echo "  stop-channel <channel> - Stop a specific channel's bot"
            echo "  stop-all           - Stop all bot instances"
            echo "  list               - List all running bot instances"
            echo "  build              - Build the Docker image"
            echo "  list-channels <bot> - List all channels using a specific bot"
            echo "  update-bot <bot>   - Update shared bot configuration"
            echo "  restart-all        - Stop and restart all bots with latest image"
            echo ""
            echo "Shortcut: ./run_bot.sh <channel> (same as start)"
        fi
        ;;
esac 