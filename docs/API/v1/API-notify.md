# API 通知（Notifications）

用户通知覆盖互动与账号安全两类事件：**被点赞、被回复/被评论、被关注、修改密码、登录**。每次触发写入 `notify` 表并实时推送给接收者。

## 功能规则

- **触发时机**：
  - 被点赞：动态被点赞、评论被点赞（自己给自己点赞不通知）；
  - 被回复/被评论：动态被评论（通知动态作者）、评论被回复（通知被回复的评论作者；自己回复自己不通知）；
  - 被关注：被其他用户关注（自己关注自己不通知）；
  - 账号安全：登录成功（含注册登录、选择账号登录、切换账号）、修改密码成功，通知账号本人。
- **通知分类**（前端分栏展示）：`account` 账号、`like` 点赞、`reply` 回复、`follow` 关注、`mention` 提及（预留）。一个分类可对应多个通知类型：`account` 同时覆盖「账号安全」与「事务事项」。
- **关联对象**（`target_type` / `target_id`）：`0` 用户（被关注通知指向关注者）、`1` 动态（被点赞/被评论的动态）、`2` 评论（被点赞/被回复的评论）。前端据此跳转：用户 → `/user/:id`，动态 → `/moment/:id`。
- **触发者**（`actor`）：互动类通知携带触发者用户信息（已注销账号展示字段统一打码）；账号安全类通知无触发者。
- **已读状态**：`0` 未读、`1` 已读、`2` 已删除（软删除，删除后不出现在列表）。标记已读/删除均为幂等操作。
- **实时推送**：所有通知写库后经 SSE（`GET /api/v1/notify/stream`）实时下发，携带写入后的最新未读数；离线期间的通知进入列表接口，登录后拉取即可。
- **在线状态**：基于 SSE 连接心跳判定——服务端每 15 秒下发一帧心跳（`event: ping`），**心跳正常 = 在线**；心跳/事件写出失败（客户端断开）立即断开连接，**超过 30 秒无成功写出即判定离线**（多标签页任一连接心跳正常即在线）。离线用户保留最近活跃时间（`last_seen`），供展示「x 分钟前在线」。客户端同理：超过 45 秒收不到心跳即判定连接离线并自动重建连接。
- **无状态化多实例部署**：在线状态与通知事件分发均存 Redis（cache DB，见 `docs/配置文件指南.md`）——心跳活跃时间写入 Redis ZSET（`notify:presence`），任何实例查询结果一致；通知写库后经 Redis pub/sub（频道 `notify:bus`）广播给全部实例，**无论用户连到哪个实例都能实时收到通知**（产生通知的实例直接投递本地连接，其他实例经总线投递，本实例发布的消息不重复投递）。通知数据始终以 MySQL 为准，Redis 状态可随时清空重建。

## 接口一览

| 接口 | 方法 | 路径 | 需登录 |
| --- | --- | --- | --- |
| 通知列表 | `GET` | `/api/v1/notify` | 是 |
| 各分类未读数 | `GET` | `/api/v1/notify/unread-count` | 是 |
| 在线状态批量查询 | `GET` | `/api/v1/notify/presence` | 是 |
| 标记单条已读 | `PATCH` | `/api/v1/notify/:id/read` | 是 |
| 全部标记已读 | `PATCH` | `/api/v1/notify/read-all` | 是 |
| 删除通知 | `DELETE` | `/api/v1/notify/:id` | 是 |
| SSE 实时推送流 | `GET` | `/api/v1/notify/stream` | 是 |

> 以上接口均要求已登录（session cookie `mwu_sess_id`）；未登录返回 `401`（"请先登录"）。

## 通知列表

### `GET /api/v1/notify`

分页获取当前用户的通知（id 倒序，不含已删除）。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `category` | string | 否 | 分类过滤：`account` / `like` / `follow` / `mention` / `reply`；缺省为全部 |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页数量，默认 `20`，最大 `100` |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "通知获取成功",
  "count": 12,
  "notifies": [
    {
      "id": 1001,
      "type": 2,
      "category": "like",
      "actor": {
        "id": 8,
        "username": "alice",
        "nickname": "爱丽丝",
        "avatar": "0f9c2b31-6a4d-4c39-9c0a-1b2f3d4e5f60",
        "status": 1
      },
      "target_type": 1,
      "target_id": 66,
      "content": "今天拍到的猫猫",
      "created_at": "2026-06-07 12:00:00",
      "read_at": null,
      "read": false
    },
    {
      "id": 990,
      "type": 0,
      "category": "account",
      "target_type": 0,
      "target_id": 0,
      "content": "你的账号刚刚完成了登录。如果这不是你本人的操作，请及时修改密码。",
      "created_at": "2026-06-07 09:30:00",
      "read_at": "2026-06-07 09:31:00",
      "read": true
    }
  ]
}
```

#### 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `type` | number | 通知类型：`0` 账号安全、`1` 事务、`2` 赞、`3` 关注、`4` 提及、`5` 回复与评论 |
| `category` | string | 前端分类（见「功能规则」） |
| `actor` | object | 触发者用户（账号安全类通知为 null），已注销账号展示字段打码 |
| `content` | string | 摘要文本：互动类为评论/回复内容或动态内容摘要（最长 200 字符），账号类为系统文案 |
| `read` | boolean | 是否已读 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `category` 非法、`offset`/`limit` 非法 |
| `401` | 未登录 |

## 各分类未读数

### `GET /api/v1/notify/unread-count`

查询当前用户各分类未读数，供导航小红点与分类页签角标展示。

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "unread": {
    "total": 7,
    "account": 1,
    "like": 3,
    "follow": 2,
    "mention": 0,
    "reply": 1
  }
}
```

