# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION}" \
    -o /app/bin/siteproof-api \
    ./cmd/api

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Install Typst for PDF generation
RUN apk add --no-cache --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community typst

# Create non-root user
RUN addgroup -g 1000 siteproof && \
    adduser -u 1000 -G siteproof -s /bin/sh -D siteproof

WORKDIR /opt/siteproof

# Copy binary from builder
COPY --from=builder /app/bin/siteproof-api ./bin/

# Copy templates and migrations
COPY templates ./templates/
COPY migrations ./migrations/

# Create directories for uploads
RUN mkdir -p /data/uploads && chown -R siteproof:siteproof /data

# Set ownership
RUN chown -R siteproof:siteproof /opt/siteproof

# Switch to non-root user
USER siteproof

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./bin/siteproof-api"]
