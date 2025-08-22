#!/bin/bash

# Enhanced EC2 Deployment Script for PerfTiltBot
# This script manages multiple Twitch bot instances on EC2 using Docker
# Mirrors the functionality of run_bot.ps1 for Linux/EC2

set -e

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

# Function to build Docker image
build_image() {
    print_status "Building Docker image..."
    
    # Build the image with lowercase name for Docker compatibility
    docker build -t "perftiltbot" .
    
    if [ $? -eq 0 ]; then
        print_status "Docker image built successfully!"
    else
        print_error "Failed to build Docker image"
        exit 1
    fi
}

# Function to extract bot name from channel config
get_bot_name_from_config() {
    local channel_config="$1"
    local bot_name=$(grep 'bot_name:' "$channel_config" | head -1 | sed 's/.*bot_name:\s*"\([^"]*\)".*/\1/')
    echo "$bot_name"
}

# Function to extract channel name from config
get_channel_name_from_config() {
    local channel_config="$1"
    local channel_name=$(grep 'channel:' "$channel_config" | head -1 | sed 's/.*channel:\s*"\([^"]*\)".*/\1/')
    echo "$channel_name"
}

# Function to list all running bot instances
list_bots() {
    print_status "Running bot instances:"
    echo "----------------------------"
    local containers=$(docker ps --format "{{.Names}}")
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r container; do
            if [[ $container =~ ^(.+)-(.+)$ ]]; then
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
    print_status "Stopping all bot instances..."
    local containers=$(docker ps -a --format "{{.Names}}")
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r container; do
            if [[ $container =~ ^(.+)-(.+)$ ]]; then
                print_status "Stopping container: $container"
                docker stop "$container" 2>/dev/null || true
                docker rm "$container" 2>/dev/null || true
            fi
        done
        print_status "All bot instances stopped and removed"
    else
        print_status "No running bot instances found"
    fi
}

# Function to stop a specific channel's bot instance
stop_channel_bot() {
    local channel="$1"
    local channel_config="configs/channels/${channel}_config_secrets.yaml"
    
    if [ ! -f "$channel_config" ]; then
        print_error "Channel configuration file not found: $channel_config"
        exit 1
    fi
    
    # Extract bot name from channel config
    local bot_name=$(get_bot_name_from_config "$channel_config")
    if [ -z "$bot_name" ]; then
        print_error "bot_name not found in $channel_config"
        exit 1
    fi
    
    local container_name="$(echo "$bot_name" | tr '[:upper:]' '[:lower:]')-$(echo "$channel" | tr '[:upper:]' '[:lower:]')"
    print_status "Stopping bot for channel: $channel"
    
    if [ "$(docker ps -a -q -f name=$container_name)" ]; then
        docker stop "$container_name"
        docker rm "$container_name"
        print_status "Bot stopped and removed for channel: $channel"
    else
        print_warning "No running bot instance found for channel: $channel"
    fi
}

