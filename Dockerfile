# Build stage
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bot cmd/bot/main.go

# Final stage
FROM alpine:latest

# Add version label
ARG VERSION=dev
LABEL version=$VERSION
LABEL maintainer="PBChatBot Team"

WORKDIR /app

# Install timezone data
RUN apk add --no-cache tzdata

# Copy the binary from builder
COPY --from=builder /app/bot .

# Create configs directory
RUN mkdir -p /app/configs

# Create data directory
RUN mkdir -p /app/data

# Set environment variables
ENV TZ=America/New_York
ENV VERSION=$VERSION
ENV CONFIG_PATH=/app/configs
ENV DATA_PATH=/app/data

# Note: For production, mount the bot auth and channel config files:
# docker run -v /path/to/bot_auth.yaml:/app/configs/bot_auth.yaml \
#          -v /path/to/channel_config.yaml:/app/configs/channel_config_secrets.yaml \
#          -v bot-data:/app/data \
#          pbchatbot

# Run the bot
CMD ["./bot"] 