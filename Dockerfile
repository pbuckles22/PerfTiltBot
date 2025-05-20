# Use the official Go image as the base image
FROM golang:latest AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot

# Use a minimal alpine image for the final image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bot .

# Create necessary directories and set permissions
RUN mkdir -p /app/configs /app/data && \
    chmod 755 /app/configs /app/data

# Create a non-root user
RUN adduser -D -g '' appuser

# Copy configuration files from the builder stage
COPY --from=builder /app/configs/bot.yaml ./configs/bot.yaml

# Set proper ownership of all directories and files
RUN chown -R appuser:appuser /app/configs /app/data

# Switch to non-root user
USER appuser

# Create a directory for secrets
RUN mkdir -p /app/configs

# Define volume for queue state data
VOLUME ["/app/data"]

# Run the bot
CMD ["./bot"]

# Note: For production, mount both secrets.yaml and channel-specific data directory:
# docker run -v /path/to/secrets.yaml:/app/configs/secrets.yaml \
#          -v channel_data:/app/data \
#          perftiltbot 