## 在线状态批量查询

### `GET /api/v1/notify/presence`

批量查询用户在线状态。**在线判定基于 SSE 连接心跳**：该用户最近 30 秒（`2 × 15s` 心跳间隔）内有成功写出的心跳/事件 → 在线；否则离线（`last_seen` 保留最近活跃时间）。多标签页多连接时任一连接心跳正常即在线。

心跳活跃时间写入 Redis（ZSET `notify:presence`，cache DB，见 `docs/配置文件指南.md`），**跨实例共享**：用户无论连到哪个实例，任何实例查询结果一致。超过 7 天（`NotifyPresenceRetention`）未活跃的记录被自动清理，不影响在线判定。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `uids` | string | 是 | 逗号分隔的用户 id 列表，去重后 1~100 个，如 `uids=1,2,3` |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "presence": [
    { "uid": 1, "online": true, "last_seen": "2026-06-07 12:00:00" },
    { "uid": 2, "online": false, "last_seen": "2026-06-07 09:30:00" },
    { "uid": 3, "online": false, "last_seen": "0001-01-01T00:00:00Z" }
  ]
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `online` | boolean | 是否在线（有活跃 SSE 连接且心跳正常） |
| `last_seen` | string | 最近活跃时间；从未连接过为 Go 零值时间（`0001-01-01T00:00:00Z`），前端按「无记录」处理 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `uids` 缺失、为空、含非法 id、去重后超过 100 个 |
| `401` | 未登录 |

## 标记已读 / 删除

### `PATCH /api/v1/notify/:id/read`

标记单条通知已读（幂等：仅未读时生效，同时记录 `read_at`）。

### `PATCH /api/v1/notify/read-all`

将当前用户全部未读通知标记为已读（幂等）。

### `DELETE /api/v1/notify/:id`

删除单条通知（软删除，幂等）：删除后不出现在列表与未读统计中。

以上三个接口成功响应均返回最新未读数（与未读数接口结构一致），供前端同步小红点：

```json
{
  "code": 200,
  "msg": "已标记为已读",
  "unread": { "total": 6, "account": 1, "like": 3, "follow": 2, "mention": 0, "reply": 0 }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `:id` 非法 |
| `401` | 未登录 |

## SSE 实时推送流

### `GET /api/v1/notify/stream`

建立 SSE（Server-Sent Events）长连接，实时接收新通知。**登录态经 session cookie 校验**（EventSource 同源请求自动携带），无需额外 ticket。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Accept` | 否 | 建议 `text/event-stream` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 成功响应

状态码：`200 OK`，`Content-Type: text/event-stream; charset=utf-8`（`Cache-Control: no-cache`，禁用反向代理缓冲）。

连接建立后：

1. 首帧下发重连间隔提示与确认注释：

```
retry: 5000

: connected

```

2. 有新通知时推送 `notification` 事件（`data` 为 JSON）：

```
event: notification
data: {"notify":{...},"unread":{"total":7,"account":1,"like":3,"follow":2,"mention":0,"reply":1}}

```

其中 `notify` 结构与通知列表接口的 `notifies` 条目一致，`unread` 为该通知写入后的最新未读数（前端收到后可直接刷新小红点）。

3. 每 15 秒下发一帧心跳（`X-Accel-Buffering: no`，反向代理不要缓冲）：

```
event: ping
data: 1765099200000

```

心跳是**双向在线判定依据**：

- **服务端**：心跳/事件写出成功即把该用户的活跃时间写入 Redis（`presence` 接口查询，跨实例一致）；写出失败立即断开连接，超过 30 秒无成功写出判定离线；
- **客户端**：收到心跳即认为连接在线；超过 45 秒收不到心跳（半开连接等 EventSource 不会主动报错的场景）判定连接离线，主动断开并重建连接。

4. 连接断开后浏览器 `EventSource` 按 `retry` 提示（5 秒）自动重连；重连同样经 session 校验，登录态失效返回 `401`（前端此时应引导登录）。

> **跨实例实时推送**：无状态化部署下产生通知的请求与用户的 SSE 连接可能落在不同实例，通知写库后经 Redis pub/sub（频道 `notify:bus`）广播给全部实例，各实例向本实例上的连接投递；产生通知的实例直接投递本地连接，总线消息由其他实例消费（本实例发布的消息跳过，不重复推送）。

> 推送为尽力而为：每连接 16 条事件缓冲，写满时该事件跳过（通知已入库，前端可通过列表/未读数接口补齐）。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `401` | 未登录或登录态失效 |
| `403` | 账号被封禁（`code` 为 `40303`）等账号状态异常 |

## 数据表

通知落库于 `notify` 表（见 `scripts/db.sql`）：`user_id` 接收者、`type` 类型、`actor_id` 触发者（账号安全类为 0）、`target_type`/`target_id` 关联对象、`content` 摘要、`status` 状态（0 未读 / 1 已读 / 2 已删除）、`received` SSE 送达标记、`read_at` 已读时间。