# Function to list channels using a specific bot
list_channels_by_bot() {
    local bot_name="$1"
    local found=false
    
    print_status "Channels using bot: $bot_name"
    echo "----------------------------"
    
    for config_file in configs/channels/*_config_secrets.yaml; do
        if [ -f "$config_file" ]; then
            local config_bot_name=$(get_bot_name_from_config "$config_file")
            if [ "$config_bot_name" = "$bot_name" ]; then
                local channel=$(basename "$config_file" _config_secrets.yaml)
                echo "Channel: $channel"
                echo "Config file: $config_file"
                echo "----------------------------"
                found=true
            fi
        fi
    done
    
    if [ "$found" = false ]; then
        print_warning "No channels found using bot: $bot_name"
    fi
}

# Function to restart all bots
restart_all_bots() {
    print_status "Starting bot restart process..."
    
    for config_file in configs/channels/*_config_secrets.yaml; do
        if [ -f "$config_file" ]; then
            local channel=$(basename "$config_file" _config_secrets.yaml)
            echo ""
            print_status "Processing channel: $channel"
            
            # Stop the specific channel
            print_status "Stopping bot for channel: $channel"
            stop_channel_bot "$channel"
            
            # Start the channel
            print_status "Starting bot for channel: $channel"
            start_bot "$channel"
            
            # Wait a moment to ensure the bot is up before moving to the next
            sleep 2
        fi
    done
    
    echo ""
    print_status "All bots have been restarted successfully!"
}

# Function to validate configuration files
validate_config() {
    print_status "Validating configuration files..."
    
    # Check if configs directory exists
    if [ ! -d "configs" ]; then
        print_error "configs directory not found. Please ensure you're in the project root."
        exit 1
    fi
    
    # Check if bots directory exists
    if [ ! -d "configs/bots" ]; then
        print_error "configs/bots directory not found."
        exit 1
    fi
    
    # Check if channels directory exists
    if [ ! -d "configs/channels" ]; then
        print_error "configs/channels directory not found."
        exit 1
    fi
    
    print_status "Configuration files validated!"
}

# Function to start a bot for a specific channel
start_bot() {
    local channel="$1"
    local channel_config="configs/channels/${channel}_config_secrets.yaml"
    
    if [ ! -f "$channel_config" ]; then
        print_error "Channel configuration file not found: $channel_config"
        exit 1
    fi
    
    # Extract bot name from channel config, preserving case
    local bot_name=$(get_bot_name_from_config "$channel_config")
    if [ -z "$bot_name" ]; then
        print_error "bot_name not found in $channel_config"
        exit 1
    fi
    
    # Extract channel name from config, preserving case
    local channel_name=$(get_channel_name_from_config "$channel_config")
    if [ -z "$channel_name" ]; then
        print_error "channel not found in $channel_config"
        exit 1
    fi
    
    # Create container name using exact case from configs, but lowercase for Docker
    local container_name="$(echo "$bot_name" | tr '[:upper:]' '[:lower:]')-$(echo "$channel_name" | tr '[:upper:]' '[:lower:]')"
    
    # Check if bot auth exists
    local bot_auth="configs/bots/${bot_name}_auth_secrets.yaml"
    if [ ! -f "$bot_auth" ]; then
        print_error "Bot authentication file not found: $bot_auth"
        exit 1
    fi
    
    # Check if container is already running or exists
    if [ "$(docker ps -a -q -f name=$container_name)" ]; then
        print_warning "Container $container_name already exists. Stopping and removing it..."
        docker stop "$container_name" 2>/dev/null || true
        docker rm "$container_name" 2>/dev/null || true
    fi
    
    # Run the container
    print_status "Starting bot for channel: $channel_name"
    
    # Get absolute paths
    local bot_auth_path="$(pwd)/$bot_auth"
    local channel_config_path="$(pwd)/$channel_config"
    
    # Run the container with proper volume mounts and CloudWatch logging
    docker run -d \
        --name "$container_name" \
        --restart unless-stopped \
        --log-driver=awslogs \
        --log-opt awslogs-region=us-west-2 \
        --log-opt awslogs-group="/ec2/perftiltbot" \
        --log-opt awslogs-stream="$container_name" \
        -e "CHANNEL_NAME=$channel_name" \
        -e "BOT_NAME=$bot_name" \
        -v "$bot_auth_path:/app/configs/bots/${bot_name}_auth_secrets.yaml" \
        -v "$channel_config_path:/app/configs/channels/${channel_name}_config_secrets.yaml" \
        -v "$(pwd)/data:/app/data" \
        "perftiltbot"
    
    if [ $? -eq 0 ]; then
        print_status "Bot started successfully!"
        print_status "Container name: $container_name"
        print_status "CloudWatch log group: /ec2/perftiltbot"
        print_status "CloudWatch log stream: $container_name"
        print_status "To view logs: docker logs $container_name"
        print_status "To stop: docker stop $container_name"
    else
        print_error "Failed to start bot"
        exit 1
    fi
}

# Function to list all running bot instances
list_bots() {
    print_status "Running bot instances:"
    echo "----------------------------"
    local containers=$(docker ps --format "{{.Names}}")
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r container; do
            if [[ $container =~ ^(.+)-(.+)$ ]]; then
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
    print_status "Stopping all bot instances..."
    local containers=$(docker ps -a --format "{{.Names}}")
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r container; do
            if [[ $container =~ ^(.+)-(.+)$ ]]; then
                print_status "Stopping container: $container"
                docker stop "$container" 2>/dev/null || true
                docker rm "$container" 2>/dev/null || true
            fi
        done
        print_status "All bot instances stopped and removed"
    else
        print_status "No running bot instances found"
    fi
}

# Function to stop a specific channel's bot instance
stop_channel_bot() {
    local channel="$1"
    local channel_config="configs/channels/${channel}_config_secrets.yaml"
    
    if [ ! -f "$channel_config" ]; then
        print_error "Channel configuration file not found: $channel_config"
        exit 1
    fi
    
    # Extract bot name from channel config
    local bot_name=$(get_bot_name_from_config "$channel_config")
    if [ -z "$bot_name" ]; then
        print_error "bot_name not found in $channel_config"
        exit 1
    fi
    
    local container_name="$(echo "$bot_name" | tr '[:upper:]' '[:lower:]')-$(echo "$channel" | tr '[:upper:]' '[:lower:]')"
    print_status "Stopping bot for channel: $channel"
    
    if [ "$(docker ps -a -q -f name=$container_name)" ]; then
        docker stop "$container_name"
        docker rm "$container_name"
        print_status "Bot stopped and removed for channel: $channel"
    else
        print_warning "No running bot instance found for channel: $channel"
    fi
}

# Function to list channels using a specific bot
list_channels_by_bot() {
    local bot_name="$1"
    local found=false
    
    print_status "Channels using bot: $bot_name"
    echo "----------------------------"
    
    for config_file in configs/channels/*_config_secrets.yaml; do
        if [ -f "$config_file" ]; then
            local config_bot_name=$(get_bot_name_from_config "$config_file")
            if [ "$config_bot_name" = "$bot_name" ]; then
                local channel=$(basename "$config_file" _config_secrets.yaml)
                echo "Channel: $channel"
                echo "Config file: $config_file"
                echo "----------------------------"
                found=true
            fi
        fi
    done
    
    if [ "$found" = false ]; then
        print_warning "No channels found using bot: $bot_name"
    fi
}

# Function to restart all bots
restart_all_bots() {
    print_status "Starting bot restart process..."
    
    for config_file in configs/channels/*_config_secrets.yaml; do
        if [ -f "$config_file" ]; then
            local channel=$(basename "$config_file" _config_secrets.yaml)
            echo ""
            print_status "Processing channel: $channel"
            
            # Stop the specific channel
            print_status "Stopping bot for channel: $channel"
            stop_channel_bot "$channel"
            
            # Start the channel
            print_status "Starting bot for channel: $channel"
            start_bot "$channel"
            
            # Wait a moment to ensure the bot is up before moving to the next
            sleep 2
        fi
    done
    
    echo ""
    print_status "All bots have been restarted successfully!"
}

# Function to show bot status
show_status() {
    local channel="$1"
    
    if [ -n "$channel" ]; then
        # Show status for specific channel
        local channel_config="configs/channels/${channel}_config_secrets.yaml"
        if [ ! -f "$channel_config" ]; then
            print_error "Channel configuration file not found: $channel_config"
            exit 1
        fi
        
        local bot_name=$(get_bot_name_from_config "$channel_config")
        local container_name="$(echo "$bot_name" | tr '[:upper:]' '[:lower:]')-$(echo "$channel" | tr '[:upper:]' '[:lower:]')"
        
        print_status "Bot Status for channel: $channel"
        echo "----------------------------"
        
        if [ "$(docker ps -q -f name=$container_name)" ]; then
            echo "✅ Bot is running"
            echo "Container: $container_name"
            echo "Status: $(docker inspect --format='{{.State.Status}}' $container_name)"
            echo "Started: $(docker inspect --format='{{.State.StartedAt}}' $container_name)"
            echo ""
            echo "📋 Recent logs:"
            docker logs --tail=10 "$container_name"
        else
            echo "❌ Bot is not running"
            if [ "$(docker ps -a -q -f name=$container_name)" ]; then
                echo "Container exists but is stopped"
                echo "Last status: $(docker inspect --format='{{.State.Status}}' $container_name)"
            fi
        fi
    else
        # Show status for all bots
        list_bots
    fi
}

# Function to view logs
view_logs() {
    local channel="$1"
    
    if [ -z "$channel" ]; then
        print_error "Channel name required for logs command"
        exit 1
    fi
    
    local channel_config="configs/channels/${channel}_config_secrets.yaml"
    if [ ! -f "$channel_config" ]; then
        print_error "Channel configuration file not found: $channel_config"
        exit 1
    fi
    
    local bot_name=$(get_bot_name_from_config "$channel_config")
    local container_name="$(echo "$bot_name" | tr '[:upper:]' '[:lower:]')-$(echo "$channel" | tr '[:upper:]' '[:lower:]')"
    
    if [ "$(docker ps -q -f name=$container_name)" ]; then
        print_status "Showing logs for $container_name (Ctrl+C to exit):"
        docker logs -f "$container_name"
    else
        print_error "Bot is not running"
        exit 1
    fi
}

# Function to view CloudWatch logs
view_cloudwatch_logs() {
    local channel="$1"
    
    if [ -z "$channel" ]; then
        print_error "Channel name required for cloudwatch-logs command"
        exit 1
    fi
    
    local channel_config="configs/channels/${channel}_config_secrets.yaml"
    if [ ! -f "$channel_config" ]; then
        print_error "Channel configuration file not found: $channel_config"
        exit 1
    fi
    
    local bot_name=$(get_bot_name_from_config "$channel_config")
    local container_name="$(echo "$bot_name" | tr '[:upper:]' '[:lower:]')-$(echo "$channel" | tr '[:upper:]' '[:lower:]')"
    
    print_status "Showing CloudWatch logs for $container_name (Ctrl+C to exit):"
    print_status "Log group: /ec2/perftiltbot"
    print_status "Log stream: $container_name"
    
    # Check if AWS CLI is available
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install it to view CloudWatch logs."
        print_status "Install with: sudo yum install -y aws-cli"
        exit 1
    fi
    
    # Show CloudWatch logs
    aws logs tail "/ec2/perftiltbot" --log-stream-names "$container_name" --follow --region us-west-2
}

# Main script logic
command=${1:-deploy}
channel=${2:-}

case "$command" in
    "start")
        if [ -z "$channel" ]; then
            print_error "Channel name required"
            exit 1
        fi
        check_prerequisites
        validate_config
        start_bot "$channel"
        ;;
    "stop-channel")
        if [ -z "$channel" ]; then
            print_error "Channel name required"
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
        check_prerequisites
        build_image
        ;;
    "list-channels")
        if [ -z "$channel" ]; then
            print_error "Bot name required"
            exit 1
        fi
        list_channels_by_bot "$channel"
        ;;
    "restart-all")
        check_prerequisites
        validate_config
        restart_all_bots
        ;;
    "status")
        show_status "$channel"
        ;;
    "logs")
        view_logs "$channel"
        ;;
    "cloudwatch-logs")
        view_cloudwatch_logs "$channel"
        ;;
    "deploy")
        if [ -z "$channel" ]; then
            print_error "Channel name required for deploy command"
            exit 1
        fi
        check_prerequisites
        validate_config
        build_image
        start_bot "$channel"
        print_status "Deployment completed successfully!"
        ;;
    *)
        if [ -n "$command" ] && [ -z "$channel" ]; then
            # If only one argument is provided, treat it as a channel name
            check_prerequisites
            validate_config
            start_bot "$command"
        else
            echo "Usage: $0 [command] [channel]"
            echo ""
            echo "Commands:"
            echo "  start <channel>     - Start a bot instance"
            echo "  stop-channel <channel> - Stop a specific channel's bot"
            echo "  stop-all           - Stop all bot instances"
            echo "  list               - List all running bot instances"
            echo "  build              - Build the Docker image"
            echo "  list-channels <bot> - List all channels using a specific bot"
            echo "  restart-all        - Stop and restart all bots with latest image"
            echo "  status [channel]   - Show bot status (all or specific channel)"
            echo "  logs <channel>     - View bot logs for a specific channel"
            echo "  cloudwatch-logs <channel> - View CloudWatch logs for a specific channel"
            echo "  deploy <channel>   - Build image and deploy a specific channel"
            echo ""
            echo "Shortcut: $0 <channel> (same as start)"
            echo ""
            echo "Examples:"
            echo "  $0 start PerfectTilt"
            echo "  $0 stop-channel PerfectTilt"
            echo "  $0 list"
            echo "  $0 restart-all"
            echo "  $0 status PerfectTilt"
            echo "  $0 logs PerfectTilt"
            echo "  $0 cloudwatch-logs PerfectTilt"
            exit 1
        fi
        ;;
esac
