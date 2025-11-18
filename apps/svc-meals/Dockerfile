# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy shared packages (needed for replace directive in go.mod)
COPY packages/shared-go/ packages/shared-go/

# Copy service go.mod and go.sum first for better caching
COPY apps/svc-meals/go.mod apps/svc-meals/go.sum* apps/svc-meals/

# Download dependencies
WORKDIR /app/apps/svc-meals
RUN go mod download

# Copy service source code
WORKDIR /app
COPY apps/svc-meals/ apps/svc-meals/

# Build the service
WORKDIR /app/apps/svc-meals
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o svc-meals .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/apps/svc-meals/svc-meals .

# Expose port
EXPOSE 8080

# Run the service
CMD ["./svc-meals"]
