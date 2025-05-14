# PerfTiltBot

A performance-focused bot for analyzing and providing insights about gameplay tilt.

## Prerequisites

- Go 1.23.1 or higher
- Git

## Setup

1. Clone the repository:
```bash
git clone https://github.com/pbuckles22/PerfTiltBot.git
cd PerfTiltBot
```

2. Install dependencies:
```bash
go mod tidy
```

## Project Structure

```
.
├── cmd/           # Application entry points
├── internal/      # Private application and library code
├── pkg/          # Library code that could be used by external applications
└── configs/      # Configuration files
```

## Development

To run the bot locally:

```bash
go run cmd/bot/main.go
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 