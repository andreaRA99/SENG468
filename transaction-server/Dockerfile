# syntax=docker/dockerfile:1

# source: https://hub.docker.com/_/golang
FROM golang:1.20 AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY cache/go.mod cache/go.sum cache/
COPY transaction-server/go.mod transaction-server/go.sum transaction-server/
RUN \
	--mount=type=cache,id=go-pkg,target=/go/pkg,sharing=shared \
	go work init \
	&& go work use cache \
	&& go work use transaction-server \
	&& go mod download

COPY cache cache
COPY transaction-server transaction-server
RUN --network=none \
	--mount=type=cache,id=go-pkg,target=/go/pkg,readonly \
	--mount=type=cache,id=go-build,target=/root/.cache/go-build,sharing=shared \
	cd transaction-server && go build -v


FROM debian:bullseye-slim

COPY --from=builder /usr/src/app/transaction-server/transaction-server ./

ENTRYPOINT ["./transaction-server"]
EXPOSE 5000
