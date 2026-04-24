.PHONY: build run test clean tidy fmt lint \
        web-install web-dev web-build run-with-web \
        dev-docker-up dev-docker-up-d dev-docker-down dev-docker-logs dev-docker-logs-backend dev-docker-logs-frontend dev-docker-rebuild \
        prod-up prod-down prod-logs prod-logs-backend prod-logs-frontend prod-restart \
        prod-update-frontend prod-update-backend prod-ps

# Build variables
BINARY_NAME=acme-console
BUILD_DIR=bin
WEB_DIR=web
DEPLOY_DIR=deploy

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build the application
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run the application (API only)
run:
	$(GOCMD) run ./cmd/server

# Run with frontend
run-with-web: web-build
	$(GOCMD) run ./cmd/server -static $(WEB_DIR)/dist

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(WEB_DIR)/dist
	rm -rf $(WEB_DIR)/node_modules
	rm -rf tmp
	rm -f coverage.out coverage.html

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Format code
fmt:
	$(GOFMT) ./...

# Download dependencies
deps:
	$(GOMOD) download

# All: format, tidy, test, build
all: fmt tidy test build

# =============================================================================
# Development with Docker + Air (Hot Reload)
# =============================================================================

# Start development environment (backend with air + frontend with vite)
dev-docker-up:
	docker compose -f $(DEPLOY_DIR)/dev/docker-compose.yaml up --build

# Start in detached mode
dev-docker-up-d:
	docker compose -f $(DEPLOY_DIR)/dev/docker-compose.yaml up --build -d

# Stop development environment
dev-docker-down:
	docker compose -f $(DEPLOY_DIR)/dev/docker-compose.yaml down

# View logs
dev-docker-logs:
	docker compose -f $(DEPLOY_DIR)/dev/docker-compose.yaml logs -f

# View backend logs only
dev-docker-logs-backend:
	docker compose -f $(DEPLOY_DIR)/dev/docker-compose.yaml logs -f backend

# View frontend logs only
dev-docker-logs-frontend:
	docker compose -f $(DEPLOY_DIR)/dev/docker-compose.yaml logs -f frontend

# Rebuild and restart
dev-docker-rebuild:
	docker compose -f $(DEPLOY_DIR)/dev/docker-compose.yaml up --build --force-recreate

# =============================================================================
# Production Docker (Nginx + Go API + MySQL)
# =============================================================================

# 启动生产环境
prod-up:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml up -d --build

# 停止生产环境
prod-down:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml down

# 查看生产环境日志
prod-logs:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml logs -f

# 查看后端日志
prod-logs-backend:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml logs -f backend

# 查看前端日志
prod-logs-frontend:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml logs -f frontend

# 重启生产环境
prod-restart:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml restart

# 只更新前端
prod-update-frontend:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml up -d --build frontend

# 只更新后端
prod-update-backend:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml up -d --build backend

# 查看生产环境状态
prod-ps:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml ps

# =============================================================================
# Frontend commands
# =============================================================================

web-install:
	cd $(WEB_DIR) && npm install

web-dev:
	cd $(WEB_DIR) && npm run dev

web-build:
	cd $(WEB_DIR) && npm install && npm run build

# =============================================================================
# Full build (backend + frontend)
# =============================================================================

build-all: web-build build
	@echo "Build complete. Run with: ./$(BUILD_DIR)/$(BINARY_NAME) -static $(WEB_DIR)/dist"
