# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download || true && go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o sunspear .

# Runtime stage
FROM alpine:3.23

RUN apk --no-cache add ca-certificates sqlite-libs wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/sunspear .

# Copy data directory structure
RUN mkdir -p /app/data/apps/compose-templates /app/data/database

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./sunspear"]
