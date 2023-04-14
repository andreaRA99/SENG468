# syntax=docker/dockerfile:1

# source: https://hub.docker.com/_/golang
FROM golang:1.20 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY cache/go.mod cache/go.sum cache/
COPY polling_service/go.mod polling_service/go.sum polling_service/
RUN go work init \
    && go work use cache \
    && go work use polling_service \
    && go mod download

COPY cache cache
COPY polling_service polling_service
RUN --network=none --mount=type=cache,target=/root/.cache/go-build cd polling_service && go build -v


FROM debian:bullseye-slim

COPY --from=builder /usr/src/app/polling_service/polling_service ./

ENTRYPOINT ["./polling_service"]