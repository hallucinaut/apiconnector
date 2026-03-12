# Build stage
FROM golang:1.21-alpine AS builder

# Install git (required for Go modules) and ca-certificates for HTTPS access
RUN apk add --no-cache git ca-certificates

# Create working directory
WORKDIR /app

# Copy go.mod and go.sum files to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking and stripped debug info
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o apiconnector ./cmd/apiconnector

# Final stage: minimal runtime image
FROM alpine:latest

# Add non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy the compiled binary from builder stage
COPY --from=builder /app/apiconnector /usr/local/bin/apiconnector

# Expose any ports if needed (optional)
# EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["apiconnector"]