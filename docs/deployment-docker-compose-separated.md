# ACME Console 生产环境 Docker Compose 部署（前后端分离）

本文档说明如何使用 Docker Compose 部署前后端分离的 ACME Console。

## 架构说明

```
                    ┌─────────────────────────────────────┐
                    │      Nginx (前端容器)               │
用户 ──► Nginx ────►│         :80                         │
         :443       │  /api/*  → 代理到后端容器            │
                    │  /*      → 静态文件服务              │
                    │           └─► SPA fallback          │
                    └─────────────────────────────────────┘
                                      │
                                      ▼
                    ┌─────────────────────────────────────┐
                    │      Go Backend (后端容器)          │
                    │         :10020                      │
                    │  /api/*     → API 处理器            │
                    │  /health    → 健康检查              │
                    └─────────────────────────────────────┘
                                      │
                                      ▼
                               MySQL Database
```

**优势**：
- 前后端完全分离，独立扩展
- 前端由 Nginx 直接服务，性能更好
- 后端专注于 API 处理
- 可以独立更新前端或后端

---

## 目录结构

```
acme-console-deploy/
├── docker-compose.yaml
├── backend/
│   └── Dockerfile
├── frontend/
│   ├── Dockerfile
│   └── nginx.conf
├── configs/
│   └── config.yaml
└── .env
```

---

## 快速开始

### 1. 创建部署目录

```bash
mkdir -p ~/acme-console-deploy
cd ~/acme-console-deploy
mkdir -p backend frontend configs
```

### 2. 克隆源码（用于构建）

```bash
git clone https://github.com/your-org/acme-console.git source
```

### 3. 创建配置文件

#### docker-compose.yaml

```yaml
version: "3.8"

services:
  # MySQL 数据库
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

  # 后端 API 服务
  backend:
    build:
      context: ./source
      dockerfile: ../backend/Dockerfile
    container_name: acme-backend
    restart: unless-stopped
    environment:
      - ACME_SERVER_HOST=0.0.0.0
      - ACME_SERVER_PORT=10020
      - ACME_DATABASE_HOST=mysql
      - ACME_DATABASE_PORT=3306
      - ACME_DATABASE_USER=acme
      - ACME_DATABASE_PASSWORD=${MYSQL_PASSWORD}
      - ACME_DATABASE_DBNAME=acme_console
      - ACME_JWT_SECRET=${JWT_SECRET}
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

  # 前端 Web 服务
  frontend:
    build:
      context: ./source
      dockerfile: ../frontend/Dockerfile
      args:
        - VITE_API_URL=http://backend:10020
    container_name: acme-frontend
    restart: unless-stopped
    ports:
      - "${FRONTEND_PORT:-80}:80"
    networks:
      - acme-network
    depends_on:
      backend:
        condition: service_healthy

volumes:
  mysql_data:
    driver: local

networks:
  acme-network:
    driver: bridge
```

#### backend/Dockerfile

```dockerfile
# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache git
ENV GOPROXY=https://goproxy.cn,direct

# 复制 Go 模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/acme-console ./cmd/server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 创建非 root 用户
RUN adduser -D -u 1000 acme

# 复制二进制文件
COPY --from=builder /app/acme-console /app/

# 创建配置目录
RUN mkdir -p /app/configs && chown -R acme:acme /app

USER acme

EXPOSE 10020

# 注意：不使用 -static 参数，只提供 API 服务
ENTRYPOINT ["/app/acme-console"]
CMD ["-config", "/app/configs/config.yaml", "-dev=false"]
```

#### frontend/Dockerfile

