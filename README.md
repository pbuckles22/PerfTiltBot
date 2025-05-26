# PBChatBot

A Twitch chat bot for managing queues and other channel features.

## Documentation

- [Windows Setup Guide](docs/windows.md) - Windows-specific setup and usage instructions
- [AWS Deployment Guide](docs/aws.md) - AWS deployment and management instructions
- [Configuration Examples](configs/examples/README.md) - Example configuration files and setup

## Prerequisites

1. **Go 1.23 or higher**
   ```bash
   # Install Go
   go version
   ```

2. **Docker**
   - Docker Desktop (Windows/macOS)
   - Docker Engine (Linux)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/pbuckles22/PBChatBot.git
   cd PBChatBot
   ```

2. **Create configuration files**
   - See [Configuration Examples](configs/examples/README.md) for details

3. **Run the bot**
   ```bash
   # Using bash script (Linux/macOS)
   ./run_bot.sh start <channel>

   # Using PowerShell script (Windows)
   .\run_bot.ps1 start <channel>
   ```

## Features

- Queue management
- Command cooldowns
- User permissions
- Channel-specific settings
- Docker support
- AWS deployment

## Development

1. **Build**
   ```bash
   # Build for current platform
   go build -o bot cmd/bot/main.go

   # Build Docker image
   docker build -t pbchatbot .
   ```

2. **Test**
   ```bash
   go test ./...
   ```

3. **Run**
   ```bash
   # Run locally
   ./bot

   # Run in Docker
   docker run -d \
       --name pbchatbot \
       -v $(pwd)/configs:/app/configs \
       -v pbchatbot-data:/app/data \
       pbchatbot
   ```

## Platform-Specific Guides

- [Windows Setup Guide](docs/windows.md) - Windows-specific setup and usage instructions
- [AWS Deployment Guide](docs/aws.md) - AWS deployment and management instructions

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 