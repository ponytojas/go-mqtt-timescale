# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy only go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mqtt-timescale ./cmd

# Final stage
FROM alpine:3.18

# Add CA certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/mqtt-timescale .

# Switch to non-root user
USER appuser

# Command to run
CMD ["./mqtt-timescale"]