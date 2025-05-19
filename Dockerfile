# Use the official Go image as the base image
FROM golang:1.21-alpine AS builder

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

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Create a directory for secrets
RUN mkdir -p /app/configs

# Copy secrets.yaml securely (ensure it is not committed to version control)
COPY --from=builder /app/configs/secrets.yaml ./configs/secrets.yaml

# Run the bot
CMD ["./bot"]

# Note: Mount secrets.yaml as a volume at runtime:
# docker run -v /path/to/secrets.yaml:/app/configs/secrets.yaml perftiltbot 