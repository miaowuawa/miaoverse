# Feed API 文档

Feed（信息流）接口：按类型拉取动态与文章的混合信息流。

## 路由分组

| 接口 | 路由 | 是否需要登录 |
| --- | --- | --- |
| 拉取时间线 | `GET /api/v1/feeds/timeline` | 否 |
| 拉取关注流 | `GET /api/v1/feeds/following` | 是（未登录返回 `40101`） |
| 拉取用户内容流 | `GET /api/v1/feeds/user/:uid` | 是 |

## 通用约定

- 基础地址：`http://{host}:{port}`
- API 前缀：`/api/v1`
- 响应格式：JSON
- 请求 ID：服务会在响应头中写入 `X-Miaoverse-ReqID`
- Session Cookie：`mwu_sess_id`
- 分页：`offset`/`limit`，`limit` 默认 `20`，最大 `50`（feed 条目含作者信息，比普通列表更重）

### 内容类型过滤

| `content` 取值 | 说明 |
| --- | --- |
| `all`（默认） | 动态 + 文章混合 |
| `moment` | 只拉取动态 |
| `article` | 只拉取文章（不含小说） |
| `novel` | 只拉取小说（`type=novel` 的根文章，含章节数；**仅用户内容流支持**，其余类型传该值返回 `400`） |

### 排序

| `sort` 取值 | 说明 |
| --- | --- |
| `time`（默认） | 按发布时间倒序 |
| `hot` | 按点赞量倒序（同点赞按发布时间倒序）；**仅用户内容流支持**，其余类型传该值返回 `400` |

### 内容可见性规则

- **屏蔽/删除内容不显示**：动态/文章被标记为屏蔽（`*StatusBlocked`）、删除（`*StatusDeleted`）、草稿等非正常状态时，一律不进入 feed。
- **拉黑/屏蔽/不想看关系过滤**：查看者与作者存在任意一方拉黑关系，或查看者屏蔽/不想看该作者时，该作者的内容不进入 feed。
- **动态可见权限**：
  - 时间线：仅公开（`permission=0`）动态进入 feed。
  - 关注流：公开动态全部可见；仅好友（`permission=1`）/仅粉丝（`permission=3`）动态需作者回关查看者才可见；仅自己（`permission=2`）动态不进入 feed。
- **文章**：仅非章节文章（`novel_id=0`）进入 feed；小说章节通过文章详情接口访问。

## 拉取时间线

### `GET /api/v1/feeds/timeline`

全站公开内容（动态 + 文章）按发布时间倒序。**无需登录**。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `content` | string | 否 | 内容类型过滤：`all`/`moment`/`article`，默认 `all`（`novel` 与 `sort=hot` 仅用户内容流支持） |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页条数，默认 `20`，最大 `50` |

#### 请求示例

