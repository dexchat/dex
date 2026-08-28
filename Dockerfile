# syntax=docker/dockerfile:1

FROM golang:1.25.4-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dex ./cmd/tui

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S dex \
    && adduser -S -G dex dex \
    && mkdir -p /data/config /data/state \
    && chown -R dex:dex /data

COPY --from=build /out/dex /usr/local/bin/dex

ENV XDG_CONFIG_HOME=/data/config \
    XDG_DATA_HOME=/data/state

VOLUME ["/data"]
USER dex
ENTRYPOINT ["/usr/local/bin/dex"]
