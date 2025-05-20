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
#
# Examples:
#   ./run_bot.sh start pbuckles
#   ./run_bot.sh stop-channel pbuckles
#   ./run_bot.sh list
#   ./run_bot.sh stop-all
#   ./run_bot.sh build
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

# Function to start a bot for a specific channel
start_bot() {
    local CHANNEL=$1
    local SECRETS_FILE="configs/${CHANNEL}_secrets.yaml"
    local CONTAINER_NAME="perftiltbot-${CHANNEL}"
    local BOT_CONFIG="configs/bot.yaml"

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

    # Copy channel-specific secrets to secrets.yaml
    echo "Setting up configuration for channel: $CHANNEL"
    cp "$SECRETS_FILE" "configs/secrets.yaml"

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

# Show usage if no arguments provided
if [ $# -eq 0 ]; then
    echo "Usage:"
    echo "  ./run_bot.sh start <channel_name>    - Start bot for a channel"
    echo "  ./run_bot.sh stop-channel <channel>  - Stop bot for a specific channel"
    echo "  ./run_bot.sh build                   - Build Docker image"
    echo "  ./run_bot.sh list                    - List running bot instances"
    echo "  ./run_bot.sh stop-all               - Stop all bot instances"
    echo ""
    echo "Examples:"
    echo "  ./run_bot.sh start pbuckles"
    echo "  ./run_bot.sh stop-channel pbuckles"
    echo "  ./run_bot.sh build"
    echo ""
    echo "Shortcut:"
    echo "  ./run_bot.sh <channel_name>         - Same as 'start <channel_name>'"
    exit 1
fi

# If only one argument is provided and it's not a known command, treat it as a channel name
if [ $# -eq 1 ] && [[ ! "$1" =~ ^(start|stop-channel|build|list|stop-all)$ ]]; then
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
    *)
        echo "Error: Unknown command '$1'"
        echo "Run ./run_bot.sh without arguments to see usage"
        exit 1
        ;;
esac 