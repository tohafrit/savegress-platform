# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY backend/go.mod backend/go.sum* ./backend/

# Download dependencies
RUN cd backend && go mod download

# Copy source code
COPY backend ./backend

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME

# Build the application
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X 'github.com/savegress/platform/backend/internal/handlers.Version=${VERSION}'" \
    -o /app/savegress-api ./cmd/api

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S savegress && \
    adduser -u 1000 -S savegress -G savegress

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/savegress-api /app/savegress-api

# Copy migrations
COPY --from=builder /app/backend/migrations /app/migrations

# Set ownership
RUN chown -R savegress:savegress /app

# Switch to non-root user
USER savegress

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# Run the application
ENTRYPOINT ["/app/savegress-api"]
