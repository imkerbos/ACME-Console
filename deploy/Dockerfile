# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Enable Go toolchain auto-download for newer versions
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://goproxy.cn,direct

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/acme-console ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/acme-console /app/acme-console
COPY --from=builder /app/configs /app/configs

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 10020

ENTRYPOINT ["/app/acme-console"]
CMD ["-config", "/app/configs/config.yaml"]
