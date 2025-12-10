# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Runtime stage
FROM ubuntu:24.04

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create directory for SSH host keys
RUN mkdir -p /app/.ssh

# Copy binary from builder
COPY --from=builder /server /app/server

# Expose SSH port
EXPOSE 2222

# Run the server
CMD ["/app/server"]
