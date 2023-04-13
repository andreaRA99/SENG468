# syntax=docker/dockerfile:1

# source: https://hub.docker.com/_/golang
FROM golang:1.20 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY cache/go.mod cache/go.sum cache/
COPY quote_server/go.mod quote_server/go.sum quote_server/
RUN go work init \
    && go work use cache \
    && go work use quote_server \
    && go mod download

COPY cache cache
COPY quote_server quote_server
RUN --network=none --mount=type=cache,target=/root/.cache/go-build cd quote_server && go build -v


FROM debian:bullseye-slim

COPY --from=builder /usr/src/app/quote_server/quote_server ./

ENTRYPOINT ["./quote_server"]