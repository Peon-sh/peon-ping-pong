# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS build
WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates git && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=0.0.1
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
  -ldflags="-s -w -X github.com/peon-sh/peon-ping-pong/internal/config.Version=${VERSION}" \
  -o /out/peon-ping-pong ./cmd/peon-ping-pong

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/peon-ping-pong /peon-ping-pong
USER nonroot:nonroot
EXPOSE 8888
ENTRYPOINT ["/peon-ping-pong"]
