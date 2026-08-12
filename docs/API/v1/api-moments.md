# 动态 API 文档

动态相关接口（发布、详情、评论/楼中楼、点赞）拆自 `API-user.md`，统一放在本文件。

## 路由分组

| 接口 | 路由 | 是否需要登录 |
| --- | --- | --- |
| 发布动态 | `POST /api/v1/moment` | 是 |
| 编辑动态 | `PATCH /api/v1/moment/:id` | 是 |
| 获取动态详情 | `GET /api/v1/moments/:id` | 否（未登录仅可查看公开动态） |
| 评论动态 | `POST /api/v1/comment/moments` | 是 |
| 回复评论（楼中楼） | `POST /api/v1/comment/moments/:id/replies` | 是 |
| 获取楼中楼完整对话 | `GET /api/v1/comment/moments/:id/conversation` | 是 |
| 给动态点赞 | `POST /api/v1/moment/likes` | 是 |

## 通用约定

- 基础地址：`http://{host}:{port}`
- API 前缀：`/api/v1`
- 请求体格式：`Content-Type: application/json`
- 响应格式：JSON
- 请求 ID：服务会在响应头中写入 `X-Miaoverse-ReqID`
- Session Cookie：`mwu_sess_id`
- 除获取动态详情外，其余接口均要求登录（`mwu_sess_id` cookie）且账号状态正常；被权限封禁的账号按各接口说明返回 `40302`。

## 发布动态

### `POST /api/v1/moment`

