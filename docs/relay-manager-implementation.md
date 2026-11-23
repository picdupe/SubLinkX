# SublinkX HAProxy 中转管理模块实施方案

> 状态：设计稿，待结合当前仓库源码审查后落地  
> 目标：在 SublinkX 内增加 HAProxy TCP 中转服务器管理能力，并通过独立 Rust Agent 管理远端 HAProxy。  
> 方案选择：保留 SublinkX 现有 Go + Gin + GORM + Vue 3 架构；Agent 使用 Rust；Agent 主动通过 HTTPS API 拉取任务。  
> 第一版范围：单级 TCP 线路 `客户端 → HAProxy 中转 → 现有落地节点`。

---

## 1. 目标与非目标

### 1.1 目标

- 在 SublinkX 后台集中登记、管理多台 HAProxy 中转服务器。
- 通过 Rust Agent 管理远端 HAProxy，不依赖 SSH 常驻管理、不开放 Agent 入站管理端口。
- 管理端维护中转线路：中转服务器监听地址/端口 → SublinkX 已登记的落地节点。
- 管理端生成 HAProxy 配置快照、版本和任务。
- Agent 下载配置后执行完整性校验、HAProxy 语法检查、原子替换、平滑 Reload、失败回滚。
- 订阅生成时从原落地节点派生中转节点，只替换客户端连接入口，不改变落地节点的协议身份参数。
- 兼容 Alpine/OpenRC、Debian/Ubuntu/Rocky/systemd。

### 1.2 第一版不做

- 不管理或安装 Xray，不修改落地服务。
- 不做多级中转、UDP/QUIC、复杂负载均衡、自动测速选路。
- 不提供任意 Shell 执行能力。
- 不允许线路目标填写任意 IP/端口；目标必须关联现有落地节点。
- 不覆盖管理员已有的 HAProxy 主配置。
- 不把 Agent 设计成通用远程运维代理。

### 1.3 核心原则

1. SublinkX 数据库是业务配置的权威来源。
2. 每台中转服务器的完整期望配置必须有不可变版本快照。
3. Agent 只执行固定、白名单化的任务类型。
4. HAProxy 主配置与 SublinkX 管理片段分离。
5. 订阅生成只输出已启用、已应用且健康的线路。
6. 任何配置失败都不能将旧配置和旧连接无保护地破坏。

---

## 2. 总体架构

```text
SublinkX 管理端（Go / Gin / GORM）
├── Relay Manager
│   ├── 中转服务器管理
│   ├── 中转线路管理
│   ├── 配置渲染与版本
│   ├── Agent API / 任务队列
│   └── 操作审计
└── Subscription Engine
    └── 从落地节点派生中转节点
              │ HTTPS Pull
       ┌──────┴──────┐
       ▼             ▼
 Rust Agent A    Rust Agent B
       │             │
    HAProxy        HAProxy
       │             │
       ▼             ▼
  现有 Xray 落地节点
```

通信由 Agent 主动发起：HTTPS 心跳、长轮询拉取任务、下载配置、上报结果。管理端不主动连接 Agent，不要求中转服务器开放 Agent 管理端口。

建议第一版使用 HTTPS + 每 Agent 独立随机 Token；后续可增加 mTLS。Token 仅在签发时展示一次，数据库只存强哈希，支持轮换和撤销。

---

## 3. 数据库设计

以下是逻辑设计，实际字段类型、时间类型、主键策略、软删除约定、表名前缀及迁移方式必须先检查当前 SublinkX 数据模型后统一。不要直接照搬 SQL 类型。

### 3.1 `relay_servers`：中转服务器

| 字段 | 类型建议 | 说明 |
|---|---|---|
| id | 主键 | 服务器 ID |
| name | varchar(100) | 显示名称 |
| public_host | varchar(255) | 客户端访问中转服务器的域名/IP |
| agent_id | varchar(64), unique | Agent 稳定 ID |
| agent_token_hash | varchar(255) | Token 强哈希，不存明文 |
| agent_token_created_at | timestamp | Token 创建时间 |
| agent_version | varchar(32) | Agent 版本 |
| os_info | varchar(255) | OS/发行版 |
| architecture | varchar(32) | CPU 架构 |
| service_manager | varchar(20) | openrc/systemd/unknown |
| haproxy_version | varchar(64) | HAProxy 版本 |
| haproxy_status | varchar(20) | running/stopped/unknown |
| desired_revision | bigint | 管理端期望版本 |
| applied_revision | bigint | Agent 已成功应用版本 |
| last_heartbeat_at | timestamp | 最近心跳 |
| enabled | bool | 是否启用 |
| created_at / updated_at | timestamp | 时间 |

