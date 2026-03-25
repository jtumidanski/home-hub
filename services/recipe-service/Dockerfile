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
COPY services/recipe-service/ services/recipe-service/

RUN cd services/recipe-service && GOWORK=off go build -o /server ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
