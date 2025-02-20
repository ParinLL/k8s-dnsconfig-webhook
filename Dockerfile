# Stage 1: Build the Go binary
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /src

# Create a non-root user for the final image
RUN adduser -D -g '' webhook-user

# Copy go mod files first to leverage Docker cache
COPY go.mod go.sum* ./

# Download dependencies (handle case where go.sum might not exist yet)
RUN go mod download || true

# Copy the source code
COPY . .

# Verify modules and generate go.sum if needed
RUN go mod verify || true
RUN go mod tidy

# Set build arguments for multi-platform support
ARG TARGETARCH
ARG TARGETOS

# Build the binary with security flags
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-w -s" \
    -o /src/webhook-server \
    ./cmd/webhook

# Stage 2: Create the final minimal image
FROM --platform=$TARGETPLATFORM alpine:3.19

# Add necessary certificates and security updates
RUN apk --no-cache add ca-certificates && \
    apk --no-cache add -U tzdata && \
    update-ca-certificates

# Create necessary directories with correct permissions
RUN mkdir -p /etc/webhook/certs && \
    chown -R nobody:nobody /etc/webhook

# Copy the binary from builder
COPY --from=builder /src/webhook-server /usr/local/bin/
COPY --from=builder /etc/passwd /etc/passwd

# Set working directory
WORKDIR /usr/local/bin

# Use non-root user
USER nobody:nobody

# Expose webhook port
EXPOSE 8443

# Define volume for certificates
VOLUME ["/etc/webhook/certs"]

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8443/health || exit 1

# Set entry point
ENTRYPOINT ["webhook-server"]