约束：
- `agent_id` 唯一。
- 服务器删除前检查关联线路和任务。
- Token 不可回显；轮换时旧 Token 应可立即撤销。
- `public_host` 不是 Agent API 地址，而是客户端连接入口。

### 3.2 `relay_routes`：中转线路

| 字段 | 类型建议 | 说明 |
|---|---|---|
| id | 主键 | 线路 ID |
| name | varchar(100) | 线路名称 |
| relay_server_id | FK | 中转服务器 |
| landing_node_id | FK | 现有 SublinkX 落地节点 |
| protocol | varchar(10) | 第一版固定 tcp |
| listen_address | varchar(64) | HAProxy bind 地址，默认 0.0.0.0 |
| listen_port | int | 中转监听端口 |
| enabled | bool | 是否启用 |
| publish_enabled | bool | 是否允许订阅输出 |
| priority | int | 输出排序 |
| description | text | 备注 |
| created_at / updated_at | timestamp | 时间 |

目标地址/端口从 `landing_node_id` 对应的现有节点模型读取，不把用户提交的 `target_host` / `target_port` 当作权威字段。若业务需要固定目标快照，应在生成 revision 时保存解析后的目标值，并记录来源节点 ID。

约束：
- `UNIQUE(relay_server_id, protocol, listen_port)`。
- `listen_port` 必须在 1–65535。
- 禁止未授权的低端口、管理端口及与系统服务冲突的端口；具体策略由管理员配置。
- `listen_address` 仅允许受支持的 IP/通配地址，不允许任意文本。
- 删除落地节点时必须检查关联中转线路。

### 3.3 `relay_config_revisions`：不可变配置版本

| 字段 | 说明 |
|---|---|
| id | 版本 ID |
| relay_server_id | 所属服务器 |
| revision | 服务器内单调递增版本号 |
| config_sha256 | 配置内容 SHA-256 |
| config_content | 完整生成配置快照 |
| change_summary | 变更说明 |
| created_by | 管理员 ID |
| created_at | 创建时间 |

约束：
- `(relay_server_id, revision)` 唯一。
- revision 创建后不可修改。
- 保存渲染后的完整配置，不仅保存模板和线路差异。
- 敏感字段不应进入 HAProxy 配置；配置下载应使用短时授权或 Agent Token 鉴权。

### 3.4 `relay_config_tasks`：Agent 任务

| 字段 | 说明 |
|---|---|
| id | UUID/唯一任务 ID |
| relay_server_id | 目标服务器 |
| revision_id | 对应配置版本 |
| task_type | 第一版固定 `apply_config` / `collect_status` |
| status | pending/leased/running/success/failed/expired |
| attempt | 执行尝试次数 |
| lease_expires_at | 任务租约到期时间 |
| error_code | 稳定错误码 |
| error_message | 脱敏错误信息 |
| started_at / finished_at | 执行时间 |
| created_at | 创建时间 |

任务拉取必须有原子领取/租约机制，避免同一任务被多个 Agent 并发领取。任务超时后可重新入队，但应用操作必须幂等。

### 3.5 `relay_agent_events`：审计事件

记录注册、Token 轮换、心跳异常、任务领取、配置校验、Reload、回滚、Agent 升级等重要事件。高频心跳不必全部写入事件表；服务器当前状态写入 `relay_servers`，普通心跳可采样或定期清理。

### 3.6 版本一致性

- 每次线路变更，在同一数据库事务中生成新的配置 revision、更新 `desired_revision`、创建 `apply_config` 任务。
- `applied_revision` 只能在 Agent 成功应用并完成健康检查后更新。
- `desired_revision != applied_revision` 表示配置未收敛。
- 旧任务不能把服务器状态回写成比当前 `desired_revision` 更新的状态。
- 对同一服务器的配置应用任务应串行执行；新 revision 可合并掉尚未开始的旧 revision 任务。

---

## 4. 管理端 API

