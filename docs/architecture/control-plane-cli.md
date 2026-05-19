# 控制面 CLI 设计

## 目标

`gamectl` 是项目的运维控制面 CLI。第一版采用“CLI 调用后台控制面 API”的方式，不直接修改服务器进程或数据库文件。

覆盖能力：

- 服务管理。
- 配置文件或配置数据更新。
- 功能和配置的 AB 测试。
- 功能灰度发布。
- 不停服滚动更新计划。
- 服务器状态和负载查看。

## 架构

```text
operator
  -> gamectl
  -> control HTTP API
  -> MySQL: 配置、AB、灰度、更新计划、服务状态
  -> Memcached: 配置和规则缓存
  -> services: 周期性拉取或订阅配置变更
```

MySQL 是控制面数据的权威存储。Memcached 只缓存配置和规则快照，不能作为权威来源。

## 当前实现

当前 demo 中已经接入控制面 API：

- `GET /control/v1/status`
- `GET /control/v1/services`
- `POST /control/v1/services/{service}/actions`
- `GET /control/v1/configs`
- `GET /control/v1/configs/{key}`
- `PUT /control/v1/configs/{key}`
- `GET /control/v1/ab-tests`
- `POST /control/v1/ab-tests`
- `GET /control/v1/rollouts`
- `POST /control/v1/rollouts`
- `GET /control/v1/updates`
- `POST /control/v1/updates`
- `GET /control/v1/versions`
- `POST /control/v1/versions`
- `GET /control/v1/releases`
- `POST /control/v1/releases`
- `GET /control/v1/rollbacks`
- `POST /control/v1/rollbacks`
- `GET /control/v1/nodes`

当前运行时使用内存 store，生产环境应替换为 MySQL repository，并在写入成功后删除或刷新 Memcached 缓存。

## CLI 用法

构建：

```bash
go build -o build/gamectl ./cmd/gamectl
```

查看控制面状态：

```bash
gamectl --addr http://127.0.0.1:8080 status
```

服务管理：

```bash
gamectl --addr http://127.0.0.1:8080 service list
gamectl --addr http://127.0.0.1:8080 service action gameplay-service restart
```

配置更新：

```bash
gamectl --addr http://127.0.0.1:8080 config set landlord.base_score 10 --scope landlord
gamectl --addr http://127.0.0.1:8080 config get landlord.base_score
gamectl --addr http://127.0.0.1:8080 config list
```

AB 测试：

```bash
gamectl --addr http://127.0.0.1:8080 abtest create \
  --name landlord_new_shuffle \
  --feature landlord.shuffle \
  --variants control,new \
  --traffic 5

gamectl --addr http://127.0.0.1:8080 abtest list
```

灰度发布：

```bash
gamectl --addr http://127.0.0.1:8080 rollout create \
  --feature landlord.super_bomb \
  --percent 10 \
  --strategy user_id_hash

gamectl --addr http://127.0.0.1:8080 rollout list
```

不停服更新计划：

```bash
gamectl --addr http://127.0.0.1:8080 update plan \
  --service gameplay-service \
  --version v1.2.3 \
  --strategy rolling

gamectl --addr http://127.0.0.1:8080 update list
```

服务版本控制：

```bash
gamectl --addr http://127.0.0.1:8080 version register \
  --service gameplay-service \
  --version v1.2.3 \
  --artifact registry/gameplay:v1.2.3 \
  --checksum sha256:abc

gamectl --addr http://127.0.0.1:8080 version list
```

发布：

```bash
gamectl --addr http://127.0.0.1:8080 release deploy \
  --service gameplay-service \
  --version v1.2.3 \
  --strategy rolling

gamectl --addr http://127.0.0.1:8080 release list
```

回滚：

```bash
gamectl --addr http://127.0.0.1:8080 rollback \
  --service gameplay-service \
  --target-version v1.2.2 \
  --reason "bad release"
```

节点负载：

```bash
gamectl --addr http://127.0.0.1:8080 nodes
```

## 生产落地要求

服务管理：

- 控制面只创建目标状态和操作记录。
- 具体 start、stop、restart、rolling update 由 agent 或编排系统执行。
- 所有操作必须记录 operator、request_id、service、action 和结果。

配置更新：

- 写 MySQL 前校验 schema。
- 写入成功后递增版本号。
- 删除或刷新 Memcached 中的配置快照。
- 服务按版本拉取配置，避免读到半更新状态。

AB 测试：

- 分流策略必须稳定，例如 user_id hash。
- 实验配置写 MySQL，服务侧缓存快照。
- 实验状态至少包含 draft、running、paused、finished。

灰度发布：

- 灰度规则按 feature_key 生效。
- 支持按用户百分比、白名单、渠道、版本和服务器分组扩展。
- 所有灰度规则必须可回滚。

不停服更新：

- 第一版只生成 rolling update 计划。
- 服务版本先注册到 `control_service_versions`，发布时只能选择已注册版本。
- 发布记录写入 `control_releases`，记录目标版本、上一版本、策略和状态。
- 回滚记录写入 `control_rollbacks`，记录来源版本、目标版本、原因和状态。
- 服务需要支持优雅退出、连接排空和健康检查。
- gameplay-service 的房间 owner 节点更新前必须停止接收新房间，并等待已有房间结束或迁移。

服务器状态和负载：

- 服务节点定期上报 cpu、mem、房间数、连接数和健康状态。
- 控制面根据节点负载辅助调度 room owner。
- 节点状态过期后应标记为 stale，不能继续分配新房间。

## 后续演进

- 将 `MemoryStore` 替换为 MySQL repository。
- 增加 operator 鉴权和审计日志。
- 增加配置 schema 校验。
- 增加服务节点 agent。
- 增加灰度规则评估 SDK。
- 增加 Memcached 配置缓存失效策略。
