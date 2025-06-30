# Build stage
FROM golang:1.24-alpine AS builder

# Build arguments for metadata
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with version information
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE} -extldflags '-static'" \
    -a -installsuffix cgo \
    -o mimir-limit-optimizer \
    ./main.go

# Verify the binary
RUN ./mimir-limit-optimizer --version || true

# Final stage
FROM gcr.io/distroless/static:nonroot

# Build arguments for labels
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# OCI Labels for metadata
LABEL org.opencontainers.image.title="Mimir Limit Optimizer"
LABEL org.opencontainers.image.description="Enterprise-grade Kubernetes controller for managing per-tenant limits in Grafana Mimir"
LABEL org.opencontainers.image.url="https://github.com/AkshayDubey29/mimir-limit-optimizer"
LABEL org.opencontainers.image.source="https://github.com/AkshayDubey29/mimir-limit-optimizer"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.vendor="Akshay Dubey"
LABEL org.opencontainers.image.authors="Akshay Dubey <akshay@example.com>"

# Copy CA certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder stage with proper permissions
COPY --from=builder --chmod=755 /app/mimir-limit-optimizer /usr/local/bin/mimir-limit-optimizer

# Use non-root user (distroless already has proper tmp directory setup)
USER 65534:65534

# Set working directory
WORKDIR /

# Environment variables for runtime configuration
ENV TZ=UTC
ENV GOGC=50
ENV GOMAXPROCS=2

# Expose metrics port (default Kubernetes controller port)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD ["/usr/local/bin/mimir-limit-optimizer", "--health-check"] || exit 1

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/mimir-limit-optimizer"] 