统一管理端前缀建议 `/api/admin/relay`，沿用 SublinkX 现有管理员认证、RBAC 和响应格式。

| 方法 | 路径 | 功能 |
|---|---|---|
| GET | `/servers` | 中转服务器列表 |
| POST | `/servers` | 创建服务器记录/生成安装凭证 |
| GET | `/servers/:id` | 服务器详情 |
| PUT | `/servers/:id` | 修改服务器元数据 |
| DELETE | `/servers/:id` | 删除（需检查线路和任务） |
| POST | `/servers/:id/rotate-token` | 轮换 Agent Token |
| POST | `/servers/:id/reconcile` | 重新生成期望配置并创建任务 |
| POST | `/servers/:id/rollback` | 选择历史 revision 回滚 |
| GET | `/servers/:id/revisions` | 配置版本历史 |
| GET | `/servers/:id/tasks` | 任务历史 |
| GET | `/routes` | 线路列表 |
| POST | `/routes` | 创建线路 |
| GET | `/routes/:id` | 线路详情 |
| PUT | `/routes/:id` | 修改线路 |
| DELETE | `/routes/:id` | 删除线路 |
| POST | `/routes/:id/enable` | 启用线路 |
| POST | `/routes/:id/disable` | 禁用线路 |

创建线路示例：

```json
{
  "name": "San Jose Relay",
  "relay_server_id": 1,
  "landing_node_id": 12,
  "protocol": "tcp",
  "listen_address": "0.0.0.0",
  "listen_port": 26281,
  "enabled": true,
  "publish_enabled": true,
  "priority": 10
}
```

API 不接受客户端传入的任意目标地址。服务端根据 `landing_node_id` 读取已登记落地节点，并验证节点协议、地址和端口可用于 TCP 转发。

---

## 5. Agent API 与通信协议

Agent API 统一前缀 `/api/agent/v1`，使用独立鉴权中间件，不复用浏览器 Session。

| 方法 | 路径 | 功能 |
|---|---|---|
| POST | `/enroll` | 首次注册/兑换安装凭证 |
| POST | `/heartbeat` | 心跳与运行状态 |
| GET | `/tasks/next?wait=25` | 长轮询领取任务 |
| POST | `/tasks/:id/ack` | 确认领取 |
| GET | `/configs/:revision` | 下载指定配置快照 |
| POST | `/tasks/:id/result` | 上报任务结果 |
| GET | `/status` | 获取 Agent 当前管理状态 |

### 5.1 心跳示例

```json
{
  "agent_id": "relay-001",
  "agent_version": "0.1.0",
  "os": "Alpine Linux",
  "architecture": "x86_64",
  "service_manager": "openrc",
  "haproxy_version": "3.x",
  "haproxy_status": "running",
  "desired_revision": 8,
  "applied_revision": 7
}
```

### 5.2 任务示例

```json
{
  "task_id": "uuid",
  "type": "apply_config",
  "revision": 8,
  "config_url": "/api/agent/v1/configs/8",
  "sha256": "hex-sha256",
  "created_at": "RFC3339 timestamp"
}
```

### 5.3 结果示例

成功：

```json
{
  "success": true,
  "revision": 8,
  "sha256": "hex-sha256",
  "validation": "passed",
  "reload": "success",
  "haproxy_status": "running",
  "duration_ms": 1250
}
```

失败并回滚成功：

```json
{
  "success": false,
  "revision": 8,
  "validation": "passed",
  "reload": "failed",
  "rollback": "success",
  "error_code": "HAPROXY_RELOAD_FAILED",
  "error_message": "Reload failed; previous configuration restored"
}
```

### 5.4 通信与安全约束

- Agent 主动建立 HTTPS 连接，不开放 Agent 入站端口。
- 每个 Agent 使用独立随机凭证；数据库只保存强哈希。
- 支持凭证轮换、撤销和过期。
- 生产环境验证服务端 TLS 证书；不允许关闭证书验证。
- 可增加请求时间戳、Nonce 或签名，防止重放。
- Agent 只接受固定任务枚举，不执行任意命令。
- 任务必须有租约、幂等 ID、超时和重试限制。
- 配置下载鉴权，避免配置快照被未授权读取。
- 结果上报必须验证任务确实属于该 Agent。
- 错误信息脱敏，不回传 Token、私钥或完整环境变量。

---

