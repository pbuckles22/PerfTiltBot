# Build stage
FROM golang:1.23-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -o perftiltbot ./cmd/bot

# Final stage
FROM alpine:latest

# Add version label
ARG VERSION=dev
LABEL version=$VERSION
LABEL maintainer="PerfTiltBot Team"

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/perftiltbot .

# Create configs directory
RUN mkdir -p /app/configs

# Create data directory
RUN mkdir -p /app/data

# Set environment variables
ENV TZ=UTC
ENV VERSION=$VERSION

# Run the application
CMD ["./perftiltbot"]

# Note: For production, mount both secrets.yaml and channel-specific data directory:
# docker run -v /path/to/secrets.yaml:/app/configs/secrets.yaml \
#          -v channel_data:/app/data \
#          perftiltbot 