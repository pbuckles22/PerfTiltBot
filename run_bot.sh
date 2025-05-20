#!/bin/bash

# Check if channel name is provided
if [ -z "$1" ]; then
    echo "Usage: ./run_bot.sh <channel_name>"
    echo "Example: ./run_bot.sh pbuckles"
    exit 1
fi

CHANNEL=$1
SECRETS_FILE="configs/${CHANNEL}_secrets.yaml"
CONTAINER_NAME="perftiltbot-${CHANNEL}"

# Check if secrets file exists
if [ ! -f "$SECRETS_FILE" ]; then
    echo "Error: Secrets file not found: $SECRETS_FILE"
    echo "Please create a secrets file at: $SECRETS_FILE"
    exit 1
fi

# Copy channel-specific secrets to secrets.yaml
echo "Setting up configuration for channel: $CHANNEL"
cp "$SECRETS_FILE" "configs/secrets.yaml"

# Check if container is already running
if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
    echo "Container $CONTAINER_NAME is already running"
    echo "Stopping and removing existing container..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
fi

# Run the container
echo "Starting bot for channel: $CHANNEL"
docker run -d \
    --name $CONTAINER_NAME \
    -v "$(pwd)/configs/secrets.yaml:/app/configs/secrets.yaml" \
    -v "perftiltbot-${CHANNEL}-data:/app/data" \
    perftiltbot

echo "Bot started successfully!"
echo "Container name: $CONTAINER_NAME"
echo "To view logs: docker logs $CONTAINER_NAME"
echo "To stop: docker stop $CONTAINER_NAME" 