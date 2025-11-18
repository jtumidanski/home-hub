# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy shared packages (needed for replace directive in go.mod)
COPY packages/shared-go/ packages/shared-go/

# Copy service go.mod and go.sum first for better caching
COPY apps/svc-ai/go.mod apps/svc-ai/go.sum apps/svc-ai/

# Download dependencies
WORKDIR /app/apps/svc-ai
RUN go mod download

# Copy service source code
WORKDIR /app
COPY apps/svc-ai/ apps/svc-ai/

# Build the service
WORKDIR /app/apps/svc-ai
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o svc-ai .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/apps/svc-ai/svc-ai .

# Expose port
EXPOSE 8080

# Run the service
CMD ["./svc-ai"]
