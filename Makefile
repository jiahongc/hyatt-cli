.PHONY: build test lint install clean

build:
	go build -o bin/hyatt-cli ./cmd/hyatt-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/hyatt-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/hyatt-mcp ./cmd/hyatt-mcp

install-mcp:
	go install ./cmd/hyatt-mcp

build-all: build build-mcp
