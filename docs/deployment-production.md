# ACME Console 生产环境部署文档

本文档详细说明如何在生产环境中部署 ACME Console。

## 目录

- [架构说明](#架构说明)
- [环境要求](#环境要求)
- [前置准备](#前置准备)
- [构建应用](#构建应用)
- [配置说明](#配置说明)
- [部署步骤](#部署步骤)
- [反向代理配置](#反向代理配置)
- [系统服务配置](#系统服务配置)
- [监控和日志](#监控和日志)
- [备份和恢复](#备份和恢复)
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
- Nginx 只需代理到后端的单一端口 (10020)

---

## 环境要求

### 硬件要求

- **CPU**: 2 核心或以上
- **内存**: 2GB 或以上
- **磁盘**: 20GB 或以上（根据证书数量调整）

### 软件要求

- **操作系统**: Linux (推荐 Ubuntu 22.04 LTS / CentOS 8+)
- **Go**: 1.23 或以上（仅构建时需要）
- **Node.js**: 20.x 或以上（仅构建前端时需要）
- **MySQL**: 8.0 或以上
- **反向代理**: Nginx 1.18+ 或 Caddy 2.x

---

## 前置准备

### 1. 安装 MySQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server

# CentOS/RHEL
sudo yum install mysql-server

# 启动 MySQL
sudo systemctl start mysql
sudo systemctl enable mysql

# 安全配置
sudo mysql_secure_installation
```

### 2. 创建数据库和用户

```bash
# 登录 MySQL
sudo mysql -u root -p

# 创建数据库
CREATE DATABASE acme_console CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 创建用户并授权
CREATE USER 'acme'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL PRIVILEGES ON acme_console.* TO 'acme'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 3. 生成加密密钥

ACME Console 使用 AES-256-GCM 加密存储敏感数据（如私钥），需要生成 32 字节的加密密钥：

```bash
# 生成加密密钥（64 个十六进制字符）
openssl rand -hex 32
```

**重要**: 请妥善保管此密钥，丢失后将无法解密已存储的私钥数据。

### 4. 准备部署目录

```bash
# 创建应用目录
sudo mkdir -p /opt/acme-console
sudo mkdir -p /opt/acme-console/configs
sudo mkdir -p /opt/acme-console/logs
sudo mkdir -p /var/lib/acme-console

# 创建运行用户
sudo useradd -r -s /bin/false acme

# 设置权限
sudo chown -R acme:acme /opt/acme-console
sudo chown -R acme:acme /var/lib/acme-console
```

---

## 构建应用

### 方式一：本地构建

```bash
# 克隆代码
git clone https://github.com/your-org/acme-console.git
cd acme-console

# 构建后端
make build

# 构建前端
make web-build

# 复制文件到部署目录
sudo cp bin/acme-console /opt/acme-console/
sudo cp -r web/dist /opt/acme-console/static
sudo chown -R acme:acme /opt/acme-console
```

### 方式二：使用 Docker 构建

**注意**：项目现有的 `deploy/Dockerfile` **不包含前端构建**，需要使用以下完整的 Dockerfile。

创建完整的生产 Dockerfile：

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
RUN apk add --no-cache ca-certificates tzdata

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

构建并导出：

```bash
# 构建镜像
docker build -f Dockerfile.prod -t acme-console:latest .

# 导出二进制文件和静态文件
docker create --name acme-console-temp acme-console:latest
docker cp acme-console-temp:/app/bin/acme-console ./bin/
docker cp acme-console-temp:/app/static ./static/
docker rm acme-console-temp

# 复制到部署目录
sudo cp bin/acme-console /opt/acme-console/
sudo cp -r static /opt/acme-console/
sudo chown -R acme:acme /opt/acme-console
```

---

## 配置说明

### 1. 创建配置文件

```bash
sudo nano /opt/acme-console/configs/config.yaml
```

### 2. 配置内容

```yaml
server:
  host: "127.0.0.1"  # 仅监听本地，通过反向代理访问
  port: 10020

database:
  host: "localhost"
  port: 3306
  user: "acme"
  password: "your_secure_password"
  dbname: "acme_console"
  # 连接池配置
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600  # 秒

jwt:
  secret: "your-jwt-secret-key-change-this"  # 至少 32 字符
  expire_hours: 24

acme:
  dns:
    resolvers: "8.8.8.8:53,1.1.1.1:53"  # DNS 验证服务器
    timeout: "10s"

encryption:
  master_key: "your-32-byte-hex-key-from-openssl-rand"  # 64 个十六进制字符
```

### 3. 使用环境变量（可选）

如果不想在配置文件中存储敏感信息，可以使用环境变量：

```bash
export ACME_DATABASE_PASSWORD="your_secure_password"
export ACME_JWT_SECRET="your-jwt-secret-key"
export ACME_ENCRYPTION_MASTER_KEY="your-32-byte-hex-key"
```

---

## 部署步骤

### 1. 测试运行

```bash
# 切换到 acme 用户
sudo -u acme /opt/acme-console/acme-console \
  -config /opt/acme-console/configs/config.yaml \
  -static /opt/acme-console/static

# 检查日志输出，确认启动成功
# 按 Ctrl+C 停止
```

### 2. 验证数据库初始化

应用首次启动时会自动：
- 创建所有数据表
- 初始化默认配置
- 创建默认管理员账户（用户名: `admin`, 密码: `admin123`）

**重要**: 首次登录后请立即修改管理员密码！

### 3. 健康检查

```bash
curl http://127.0.0.1:10020/health
# 应返回: {"status":"ok"}
```

---

## 反向代理配置

**架构说明**：后端服务同时处理 API 和静态文件，因此 Nginx 只需将所有请求代理到后端的 10020 端口。后端会根据路径自动分发：
- `/api/*` → API 处理器
- `/health` → 健康检查
- 其他路径 → 静态文件服务（SPA 支持）

### Nginx 配置

创建 Nginx 配置文件：

```bash
sudo nano /etc/nginx/sites-available/acme-console
```

配置内容：

```nginx
server {
    listen 80;
    server_name acme.example.com;  # 修改为你的域名

    # 重定向到 HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name acme.example.com;  # 修改为你的域名

    # SSL 证书配置
    ssl_certificate /etc/ssl/certs/acme-console.crt;
    ssl_certificate_key /etc/ssl/private/acme-console.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 日志
    access_log /var/log/nginx/acme-console-access.log;
    error_log /var/log/nginx/acme-console-error.log;

    # 客户端上传大小限制（证书文件）
    client_max_body_size 10M;

    # 代理到后端（后端同时处理 API 和静态文件）
    location / {
        proxy_pass http://127.0.0.1:10020;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持（如果需要）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

**性能优化（可选）**：如果需要更好的静态文件性能，可以让 Nginx 直接服务静态文件：

```nginx
server {
    listen 443 ssl http2;
    server_name acme.example.com;

    # SSL 配置...

    # 静态文件由 Nginx 直接服务
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        root /opt/acme-console/static;
        expires 1y;
        add_header Cache-Control "public, immutable";
        try_files $uri =404;
    }

    # API 和其他请求代理到后端
    location / {
        proxy_pass http://127.0.0.1:10020;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

启用配置：

```bash
# 创建软链接
sudo ln -s /etc/nginx/sites-available/acme-console /etc/nginx/sites-enabled/

# 测试配置
sudo nginx -t

# 重载 Nginx
sudo systemctl reload nginx
```

### Caddy 配置

创建 Caddyfile：

```bash
sudo nano /etc/caddy/Caddyfile
```

配置内容：

```caddy
acme.example.com {
    # 自动 HTTPS
    reverse_proxy 127.0.0.1:10020

    # 日志
    log {
        output file /var/log/caddy/acme-console.log
    }
}
```

重载 Caddy：

```bash
sudo systemctl reload caddy
```

---

## 系统服务配置

### 创建 systemd 服务

```bash
sudo nano /etc/systemd/system/acme-console.service
```

服务配置：

```ini
[Unit]
Description=ACME Console Service
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=acme
Group=acme
WorkingDirectory=/opt/acme-console

# 环境变量（可选）
Environment="ACME_DATABASE_PASSWORD=your_secure_password"
Environment="ACME_JWT_SECRET=your-jwt-secret-key"
Environment="ACME_ENCRYPTION_MASTER_KEY=your-32-byte-hex-key"

# 启动命令
ExecStart=/opt/acme-console/acme-console \
    -config /opt/acme-console/configs/config.yaml \
    -static /opt/acme-console/static

# 重启策略
Restart=always
RestartSec=10

# 资源限制
LimitNOFILE=65536

# 日志
StandardOutput=journal
StandardError=journal
SyslogIdentifier=acme-console

[Install]
WantedBy=multi-user.target
```

### 启动服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start acme-console

# 设置开机自启
sudo systemctl enable acme-console

# 查看状态
sudo systemctl status acme-console
```

### 服务管理命令

```bash
# 启动服务
sudo systemctl start acme-console

# 停止服务
sudo systemctl stop acme-console

# 重启服务
sudo systemctl restart acme-console

# 查看日志
sudo journalctl -u acme-console -f

# 查看最近 100 行日志
sudo journalctl -u acme-console -n 100
```

---

## 监控和日志

### 1. 应用日志

应用日志通过 systemd journal 管理：

```bash
# 实时查看日志
sudo journalctl -u acme-console -f

# 查看今天的日志
sudo journalctl -u acme-console --since today

# 查看最近 1 小时的日志
sudo journalctl -u acme-console --since "1 hour ago"

# 导出日志到文件
sudo journalctl -u acme-console > /tmp/acme-console.log
```

### 2. Nginx 日志

```bash
# 访问日志
tail -f /var/log/nginx/acme-console-access.log

# 错误日志
tail -f /var/log/nginx/acme-console-error.log
```

### 3. 健康检查

设置定时健康检查：

```bash
# 创建健康检查脚本
sudo nano /opt/acme-console/healthcheck.sh
```

脚本内容：

```bash
#!/bin/bash

HEALTH_URL="http://127.0.0.1:10020/health"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_URL)

if [ "$RESPONSE" != "200" ]; then
    echo "$(date): Health check failed with status $RESPONSE" >> /var/log/acme-console-health.log
    # 可选：发送告警通知
    # curl -X POST "https://your-webhook-url" -d "ACME Console health check failed"
fi
```

添加到 crontab：

```bash
# 每 5 分钟检查一次
sudo crontab -e
*/5 * * * * /opt/acme-console/healthcheck.sh
```

### 4. 监控指标

推荐使用以下工具监控：

- **Prometheus + Grafana**: 应用指标监控
- **Netdata**: 系统资源监控
- **Uptime Kuma**: 服务可用性监控

---

## 备份和恢复

### 1. 数据库备份

#### 自动备份脚本

```bash
sudo nano /opt/acme-console/backup.sh
```

脚本内容：

```bash
#!/bin/bash

BACKUP_DIR="/var/backups/acme-console"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="acme_console"
DB_USER="acme"
DB_PASS="your_secure_password"

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份数据库
mysqldump -u $DB_USER -p$DB_PASS $DB_NAME | gzip > $BACKUP_DIR/acme_console_$DATE.sql.gz

# 删除 30 天前的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "$(date): Database backup completed: acme_console_$DATE.sql.gz"
```

设置权限并添加到 crontab：

```bash
sudo chmod +x /opt/acme-console/backup.sh

# 每天凌晨 2 点备份
sudo crontab -e
0 2 * * * /opt/acme-console/backup.sh >> /var/log/acme-console-backup.log 2>&1
```

#### 手动备份

```bash
# 备份数据库
mysqldump -u acme -p acme_console | gzip > acme_console_backup.sql.gz

# 备份配置文件
tar -czf acme_console_config_backup.tar.gz /opt/acme-console/configs
```

### 2. 数据恢复

```bash
# 恢复数据库
gunzip < acme_console_backup.sql.gz | mysql -u acme -p acme_console

# 恢复配置文件
tar -xzf acme_console_config_backup.tar.gz -C /
```

### 3. 加密密钥备份

**重要**: 加密密钥 (`encryption.master_key`) 必须妥善备份，建议：

- 存储在密码管理器中（如 1Password、Bitwarden）
- 打印并存放在安全的物理位置
- 使用多个备份位置

---

## 故障排查

### 1. 服务无法启动

#### 检查日志

```bash
sudo journalctl -u acme-console -n 50
```

#### 常见问题

**问题**: `dial tcp: connect: connection refused`

**原因**: 无法连接到 MySQL

**解决**:
```bash
# 检查 MySQL 是否运行
sudo systemctl status mysql

# 检查配置文件中的数据库连接信息
cat /opt/acme-console/configs/config.yaml

# 测试数据库连接
mysql -u acme -p -h localhost acme_console
```

**问题**: `permission denied`

**原因**: 文件权限不正确

**解决**:
```bash
sudo chown -R acme:acme /opt/acme-console
sudo chmod +x /opt/acme-console/acme-console
```

### 2. 无法访问 Web 界面

#### 检查后端是否启用了静态文件服务

```bash
# 检查启动命令是否包含 -static 参数
sudo systemctl status acme-console

# 应该看到类似这样的命令：
# ExecStart=/opt/acme-console/acme-console -config ... -static /opt/acme-console/static
```

**问题**: 访问首页返回 404

**原因**: 后端没有启用静态文件服务

**解决**:
```bash
# 确保启动命令包含 -static 参数
sudo nano /etc/systemd/system/acme-console.service

# ExecStart 应该是：
# ExecStart=/opt/acme-console/acme-console \
#     -config /opt/acme-console/configs/config.yaml \
#     -static /opt/acme-console/static

# 重启服务
sudo systemctl daemon-reload
sudo systemctl restart acme-console
```

#### 检查反向代理

```bash
# Nginx
sudo nginx -t
sudo systemctl status nginx

# Caddy
sudo systemctl status caddy
```

#### 检查防火墙

```bash
# Ubuntu/Debian (ufw)
sudo ufw status
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# CentOS/RHEL (firewalld)
sudo firewall-cmd --list-all
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 3. 证书签发失败

#### 检查 DNS 解析

```bash
# 检查 DNS 记录
dig _acme-challenge.example.com TXT

# 使用指定的 DNS 服务器
dig @8.8.8.8 _acme-challenge.example.com TXT
```

#### 检查 ACME 日志

```bash
sudo journalctl -u acme-console | grep -i "acme\|challenge\|verify"
```

### 4. 数据库连接池耗尽

#### 检查连接数

```bash
mysql -u root -p -e "SHOW PROCESSLIST;"
```

#### 调整连接池配置

编辑 `config.yaml`:

```yaml
database:
  max_idle_conns: 20
  max_open_conns: 200
  conn_max_lifetime: 3600
```

### 5. 内存占用过高

#### 检查内存使用

```bash
# 查看进程内存
ps aux | grep acme-console

# 查看系统内存
free -h
```

#### 限制内存使用

编辑 systemd 服务文件：

```ini
[Service]
MemoryLimit=1G
```

### 6. 获取详细日志

如果需要更详细的调试信息，可以临时启用调试模式：

```bash
# 停止服务
sudo systemctl stop acme-console

# 手动运行并查看详细日志
sudo -u acme /opt/acme-console/acme-console \
  -config /opt/acme-console/configs/config.yaml \
  -static /opt/acme-console/static
```

---

## 安全建议

### 1. 定期更新

```bash
# 更新系统
sudo apt update && sudo apt upgrade  # Ubuntu/Debian
sudo yum update                       # CentOS/RHEL

# 更新应用
cd acme-console
git pull
make build
sudo systemctl restart acme-console
```

### 2. 防火墙配置

只开放必要的端口：

```bash
# 仅允许 80 和 443 端口
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable
```

### 3. 定期审计

- 定期检查用户账户和权限
- 审查证书签发日志
- 监控异常登录行为

### 4. 密钥轮换

定期更换 JWT 密钥和加密密钥（需要重新加密数据）。

---

## 性能优化

### 1. 数据库优化

```sql
-- 添加索引（如果需要）
CREATE INDEX idx_certificates_expires_at ON certificates(expires_at);
CREATE INDEX idx_challenges_status ON challenges(status);
```

### 2. 连接池调优

根据实际负载调整 `config.yaml` 中的连接池参数。

### 3. 反向代理缓存

对于静态资源，可以在 Nginx 中启用缓存：

```nginx
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

---

## 升级指南

### 1. 备份数据

```bash
# 备份数据库
/opt/acme-console/backup.sh

# 备份配置
cp -r /opt/acme-console/configs /tmp/acme-console-configs-backup
```

### 2. 停止服务

```bash
sudo systemctl stop acme-console
```

### 3. 更新应用

```bash
# 构建新版本
cd acme-console
git pull
make build
make web-build

# 替换二进制文件
sudo cp bin/acme-console /opt/acme-console/
sudo cp -r web/dist /opt/acme-console/static
sudo chown -R acme:acme /opt/acme-console
```

### 4. 启动服务

```bash
sudo systemctl start acme-console
sudo systemctl status acme-console
```

### 5. 验证升级

```bash
# 检查健康状态
curl http://127.0.0.1:10020/health

# 检查日志
sudo journalctl -u acme-console -f
```

---

## 支持和反馈

如有问题，请：

1. 查看日志: `sudo journalctl -u acme-console -n 100`
2. 检查 GitHub Issues: https://github.com/your-org/acme-console/issues
3. 提交新 Issue 并附上日志和配置信息

---

**文档版本**: 1.0
**最后更新**: 2026-02-08

