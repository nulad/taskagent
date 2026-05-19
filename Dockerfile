# ---- Build stage ----
FROM golang:1.25-alpine AS builder

RUN apk --no-cache add git gcc musl-dev

WORKDIR /app

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o /taskagent ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary
COPY --from=builder /taskagent /app/taskagent

# Create data directory for persistent SQLite storage
RUN mkdir -p /data

EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/taskagent"]