## 6. Rust Agent 设计

建议独立仓库/目录 `sublinkx-relay-agent`，Rust 稳定版，第一版支持 Linux x86_64 和 aarch64。

建议目录：

```text
sublinkx-relay-agent/
├── Cargo.toml
├── src/
│   ├── main.rs
│   ├── config.rs
│   ├── client/
│   │   ├── auth.rs
│   │   ├── heartbeat.rs
│   │   └── tasks.rs
│   ├── haproxy/
│   │   ├── config.rs
│   │   ├── validate.rs
│   │   ├── reload.rs
│   │   ├── status.rs
│   │   └── rollback.rs
│   ├── system/
│   │   ├── service.rs
│   │   ├── openrc.rs
│   │   └── systemd.rs
│   └── storage/
│       └── state.rs
├── tests/
└── packaging/
    ├── systemd/
    └── openrc/
```

建议依赖：tokio、reqwest（rustls）、serde、serde_json、sha2、uuid、tracing、tracing-subscriber、thiserror、anyhow。

Agent 配置示例：

```toml
[agent]
id = "relay-001"
server_url = "https://panel.example.com"
token_file = "/etc/sublinkx-agent/token"
heartbeat_interval = 30
poll_timeout = 25

[haproxy]
binary = "/usr/sbin/haproxy"
main_config = "/etc/haproxy/haproxy.cfg"
managed_config = "/etc/haproxy/sublinkx/relay.cfg"
service_name = "haproxy"

[system]
service_manager = "auto"
state_file = "/var/lib/sublinkx-agent/state.json"
backup_dir = "/var/lib/sublinkx-agent/backups"
```

`service_manager=auto` 检测 systemd/OpenRC。Alpine/OpenRC 必须作为第一优先适配和验收环境。

Agent 权限应最小化：安装时明确授予管理 HAProxy 配置、执行指定 HAProxy 校验和触发指定服务 Reload 所需权限。不得把 Agent 做成通用 root Shell。Token 文件权限建议 `0600`，状态/备份目录仅 Agent 管理用户可读写。若 HAProxy 需要 root 权限，考虑将特权操作收敛到固定 helper，而不是让整个网络 Agent 进程长期以 root 运行。

---

## 7. HAProxy 配置生成

### 7.1 配置文件隔离

不覆盖现有 `/etc/haproxy/haproxy.cfg`。SublinkX 只维护独立片段：

```text
/etc/haproxy/
├── haproxy.cfg
└── sublinkx/
    └── relay.cfg
```

主配置通过第二个 `-f` 参数加载管理片段。部署前必须检查实际 systemd/OpenRC 服务启动参数，确保校验命令与运行服务加载的是同一组配置文件。

校验示例：

```bash
haproxy -c \
  -f /etc/haproxy/haproxy.cfg \
  -f /etc/haproxy/sublinkx/relay.cfg
```

不能只执行上述校验却没有让服务实际加载 `relay.cfg`。如果服务启动配置需要调整，必须先备份、检查、验证并可回滚。

### 7.2 TCP 转发模板

示例：中转监听 `26281`，落地节点 `155.94.155.178:26281`。

```haproxy
frontend sublinkx_relay_in_1
    bind 0.0.0.0:26281
    mode tcp
    default_backend sublinkx_relay_out_1

backend sublinkx_relay_out_1
    mode tcp
    server landing_1 155.94.155.178:26281 check inter 5s fall 3 rise 2
```

第一版：
- 纯 TCP 透传，不终止 TLS/REALITY。
- 不添加 PROXY Protocol。
- 不改写 SNI、UUID、REALITY 公钥或 Short ID。
- 健康检查先使用 TCP connect check。
- 线路 ID 生成稳定且合法的 HAProxy section 名称，不使用用户输入直接拼接配置标识符。
- 对 `target_host` 做 DNS/IP 解析及地址策略校验，禁止回环、链路本地、管理网段和未授权内网目标；目标只来源于受信任落地节点记录。

多条线路生成独立 frontend/backend，监听端口唯一。

### 7.3 配置拼接注意

HAProxy 的多个 `-f` 配置文件会按顺序加载，但配置文件片段的 section 作用域、重复 `global/defaults`、继承关系必须按目标 HAProxy 版本验证。应生成不含重复 `global/defaults` 的管理片段，并在实际安装版本上做集成测试。

