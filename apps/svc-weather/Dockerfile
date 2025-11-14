# ---- build stage ----
FROM golang:1.24.5-alpine3.21 AS build-env

RUN apk add --no-cache git

WORKDIR /src
COPY apps/svc-weather/ apps/svc-weather/
COPY packages/shared-go/ packages/shared-go/

WORKDIR /src/apps/svc-weather

# Build without go.work - each service is independent
# Use go mod edit to replace the shared-go module with local path
RUN go mod edit -replace github.com/jtumidanski/home-hub/packages/shared-go=../../packages/shared-go
RUN go mod tidy
RUN go mod download

RUN go build -o /server

# ---- runtime stage ----
FROM alpine:3.22

EXPOSE 8080

RUN apk add --no-cache libc6-compat ca-certificates

WORKDIR /

COPY --from=build-env /server /

CMD ["/server"]
