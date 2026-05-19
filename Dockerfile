# ---- Build Stage ----
# Use an Alpine Go image matching the module directive in go.mod
FROM golang:1.25.6-alpine AS builder

WORKDIR /build

# Copy module files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy remaining source
COPY . .

# Build the server binary with CGO disabled for a fully static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/taskagent ./cmd/server

# ---- Runtime Stage ----
# Minimal Alpine image with only ca-certificates for TLS
FROM alpine:3.20

# Install CA certificates for TLS verification
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create the data directory and set ownership for the non-root user
RUN mkdir -p /data && chown appuser:appgroup /data

# Copy the built binary from the builder stage
COPY --from=builder /out/taskagent /usr/local/bin/taskagent

# Set environment variables for container operation
ENV TASKAGENT_LISTEN_ADDR=:8080
ENV TASKAGENT_DB_PATH=/data/taskagent.db

# Expose the HTTP port
EXPOSE 8080

# Health check using wget (available in Alpine)
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Switch to non-root user
USER appuser

# Run the server binary
ENTRYPOINT ["/usr/local/bin/taskagent"]
CMD []
