# 部署文档更新说明

## 问题发现

在审查部署配置时，发现项目现有的 `deploy/Dockerfile` 存在以下问题：

### 1. 缺少前端构建

**现有 Dockerfile** (`deploy/Dockerfile`):
```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
# ... 只构建了后端
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/acme-console ./cmd/server

# Runtime stage
FROM alpine:3.19
COPY --from=builder /app/bin/acme-console /app/acme-console
# ⚠️ 没有前端静态文件
CMD ["-config", "/app/configs/config.yaml"]
# ⚠️ 没有 -static 参数
```

**问题**：
- 只构建了 Go 后端，没有构建前端（Vue 3）
- 启动命令缺少 `-static` 参数
- 容器只能提供 API 服务，无法访问 Web 控制台

### 2. 架构说明

ACME Console 采用**前后端一体化部署**架构：

```
后端 Go 服务 (cmd/server/main.go)
├── 当使用 -static 参数时
│   ├── /api/*     → API 处理器
│   ├── /health    → 健康检查
│   └── /*         → 静态文件服务 (前端)
│                    └─► SPA fallback 到 index.html
└── 路由逻辑在 internal/router/router.go
```

**关键代码** (`internal/router/router.go:126-128`):
```go
// Serve static files if provided
if staticFS != nil {
    r.Use(staticFileHandler(staticFS))
}
```

**关键代码** (`cmd/server/main.go:112-115`):
```go
var staticFS fs.FS
if *staticDir != "" {
    staticFS = os.DirFS(*staticDir)
    logger.Info("Serving static files", logger.String("path", *staticDir))
}
```

## 解决方案

### 完整的生产 Dockerfile

已在部署文档中提供完整的 Dockerfile，包含：

1. **多阶段构建**：
   - Stage 1: 构建 Go 后端
   - Stage 2: 构建 Vue 前端
   - Stage 3: 运行时镜像

2. **正确的启动命令**：
   ```dockerfile
   CMD ["-config", "/app/configs/config.yaml", "-static", "/app/static"]
   ```

3. **完整的文件结构**：
   ```
   /app/
   ├── bin/
   │   └── acme-console
   ├── static/          # 前端构建产物
   │   ├── index.html
   │   ├── assets/
   │   └── ...
   └── configs/
       └── config.yaml
   ```

## 部署文档更新

已创建/更新以下文档：

### 1. `docs/deployment-production.md` - 生产环境部署文档

**新增内容**：
- ✅ 架构说明（前后端一体化部署）
- ✅ 完整的生产 Dockerfile（包含前端构建）
- ✅ Nginx 配置说明（为什么只需代理到一个端口）
- ✅ 性能优化选项（Nginx 直接服务静态文件）
- ✅ 故障排查（检查 -static 参数）

### 2. `docs/deployment-docker-compose.md` - Docker Compose 部署文档

**新增内容**：
- ✅ 架构说明（前后端一体化部署）
- ✅ 指出现有 Dockerfile 的问题
- ✅ 提供完整的 Dockerfile
- ✅ 反向代理配置说明
- ✅ 故障排查（检查静态文件和 -static 参数）

## Nginx 配置说明

### 为什么 Nginx 只需代理到一个端口？

因为后端服务同时处理 API 和静态文件：

```nginx
# 正确的配置
location / {
    proxy_pass http://127.0.0.1:10020;
    # 后端会根据路径自动分发：
    # /api/* → API 处理器
    # /health → 健康检查
    # /* → 静态文件服务（SPA 支持）
}
```

### 性能优化（可选）

如果需要更好的静态文件性能，可以让 Nginx 直接服务静态文件：

```nginx
# 静态资源由 Nginx 直接服务
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
    root /opt/acme-console/static;
    expires 1y;
    add_header Cache-Control "public, immutable";
}

# API 和 HTML 请求代理到后端
location / {
    proxy_pass http://127.0.0.1:10020;
}
```

## 验证清单

部署后请验证：

- [ ] 访问 `http://localhost:10020` 能看到前端页面
- [ ] 访问 `http://localhost:10020/health` 返回 `{"status":"ok"}`
- [ ] 访问 `http://localhost:10020/api/v1/auth/login` 能调用 API
- [ ] 前端路由正常工作（刷新页面不会 404）
- [ ] 静态资源正常加载（JS、CSS、图片等）

## 后续建议

### 1. 更新项目的 Dockerfile

建议更新 `deploy/Dockerfile` 为完整版本：

```bash
# 备份现有文件
mv deploy/Dockerfile deploy/Dockerfile.old

# 使用文档中的完整 Dockerfile
cp docs/deployment-docker-compose.md deploy/Dockerfile
# （提取 Dockerfile 部分）
```

### 2. 更新 Makefile

确保 `make docker-build` 使用正确的 Dockerfile：

```makefile
docker-build:
    docker build -f deploy/Dockerfile.prod -t acme-console:latest .
```

### 3. 添加 CI/CD

在 GitHub Actions 中添加构建验证：

```yaml
- name: Build Docker image
  run: docker build -f deploy/Dockerfile.prod -t acme-console:test .

- name: Test image
  run: |
    docker run -d --name test acme-console:test
    docker exec test ls -la /app/static  # 验证静态文件存在
    docker exec test /app/bin/acme-console -help  # 验证二进制文件可执行
```

---

**文档版本**: 1.0
**创建日期**: 2026-02-08
**作者**: Claude Code
