FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o govent ./cmd/govent

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/govent .

# Copy config directory
COPY configs/ ./configs/

# Expose port
EXPOSE 8080

# Environment variables
ENV CONFIG_PATH=/app/configs/triggers.yaml

# Run the binary
CMD ["./govent"]
