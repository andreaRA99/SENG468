# syntax=docker/dockerfile:1

FROM golang:1.20 AS builder

WORKDIR /usr/src/app

COPY . .
RUN --network=none --mount=type=cache,target=/root/.cache/go-build go build -v


FROM debian:bullseye-slim

COPY --from=builder /usr/src/app/quote_server ./

ENTRYPOINT ["./quote_server"]