```bash
curl -i "http://localhost:3000/api/v1/feeds/timeline?content=all&offset=0&limit=20"
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "count": 2,
  "items": [
    {
      "id": 2,
      "type": "moment",
      "user_id": 10001,
      "title": "",
      "content": "今天天气真好",
      "description": "",
      "cover": "",
      "images": ["15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"],
      "article_type": 0,
      "novel_id": 0,
      "chapter_id": 0,
      "status": 0,
      "permission": 0,
      "created_at": "2026-08-12T12:00:00+08:00",
      "updated_at": "2026-08-12T12:00:00+08:00",
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
      "is_following": false,
      "full": true
    },
    {
      "id": 1,
      "type": "article",
      "user_id": 10002,
      "title": "文章标题",
      "content": "",
      "description": "摘要",
      "cover": "cover.jpg",
      "article_type": 0,
      "novel_id": 0,
      "chapter_id": 0,
      "status": 0,
      "permission": 0,
      "created_at": "2026-08-12T11:00:00+08:00",
      "updated_at": "2026-08-12T11:00:00+08:00",
      "author": {
        "id": 10002,
        "username": "anon",
        "nickname": "爱音",
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
      "is_following": false,
      "full": false
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `count` | number | 当前过滤条件下的内容总数（用于分页） |
| `items` | array | feed 条目列表 |
| `items[].type` | string | 内容类型：`moment` 动态 / `article` 文章 |
| `items[].content` | string | 动态为正文；文章为空（feed 不触达 MongoDB 正文，仅展示元数据与预览） |
| `items[].images` | string[] | 动态图片文件 UUID 列表（仅动态有值，文章为空数组）；原始存储 URL 不下发，需通过 `GET /api/v1/user/files/:uuid/shared-link` 换取临时访问链接 |
| `items[].full` | boolean | 动态恒为 `true`；文章恒为 `false`（正文需通过详情接口获取） |
| `items[].author` | object | 作者公开资料；作者已注销时 `username`/`bio` 为固定打码文案 |
| `items[].stats` | object | 互动计数：`likes` 点赞数、`comments` 评论数、`shares` 分享数 |
| `items[].is_liked` | boolean | 当前登录用户是否已点赞该内容（未登录恒为 `false`） |
| `items[].is_following` | boolean | 当前登录用户是否已关注作者（未登录恒为 `false`） |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `content` 取值非法、`offset`/`limit` 取值非法 |
| `500` | 数据库查询异常 |

## 拉取关注流

### `GET /api/v1/feeds/following`

当前登录用户关注用户的内容（动态 + 文章）按发布时间倒序。**需要登录**。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `content` | string | 否 | 内容类型过滤：`all`/`moment`/`article`，默认 `all`（`novel` 与 `sort=hot` 仅用户内容流支持） |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页条数，默认 `20`，最大 `50` |

#### 请求示例

```bash
curl -i "http://localhost:3000/api/v1/feeds/following?content=all&offset=0&limit=20" \
  -b cookie.txt
```

#### 成功响应

与时间线一致（`200 OK`，`items` 为关注用户的内容）。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `content` 取值非法、`offset`/`limit` 取值非法 |
| `401` | 未登录，body 中 `code` 为 `40101`（见 `API-errors.md`） |
| `500` | 数据库查询异常 |

## 拉取用户内容流

### `GET /api/v1/feeds/user/:uid`

拉取指定用户发布的全部内容（动态 + 文章 + 小说元数据，按 `content` 过滤），按 `sort` 排序。**需要登录**。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `uid` | number | 是 | 目标用户 ID（路径参数），大于 0；可为当前登录用户自己 |
| `content` | string | 否 | 内容类型过滤：`all`/`moment`/`article`/`novel`，默认 `all` |
| `sort` | string | 否 | 排序：`time` 按发布时间倒序（默认）、`hot` 按点赞量倒序（同点赞按发布时间倒序） |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页条数，默认 `20`，最大 `50` |

#### 请求示例

```bash
curl -i "http://localhost:3000/api/v1/feeds/user/20002?content=moment&sort=hot&offset=0&limit=20" \
  -b cookie.txt

curl -i "http://localhost:3000/api/v1/feeds/user/20002?content=novel&sort=time&offset=0&limit=20" \
  -b cookie.txt
```

#### 成功响应

与时间线一致（`200 OK`），额外说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `items[].type` | string | 内容类型：`moment` 动态 / `article` 文章（`content=novel` 时指小说根文章） |
| `items[].chapter_count` | number | 小说根文章（`article_type=2` 且 `novel_id=0`）的已发布章节数，其余为 `0` |
| `items[].stats.likes` | number | 点赞数（`sort=hot` 时按此降序） |

#### 可见性规则

- 动态可见性与内容列表一致：公开全部可见；仅好友/仅粉丝需作者回关查看者；仅自己（`permission=2`）仅本人可见。
- 文章/小说仅返回 `status=normal` 的非章节记录（`novel_id=0`）；小说章节通过文章详情/分段接口获取。
- 拉黑/被拉黑/屏蔽/不想看关系存在时返回 `40301`；目标用户账号被封禁（不允许登录）返回 `40303`。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `content`/`sort`/`offset`/`limit` 取值非法、`uid` 非法 |
| `401` | 未登录，body 中 `code` 为 `40101`（见 `API-errors.md`） |
| `403` | 与目标用户存在拉黑/被拉黑关系（`code` 为 `40301`）；目标用户账号被封禁（`code` 为 `40303`） |
| `404` | 目标用户不存在 |
| `500` | 数据库查询异常 |
