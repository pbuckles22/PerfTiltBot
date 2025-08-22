#!/bin/bash

# Setup script for Twitch OAuth Web Helper
# This script installs dependencies and runs the web helper

set -e

echo "🤖 Setting up Twitch OAuth Web Helper..."

# Check if Python 3 is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed. Installing..."
    sudo yum update -y
    sudo yum install -y python3 python3-pip
fi

# Install pip if not available
if ! command -v pip3 &> /dev/null; then
    echo "❌ pip3 is not installed. Installing..."
    sudo yum install -y python3-pip
fi

# Install Flask and requests
echo "📦 Installing Python dependencies..."
pip3 install Flask==2.3.3 requests==2.31.0

# Make the web helper executable
chmod +x twitch_oauth_web.py

echo "✅ Setup complete!"
echo ""
echo "🌐 To start the web helper:"
echo "   python3 twitch_oauth_web.py"
echo ""
echo "🔧 Options:"
echo "   --port 3000     # Change port (default: 3000)"
echo "   --debug         # Enable debug mode"
echo ""
echo "📝 Remember to:"
echo "   1. Add port 3000 to your EC2 security group"
echo "   2. Access via: http://YOUR_EC2_IP:3000"
echo "   3. Only your IP addresses should have access"
