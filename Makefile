.PHONY: all deps fmt lint build test test-unit test-integration man install check help clean

VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo dev)

all: check

deps:
	go mod download && go mod tidy

fmt:
	go fmt ./...

lint:
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@PATH="$(shell go env GOPATH)/bin:$(PATH)" golangci-lint run

build:
	go build -ldflags "-X github.com/roboalchemist/desearch-cli/cmd.version=$(VERSION)" -o desearch .

test: build
	./desearch --help
	./desearch version
	./desearch docs > /dev/null
	./desearch skill print > /dev/null
	./desearch completion --help
	./desearch config --help

test-unit:
	go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep total || true
	@echo "Checking pkg/ coverage..."
	@total=$$(go tool cover -func=coverage.out | grep "total" | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$total" != "" ] && awk "BEGIN {exit !($$total < 75)}"; then \
		echo "WARNING: Overall coverage is $${total}%, which is below target"; \
	fi

test-integration:
	go test -v -tags=integration -run TestIntegration ./...
	@echo "Integration tests passed"

test-integration-live:
	go test -v -tags=integration -run TestIntegration_LiveAPI ./... \
		-DESEARCH_API_KEY=$$DESEARCH_API_KEY

man:
	go run ./cmd/gendocs

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
	@echo "  test-integration-live - integration tests with live API (requires DESEARCH_API_KEY)"
	@echo "  man               - generate man pages"
	@echo "  install           - install to /usr/local/bin/"
	@echo "  check             - fmt + lint + test + test-unit"
	@echo "  clean             - remove build artifacts"
