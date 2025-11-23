# Savegress Platform Makefile

# Variables
APP_NAME := savegress-api
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X 'github.com/savegress/platform/backend/internal/handlers.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'"

# Server configuration
SERVER_USER := infra
SERVER_IP := 88.198.170.206
SERVER_PATH := /home/infra/savegress-platform

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# Paths
BACKEND_PATH := ./backend
CMD_PATH := $(BACKEND_PATH)/cmd/api
BIN_PATH := ./bin

.PHONY: help build build-linux clean deps test test-coverage lint fmt run dev \
        docker-dev docker-dev-build docker-dev-down docker-dev-logs docker-dev-clean \
        frontend deploy deploy-quick server-logs server-status server-ssh server-restart

# Default target
help:
	@echo "Savegress Platform - Available commands:"
	@echo ""
	@echo "  Backend (Go):"
	@echo "    make build        - Build backend binary"
	@echo "    make build-linux  - Build for Linux amd64"
	@echo "    make run          - Build and run backend"
	@echo "    make dev          - Run backend in dev mode (go run)"
	@echo "    make test         - Run tests"
	@echo "    make test-coverage- Run tests with coverage"
	@echo "    make lint         - Run linter"
	@echo "    make fmt          - Format code"
	@echo "    make deps         - Download dependencies"
	@echo "    make clean        - Clean build artifacts"
	@echo ""
	@echo "  Frontend:"
	@echo "    make frontend     - Run frontend dev server"
	@echo "    make frontend-install - Install frontend dependencies"
	@echo ""
	@echo "  Docker (Local Development):"
	@echo "    make docker-dev       - Start all services locally"
	@echo "    make docker-dev-build - Build and start all services"
	@echo "    make docker-dev-down  - Stop all services"
	@echo "    make docker-dev-logs  - View logs from all services"
	@echo "    make docker-dev-clean - Stop and remove containers, volumes"
	@echo ""
	@echo "  Deployment:"
	@echo "    make deploy       - Full deploy (sync + rebuild)"
	@echo "    make deploy-quick - Quick deploy (sync + restart)"
	@echo "    make server-logs  - View server logs"
	@echo "    make server-status- Check server status"
	@echo "    make server-ssh   - SSH into server"
	@echo "    make server-restart - Restart services"

# ==================== Backend (Go) ====================

build:
	@echo "Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BIN_PATH)
	cd $(BACKEND_PATH) && $(GOBUILD) $(LDFLAGS) -o ../$(BIN_PATH)/$(APP_NAME) ./cmd/api

build-linux:
	@echo "Building $(APP_NAME) for Linux amd64..."
	@mkdir -p $(BIN_PATH)
	cd $(BACKEND_PATH) && GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o ../$(BIN_PATH)/$(APP_NAME)-linux-amd64 ./cmd/api

build-linux-arm64:
	@echo "Building $(APP_NAME) for Linux arm64..."
	@mkdir -p $(BIN_PATH)
	cd $(BACKEND_PATH) && GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o ../$(BIN_PATH)/$(APP_NAME)-linux-arm64 ./cmd/api

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_PATH)
	cd $(BACKEND_PATH) && $(GOCMD) clean -testcache

deps:
	@echo "Downloading dependencies..."
	cd $(BACKEND_PATH) && $(GOMOD) download && $(GOMOD) tidy

test:
	@echo "Running tests..."
	cd $(BACKEND_PATH) && $(GOTEST) -v -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	cd $(BACKEND_PATH) && $(GOTEST) -v -race -coverprofile=coverage.out ./...
	cd $(BACKEND_PATH) && $(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: backend/coverage.html"

lint:
	@echo "Running linter..."
	cd $(BACKEND_PATH) && $(GOLINT) run ./...

fmt:
	@echo "Formatting code..."
	cd $(BACKEND_PATH) && $(GOFMT) -s -w .

run: build
	@echo "Running $(APP_NAME)..."
	$(BIN_PATH)/$(APP_NAME)

dev:
	@echo "Running backend in development mode..."
	cd $(BACKEND_PATH) && $(GOCMD) run ./cmd/api

# ==================== Frontend ====================

frontend:
	cd frontend && npm run dev

frontend-install:
	cd frontend && npm install

# ==================== Docker (Local Development) ====================

docker-dev:
	cd docker && docker compose -f docker-compose.local.yml up

docker-dev-build:
	cd docker && docker compose -f docker-compose.local.yml up --build

docker-dev-down:
	cd docker && docker compose -f docker-compose.local.yml down

docker-dev-logs:
	cd docker && docker compose -f docker-compose.local.yml logs -f

docker-dev-clean:
	cd docker && docker compose -f docker-compose.local.yml down -v --rmi local

# ==================== Deployment ====================

deploy:
	@echo ">>> Deploying to $(SERVER_IP)..."
	./scripts/deploy.sh

deploy-quick:
	@echo ">>> Quick deploy to $(SERVER_IP)..."
	./scripts/deploy-quick.sh

server-logs:
	ssh $(SERVER_USER)@$(SERVER_IP) "cd $(SERVER_PATH)/docker && docker compose logs -f"

server-status:
	ssh $(SERVER_USER)@$(SERVER_IP) "cd $(SERVER_PATH)/docker && docker compose ps"

server-ssh:
	ssh $(SERVER_USER)@$(SERVER_IP)

server-restart:
	ssh $(SERVER_USER)@$(SERVER_IP) "cd $(SERVER_PATH)/docker && docker compose restart"

# ==================== Database ====================

migrate-create:
	@echo "Creating new migration..."
	@read -p "Migration name: " name; \
	touch $(BACKEND_PATH)/migrations/$$(date +%Y%m%d%H%M%S)_$$name.up.sql; \
	touch $(BACKEND_PATH)/migrations/$$(date +%Y%m%d%H%M%S)_$$name.down.sql

# ==================== Misc ====================

generate:
	@echo "Running go generate..."
	cd $(BACKEND_PATH) && $(GOCMD) generate ./...

security:
	@echo "Running security scan..."
	cd $(BACKEND_PATH) && gosec ./...
