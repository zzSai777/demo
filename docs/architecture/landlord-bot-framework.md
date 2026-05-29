# 斗地主机器人框架设计

## 目标与边界

斗地主机器人用于提升匹配成功率、补齐低峰期房间、承接托管出牌，并为后续 AI 策略演进预留接口。

第一版目标：

- 支持匹配超时后的机器人补位。
- 支持玩家断线或超时后的托管出牌。
- 支持按场次、房间等级、用户段位配置机器人比例和策略强度。
- 机器人行为可观测、可回放、可调参。
- 不影响资产结算一致性，不绕过 gameplay-service 的房间规则。

边界：

- 第一版不做复杂 AI，不做深度学习模型。
- 机器人不直接操作 MySQL、Memcached、ClickHouse。
- 机器人不直接调用 wallet-service。
- 机器人所有动作都必须以房间命令形式进入 room actor。
- 机器人策略只能建议动作，最终合法性由 landlord 规则引擎校验。

## 架构位置

第一版机器人作为 `gameplay-service` 内部模块存在，不单独部署 `bot-service`。

原因：

- 机器人补位、托管、出牌都强依赖房间状态。
- 和 room actor 同进程可减少跨服务延迟。
- 第一版策略简单，独立服务收益不高。
- 后续机器人策略复杂或资源消耗明显时，再拆出独立 `bot-service`。

```text
gameplay-service
  match
    -> bot seat filler
  room
    -> room actor
       -> bot controller
       -> landlord rule engine
  bot
    -> policy
    -> profile
    -> decision
```

## 核心流程

### 匹配补位

```text
玩家进入匹配队列
  -> 等待真人匹配
  -> 超过 bot_fill_timeout
  -> match 模块按配置选择机器人档位
  -> 创建机器人 seat
  -> 创建房间
  -> room actor 接管机器人动作
```

要求：

- 机器人补位必须受配置控制。
- 高价值场、比赛场可默认关闭机器人。
- 同一房间机器人数量不能超过配置上限。
- 机器人补位事件必须写入牌局事件日志。

### 托管出牌

```text
玩家超时 / 断线
  -> room actor 标记托管状态
  -> bot controller 为该座位生成动作
  -> 动作进入 room actor 命令队列
  -> landlord rule engine 校验动作
  -> 广播托管动作事件
```

要求：

- 托管动作也必须走正常规则校验。
- 托管不能直接改 `GameState`。
- 玩家恢复在线后，可按规则取消托管。
- 托管动作需要保留原因：超时、断线、主动托管。

## Actor 集成

每个房间由一个 room actor 串行处理消息。机器人动作作为普通命令进入该 actor 的 mailbox。

推荐消息：

```text
BotSeatRequested
BotJoined
BotThinkRequested
BotCommandReady
BotCommandRejected
BotTakeoverStarted
BotTakeoverStopped
```

处理原则：

- room actor 只调度机器人，不在 actor 内执行耗时策略。
- 简单策略可同步返回，但必须设置耗时上限。
- 复杂策略应异步计算，完成后投递 `BotCommandReady`。
- actor mailbox 必须有长度监控，避免机器人消息堆积。

## 策略接口

机器人策略接口建议保持纯函数化，便于测试和替换。

```go
type BotPolicy interface {
    Decide(ctx context.Context, input DecisionInput) (Decision, error)
}
```

输入：

```text
game_code
room_level
seat_id
game_state_snapshot
legal_actions
bot_profile
timeout_ms
random_seed
```

输出：

```text
command_type
cards
confidence
think_duration_ms
reason
```

原则：

- 策略只从 `legal_actions` 中选择动作。
- 策略不直接构造非法牌型。
- 策略不能访问房间可变状态指针，只能使用快照。
- 随机数必须可注入 seed，方便回放和测试。

## 策略分层

第一版策略分三层。

### 托管策略

用于真人超时或断线后的托管。

目标：

- 尽量不破坏玩家体验。
- 优先选择最小合法动作。
- 必要时自动过牌。

示例：

```text
如果可以过牌 -> 过牌
如果必须出牌 -> 出最小单牌或最小可压制牌
如果只剩可一手出完 -> 出完
```

### 补位机器人策略

用于匹配补位机器人。

目标：

- 行为接近低到中等水平真人。
- 不明显作弊。
- 出牌有轻微随机性。

策略：

- 按牌力评分决定叫地主/抢地主倾向。
- 出牌优先保留炸弹和关键组合。
- 允许按配置调整激进程度。

