#!/bin/bash

# EC2 Deployment Script for PerfTiltBot
# This script deploys the Twitch bot on an EC2 instance using Docker

set -e

# Configuration
BOT_NAME=${BOT_NAME:-"perftiltbot"}
CHANNEL_NAME=${CHANNEL_NAME:-"your_channel"}
CONTAINER_NAME="${BOT_NAME}-${CHANNEL_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please run install_docker_ec2.sh first."
        exit 1
    fi
    
    # Check if Docker service is running
    if ! sudo systemctl is-active --quiet docker; then
        print_error "Docker service is not running. Starting Docker..."
        sudo systemctl start docker
    fi
    
    # Check if user is in docker group
    if ! groups $USER | grep -q docker; then
        print_warning "User is not in docker group. You may need to run 'newgrp docker' or log out and back in."
    fi
    
    print_status "Prerequisites check passed!"
}

# Function to validate configuration files
validate_config() {
    print_status "Validating configuration files..."
    
    # Check if secrets directory exists
    if [ ! -d "configs" ]; then
        print_error "configs directory not found. Please ensure you're in the project root."
        exit 1
    fi
    
    # Check for bot auth file
    local bot_auth="configs/bots/${BOT_NAME}_auth_secrets.yaml"
    if [ ! -f "$bot_auth" ]; then
        print_error "Bot authentication file not found: $bot_auth"
        print_status "Please create this file with your Twitch bot credentials."
        exit 1
    fi
    
    # Check for channel config file
    local channel_config="configs/channels/${CHANNEL_NAME}_config_secrets.yaml"
    if [ ! -f "$channel_config" ]; then
        print_error "Channel configuration file not found: $channel_config"
        print_status "Please create this file with your channel settings."
        exit 1
    fi
    
    print_status "Configuration files validated!"
}

# Function to build Docker image
build_image() {
    print_status "Building Docker image..."
    
    # Build the image
    docker build -t "$BOT_NAME" .
    
    if [ $? -eq 0 ]; then
        print_status "Docker image built successfully!"
    else
        print_error "Failed to build Docker image"
        exit 1
    fi
}

# Function to stop existing container
stop_existing_container() {
    print_status "Checking for existing container..."
    
    if [ "$(docker ps -a -q -f name=$CONTAINER_NAME)" ]; then
        print_warning "Container $CONTAINER_NAME already exists. Stopping and removing it..."
        docker stop "$CONTAINER_NAME" 2>/dev/null || true
        docker rm "$CONTAINER_NAME" 2>/dev/null || true
        print_status "Existing container removed."
    fi
}

# Function to run the bot
run_bot() {
    print_status "Starting bot for channel: $CHANNEL_NAME"
    
    # Get absolute paths
    local bot_auth="$(pwd)/configs/bots/${BOT_NAME}_auth_secrets.yaml"
    local channel_config="$(pwd)/configs/channels/${CHANNEL_NAME}_config_secrets.yaml"
    
    # Run the container
    docker run -d \
        --name "$CONTAINER_NAME" \
        --restart unless-stopped \
        -e "CHANNEL_NAME=$CHANNEL_NAME" \
        -v "$bot_auth:/app/configs/bots/${BOT_NAME}_auth_secrets.yaml" \
        -v "$channel_config:/app/configs/channels/${CHANNEL_NAME}_config_secrets.yaml" \
        -v "${BOT_NAME}-${CHANNEL_NAME}-data:/app/data" \
        "$BOT_NAME"
    
    if [ $? -eq 0 ]; then
        print_status "Bot started successfully!"
        print_status "Container name: $CONTAINER_NAME"
        print_status "To view logs: docker logs $CONTAINER_NAME"
        print_status "To stop: docker stop $CONTAINER_NAME"
    else
        print_error "Failed to start bot"
        exit 1
    fi
}

# Function to show bot status
show_status() {
    print_status "Bot Status:"
    echo "----------------------------"
    
    if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
        echo "✅ Bot is running"
        echo "Container: $CONTAINER_NAME"
        echo "Status: $(docker inspect --format='{{.State.Status}}' $CONTAINER_NAME)"
        echo "Started: $(docker inspect --format='{{.State.StartedAt}}' $CONTAINER_NAME)"
        echo ""
        echo "📋 Recent logs:"
        docker logs --tail=10 "$CONTAINER_NAME"
    else
        echo "❌ Bot is not running"
        if [ "$(docker ps -a -q -f name=$CONTAINER_NAME)" ]; then
            echo "Container exists but is stopped"
            echo "Last status: $(docker inspect --format='{{.State.Status}}' $CONTAINER_NAME)"
        fi
    fi
}

# Function to stop the bot
stop_bot() {
    print_status "Stopping bot..."
    
    if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
        docker stop "$CONTAINER_NAME"
        print_status "Bot stopped successfully!"
    else
        print_warning "Bot is not currently running"
    fi
}

# Function to restart the bot
restart_bot() {
    print_status "Restarting bot..."
    
    stop_bot
    sleep 2
    run_bot
}

# Function to view logs
view_logs() {
    if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
        print_status "Showing logs for $CONTAINER_NAME (Ctrl+C to exit):"
        docker logs -f "$CONTAINER_NAME"
    else
        print_error "Bot is not running"
        exit 1
    fi
}

# Function to deploy everything
deploy() {
    print_step "Starting EC2 deployment..."
    
    check_prerequisites
    validate_config
    build_image
    stop_existing_container
    run_bot
    
    print_status "Deployment completed successfully!"
    print_status "Your bot is now running on EC2"
    echo ""
    print_status "Useful commands:"
    echo "  ./deploy_ec2.sh status    - Check bot status"
    echo "  ./deploy_ec2.sh logs      - View bot logs"
    echo "  ./deploy_ec2.sh stop      - Stop the bot"
    echo "  ./deploy_ec2.sh restart   - Restart the bot"
}

# Main script logic
command=${1:-deploy}

case "$command" in
    "deploy")
        deploy
        ;;
    "status")
        show_status
        ;;
    "logs")
        view_logs
        ;;
    "stop")
        stop_bot
        ;;
    "restart")
        restart_bot
        ;;
    "build")
        check_prerequisites
        build_image
        ;;
    *)
        echo "Usage: $0 [deploy|status|logs|stop|restart|build]"
        echo ""
        echo "Commands:"
        echo "  deploy   - Deploy the bot (default)"
        echo "  status   - Show bot status"
        echo "  logs     - View bot logs"
        echo "  stop     - Stop the bot"
        echo "  restart  - Restart the bot"
        echo "  build    - Build Docker image only"
        echo ""
        echo "Environment variables:"
        echo "  BOT_NAME      - Bot name (default: perftiltbot)"
        echo "  CHANNEL_NAME  - Channel name (default: your_channel)"
        echo ""
        echo "Examples:"
        echo "  BOT_NAME=mybot CHANNEL_NAME=mychannel ./deploy_ec2.sh deploy"
        echo "  ./deploy_ec2.sh status"
        echo "  ./deploy_ec2.sh logs"
        exit 1
        ;;
esac