发布当前登录用户的动态。支持设置发布状态（正常/草稿）、可见权限（公开/仅好友/仅自己/仅粉丝）、评论权限（全部可评论/仅好友可评论/仅粉丝可评论/全部不可评论）以及个人置顶。全站置顶（`top=100`）不允许普通用户设置，服务端会直接拒绝。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "content": "今天天气真好",
  "status": 0,
  "permission": 0,
  "comment_permission": 0,
  "top": 0
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `content` | string | 是 | 非空，最长 5000 字符 | 动态正文 |
| `status` | number | 否 | `0` 正常，`2` 草稿 | 发布状态，默认 `0` |
| `permission` | number | 否 | `0` 公开，`1` 仅好友，`2` 仅自己，`3` 仅粉丝 | 可见权限，默认 `0` |
| `comment_permission` | number | 否 | `0` 全部可评论，`1` 仅好友（互相关注），`2` 仅粉丝，`3` 全部不可评论 | 评论权限，默认 `0` |
| `top` | number | 否 | `0` 不置顶，`1` 个人置顶 | 置顶状态，默认 `0`；`100` 全站置顶不允许设置 |

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/moment \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{"content":"今天天气真好","status":0,"permission":0,"comment_permission":0,"top":0}'
```

#### 成功响应

状态码：`201 Created`

```json
{
  "code": 201,
  "msg": "发布成功",
  "moment": {
    "id": 1,
    "user_id": 10001,
    "title": "",
    "content": "今天天气真好",
    "status": 0,
    "permission": 0,
    "comment_permission": 0,
    "top": 0,
    "created_at": "2026-06-07T12:00:00+08:00",
    "updated_at": "2026-06-07T12:00:00+08:00"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、`content` 为空或超长、`status`/`permission`/`comment_permission`/`top` 取值非法（含设置全站置顶） |
| `401` | 未登录或 session 中没有 `UID` |
| `500` | 数据库写入异常 |

## 编辑动态

### `PATCH /api/v1/moment/:id`

编辑当前登录用户自己的动态，支持部分更新：请求体中出现的字段才会被修改，未出现的字段保持不变。可修改内容、发布状态、可见权限、评论权限以及个人置顶。全站置顶（`top=100`）不允许普通用户设置，服务端会直接拒绝。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | number | 是 | 动态 ID，大于 0 |

#### 请求体

```json
{
  "content": "修改后的内容",
  "permission": 1,
  "comment_permission": 2,
  "top": 1
}
```

字段说明（均为可选，出现才修改）：

| 字段 | 类型 | 校验规则 | 说明 |
| --- | --- | --- | --- |
| `title` | string | 最长 255 字符 | 动态标题 |
| `content` | string | 非空，最长 5000 字符 | 动态正文 |
| `status` | number | `0` 正常，`2` 草稿 | 发布状态 |
| `permission` | number | `0` 公开，`1` 仅好友，`2` 仅自己，`3` 仅粉丝 | 可见权限 |
| `comment_permission` | number | `0` 全部可评论，`1` 仅好友（互相关注），`2` 仅粉丝，`3` 全部不可评论 | 评论权限 |
| `top` | number | `0` 不置顶，`1` 个人置顶 | 置顶状态；`100` 全站置顶不允许设置 |

#### 请求示例

```bash
curl -i -X PATCH http://localhost:3000/api/v1/moment/1 \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{"content":"修改后的内容","permission":1,"comment_permission":2,"top":1}'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "修改成功",
  "moment": {
    "id": 1,
    "user_id": 10001,
    "title": "",
    "content": "修改后的内容",
    "status": 0,
    "permission": 1,
    "comment_permission": 2,
    "top": 1,
    "created_at": "2026-06-07T12:00:00+08:00",
    "updated_at": "2026-06-07T12:05:00+08:00"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、请求体为空（无任何可修改字段）、字段取值非法（含设置全站置顶）、`:id` 非法 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 发布权限封禁（`code` 为 `40302`）；与动态作者存在拉黑/被拉黑关系（`code` 为 `40301`） |
| `404` | 动态不存在、已删除/草稿等非正常状态，或当前用户不是动态作者 |
| `451` | 动态被屏蔽（`status=4`），body 中 `code` 为 `45101` |
| `500` | 数据库异常 |

## 获取动态详情

### `GET /api/v1/moments/:id`

获取指定动态的详情，包含动态本体、作者信息、互动计数以及当前登录用户对该动态的点赞/关注状态。**无需登录**：未登录用户仅可查看公开（`permission=0`）动态，`is_liked`/`is_following` 恒为 `false`；登录用户按可见性规则查看。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 否 | 已登录 session 的 `mwu_sess_id`；未登录时仅可查看公开动态 |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | number | 是 | 动态 ID，大于 0 |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/moments/1 \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "moment": {
    "id": 1,
    "user_id": 10001,
    "title": "",
    "content": "今天天气真好",
    "status": 0,
    "permission": 0,
    "comment_permission": 0,
    "top": 0,
    "created_at": "2026-06-07T12:00:00+08:00",
    "updated_at": "2026-06-07T12:00:00+08:00",
    "author": {
      "id": 10001,
      "username": "miao",
      "nickname": "喵",
      "region": 86,
      "avatar": "",
      "bio": "",
      "gender": 0,
      "status": 1,
      "created_at": "2026-06-01T12:00:00+08:00",
      "updated_at": "2026-06-01T12:00:00+08:00"
    },
    "stats": {
      "likes": 5,
      "comments": 3,
      "shares": 0
    },
    "is_liked": false,
    "is_following": false
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `moment` | object | 动态详情 |
| `moment.author` | object | 作者公开资料；作者已注销时 `username`/`bio` 为固定打码文案 |
| `moment.stats` | object | 互动计数：`likes` 点赞数、`comments` 评论数、`shares` 分享数 |
| `moment.is_liked` | boolean | 当前登录用户是否已点赞该动态 |
| `moment.is_following` | boolean | 当前登录用户是否已关注作者 |

#### 可见性规则

与内容列表一致：

- 公开（`permission=0`）动态对所有人可见（含未登录用户）。
- 仅好友（`permission=1`）动态仅对互相关注的查看者可见。
- 仅粉丝（`permission=3`）动态仅对目标用户关注了的查看者可见。
- 仅自己（`permission=2`）动态仅本人可见。
- 草稿、已删除等非正常状态动态不返回。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `id` 非法（非数字或为 0） |
| `403` | 登录用户与动态作者存在任意一方的拉黑关系，body 中 `code` 为 `40301`（见 `API-errors.md`） |
| `404` | 动态不存在、已删除/草稿等非正常状态，或对当前用户不可见（未登录用户查看非公开动态） |
| `451` | 动态被屏蔽（`status=4`），body 中 `code` 为 `45101`（见 `API-errors.md`） |
| `500` | 数据库查询异常 |

## 评论动态

### `POST /api/v1/comment/moments`

给动态发送评论。会写入 `comment` 表和 `interacts` 表（type=103 对内容评论），并原子自增 `moment_interact_count.comment_count`。被拉黑/拉黑对方、评论权限不允许、或被封禁评论权限时拒绝。

> 评论接口统一挂在 `/comment/` 分组下，并按被评论内容类型划分子路径：评论动态为 `/comment/moments`，后续文章等类型为 `/comment/articles`，不提供通用 `/comment` 根接口。

> 回复与楼中楼对话接口见下文「回复评论（楼中楼）」与「获取楼中楼完整对话」。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "moment_id": 1,
  "content": "写得好"
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `moment_id` | number | 是 | 大于 0 | 目标动态 ID |
| `content` | string | 是 | 非空，最长 1000 字符 | 评论内容 |

#### 成功响应

状态码：`201 Created`

```json
{
  "code": 201,
  "msg": "评论成功",
  "comment": {
    "id": 1,
    "user_id": 10001,
    "moment_id": 1,
    "content": "写得好",
    "status": 0,
    "created_at": "2026-06-07 12:00:00"
  }
}
```

#### 评论权限规则

- 动态评论权限 `0`（全部可评论）：任何登录用户可评论。
- `1`（仅好友）：仅与动态作者互相关注的用户可评论。
- `2`（仅粉丝）：仅动态作者关注了的用户可评论。
- `3`（全部不可评论）：所有人都不能评论。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、`moment_id` 为 0、`content` 为空或超长 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 评论权限封禁（`code` 为 `40302`）；拉黑/被拉黑关系（`code` 为 `40301`）；评论权限不允许（普通 `403`） |
| `404` | 动态不存在或已删除 |
| `451` | 动态被屏蔽（`status=4`），body 中 `code` 为 `45101` |
| `500` | 数据库异常 |

## 回复评论（楼中楼）

### `POST /api/v1/comment/moments/:id/replies`

回复动态下的某条评论（楼中楼）。回复同样写入 `comment` 表（`target_type=2`，`target_id` 为被回复的评论 id），并在 `interacts` 表写入 `type=104 回复评论` 记录（`user_to` 为被回复评论作者，`reference_id` 为楼中楼首条评论 id）。

规则：

- 需要登录且账号状态正常。
- 评论权限封禁（`PermComment`）期间不可回复，返回 `40302`。
- 与动态作者或被回复评论作者存在任意一方拉黑关系时不可回复，返回 `40301`。
- 评论权限按所属动态的 `comment_permission` 校验（与评论动态一致）。
- 楼中楼回复不计入 `moment_interact_count.comment_count`（该计数只统计一级评论），完整回复列表通过对话接口获取。
- 支持回复自己的评论、回复自己动态下的评论。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "content": "同意"
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `content` | string | 是 | 非空，最长 1000 字符 | 回复内容 |

#### 成功响应

状态码：`201 Created`

```json
{
  "code": 201,
  "msg": "回复成功",
  "reply": {
    "id": 2,
    "user_id": 10002,
    "moment_id": 1,
    "reply_to_id": 1,
    "reply_to_user_id": 10001,
    "content": "同意",
    "status": 0,
    "created_at": "2026-06-07 12:01:00"
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `reply_to_id` | number | 被回复的评论 id |
| `reply_to_user_id` | number | 被回复的评论作者 id |
| `moment_id` | number | 所属动态 id |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、`content` 为空或超长、`:id` 非法或评论不存在/已删除 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 评论权限封禁（`code` 为 `40302`）；与动态作者或被回复评论作者存在拉黑/被拉黑关系（`code` 为 `40301`）；评论权限不允许（普通 `403`） |
| `451` | 所属动态或被回复评论被屏蔽（`status=4`），body 中 `code` 为 `45101` |
| `500` | 数据库异常 |

## 获取楼中楼完整对话

### `GET /api/v1/comment/moments/:id/conversation`

传入楼中楼首条评论 id，返回该评论（`root`）及其全部子孙回复（`replies`，扁平列表按时间正序）。`count` 为该楼的全部回复总数。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `offset` | number | 否 | 回复列表偏移量，默认 `0` |
| `limit` | number | 否 | 每页回复数量，默认 `20`，最大 `100` |

#### 请求示例

```bash
curl -i "http://localhost:3000/api/v1/comment/moments/1/conversation?offset=0&limit=20" \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "conversation": {
    "root": {
      "id": 1,
      "user_id": 10001,
      "moment_id": 1,
      "content": "写得好",
      "status": 0,
      "created_at": "2026-06-07 12:00:00"
    },
    "count": 2,
    "replies": [
      {
        "id": 2,
        "user_id": 10002,
        "moment_id": 1,
        "reply_to_id": 1,
        "reply_to_user_id": 10001,
        "content": "同意",
        "status": 0,
        "created_at": "2026-06-07 12:01:00"
      }
    ]
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `:id` 非法或评论不存在/已删除、`offset`/`limit` 取值非法 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 与动态作者或首条评论作者存在拉黑/被拉黑关系，body 中 `code` 为 `40301` |
| `451` | 所属动态或上溯链上的评论被屏蔽（`status=4`），body 中 `code` 为 `45101` |
| `500` | 数据库查询异常 |

## 给动态点赞

### `POST /api/v1/moment/likes`

给动态点赞。被拉黑/拉黑对方不能点赞。点赞会同步自增动态的点赞计数。

> 动态相关接口统一挂在 `/moment/` 分组下。后续文章等类型使用各自的子路径（如 `/article/likes`），不提供通用的 `/likes`。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "moment_id": 1
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `moment_id` | number | 是 | 大于 0 | 要点赞的动态 ID |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "操作成功",
  "target": 1,
  "action": "like"
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、`moment_id` 为 0 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 社交互动权限封禁（`code` 为 `40302`）；拉黑/被拉黑关系（`code` 为 `40301`） |
| `404` | 动态不存在或已删除 |
| `451` | 动态被屏蔽（`status=4`），body 中 `code` 为 `45101` |
| `500` | 数据库异常 |