### 调试策略

用于测试和压测。

目标：

- 行为确定。
- 可重复回放。
- 便于覆盖边界场景。

策略：

- 固定出最小合法牌。
- 固定过牌。
- 固定抢地主或不抢地主。

## 机器人画像

机器人不需要完整用户账号体系，但需要可识别、可配置、可审计。

建议字段：

```text
bot_id
nickname
avatar_id
skill_level
aggressiveness
think_time_min_ms
think_time_max_ms
mistake_rate
enabled_room_levels
```

画像来源：

- 第一版从配置读取。
- 后续可从 MySQL 配置表读取。
- 热点配置可放 Memcached 快照。

## 配置项

建议按玩法和场次配置。

```text
bot.enabled
bot.fill_timeout_ms
bot.max_bots_per_room
bot.allowed_room_levels
bot.skill_level
bot.think_time_min_ms
bot.think_time_max_ms
bot.mistake_rate
bot.takeover_enabled
bot.takeover_timeout_ms
```

约束：

- 配置以 MySQL 为权威来源。
- Memcached 只缓存配置快照。
- 配置变更通过 Control API 和 `gamectl config set` 发布。
- 灰度策略可按 room_level、渠道、版本或用户段位启用。

## 数据记录

机器人相关事件必须可追溯。

MySQL 关键事件：

```text
bot_joined
bot_takeover_started
bot_takeover_stopped
bot_command_selected
bot_command_rejected
bot_settlement_result
```

ClickHouse 分析事件：

```text
bot_fill_rate
bot_win_rate
bot_avg_think_time
bot_command_distribution
bot_takeover_rate
bot_user_retention_after_match
```

注意：

- 结算事实仍由 wallet-service 和 MySQL 资产流水确定。
- ClickHouse 只做分析，不作为补偿依据的唯一来源。

## 风控与体验

机器人必须避免破坏公平性和用户体验。

要求：

- 机器人不能读取规则外不可见信息。
- 机器人策略只能基于当前座位可见信息和公开信息决策。
- 机器人胜率需要按场次监控。
- 机器人补位比例需要可配置、可下调、可关闭。
- 不同技能等级机器人应有差异，不要所有机器人行为一致。
- 用户举报或异常牌局需要能从事件日志回放。

## 监控指标

必须监控：

- 机器人补位率。
- 机器人参与房间数。
- 每局机器人数量分布。
- 机器人胜率。
- 机器人平均思考时间。
- 机器人动作失败率。
- 托管触发率。
- 托管后玩家回连率。
- bot actor / policy 决策耗时。
- bot 消息队列积压。

告警：

- 机器人胜率异常升高。
- 机器人补位率异常升高。
- 机器人动作失败率升高。
- 机器人决策耗时 P99 升高。
- 托管触发率异常升高。

## 测试策略

单元测试：

- 托管策略选择最小合法动作。
- 无合法出牌时选择过牌。
- 叫地主策略按牌力评分变化。
- 随机 seed 固定时输出确定。
- 策略只从 legal_actions 中选择动作。

集成测试：

- 真人不足时机器人补位。
- 玩家超时后托管出牌。
- 玩家重连后取消托管。
- 机器人参与完整一局并完成结算。
- 机器人动作被规则引擎拒绝时不破坏房间状态。

压测：

- 大量机器人补位创建房间。
- 大量托管动作并发。
- bot policy 决策耗时。
- room actor mailbox 长度。

## 演进路线

第一阶段：规则策略。

- 最小合法动作。
- 简单牌力评分。
- 基础思考时间随机。
- 配置驱动补位和托管。

第二阶段：画像策略。

- 不同 skill_level。
- 不同 aggressiveness。
- mistake_rate。
- 按场次和段位配置策略。

第三阶段：策略服务化。

- 当策略复杂或 CPU 消耗较高时，拆出 `bot-service`。
- gameplay-service 仍负责房间 owner 和规则校验。
- bot-service 只返回建议动作。

第四阶段：智能策略。

- 基于历史牌局训练策略。
- 引入策略版本和 AB 实验。
- 按灰度逐步放量。

## 关键结论

- 第一版机器人留在 gameplay-service 内部，不单独拆服务。
- 房间 actor 是机器人动作的唯一入口。
- 策略只做建议，规则引擎做最终合法性判断。
- MySQL 记录关键事实，ClickHouse 做分析。
- 配置、AB 和灰度走 Control API。
- 后续只有当策略复杂度或资源消耗上来后，才拆独立 bot-service。
