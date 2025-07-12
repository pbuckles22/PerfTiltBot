# Multi-Channel Bot Runner Script
# This script runs the bot with multiple channels

# Set environment variables
$env:BOT_NAME = "pbtestbot"
$env:CHANNEL_NAMES = "PerfectTilt,pbuckles,PerfectZombified"

Write-Host "Starting Multi-Channel Bot..."
Write-Host "Bot: $env:BOT_NAME"
Write-Host "Channels: $env:CHANNEL_NAMES"
Write-Host ""

# Run the multi-channel bot
go run ./cmd/multibot/ 