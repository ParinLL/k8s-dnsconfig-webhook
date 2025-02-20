# Stage 1: Build the Go binary
FROM golang:1.21-alpine  --platform=$TARGETPLATFORM AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Create a non-root user for the final image
RUN adduser -D -g '' webhook-user

# Copy go mod files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the binary with security flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/webhook-server \
    ./cmd/webhook

# Stage 2: Create the final minimal image
FROM alpine:3.19

# Add necessary certificates and security updates
RUN apk --no-cache add ca-certificates && \
    apk --no-cache add -U tzdata && \
    update-ca-certificates

# Create necessary directories with correct permissions
RUN mkdir -p /etc/webhook/certs && \
    chown -R nobody:nobody /etc/webhook

# Copy the binary from builder
COPY --from=builder /app/webhook-server /usr/local/bin/
COPY --from=builder /etc/passwd /etc/passwd

# Set working directory
WORKDIR /usr/local/bin

# Use non-root user
USER nobody:nobody

# Expose webhook port
EXPOSE 8443

# Define volume for certificates
VOLUME ["/etc/webhook/certs"]

# Set entry point
ENTRYPOINT ["webhook-server"]