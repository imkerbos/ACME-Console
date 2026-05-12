.PHONY: build run test clean tidy fmt deps all build-all \
        web-install web-dev web-build run-with-web \
        dev-docker-up dev-docker-up-d dev-docker-down dev-docker-logs dev-docker-logs-backend dev-docker-logs-frontend dev-docker-rebuild \
        prod-config prod-docker-up prod-docker-up-d prod-docker-down prod-docker-logs prod-docker-logs-backend prod-docker-logs-frontend prod-docker-rebuild

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
# Production with Docker (Nginx + Go API + MySQL)
# =============================================================================

# 从 .env 渲染配置 + 自动生成自签名证书
prod-config:
	@if [ ! -f $(DEPLOY_DIR)/.env ]; then \
		echo "错误: 缺少 deploy/.env，请先配置:"; \
		echo "  cp deploy/.env.example deploy/.env && vim deploy/.env"; \
		exit 1; \
	fi
	@mkdir -p $(DEPLOY_DIR)/configs $(DEPLOY_DIR)/certs
	@set -a && . $(DEPLOY_DIR)/.env && set +a && \
		if [ -z "$$JWT_SECRET" ]; then JWT_SECRET=$$(openssl rand -hex 32); \
			sed -i'' -e "s|^JWT_SECRET=.*|JWT_SECRET=$$JWT_SECRET|" $(DEPLOY_DIR)/.env; \
			echo "✓ 已自动生成 JWT_SECRET"; \
		fi && \
		if [ -z "$$ENCRYPTION_MASTER_KEY" ]; then ENCRYPTION_MASTER_KEY=$$(openssl rand -hex 32); \
			sed -i'' -e "s|^ENCRYPTION_MASTER_KEY=.*|ENCRYPTION_MASTER_KEY=$$ENCRYPTION_MASTER_KEY|" $(DEPLOY_DIR)/.env; \
			echo "✓ 已自动生成 ENCRYPTION_MASTER_KEY"; \
		fi && \
		sed -e "s|\$${MYSQL_USER}|$$MYSQL_USER|g" \
		    -e "s|\$${MYSQL_PASSWORD}|$$MYSQL_PASSWORD|g" \
		    -e "s|\$${MYSQL_DATABASE}|$$MYSQL_DATABASE|g" \
		    -e "s|\$${JWT_SECRET}|$$JWT_SECRET|g" \
		    -e "s|\$${ENCRYPTION_MASTER_KEY}|$$ENCRYPTION_MASTER_KEY|g" \
		    $(DEPLOY_DIR)/configs/config.yaml.tpl > $(DEPLOY_DIR)/configs/config.yaml
	@echo "✓ deploy/configs/config.yaml"
	@set -a && . $(DEPLOY_DIR)/.env && set +a && \
		NGINX_SERVER_NAME=$${SERVER_NAME:-_} && \
		if [ "$$REDIRECT_HTTP_TO_HTTPS" = "true" ]; then \
			REDIRECT_LINE="return 301 https://\$$host\$$request_uri;"; \
		else \
			REDIRECT_LINE=""; \
		fi && \
		sed -e "s|{{SERVER_NAME}}|$$NGINX_SERVER_NAME|g" \
		    -e "s|{{REDIRECT_BLOCK}}|$$REDIRECT_LINE|g" \
		    $(DEPLOY_DIR)/nginx.conf.tpl > $(DEPLOY_DIR)/nginx.conf
	@echo "✓ deploy/nginx.conf"
	@if [ ! -f $(DEPLOY_DIR)/certs/server.crt ] || [ ! -f $(DEPLOY_DIR)/certs/server.key ]; then \
		echo "生成自签名证书..."; \
		set -a && . $(DEPLOY_DIR)/.env && set +a && \
		SAN_NAME=$${SERVER_NAME:-localhost} && \
		openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
			-keyout $(DEPLOY_DIR)/certs/server.key \
			-out $(DEPLOY_DIR)/certs/server.crt \
			-subj "/CN=$$SAN_NAME" \
			-addext "subjectAltName=DNS:$$SAN_NAME,IP:127.0.0.1" 2>/dev/null; \
		echo "✓ deploy/certs/server.crt (self-signed)"; \
	else \
		echo "✓ deploy/certs/ (已有证书，跳过生成)"; \
	fi

# Start production environment (foreground)
prod-docker-up: prod-config
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml up --build

# Start in detached mode
prod-docker-up-d: prod-config
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml up --build -d

# Stop production environment
prod-docker-down:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml down

# View logs
prod-docker-logs:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml logs -f

# View backend logs only
prod-docker-logs-backend:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml logs -f backend

# View frontend logs only
prod-docker-logs-frontend:
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml logs -f frontend

# Rebuild and restart
prod-docker-rebuild: prod-config
	docker compose -f $(DEPLOY_DIR)/docker-compose.yaml up --build --force-recreate

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
