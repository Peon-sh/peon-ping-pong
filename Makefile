.PHONY: test lint build docker mock run-smoke tidy

VERSION ?= 0.0.1
IMAGE ?= ghcr.io/peon-sh/peon-ping-pong:$(VERSION)
LDFLAGS := -ldflags="-s -w -X github.com/peon-sh/peon-ping-pong/internal/config.Version=$(VERSION)"

tidy:
	go mod tidy

test:
	go test ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || go vet ./...

build:
	go build $(LDFLAGS) -o bin/peon-ping-pong ./cmd/peon-ping-pong
	go build -o bin/mock-receiver ./cmd/mock-receiver

docker:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

mock:
	go run ./cmd/mock-receiver

# Local smoke: start mock receiver, then agent against it (requires Docker socket).
run-smoke: build
	@echo "Start mock in another terminal: make mock"
	@echo "Then:"
	@echo "  TOKEN=dev PUSH_ENDPOINT=http://127.0.0.1:9090/api/v1/agents/push SERVER_ID=local-dev PUSH_INTERVAL_SECONDS=10 ./bin/peon-ping-pong"
