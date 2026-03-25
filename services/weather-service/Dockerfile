FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy shared modules first for dependency caching
COPY shared/go/model/ shared/go/model/
COPY shared/go/tenant/ shared/go/tenant/
COPY shared/go/logging/ shared/go/logging/
COPY shared/go/database/ shared/go/database/
COPY shared/go/server/ shared/go/server/
COPY shared/go/auth/ shared/go/auth/

# Copy service
COPY services/weather-service/ services/weather-service/

RUN cd services/weather-service && GOWORK=off go build -o /server ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy custom CA certificates (supports environments with TLS inspection proxies like Zscaler).
# To use: place .crt files in services/weather-service/certs/ (gitignored).
# No-op when directory contains no .crt files.
COPY services/weather-service/certs/ /tmp/custom-certs/
RUN cp /tmp/custom-certs/*.crt /usr/local/share/ca-certificates/ 2>/dev/null; \
    update-ca-certificates 2>/dev/null; \
    rm -rf /tmp/custom-certs

COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
