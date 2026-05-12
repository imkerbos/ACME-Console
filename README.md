# ACME Console

> A self-hosted ACME certificate management console for real-world operations.

`acme-console` 是一个面向 **运维 / SRE / 平台团队** 的证书管理平台，
用于解决 **acme.sh 纯命令行在多域名、手工 DNS、泛域证书、CDN 场景下的可视化与流程管理问题**。

它不是一个新的 ACME 客户端，
而是一个 **基于 ACME 协议的证书管理与编排层（Console）**。

---

## Features

- **证书全生命周期管理** — 申请、验证、签发、续期、吊销、下载，一站式完成
- **DNS-01 验证** — 支持手工 TXT 模式，适配 DNS 不在自己手里的真实场景
- **HTTP-01 验证** — 在 Web 服务器放置验证文件，无需修改 DNS（不支持通配符）
- **泛域名 & 多 SAN** — `example.com + *.example.com` 开箱即用
- **RSA / ECC 双算法** — 按需选择密钥类型
- **失败重试** — 验证失败后可重新创建 ACME 订单，无需重新填写信息
- **证书吊销** — 支持向 CA 申请吊销已签发证书
- **Production / Staging** — 支持切换 Let's Encrypt 正式/测试环境
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

### Development

```bash
make dev-docker-up      # 前台运行
make dev-docker-up-d    # 后台运行
```

前端 http://localhost:3200 | 后端 http://localhost:10020 | 代码修改自动热更新

### Production

```bash
# 1. 配置（只需编辑这一个文件）
cp deploy/.env.example deploy/.env && vim deploy/.env

# 2. 启动（自动渲染 config.yaml 并启动所有服务）
make prod-docker-up-d
```

Default login: `admin` / `admin123` (please change after first login)

详见 [部署文档](docs/deployment-production.md)。

---

## Certificate Lifecycle

```text
Issue → Challenge → Verify → Renew → Package → Deploy
```

1. **Issue** — 创建证书，填写域名，选择验证方式（DNS-01 / HTTP-01）
2. **Challenge** — DNS-01：系统生成 TXT 记录；HTTP-01：系统生成验证文件路径和内容
3. **Verify** — 配置好验证信息后，点击「验证并签发」
4. **Renew** — 到期前自动/手动续期
5. **Package** — 下载证书包（PEM / PFX / ZIP）
6. **Deploy** — 部署到 CDN / LB / Ingress

> 验证失败？点击「重试」即可重新创建 ACME 订单，无需重新填写信息。

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
