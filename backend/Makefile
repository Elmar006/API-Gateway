SHELL          := /usr/bin/env bash
BIN            := bin/gateway
PKG            := ./...
GOFLAGS        ?=
COVER_OUT      := coverage.out
COVER_HTML     := coverage.html
DB_DSN_DEFAULT := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATE_DSN    ?= postgres://$${DB_USER:-postgres}:$${DB_PASSWORD:-postgres}@$${DB_HOST:-localhost}:$${DB_PORT:-5432}/$${DB_NAME:-postgres}?sslmode=$${DB_SSLMODE:-disable}

.PHONY: help build run tidy test test-integration cover lint vet fmt clean migrate-up migrate-down

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'

build: ## Build the gateway binary into bin/gateway
	@mkdir -p bin
	go build -trimpath -ldflags='-s -w' -o $(BIN) ./cmd

run: ## Run gateway with .env
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi && go run ./cmd

tidy: ## go mod tidy
	go mod tidy

fmt: ## gofmt -s -w .
	gofmt -s -w .

vet: ## go vet
	go vet $(PKG)

lint: vet ## go vet + golangci-lint (skips lint if not installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run $(PKG); \
	else \
		echo "golangci-lint not installed: install via 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"; \
	fi

test: ## Run unit tests with coverage
	go test $(GOFLAGS) -count=1 -race -cover $(PKG)

test-integration: ## Run integration tests (requires Docker)
	go test $(GOFLAGS) -count=1 -tags=integration $(PKG)

cover: ## HTML coverage report (coverage.html)
	go test -count=1 -coverprofile=$(COVER_OUT) $(PKG)
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo "open $(COVER_HTML)"

migrate-up: ## Apply all pending migrations
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi && \
	migrate -source file://migrations -database "$(MIGRATE_DSN)" up

migrate-down: ## Rollback one migration
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi && \
	migrate -source file://migrations -database "$(MIGRATE_DSN)" down 1

clean: ## Remove build artefacts
	rm -rf bin $(COVER_OUT) $(COVER_HTML)
