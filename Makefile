.PHONY: all deps fmt lint build test test-unit test-integration man install check help clean

VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo dev)

all: check

deps:
	go mod download && go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run

build:
	go build -ldflags "-X github.com/roboalchemist/desearch-cli/cmd.version=$(VERSION)" -o desearch .

test: build
	./desearch --help
	./desearch version
	./desearch completion --help
	./desearch config --help

test-unit:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

test-integration:
	./desearch search "test" --dry-run

man:
	go run github.com/spf13/cobra-cli/cmd gendocs --help 2>/dev/null || true
	# generate man pages if cobra-doc available

install:
	sudo install -m 755 desearch /usr/local/bin/

check: fmt lint test test-unit

clean:
	rm -f desearch coverage.out
	rm -rf dist/

help:
	@echo "Available targets:"
	@echo "  deps              - download and tidy Go modules"
	@echo "  fmt               - format code"
	@echo "  lint              - run golangci-lint"
	@echo "  build             - build binary with version $(VERSION)"
	@echo "  test              - smoke tests (no API key needed)"
	@echo "  test-unit         - unit tests with coverage"
	@echo "  test-integration  - integration tests"
	@echo "  man               - generate man pages"
	@echo "  install           - install to /usr/local/bin/"
	@echo "  check             - fmt + lint + test + test-unit"
	@echo "  clean             - remove build artifacts"
