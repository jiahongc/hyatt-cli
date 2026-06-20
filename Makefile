.PHONY: build test lint install clean

build:
	go build -o bin/hyatt-pp-cli ./cmd/hyatt-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/hyatt-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/hyatt-pp-mcp ./cmd/hyatt-pp-mcp

install-mcp:
	go install ./cmd/hyatt-pp-mcp

build-all: build build-mcp