---

## 8. 原子更新、Reload 与回滚

禁止直接覆盖配置后立即重启服务。

### 8.1 推荐流程

1. 下载配置到受控临时文件。
2. 校验 SHA-256。
3. 检查 revision 单调性及任务 ID。
4. 写入与目标文件同一文件系统的临时文件。
5. 设置 owner/mode，执行 fsync。
6. 使用实际服务加载的完整配置参数执行 `haproxy -c`。
7. 保存当前已应用配置和必要的服务启动配置备份。
8. 原子 Rename 替换管理片段。
9. 执行受控 Reload。
10. 检查 HAProxy 进程、服务状态及监听端口。
11. 成功后写入本地 `applied_revision`，上报结果。
12. 失败则恢复上一份文件，再次校验并 Reload，报告回滚结果。

保留最近 10 个配置备份，按 SHA-256 校验。

### 8.2 Reload 方式

- Debian/Ubuntu/Rocky：systemd 的 Reload 操作必须先确认 unit 文件确实实现了正确的 HAProxy 平滑 Reload。
- Alpine：检查 OpenRC 脚本是否提供可靠 reload；没有时使用经过验证的 HAProxy master-worker / Master CLI 方案。
- 不假设所有系统都支持同名 `reload` 命令。
- 优先平滑 Reload，不使用无必要的 stop/start。
- HAProxy Runtime API 的动态修改不作为权威配置来源，因为运行时变更未必持久化，Reload 后可能丢失。

### 8.3 回滚边界

如果 Reload 命令返回失败，但新进程实际已经启动，不能仅凭退出码盲目覆盖。Agent 必须先检查实际运行状态和配置版本，再决定是否恢复。回滚失败时标记 `degraded`，保留现场和备份，并在管理端告警；不得无限重试。

---

## 9. 订阅生成集成

中转线路只改变客户端连接入口，REALITY/协议身份参数仍来自落地节点。

原始节点示例：

```yaml
- name: San Jose
  type: vless
  server: 155.94.155.178
  port: 26281
  uuid: YOUR_UUID
  network: tcp
  tls: true
  servername: tesla.com
  client-fingerprint: chrome
  reality-opts:
    public-key: YOUR_PUBLIC_KEY
    short-id: YOUR_SHORT_ID
```

中转派生节点：

```yaml
- name: San Jose - Relay
  type: vless
  server: relay.example.com
  port: 26281
  uuid: YOUR_UUID
  network: tcp
  tls: true
  servername: tesla.com
  client-fingerprint: chrome
  reality-opts:
    public-key: YOUR_PUBLIC_KEY
    short-id: YOUR_SHORT_ID
```

只改变：
- `server`：中转服务器 `public_host`。
- `port`：中转线路 `listen_port`。
- `name`：派生名称。

保留：
- UUID、协议、传输层、TLS/REALITY 参数、SNI、指纹、公钥、Short ID、flow 等落地节点身份参数。

建议增加独立派生函数（示意）：

```go
func BuildRelayNode(
    node Node,
    route RelayRoute,
    relay RelayServer,
) (Node, error)
```

实现时深拷贝原节点，禁止修改数据库中的原始 Node 对象或共享缓存对象。

### 9.1 发布条件

只有满足以下条件才允许订阅输出：
- 中转服务器启用。
- Agent 最近心跳在允许窗口内。
- HAProxy 状态正常。
- `applied_revision == desired_revision`。
- 线路启用且 `publish_enabled=true`。
- 落地节点仍存在且可用。
- 用户有权限使用对应节点/线路。

默认不健康线路从订阅中排除，不自动退回直连，避免意外暴露落地地址。直连回退如未来需要，必须做成显式策略并让用户知情。

### 9.2 协议范围

第一版只发布经过验证的 TCP 节点。Hysteria2、TUIC 等 UDP/QUIC 协议不得通过 TCP HAProxy 伪装成可用中转节点。每种客户端格式应分别做输出兼容测试。

---

## 10. 前端页面

在 SublinkX 管理侧栏增加「中转管理」：

1. **中转服务器**
   - 名称、域名/IP、Agent 在线状态、Agent 版本、HAProxy 版本。
   - 最近心跳、CPU/内存（若 Agent 提供）。
   - desired/applied revision。
   - 安装命令/一次性注册凭证、Token 轮换。

