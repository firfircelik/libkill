.PHONY: build test lint clean install

BINARY := libkill
CMD_DIR := ./cmd/libkill

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build:
	go build -o $(BINARY) $(CMD_DIR)

test:
	go test ./... -race -count=1

test-cover:
	go test ./... -race -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY) coverage.out

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY) $(HOME)/.local/bin/

build-all:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -o dist/libkill-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o dist/libkill-darwin-arm64 $(CMD_DIR)
	GOOS=linux GOARCH=amd64 go build -o dist/libkill-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build -o dist/libkill-linux-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build -o dist/libkill-windows-amd64.exe $(CMD_DIR)
	@echo "Built all targets in dist/"
