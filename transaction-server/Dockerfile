# syntax=docker/dockerfile:1

# source: https://hub.docker.com/_/golang
FROM golang:1.20 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN --network=none --mount=type=cache,target=/root/.cache/go-build go build -v


FROM debian:bullseye-slim

COPY --from=builder /usr/src/app/transaction-server ./

ENTRYPOINT ["./transaction-server"]
