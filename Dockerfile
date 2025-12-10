# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS connections
RUN apk add --no-cache ca-certificates

# Create directory for SSH host keys
RUN mkdir -p /app/.ssh

# Copy binary from builder
COPY --from=builder /server /app/server

# Expose SSH port
EXPOSE 2222

# Run the server
CMD ["/app/server"]
