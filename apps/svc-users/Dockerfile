# ---- build stage ----
FROM golang:1.24.5-alpine3.21 AS build-env

RUN apk add --no-cache git

WORKDIR /src
COPY go.work go.work
COPY apps/svc-users/ apps/svc-users/
COPY packages/shared-go/ packages/shared-go/

WORKDIR /src/apps/svc-users

RUN go mod download

RUN go build -o /server

# ---- runtime stage ----
FROM alpine:3.22

EXPOSE 8080

RUN apk add --no-cache libc6-compat

WORKDIR /

COPY --from=build-env /server /

CMD ["/server"]
