# 文章 API 文档

文章按"元数据（MySQL `article_meta`）+ 正文（MongoDB `article`）"跨库存储，本接口负责跨库拉取并组装详情。

## 路由分组

| 接口 | 路由 | 是否需要登录 |
| --- | --- | --- |
| 获取文章详情 | `GET /api/v1/articles/:id` | 否（未登录仅可见前 60%/前 2 章） |
| 获取文章正文分段 | `GET /api/v1/articles/:id/segments/:seq` | 否（未登录仅可拉前 60%/前 2 章范围） |

## 通用约定

- 文章 ID 为 MySQL 自增 id（对外暴露），正文 MongoDB `_id` 为内部关联键，不对外返回。
- 小说按章节存储：根文章 `type=2, novel_id=0`，每章一条独立文章记录（`novel_id` 指向根文章、`chapter_id` 从 1 起）。接口直接以章节文章 id 访问即可获取该章。
- 正文截断按字符（rune）计算，不会截出半个多字节字符。
- 分段长度上限：`consts.ArticleSegmentSize`（100 千字）。

### 正文可见性规则（未登录）

- 普通文章/小说根文章（`chapter_id=0`）：仅返回正文前 60%，`full=false`，body `code` 为 `20001`。
- 小说章节：前 2 章（`chapter_id <= 2`）返回完整正文（`full=true`，无 20001）；第 3 章起正文为空，仅返回元数据，body `code` 为 `20001`。
- 登录用户：始终返回完整正文。

### 正文分段规则

- 交付正文（截断后的）长度超过一个分段（100 千字）时，详情接口不返回完整正文，body `code` 为 `20006`，`article.segments` 为分段总数，`article.content` 仅为首段；客户端应改用分段接口逐段拉取。
- 分段接口 `seq` 从 1 起，逐段拉取；`seq` 超出交付范围（含未登录第 3 章起的不可见章节）返回 `400`。
- 详情与分段接口共享同一截断/分段口径，保证按 `segments` 拼接后与详情展示一致。

## 获取文章详情

### `GET /api/v1/articles/:id`

获取文章详情：元数据 + 正文 + 作者信息 + 互动计数 + 当前登录用户的点赞/关注状态。**无需登录**，未登录按上述 60%/前 2 章规则截断正文。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 否 | 已登录 session 的 `mwu_sess_id`；未登录时正文受限 |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | number | 是 | 文章 ID，大于 0 |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/articles/1 \
  -b cookie.txt
```

#### 成功响应（登录用户，正文未超长）

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "article": {
    "id": 1,
    "user_id": 10001,
    "title": "标题",
    "description": "摘要",
    "cover": "",
    "type": 0,
    "novel_id": 0,
    "chapter_id": 0,
    "content": "全文……",
    "status": 0,
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
      "shares": 0,
      "views": 100
    },
    "is_liked": false,
    "is_following": false,
    "full": true,
    "segments": 0
  }
}
```

#### 成功响应（未登录，正文被截断）

状态码：`200 OK`，body `code` 为 `20001`（登录后查看完整内容），`article.full` 为 `false`，`article.content` 为前 60% 正文（小说第 3 章起为空字符串）。

```json
{
  "code": 20001,
  "msg": "登录后查看完整内容",
  "article": {
    "id": 1,
    "title": "标题",
    "content": "前60%……",
    "full": false,
    "segments": 0
  }
}
```

#### 成功响应（正文超长，需分段拉取）

状态码：`200 OK`，body `code` 为 `20006`，`article.segments` 为分段总数（>1），`article.content` 仅为首段，客户端应调用分段接口逐段拉取。

```json
{
  "code": 20006,
  "msg": "文章过长，请使用分段接口获取正文",
  "article": {
    "id": 1,
    "title": "超长标题",
    "content": "首段……",
    "full": true,
    "segments": 12
  }
}
```

#### 字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `article` | object | 文章详情 |
| `article.author` | object | 作者公开资料；作者已注销时 `username`/`bio` 为固定打码文案 |
| `article.stats` | object | 互动计数：`likes` 点赞、`comments` 评论、`shares` 分享、`views` 浏览 |
| `article.is_liked` | boolean | 当前登录用户是否已点赞（未登录恒为 `false`） |
| `article.is_following` | boolean | 当前登录用户是否已关注作者（未登录恒为 `false`） |
| `article.full` | boolean | `content` 是否为完整正文（未登录截断时为 `false`） |
| `article.segments` | number | 交付正文的分段总数；0 表示无需分段；>0 表示需用分段接口 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `id` 非法（非数字或为 0） |
| `403` | 登录用户与文章作者存在任意一方的拉黑关系，body 中 `code` 为 `40301`（见 `API-errors.md`） |
| `404` | 文章不存在、已删除/草稿/限制传播等非正常状态 |
| `451` | 文章被屏蔽（`status=4`），body 中 `code` 为 `45101`（见 `API-errors.md`） |
| `500` | 数据库查询异常 |

## 获取文章正文分段

### `GET /api/v1/articles/:id/segments/:seq`

获取文章正文指定分段（`seq` 从 1 起）。与详情接口共用屏蔽/拉黑校验与截断口径；未登录仅可拉取正文前 60%（小说前 2 章）范围内的分段。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | number | 是 | 文章 ID，大于 0 |
| `seq` | number | 是 | 分段序号，从 1 起 |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/articles/1/segments/2 \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "segment": {
    "id": 1,
    "seq": 2,
    "content": "第二段……",
    "full": true
  }
}
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `segment.id` | number | 文章 ID |
| `segment.seq` | number | 分段序号 |
| `segment.content` | string | 该段正文 |
| `segment.full` | boolean | 当前身份下交付的正文是否为完整正文（未登录截断时为 `false`） |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `id`/`seq` 非法，或 `seq` 超出交付正文的分段范围（含未登录访问小说第 3 章起的不可见章节） |
| `403` | 登录用户与文章作者存在任意一方的拉黑关系，body 中 `code` 为 `40301` |
| `404` | 文章不存在、已删除/草稿等非正常状态 |
| `451` | 文章被屏蔽（`status=4`），body 中 `code` 为 `45101` |
| `500` | 数据库查询异常 |
