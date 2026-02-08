# ACME Console Docker Compose 部署文档

本文档说明如何使用 Docker Compose 快速部署 ACME Console。

## 目录

- [架构说明](#架构说明)
- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [完整部署示例](#完整部署示例)
- [数据持久化](#数据持久化)
- [日志管理](#日志管理)
- [更新和维护](#更新和维护)
- [故障排查](#故障排查)

---

## 架构说明

ACME Console 采用**前后端一体化部署**架构：

```
                    ┌─────────────────────────────────────┐
                    │         ACME Console (Go)           │
                    │              :10020                 │
用户 ──► Nginx ────►│                                     │
         :443       │  /api/*     → API 处理器            │
                    │  /health    → 健康检查              │
                    │  /*         → 静态文件服务 (前端)    │
                    │              └─► SPA fallback       │
                    └─────────────────────────────────────┘
                                      │
                                      ▼
                               MySQL Database
```

**关键点**：
- 后端 Go 服务同时提供 **API** 和 **前端静态文件**
- 使用 `-static` 参数指定前端构建产物目录
- 后端内置 SPA 路由支持，自动 fallback 到 `index.html`
- 反向代理只需代理到后端的单一端口 (10020)

**重要提示**：项目现有的 `deploy/Dockerfile` **不包含前端构建**，本文档提供了完整的生产 Dockerfile。

---

## 环境要求

- **Docker**: 20.10 或以上
- **Docker Compose**: v2.0 或以上
- **内存**: 2GB 或以上
- **磁盘**: 10GB 或以上

### 安装 Docker

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 安装 Docker Compose (已包含在 Docker Desktop 中)
# 如果是 Linux 服务器，Docker Compose v2 已内置于 Docker CLI
docker compose version
```

---

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/your-org/acme-console.git
cd acme-console
```

### 2. 准备配置

```bash
# 复制示例配置
cp configs/config.yaml configs/config.prod.yaml

# 生成加密密钥
openssl rand -hex 32
# 将生成的密钥填入 config.prod.yaml 的 encryption.master_key
```

### 3. 启动服务

```bash
# 使用 Makefile（推荐）
make docker-up

# 或直接使用 docker compose
cd deploy
docker compose up -d
```

### 4. 访问服务

- **Web 控制台**: http://localhost:10020
- **默认账户**: admin / admin123

**重要**: 首次登录后请立即修改管理员密码！

---

## 配置说明

### 项目自带的 docker-compose.yaml

**重要提示**：项目在 `deploy/docker-compose.yaml` 中提供的配置**不完整**：

```yaml
services:
  app:
    build:
      context: ..
      dockerfile: deploy/Dockerfile  # ⚠️ 此 Dockerfile 不包含前端构建
    ports:
      - "10020:10020"
    extra_hosts:
      - "host.docker.internal:host-gateway"
    volumes:
      - ../configs:/app/configs:ro
    restart: unless-stopped
```

**问题**：
1. `deploy/Dockerfile` 只构建了后端，没有构建前端
2. 启动命令没有 `-static` 参数，无法服务前端页面
3. 只能提供 API 服务，无法访问 Web 控制台

**解决方案**：使用本文档提供的完整 Dockerfile（见下文"完整部署示例"）。

### 环境变量配置

可以通过环境变量覆盖配置文件中的设置：

| 环境变量 | 说明 | 示例 |
|---------|------|------|
| `ACME_SERVER_HOST` | 服务监听地址 | `0.0.0.0` |
| `ACME_SERVER_PORT` | 服务监听端口 | `10020` |
| `ACME_DATABASE_HOST` | 数据库地址 | `mysql` |
| `ACME_DATABASE_PORT` | 数据库端口 | `3306` |
| `ACME_DATABASE_USER` | 数据库用户 | `acme` |
| `ACME_DATABASE_PASSWORD` | 数据库密码 | `secret` |
| `ACME_DATABASE_DBNAME` | 数据库名称 | `acme_console` |
| `ACME_JWT_SECRET` | JWT 密钥 | `your-secret` |
| `ACME_ENCRYPTION_MASTER_KEY` | 加密密钥 | `64位十六进制` |

---

## 完整部署示例

以下是包含 MySQL 的完整 Docker Compose 配置，适合一键部署。

### 1. 创建部署目录

```bash
mkdir -p ~/acme-console-deploy
cd ~/acme-console-deploy
```

### 2. 创建 docker-compose.yaml

```yaml
version: "3.8"

services:
  # MySQL 数据库
  mysql:
    image: mysql:8.0
    container_name: acme-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD:-rootpassword}
      MYSQL_DATABASE: acme_console
      MYSQL_USER: acme
      MYSQL_PASSWORD: ${MYSQL_PASSWORD:-acmepassword}
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - acme-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-p${MYSQL_ROOT_PASSWORD:-rootpassword}"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  # ACME Console 应用
  app:
    image: acme-console:latest
    build:
      context: .
      dockerfile: Dockerfile
    container_name: acme-console
    restart: unless-stopped
    ports:
      - "${APP_PORT:-10020}:10020"
    environment:
      - ACME_DATABASE_HOST=mysql
      - ACME_DATABASE_PORT=3306
      - ACME_DATABASE_USER=acme
      - ACME_DATABASE_PASSWORD=${MYSQL_PASSWORD:-acmepassword}
      - ACME_DATABASE_DBNAME=acme_console
      - ACME_JWT_SECRET=${JWT_SECRET:-change-this-jwt-secret-in-production}
      - ACME_ENCRYPTION_MASTER_KEY=${ENCRYPTION_KEY}
    volumes:
      - ./configs:/app/configs:ro
    networks:
      - acme-network
    depends_on:
      mysql:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:10020/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

volumes:
  mysql_data:
    driver: local

networks:
  acme-network:
    driver: bridge
```

### 3. 创建 Dockerfile

**重要**：项目现有的 `deploy/Dockerfile` 不包含前端构建，需要创建完整的生产 Dockerfile。

在部署目录创建 `Dockerfile`：

```dockerfile
# 构建阶段 - 后端
FROM golang:1.23-alpine AS backend-builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git
ENV GOPROXY=https://goproxy.cn,direct

# 复制 Go 模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/acme-console ./cmd/server

# 构建阶段 - 前端
FROM node:20-alpine AS frontend-builder

WORKDIR /build

# 复制前端文件
COPY web/package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com

COPY web/ ./
RUN npm run build

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 创建非 root 用户
RUN adduser -D -u 1000 acme

# 复制构建产物
COPY --from=backend-builder /app/bin/acme-console /app/bin/
COPY --from=frontend-builder /build/dist /app/static

# 创建配置目录
RUN mkdir -p /app/configs && chown -R acme:acme /app

USER acme

EXPOSE 10020

ENTRYPOINT ["/app/bin/acme-console"]
CMD ["-config", "/app/configs/config.yaml", "-static", "/app/static"]
```

**说明**：
- 多阶段构建：分别构建后端和前端
- 后端构建产物：`/app/bin/acme-console`
- 前端构建产物：`/app/static`（来自 `web/dist`）
- 启动命令包含 `-static /app/static` 参数，启用静态文件服务

### 4. 创建配置文件

```bash
mkdir -p configs
cat > configs/config.yaml << 'EOF'
server:
  host: "0.0.0.0"
  port: 10020

database:
  host: "mysql"
  port: 3306
  user: "acme"
  password: "acmepassword"
  dbname: "acme_console"

jwt:
  secret: "change-this-jwt-secret-in-production"
  expire_hours: 24

acme:
  dns:
    resolvers: "8.8.8.8:53,1.1.1.1:53"
    timeout: "10s"

encryption:
  master_key: ""
EOF
```

### 5. 创建环境变量文件

```bash
cat > .env << 'EOF'
# MySQL 配置
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_PASSWORD=your_acme_password

# 应用配置
APP_PORT=10020
JWT_SECRET=your-jwt-secret-at-least-32-characters

# 加密密钥（使用 openssl rand -hex 32 生成）
ENCRYPTION_KEY=your_64_character_hex_encryption_key
EOF
```

生成加密密钥：

```bash
# 生成并更新 .env 文件
ENCRYPTION_KEY=$(openssl rand -hex 32)
sed -i "s/ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" .env
echo "Generated encryption key: $ENCRYPTION_KEY"
```

### 6. 启动服务

```bash
# 构建并启动
docker compose up -d --build

# 查看日志
docker compose logs -f

# 检查服务状态
docker compose ps
```

---

## 使用预构建镜像

如果不想从源码构建，可以使用预构建的镜像。

### docker-compose.yaml（使用预构建镜像）

```yaml
version: "3.8"

services:
  mysql:
    image: mysql:8.0
    container_name: acme-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: acme_console
      MYSQL_USER: acme
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - acme-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  app:
    image: ghcr.io/your-org/acme-console:latest  # 替换为实际镜像地址
    container_name: acme-console
    restart: unless-stopped
    ports:
      - "10020:10020"
    environment:
      - ACME_DATABASE_HOST=mysql
      - ACME_DATABASE_PORT=3306
      - ACME_DATABASE_USER=acme
      - ACME_DATABASE_PASSWORD=${MYSQL_PASSWORD}
      - ACME_DATABASE_DBNAME=acme_console
      - ACME_JWT_SECRET=${JWT_SECRET}
      - ACME_ENCRYPTION_MASTER_KEY=${ENCRYPTION_KEY}
    networks:
      - acme-network
    depends_on:
      mysql:
        condition: service_healthy

volumes:
  mysql_data:

networks:
  acme-network:
```

---

## 数据持久化

### 数据卷说明

| 卷名 | 用途 | 重要性 |
|-----|------|-------|
| `mysql_data` | MySQL 数据库文件 | **关键** |

### 备份数据

```bash
# 备份 MySQL 数据
docker compose exec mysql mysqldump -u root -p acme_console > backup_$(date +%Y%m%d).sql

# 或使用 docker 卷备份
docker run --rm -v acme-console-deploy_mysql_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/mysql_data_backup.tar.gz /data
```

### 恢复数据

```bash
# 恢复 MySQL 数据
docker compose exec -T mysql mysql -u root -p acme_console < backup_20240101.sql

# 或恢复 docker 卷
docker run --rm -v acme-console-deploy_mysql_data:/data -v $(pwd):/backup alpine \
  tar xzf /backup/mysql_data_backup.tar.gz -C /
```

---

## 日志管理

### 查看日志

```bash
# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f app
docker compose logs -f mysql

# 查看最近 100 行日志
docker compose logs --tail=100 app
```

### 日志配置

在 `docker-compose.yaml` 中配置日志驱动：

```yaml
services:
  app:
    logging:
      driver: "json-file"
      options:
        max-size: "100m"
        max-file: "3"
```

### 导出日志

```bash
# 导出应用日志
docker compose logs app > app_logs_$(date +%Y%m%d).log
```

---

## 更新和维护

### 更新应用

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker compose up -d --build

# 或者如果使用预构建镜像
docker compose pull
docker compose up -d
```

### 清理资源

```bash
# 停止并删除容器
docker compose down

# 停止并删除容器和卷（会删除数据！）
docker compose down -v

# 清理未使用的镜像
docker image prune -f

# 清理所有未使用的资源
docker system prune -f
```

### 重启服务

```bash
# 重启所有服务
docker compose restart

# 重启特定服务
docker compose restart app
```

---

## 反向代理配置

**架构说明**：后端服务同时处理 API 和静态文件，因此反向代理只需将所有请求代理到后端的 10020 端口。后端会根据路径自动分发：
- `/api/*` → API 处理器
- `/health` → 健康检查
- 其他路径 → 静态文件服务（SPA 支持）

### 使用 Nginx

在 `docker-compose.yaml` 中添加 Nginx 服务：

```yaml
services:
  nginx:
    image: nginx:alpine
    container_name: acme-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    networks:
      - acme-network
    depends_on:
      - app
```

创建 `nginx.conf`：

```nginx
server {
    listen 80;
    server_name acme.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name acme.example.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    # 所有请求代理到后端（后端同时处理 API 和静态文件）
    location / {
        proxy_pass http://app:10020;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 使用 Traefik

```yaml
services:
  traefik:
    image: traefik:v2.10
    container_name: traefik
    restart: unless-stopped
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - acme-network

  app:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.acme.rule=Host(`acme.example.com`)"
      - "traefik.http.routers.acme.entrypoints=websecure"
      - "traefik.http.services.acme.loadbalancer.server.port=10020"
```

---

## 故障排查

### 1. 容器无法启动

```bash
# 查看容器状态
docker compose ps -a

# 查看容器日志
docker compose logs app

# 检查容器详情
docker inspect acme-console
```

### 2. 数据库连接失败

```bash
# 检查 MySQL 是否就绪
docker compose exec mysql mysqladmin ping -h localhost

# 检查网络连接
docker compose exec app ping mysql

# 查看 MySQL 日志
docker compose logs mysql
```

### 3. 健康检查失败

```bash
# 手动执行健康检查
docker compose exec app wget -q --spider http://localhost:10020/health

# 查看健康检查状态
docker inspect --format='{{json .State.Health}}' acme-console | jq
```

**问题**: 访问首页返回 404

**原因**: 容器没有包含前端静态文件或未启用 `-static` 参数

**解决**:
```bash
# 检查容器内是否有静态文件
docker compose exec app ls -la /app/static

# 检查启动命令
docker compose exec app ps aux | grep acme-console

# 应该看到: /app/bin/acme-console -config ... -static /app/static

# 如果没有，需要使用本文档提供的完整 Dockerfile 重新构建
docker compose down
docker compose up -d --build
```

### 4. 端口冲突

```bash
# 检查端口占用
sudo lsof -i :10020
sudo netstat -tlnp | grep 10020

# 修改端口映射
# 在 .env 中设置 APP_PORT=10021
```

### 5. 磁盘空间不足

```bash
# 检查 Docker 磁盘使用
docker system df

# 清理未使用的资源
docker system prune -a --volumes
```

### 6. 进入容器调试

```bash
# 进入应用容器
docker compose exec app sh

# 进入 MySQL 容器
docker compose exec mysql bash
docker compose exec mysql mysql -u root -p
```

---

## 生产环境检查清单

部署到生产环境前，请确认以下事项：

- [ ] 修改默认管理员密码
- [ ] 设置强密码的 MySQL root 和 acme 用户
- [ ] 生成并安全保存加密密钥 (`ENCRYPTION_KEY`)
- [ ] 设置强 JWT 密钥 (`JWT_SECRET`)
- [ ] 配置 HTTPS（通过反向代理）
- [ ] 配置防火墙规则
- [ ] 设置数据库自动备份
- [ ] 配置日志轮转
- [ ] 设置监控告警

---

## 常用命令速查

```bash
# 启动服务
docker compose up -d

# 停止服务
docker compose down

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose restart

# 重新构建
docker compose up -d --build

# 进入容器
docker compose exec app sh

# 备份数据库
docker compose exec mysql mysqldump -u root -p acme_console > backup.sql

# 恢复数据库
docker compose exec -T mysql mysql -u root -p acme_console < backup.sql
```

---

**文档版本**: 1.0
**最后更新**: 2026-02-08