2. **中转线路**
   - 中转服务器、监听地址、监听端口。
   - 选择现有落地节点。
   - 启用状态、订阅发布状态、排序。
   - 端口冲突和目标节点校验。

3. **配置版本**
   - 当前版本、历史版本、变更摘要、配置摘要。
   - 应用状态、校验结果、Reload 结果、回滚按钮。

4. **操作日志**
   - Agent 注册、任务领取、配置应用、失败原因、回滚结果。
   - 对敏感信息脱敏。

所有危险操作（删除服务器、轮换 Token、回滚配置）需要权限校验和二次确认。

---

## 11. 实施阶段与验收标准

| 阶段 | 内容 | 验收 |
|---|---|---|
| P0 | 审查现有仓库结构、数据库、Node 模型、订阅转换链路、HAProxy 实际服务启动方式 | 输出与当前代码一致的实施映射，不猜字段 |
| P1 | GORM Models、迁移、管理 API、权限 | 可创建/修改/删除服务器和线路，约束正确 |
| P2 | Rust Agent 注册、心跳、Pull 任务、状态存储 | Alpine/OpenRC 和 systemd 可稳定连接 |
| P3 | HAProxy 配置渲染、校验、应用 | 不覆盖主配置，可创建 TCP 转发 |
| P4 | 版本、原子更新、Reload、回滚 | 错误配置不影响旧线路，回滚可验证 |
| P5 | 订阅派生 | 只改变连接地址/端口，协议参数保持一致 |
| P6 | 前端页面、部署包、文档和端到端测试 | 完整管理流程可操作 |

### 必须覆盖的测试

- 监听端口冲突。
- 落地节点删除或禁用。
- 目标地址不可达。
- DNS 解析异常或解析到禁止网段。
- 配置语法错误。
- 文件权限错误、磁盘写入失败。
- Reload 失败、Reload 返回异常但进程状态已变化。
- Agent 离线、重启、任务重复领取。
- 连续产生两个配置 revision。
- 旧任务迟到上报。
- 回滚成功与回滚失败。
- 订阅派生前后 REALITY 参数一致。
- 非 TCP 协议不能错误发布为 TCP 中转。

---

## 12. 当前仓库实施前必须核实的事项

本文件是设计方案，不代表已对当前仓库完成代码级核实。开始 P1 前，Codex 应先检查：

- 当前 Go 模块路径、目录结构和路由注册方式。
- GORM 数据库类型、迁移机制、主键和软删除约定。
- 现有 Node 模型中 server、port、protocol、UUID、REALITY、TLS、flow 等字段的真实结构。
- 订阅转换函数及各输出格式的实现入口。
- 当前用户权限/角色模型。
- 现有 HAProxy 中转主配置、启动参数、OpenRC/systemd 服务脚本。
- 项目现有测试命令和 CI 约定。

如实际结构与本设计不同，应先提出映射和差异，再实施；不要为了匹配本设计而破坏现有兼容性。

## 13. 建议的 Codex 首轮任务提示

```text
请先阅读 docs/relay-manager-implementation.md，暂时不要修改代码。

审查当前 SublinkX 仓库：
1. Go 后端目录、Gin 路由注册、GORM 模型和数据库迁移机制。
2. 现有 Node 数据模型及订阅生成/转换调用链。
3. 用户认证、角色权限和审计日志机制。
4. 当前测试结构和 CI。

将设计文档中的逻辑表、API 和订阅派生方案映射到真实代码。
特别检查不同数据库方言的迁移兼容性，以及节点对象深拷贝和缓存行为。

输出：
- 现有代码结构摘要；
- 设计与现有实现的差异；
- 需要新增/修改的文件清单；
- 数据库迁移风险；
- API 兼容性风险；
- 分阶段实施计划和测试计划。

未经确认，不要开始实现，不要修改现有订阅行为。
```

---

## 14. 参考资料

- HAProxy Configuration Manual：<https://docs.haproxy.org/>
- HAProxy Runtime API：<https://www.haproxy.com/documentation/haproxy-runtime-api/>
- HAProxy Management Guide：<https://www.haproxy.com/documentation/haproxy-configuration-tutorials/management/>
- Rust Tokio：<https://tokio.rs/>
- Reqwest：<https://docs.rs/reqwest/>
