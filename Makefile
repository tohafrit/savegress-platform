.PHONY: help dev dev-build dev-down dev-logs dev-clean prod prod-build prod-down prod-logs install db-studio deploy deploy-quick server-logs server-status server-ssh

# Server configuration
SERVER_USER := infra
SERVER_IP := 88.198.170.206
SERVER_PATH := /home/infra/savegress-platform

# Default target
help:
	@echo "Savegress Platform - Available commands:"
	@echo ""
	@echo "  Local Development (Docker):"
	@echo "    make dev          - Start all services locally"
	@echo "    make dev-build    - Build and start all services"
	@echo "    make dev-down     - Stop all services"
	@echo "    make dev-logs     - View logs from all services"
	@echo "    make dev-clean    - Stop and remove all containers, volumes"
	@echo ""
	@echo "  Local Development (Native):"
	@echo "    make install      - Install dependencies for frontend and backend"
	@echo "    make frontend     - Run frontend in dev mode (npm run dev)"
	@echo "    make backend      - Run backend in dev mode (npm run dev)"
	@echo ""
	@echo "  Database:"
	@echo "    make db-studio    - Open Prisma Studio"
	@echo "    make db-migrate   - Run database migrations"
	@echo "    make db-reset     - Reset database (WARNING: deletes data)"
	@echo ""
	@echo "  Deployment:"
	@echo "    make deploy       - Full deploy (sync + rebuild)"
	@echo "    make deploy-quick - Quick deploy (sync + restart only)"
	@echo "    make server-logs  - View server logs"
	@echo "    make server-status- Check server status"
	@echo "    make server-ssh   - SSH into server"

# ==================== Local Development (Docker) ====================

dev:
	cd docker && docker-compose -f docker-compose.local.yml up

dev-build:
	cd docker && docker-compose -f docker-compose.local.yml up --build

dev-down:
	cd docker && docker-compose -f docker-compose.local.yml down

dev-logs:
	cd docker && docker-compose -f docker-compose.local.yml logs -f

dev-clean:
	cd docker && docker-compose -f docker-compose.local.yml down -v --rmi local

# ==================== Local Development (Native) ====================

install:
	cd frontend && npm install
	cd backend && npm install

frontend:
	cd frontend && npm run dev

backend:
	cd backend && npm run dev

# ==================== Database ====================

db-studio:
	cd backend && npx prisma studio

db-migrate:
	cd backend && npx prisma migrate dev

db-reset:
	cd backend && npx prisma migrate reset

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
