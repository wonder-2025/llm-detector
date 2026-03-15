# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o llm-detector cmd/detector/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary
COPY --from=builder /app/llm-detector /usr/local/bin/llm-detector

# Copy fingerprint data
COPY --from=builder /app/pkg/fingerprints/data /app/fingerprints

# Set environment variable for fingerprint path
ENV LLM_DETECTOR_FINGERPRINT_PATH=/app/fingerprints

# Run as non-root user
RUN adduser -D -s /bin/sh detector
USER detector

ENTRYPOINT ["llm-detector"]
CMD ["--help"]
