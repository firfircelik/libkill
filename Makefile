.PHONY: build test lint clean install build-all build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

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
	@echo "build-all requires C cross-compilers for each target (CGO is mandatory)"
	@echo "Use 'make build-release' to build only native platform, or see .github/workflows/release.yml"
	@echo ""

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o dist/libkill-linux-amd64 $(CMD_DIR)

build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o dist/libkill-linux-arm64 $(CMD_DIR)

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o dist/libkill-darwin-amd64 $(CMD_DIR)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o dist/libkill-darwin-arm64 $(CMD_DIR)

build-windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o dist/libkill-windows-amd64.exe $(CMD_DIR)