```dockerfile
# 构建阶段
FROM node:20-alpine AS builder

WORKDIR /build

# 复制前端文件
COPY web/package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com

# 复制源码并构建
COPY web/ ./

# 构建参数（API 地址）
ARG VITE_API_URL
ENV VITE_API_URL=${VITE_API_URL}

RUN npm run build

# 运行阶段
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /build/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

#### frontend/nginx.conf

```nginx
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # 日志
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;

    # API 代理到后端
    location /api/ {
        proxy_pass http://backend:10020;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 健康检查代理到后端
    location /health {
        proxy_pass http://backend:10020;
        proxy_set_header Host $host;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        try_files $uri =404;
    }

    # SPA 路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

#### configs/config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 10020

database:
  host: "mysql"
  port: 3306
  user: "acme"
  password: "acmepassword"  # 将被环境变量覆盖
  dbname: "acme_console"

jwt:
  secret: "change-this-jwt-secret-in-production"  # 将被环境变量覆盖
  expire_hours: 24

acme:
  dns:
    resolvers: "8.8.8.8:53,1.1.1.1:53"
    timeout: "10s"

encryption:
  master_key: ""  # 将被环境变量覆盖
```

#### .env

```bash
# MySQL 配置
MYSQL_ROOT_PASSWORD=your_root_password_change_this
MYSQL_PASSWORD=your_acme_password_change_this

# 应用配置
FRONTEND_PORT=80
JWT_SECRET=your-jwt-secret-at-least-32-characters-change-this

# 加密密钥（使用 openssl rand -hex 32 生成）
ENCRYPTION_KEY=
```

### 4. 生成加密密钥

```bash
# 生成加密密钥并更新 .env
ENCRYPTION_KEY=$(openssl rand -hex 32)
echo "ENCRYPTION_KEY=$ENCRYPTION_KEY" >> .env
echo "Generated encryption key: $ENCRYPTION_KEY"
```

### 5. 启动服务

```bash
# 构建并启动所有服务
docker compose up -d --build

# 查看日志
docker compose logs -f

# 检查服务状态
docker compose ps
```

### 6. 访问服务

- **Web 控制台**: http://localhost
- **后端 API**: http://localhost/api/
- **健康检查**: http://localhost/health
- **默认账户**: admin / admin123

**重要**: 首次登录后请立即修改管理员密码！

---

## 使用外部反向代理（推荐）

如果你有外部的 Nginx 或 Traefik，可以只暴露前端容器的端口。

### docker-compose.yaml（不暴露端口）

```yaml
services:
  frontend:
    # ... 其他配置
    # 不暴露端口，由外部反向代理访问
    expose:
      - "80"
    networks:
      - acme-network
      - proxy-network  # 外部反向代理网络

networks:
  acme-network:
    driver: bridge
  proxy-network:
    external: true  # 外部网络
```

### 外部 Nginx 配置

```nginx
upstream acme_frontend {
    server acme-frontend:80;
}

server {
    listen 80;
    server_name acme.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name acme.example.com;

    ssl_certificate /etc/ssl/certs/acme-console.crt;
    ssl_certificate_key /etc/ssl/private/acme-console.key;

    location / {
        proxy_pass http://acme_frontend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 独立更新前端或后端

### 只更新前端

```bash
# 拉取最新代码
cd source && git pull && cd ..

# 只重建前端
docker compose up -d --build frontend

# 查看前端日志
docker compose logs -f frontend
```

### 只更新后端

```bash
# 拉取最新代码
cd source && git pull && cd ..

# 只重建后端
docker compose up -d --build backend

# 查看后端日志
docker compose logs -f backend
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
docker compose exec mysql mysqldump -u root -p${MYSQL_ROOT_PASSWORD} acme_console > backup_$(date +%Y%m%d).sql

# 备份配置文件
tar czf configs_backup_$(date +%Y%m%d).tar.gz configs/

# 备份环境变量（注意安全）
cp .env .env.backup
```

### 恢复数据

```bash
# 恢复 MySQL 数据
docker compose exec -T mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} acme_console < backup_20240101.sql
```

---

## 监控和日志

### 查看日志

```bash
# 查看所有服务日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f frontend
docker compose logs -f backend
docker compose logs -f mysql

# 查看最近 100 行日志
docker compose logs --tail=100 backend
```

### 健康检查

```bash
# 检查所有服务状态
docker compose ps

# 检查后端健康
curl http://localhost/health

# 检查前端
curl -I http://localhost/
```

---

## 性能优化

### 1. 前端静态资源 CDN

如果使用 CDN，可以修改 `frontend/nginx.conf`：

```nginx
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
    # 添加 CORS 头
    add_header Access-Control-Allow-Origin *;
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

### 2. 后端水平扩展

```yaml
services:
  backend:
    # ... 其他配置
    deploy:
      replicas: 3  # 启动 3 个后端实例
```

然后在前端 Nginx 配置中添加负载均衡：

```nginx
upstream backend_cluster {
    server backend:10020;
    # Docker Compose 会自动负载均衡到多个副本
}

location /api/ {
    proxy_pass http://backend_cluster;
}
```

### 3. 数据库连接池优化

编辑 `configs/config.yaml`：

```yaml
database:
  max_idle_conns: 20
  max_open_conns: 200
  conn_max_lifetime: 3600
```

---

## 故障排查

### 1. 前端无法访问后端 API

**问题**: 浏览器控制台显示 CORS 错误或 API 请求失败

**原因**: 前端容器无法连接到后端容器

**解决**:
```bash
# 检查网络连接
docker compose exec frontend ping backend

# 检查后端是否健康
docker compose exec frontend wget -q --spider http://backend:10020/health

# 查看后端日志
docker compose logs backend
```

### 2. 前端页面空白

**问题**: 访问首页显示空白或 404

**原因**: 前端构建失败或 Nginx 配置错误

**解决**:
```bash
# 检查前端容器内的文件
docker compose exec frontend ls -la /usr/share/nginx/html

# 应该看到 index.html 和 assets/ 目录

# 检查 Nginx 配置
docker compose exec frontend nginx -t

# 查看前端日志
docker compose logs frontend
```

### 3. API 请求返回 502

**问题**: 前端可以访问，但 API 请求返回 502 Bad Gateway

**原因**: 后端服务未启动或崩溃

**解决**:
```bash
# 检查后端容器状态
docker compose ps backend

# 查看后端日志
docker compose logs backend

# 重启后端
docker compose restart backend
```

### 4. 数据库连接失败

**问题**: 后端日志显示无法连接到 MySQL

**原因**: MySQL 未就绪或密码错误

**解决**:
```bash
# 检查 MySQL 状态
docker compose ps mysql

# 测试数据库连接
docker compose exec backend wget -q --spider http://mysql:3306

# 查看 MySQL 日志
docker compose logs mysql

# 手动连接测试
docker compose exec mysql mysql -u acme -p${MYSQL_PASSWORD} acme_console
```

---

## 生产环境检查清单

部署到生产环境前，请确认：

- [ ] 修改 `.env` 中的所有默认密码
- [ ] 生成并安全保存加密密钥
- [ ] 配置外部 HTTPS 反向代理
- [ ] 设置防火墙规则（只开放 80/443）
- [ ] 配置数据库自动备份
- [ ] 设置日志轮转
- [ ] 配置监控告警
- [ ] 测试前后端独立更新流程
- [ ] 验证健康检查正常工作
- [ ] 首次登录后修改管理员密码

---

## 常用命令速查

```bash
# 启动所有服务
docker compose up -d

# 启动并重建
docker compose up -d --build

# 停止所有服务
docker compose down

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f

# 只重建前端
docker compose up -d --build frontend

# 只重建后端
docker compose up -d --build backend

# 重启服务
docker compose restart frontend
docker compose restart backend

# 进入容器
docker compose exec frontend sh
docker compose exec backend sh

# 备份数据库
docker compose exec mysql mysqldump -u root -p acme_console > backup.sql

# 清理资源
docker compose down -v  # 警告：会删除数据！
```

---

**文档版本**: 2.0 (前后端分离版本)
**最后更新**: 2026-02-08
