.PHONY: help dev up down build clean test backend frontend db-up db-down logs

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment with hot reload
	docker-compose up --build

up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

build: ## Build all services
	docker-compose build

clean: ## Clean up containers and volumes
	docker-compose down -v --remove-orphans
	docker system prune -f

test: ## Run all tests
	cd backend && go test ./...
	cd frontend && npm test

backend: ## Start only backend services (db, redis, api)
	docker-compose up -d postgres redis backend

frontend: ## Start only frontend
	docker-compose up -d frontend

db-up: ## Start only database services
	docker-compose up -d postgres redis

db-down: ## Stop database services
	docker-compose stop postgres redis

logs: ## Show logs from all services
	docker-compose logs -f

logs-backend: ## Show backend logs
	docker-compose logs -f backend

logs-frontend: ## Show frontend logs
	docker-compose logs -f frontend

# Go specific commands
go-mod: ## Update Go dependencies
	cd backend && go mod tidy

go-test: ## Run Go tests
	cd backend && go test -v ./...

# Frontend specific commands
npm-install: ## Install npm dependencies
	cd frontend && npm install

npm-build: ## Build frontend
	cd frontend && npm run build