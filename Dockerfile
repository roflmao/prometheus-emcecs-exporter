# Dockerfile for Prometheus EMC ECS Exporter
# Multi-stage build for minimal final image

# Stage 1: Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Use build flags from Makefile for version information
ARG BUILD_TIME
ARG RELEASE
ARG COMMIT

RUN BUILD_TIME=${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')} \
    RELEASE=${RELEASE:-dev} \
    COMMIT=${COMMIT:-unknown} \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags="-w -s \
    -X main.commit=${COMMIT} \
    -X main.date=${BUILD_TIME} \
    -X main.version=${RELEASE}" \
    -o prometheus-emcecs-exporter \
    ./cmd

# Stage 2: Final minimal image
FROM scratch

# Copy CA certificates for HTTPS connections to ECS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder
COPY --from=builder /build/prometheus-emcecs-exporter /prometheus-emcecs-exporter

# Expose exporter port
EXPOSE 9438

# Set default bind address to allow external access in container
ENV ECSENV_BIND_ADDRESS=0.0.0.0

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/prometheus-emcecs-exporter", "--help"]

# Run as non-root (using numeric UID for scratch image)
USER 65534:65534

# Set entrypoint
ENTRYPOINT ["/prometheus-emcecs-exporter"]

# Default command (can be overridden)
CMD []
