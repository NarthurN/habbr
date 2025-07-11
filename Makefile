.PHONY: help build run test clean generate docker-build docker-run lint format

# Variables
APP_NAME=posts-comments-api
VERSION?=v1.0.0
DOCKER_IMAGE=habbr/$(APP_NAME):$(VERSION)
GO_VERSION=1.23.4

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
generate: ## Generate GraphQL code
	go run github.com/99designs/gqlgen generate

build: ## Build the application
	go mod tidy
	go build -o bin/$(APP_NAME) ./cmd/server

run: ## Run the application
	go run ./cmd/server

dev: generate run ## Generate code and run in development mode

# Testing
test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...

test-unit: ## Run only unit tests
	go test -v -short ./...

test-integration: ## Run integration tests (requires PostgreSQL)
	go test -v -tags=integration ./internal/repository/postgres/

test-coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...
	goimports -w .

# Docker
docker-build: ## Build Docker image (production)
	docker build -t $(DOCKER_IMAGE) --target production .

docker-build-dev: ## Build Docker image (development)
	docker build -t habbr/$(APP_NAME):dev --target development .

docker-run: ## Run Docker container
	docker run -p 8080:8080 -e DATABASE_TYPE=memory $(DOCKER_IMAGE)

docker-compose-up: ## Start with docker-compose
	docker compose up --build

docker-compose-down: ## Stop docker-compose
	docker compose down -v

docker-dev: ## Start development environment
	./scripts/docker-dev.sh

docker-dev-tools: ## Start development environment with tools
	./scripts/docker-dev.sh --with-tools

docker-prod: ## Start production environment
	./scripts/docker-prod.sh

docker-prod-logs: ## Start production environment with logs
	./scripts/docker-prod.sh --logs

docker-clean: ## Clean Docker resources
	docker compose down -v --remove-orphans
	docker system prune -f

# Database
db-up: ## Start PostgreSQL with docker-compose
	docker compose up -d postgres

db-down: ## Stop PostgreSQL
	docker compose down postgres

# Cleanup
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

# Tools
install-tools: ## Install development tools
	go install github.com/99designs/gqlgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# All in one
setup: deps install-tools generate ## Setup development environment
