# ACME Console

> A self-hosted ACME certificate management console for real-world operations.

`acme-console` 是一个面向 **运维 / SRE / 平台团队** 的证书管理平台，
用于解决 **acme.sh 纯命令行在多域名、手工 DNS、泛域证书、CDN 场景下的可视化与流程管理问题**。

它不是一个新的 ACME 客户端，
而是一个 **基于 ACME 协议的证书管理与编排层（Console）**。

---

## Features

- **证书全生命周期管理** — 申请、验证、签发、续期、下载，一站式完成
- **DNS-01 验证** — 支持手工 TXT 模式，适配 DNS 不在自己手里的真实场景
- **泛域名 & 多 SAN** — `example.com + *.example.com` 开箱即用
- **RSA / ECC 双算法** — 按需选择密钥类型
- **TXT 记录汇总导出** — 人类可读，可直接发给客户 / 业务方
- **工作空间** — 团队协作，按项目 / 客户隔离证书
- **到期通知** — 支持 Webhook、Telegram、飞书
- **多语言** — 中文 / English
- **用户管理** — 管理员 / 普通用户角色
- **系统设置** — 自定义网站标题等

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go (Gin + GORM) |
| Frontend | Vue.js 3 + Vue Router + Vue I18n |
| Database | MySQL |
| ACME | lego (Go library) |
| Auth | JWT |
| Deploy | Docker / Docker Compose |

---

## Architecture

```text
+-------------+        +------------------+        +-------------+
| Web Console | -----> | acme-console API | -----> |    lego     |
|  (Vue.js)   |        |    (Go/Gin)      |        | (ACME lib)  |
+-------------+        +------------------+        +-------------+
                               |                          |
                               v                          v
                        +------------------+      +------------------+
                        |     Storage      |      |     ACME CA      |
                        |      MySQL       |      | Let's Encrypt /  |
                        +------------------+      |     ZeroSSL      |
                                                  +------------------+
```

---

## Quick Start

### Docker Compose (Recommended)

```bash
# Clone the repo
git clone https://github.com/imkerbos/ACME-Console.git
cd ACME-Console

# Copy and edit config
cp deploy/docker/config.yaml deploy/docker/config.local.yaml
# Edit config.local.yaml: set encryption.master_key, jwt.secret, etc.

# Start
cd deploy/docker
docker compose up -d
```

Default login: `admin` / `admin123` (please change after first login)

See [Deployment Guide](docs/deployment-production.md) for production setup details.

---

## Certificate Lifecycle

```text
Issue → TXT → Verify → Renew → Package → Deploy
```

1. **Issue** — 创建证书，填写域名
2. **TXT** — 系统生成 DNS-01 TXT 记录
3. **Verify** — 添加 TXT 后点击验证并签发
4. **Renew** — 到期前续期
5. **Package** — 下载证书包（含 cert / key / chain）
6. **Deploy** — 部署到 CDN / LB / Ingress

---

## Design Philosophy

- **Reality-first** — 承认 DNS 不可控，不做"全自动幻想"
- **Human-friendly** — TXT 记录要给人看，模板要能直接发客户
- **Ops-oriented** — 失败是常态，一切必须可追踪
- **Composable** — API 和 Web 界面都能用

---

## License

MIT

---

## Author

Built by [Kerbos](https://github.com/imkerbos), for operators.
