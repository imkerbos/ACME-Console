# acme-console

> A self-hosted ACME certificate management console for real-world operations.

`acme-console` 是一个面向 **运维 / SRE / 平台团队** 的证书管理平台，  
用于解决 **acme.sh 纯命令行在多域名、手工 DNS、泛域证书、CDN 场景下的可视化与流程管理问题**。

它不是一个新的 ACME 客户端，  
而是一个 **基于现有 ACME 客户端（如 acme.sh）的管理与编排层（Console）**。

---

## ✨ Why acme-console?

在真实生产环境中，证书管理通常面临这些问题：

- 多个域名 / 泛域名（`example.com + *.example.com`）
- DNS 不在自己手里，只能 **人工加 TXT**
- 客户 / 业务方需要清晰的 TXT 配置指引
- renew 失败 / 成功没有统一视图
- 证书最终要 **上传到 CDN / LB / WAF**
- 全流程靠 shell + 人肉复制，**不可审计、不可回溯**

而现有工具要么：

- 太底层（`acme.sh / certbot`，CLI only）
- 太封闭（云厂商证书托管，证书不可导出）
- 太“面板化”，不适合批量和企业流程

👉 **acme-console 的目标就是补上这一层。**

---

## 🎯 Project Goals

- 提供 **ACME 证书生命周期的可视化管理**
- 适配 **DNS 手工模式（TXT）** 的真实企业场景
- 支持 **泛域名 / 多域名 / 批量操作**
- 输出 **人类可读、可直接发给客户的 TXT 指引**
- 让证书从“命令行黑盒”变成“可管理资产”

---

## 🚫 Non-Goals（刻意不做的事）

- ❌ 不重写 ACME 协议
- ❌ 不取代 acme.sh / certbot / lego
- ❌ 不强依赖某一家 DNS / CDN 厂商
- ❌ 不做“全自动幻想”（没有 DNS API 就不装自动）

---

## 🧩 Core Concepts

### Certificate
- 一个证书实体
- 支持：
  - RSA / ECC
  - 根域 + 泛域
  - 多 SAN

### Challenge
- ACME Challenge 记录
- 重点支持：
  - DNS-01（手工 / API）
- 可生成：
  - TXT 汇总表
  - 客户填写模板

## Flow

证书的生命周期流程：

```text
Issue -> TXT -> Verify -> Renew -> Package -> Deploy
```

### Target
- 证书最终使用位置
- CDN
- LoadBalancer
- Ingress
- File Export

---

## 🧪 MVP Features (Phase 1)

- [ ] 证书申请（基于 acme.sh）
- [ ] 泛域名支持（root + wildcard）
- [ ] TXT Challenge 自动解析与汇总
- [ ] TXT 配置模板导出（Markdown / Text）
- [ ] renew 执行与结果记录
- [ ] 证书文件打包（zip）
- [ ] 证书状态列表（Pending / Ready / Failed）

---

## 🔜 Roadmap

### Phase 2
- DNS Alias Mode 支持（CNAME 托管）
- renew 失败原因可视化
- TXT 生效检测（dig 校验）
- 到期提醒（Webhook / Email）

### Phase 3
- CDN / LB 自动部署插件
- 多环境（Stage / UAT / Prod）
- RBAC / 审计日志
- CLI + Web 混合模式

---

## 🛠️ Tech Stack (Proposed)

- Backend: Go
- ACME Client: acme.sh (external)
- API: REST
- UI: Web Console (TBD)
- Storage: SQLite / PostgreSQL
- Auth: Local / OIDC (future)

---

## 📂 Architecture (High Level)


```text
+-------------+        +------------------+        +-------------+
| Web Console | -----> | acme-console API | -----> | acme.sh CLI |
+-------------+        +------------------+        +-------------+
                               |
                               v
                        +------------------+
                        |     Storage      |
                        |  SQLite / PG     |
                        +------------------+
                               |
                               v
                        +------------------+
                        |     ACME CA      |
                        | ZeroSSL / LE     |
                        +------------------+

```


---

## 🧠 Design Philosophy

- **Reality-first**：承认 DNS 不可控
- **Human-friendly**：TXT 要给人看
- **Ops-oriented**：失败是常态，必须可追踪
- **Composable**：CLI / API / UI 都能用

---

## 📜 License

MIT

---

## 👤 Author

Built by operators, for operators.

