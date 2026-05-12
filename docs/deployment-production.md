# 部署文档

## 开发环境

Docker Compose + Air 热更新，代码修改后自动重新编译。

```bash
# 前台运行（看日志）
make dev-docker-up

# 后台运行
make dev-docker-up-d
```

- 前端 (Vite): http://localhost:3200
- 后端 (API): http://localhost:10020

```bash
make dev-docker-down      # 停止
make dev-docker-rebuild   # 重建
make dev-docker-logs      # 查看日志
```

---

## 生产环境

### 1. 配置 .env

```bash
cp deploy/.env.example deploy/.env
vim deploy/.env
```

所有配置集中在这一个文件：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MYSQL_ROOT_PASSWORD` | MySQL root 密码 | - |
| `MYSQL_DATABASE` | 数据库名 | `acme_console` |
| `MYSQL_USER` | 数据库用户 | `acme` |
| `MYSQL_PASSWORD` | 数据库密码 | - |
| `MYSQL_PORT` | MySQL 端口 | `3306` |
| `API_PORT` | 后端 API 宿主机映射端口 | `10020` |
| `HTTP_PORT` | HTTP 端口 | `80` |
| `HTTPS_PORT` | HTTPS 端口 | `443` |
| `JWT_SECRET` | JWT 密钥 | `openssl rand -hex 32` |
| `ENCRYPTION_MASTER_KEY` | 加密密钥 | `openssl rand -hex 32` |
| `SERVER_NAME` | 域名（留空则用 IP 访问） | 空 |
| `REDIRECT_HTTP_TO_HTTPS` | HTTP 跳转 HTTPS | `true` |

### 2. 启动

```bash
# 前台
make prod-docker-up

# 后台
make prod-docker-up-d
```

`prod-docker-up` 自动完成：
1. 从 `.env` 渲染 `deploy/configs/config.yaml`
2. 从 `.env` 渲染 `deploy/nginx.conf`（含域名和跳转配置）
3. 如果 `deploy/certs/` 下没有证书，自动生成自签名证书
4. 构建镜像并启动所有服务

### 3. HTTPS 证书

**默认**：自动生成自签名证书，用 IP + 443 端口访问，浏览器会提示不安全，忽略即可。

**自定义证书**：将你的证书放到 `deploy/certs/` 目录下：
```
deploy/certs/server.crt   # 证书
deploy/certs/server.key   # 私钥
```
同时在 `.env` 里填上域名：
```
SERVER_NAME=console.example.com
```

再次 `make prod-docker-up-d` 即可，已有证书不会被覆盖。

### 4. 日常运维

```bash
make prod-docker-logs            # 查看日志
make prod-docker-logs-backend    # 只看后端
make prod-docker-logs-frontend   # 只看前端
make prod-docker-down            # 停止
make prod-docker-rebuild         # 重建
```

### 5. 更新版本

```bash
git pull
make prod-docker-up-d
```

---

## 验证方式

ACME Console 支持两种域名验证方式：

### DNS-01（默认）

通过添加 DNS TXT 记录验证域名所有权。

- 支持通配符证书（`*.example.com`）
- 适合 DNS 服务商不提供 API 的手工场景
- 创建证书后，系统会生成 TXT 主机名和值，添加到 DNS 后点击「验证并签发」

### HTTP-01

通过在 Web 服务器放置验证文件来验证域名所有权。

- **不支持通配符证书**
- 适合有 Web 服务器控制权但不方便修改 DNS 的场景
- 创建证书后，系统会生成文件路径和内容，需要在域名对应的 Web 服务器上配置：

```nginx
# Nginx 示例 — 让 ACME 验证路径可访问
location /.well-known/acme-challenge/ {
    root /var/www/acme;
}
```

然后将系统生成的文件内容写入对应路径，例如：
```bash
# 路径和内容从 ACME Console 页面复制
mkdir -p /var/www/acme/.well-known/acme-challenge/
echo -n "验证内容" > /var/www/acme/.well-known/acme-challenge/token文件名
```

确保 `http://你的域名/.well-known/acme-challenge/token` 能正常访问后，点击「验证并签发」。

### CA 环境

在创建证书页面的「高级选项」中可以切换：

| 环境 | 说明 |
|------|------|
| **Production** | 正式环境，签发受浏览器信任的证书（有速率限制） |
| **Staging** | 测试环境，签发不受信任的测试证书（无速率限制，适合调试） |

---

## 架构

```
            HTTP(:80)  ─┐
用户 ───────────────────┤→ Nginx ──→ Backend(:10020) ──→ MySQL(:3306)
           HTTPS(:443) ─┘     ↓
                         静态文件 (SPA)
```

数据持久化在 `deploy/data/` 目录下。

---

## 备份

```bash
# 备份
docker compose -f deploy/docker-compose.yaml exec mysql \
  mysqldump -u root -p acme_console > backup.sql

# 恢复
docker compose -f deploy/docker-compose.yaml exec -i mysql \
  mysql -u root -p acme_console < backup.sql
```
