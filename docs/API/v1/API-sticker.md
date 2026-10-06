# API 贴纸（Stickers）

贴纸是可内嵌在评论文字中穿插展示的图片。本文档描述贴纸上传、个人贴纸收藏夹、贴纸包与贴纸包收藏接口。

> 评论中使用贴纸的方式见 `api-moments.md`：评论/回复的 `content` 中携带内嵌标记 `[sticker:<uuid>]`，标记随文本在任意位置穿插展示；**一条评论最多使用 25 张贴纸**。

## 功能规则

- **上传贴纸**：登录且已绑定手机号、未被封禁上传权限（`PermUploadFile`）的用户可上传贴纸。单张贴纸最大 10MB（配置 `upload.max_sticker_size_bytes`，见 `docs/配置文件指南.md`）。仅接受 jpg/png/gif/webp 安全栅格图片（按文件头魔数嗅探，拒绝 SVG/HTML/JS 等可携带脚本的格式）。
- **使用贴纸**：登录且已绑定手机号的用户，可使用本人上传的贴纸、已加入个人收藏夹的贴纸、以及已收藏贴纸包内的贴纸。
- **个人贴纸收藏夹**：本人上传的贴纸自动加入收藏夹（`source=1`）；他人贴纸可手动收藏（`source=2`），**收藏夹最多添加 500 张收藏贴纸**。收藏夹内贴纸可置顶（置顶优先展示）。
- **贴纸包**：登录、绑定手机号且有发布评论权限（未被封禁 `PermComment`）的用户可创建贴纸包，**单个贴纸包最多 100 张贴纸**。其他用户可整包收藏/取消收藏；**收藏的贴纸包内容更新自动同步**（包内容实时查询）。
- **贴纸包封禁**：贴纸包有封禁标记 `sticker_pack.banned`（运营处置字段，与内容屏蔽状态一致由管理侧写入，无公开接口）。**贴纸包被封禁后，包内贴纸在所有使用处（评论等）无法显示**，前端在对应评论下灰字提示「部分贴纸未显示」。
- **贴纸图片展示**：所有接口只下发贴纸图片的文件 UUID（`file_uuid`），不下发原始存储 URL；展示前通过 `GET /api/v1/user/files/:uuid/shared-link` 换取临时访问链接。

## 接口一览

| 接口 | 方法 | 路径 | 需登录 |
| --- | --- | --- | --- |
| 上传贴纸 | `POST` | `/api/v1/stickers` | 是（需绑定手机号） |
| 我的贴纸收藏夹 | `GET` | `/api/v1/stickers/collection` | 是 |
| 添加贴纸到收藏夹 | `POST` | `/api/v1/stickers/collection` | 是 |
| 从收藏夹移除贴纸 | `DELETE` | `/api/v1/stickers/collection` | 是 |
| 设置/取消贴纸置顶 | `PATCH` | `/api/v1/stickers/:uuid/top` | 是 |
| 删除本人贴纸 | `DELETE` | `/api/v1/stickers/:uuid` | 是 |
| 贴纸包列表 | `GET` | `/api/v1/stickers/packs` | 是 |
| 创建贴纸包 | `POST` | `/api/v1/stickers/packs` | 是（需绑定手机号 + 评论权限） |
| 收藏贴纸包 | `POST` | `/api/v1/stickers/packs/favorites` | 是 |
| 取消收藏贴纸包 | `DELETE` | `/api/v1/stickers/packs/favorites` | 是 |
| 贴纸包详情 | `GET` | `/api/v1/stickers/packs/:id` | 是 |
| 添加贴纸到贴纸包 | `POST` | `/api/v1/stickers/packs/:id/stickers` | 是（包主） |
| 移出贴纸包 | `DELETE` | `/api/v1/stickers/packs/:id/stickers/:sticker_uuid` | 是（包主） |

> 上传/使用贴纸要求已绑定手机号；未绑定返回 `403`（"先绑定手机号再操作哦～"）。

## 上传贴纸

### `POST /api/v1/stickers`

上传贴纸图片，自动加入本人贴纸收藏夹（`source=1`）。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `multipart/form-data` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求参数（表单字段）

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `file` | file | 是 | jpg/png/gif/webp 图片，单张最大 10MB | 贴纸图片（按文件头魔数嗅探校验，拒绝 SVG 等） |
| `name` | string | 否 | 最长 64 字符 | 贴纸展示名，缺省取原文件名 |

#### 成功响应

状态码：`201 Created`

```json
{
  "code": 201,
  "msg": "贴纸上传成功",
  "sticker": {
    "uuid": "123e4567-e89b-12d3-a456-426614174000",
    "file_uuid": "0f9c2b31-6a4d-4c39-9c0a-1b2f3d4e5f60",
    "pack_id": 0,
    "name": "好耶",
    "source": 1,
    "top": 0,
    "created_at": "2026-06-07 12:00:00"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 缺少 `file`、文件名非法、图片不是安全栅格格式（SVG/HTML/JS 等），body 中 `code` 与 msg 区分 `贴纸仅支持 jpg/png/gif/webp 图片` |
| `401` | 未登录 |
| `403` | 未绑定手机号；上传权限被封禁（`code` 为 `40302`） |
| `413` | 贴纸超过单张大小限制（默认 10MB） |
| `503` | 文件存储服务不可用 |

## 贴纸收藏夹（个人）

### `GET /api/v1/stickers/collection`

获取当前用户贴纸收藏夹（我上传的 + 收藏的贴纸）。**置顶（`top=1`）的贴纸排在最前**，其余按加入时间升序。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页数量，默认 `20`，最大 `100` |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "count": 2,
  "stickers": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "file_uuid": "0f9c2b31-6a4d-4c39-9c0a-1b2f3d4e5f60",
      "pack_id": 0,
      "name": "好耶",
      "source": 1,
      "top": 1,
      "created_at": "2026-06-07 12:00:00"
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `source` | number | `1` 本人上传，`2` 收藏他人贴纸 |
| `top` | number | `1` 已置顶（收藏夹内优先展示），`0` 未置顶 |

### `POST /api/v1/stickers/collection`

把他人贴纸添加到收藏夹（幂等：已在收藏夹时直接返回成功）。**收藏（`source=2`）上限 500 张**，超限返回 `400`（"收藏夹最多添加 500 张贴纸"）。本人上传的贴纸自动在收藏夹中，不占该额度。

#### 请求体

```json
{ "sticker_uuid": "123e4567-e89b-12d3-a456-426614174000" }
```

#### 成功响应

状态码：`200 OK`，返回 `sticker` 对象（同上传响应）。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `sticker_uuid` 非法；收藏数量达到 500 上限 |
| `403` | 贴纸所在贴纸包被封禁 |
| `404` | 贴纸不存在或已删除 |

### `DELETE /api/v1/stickers/collection`

从收藏夹移除收藏的贴纸（幂等）。请求体同添加收藏（`sticker_uuid`）。本人上传的贴纸条目不会被移除（使用删除贴纸接口）。

### `PATCH /api/v1/stickers/:uuid/top`

设置/取消收藏夹贴纸置顶。请求体：`{ "top": 1 }` 置顶，`{ "top": 0 }` 取消置顶。贴纸必须在当前用户收藏夹中，否则 `404`。

### `DELETE /api/v1/stickers/:uuid`

删除本人上传的贴纸（软删除，同时清理收藏夹条目）。非本人贴纸一律按 `404` 处理。删除后历史评论中的贴纸按「贴纸未显示」处理。

## 贴纸包

### `GET /api/v1/stickers/packs`

贴纸包列表（不含被封禁的贴纸包）。`filter=all`（默认）全部贴纸包；`filter=favorite` 当前用户已整包收藏的贴纸包。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `filter` | string | 否 | `all`（默认）/ `favorite` |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页数量，默认 `20`，最大 `100` |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "count": 1,
  "packs": [
    {
      "id": 3,
      "uuid": "9d1f8a2c-1111-4a2b-8c3d-222233334444",
      "user_id": 10001,
      "name": "小猫贴纸",
      "description": "可爱小猫",
      "banned": false,
      "sticker_count": 12,
      "is_favorite": true,
      "created_at": "2026-06-07 12:00:00"
    }
  ]
}
```

### `POST /api/v1/stickers/packs`

创建贴纸包。要求登录、绑定手机号且未被封禁评论权限（`PermComment`）。

#### 请求体

```json
{ "name": "小猫贴纸", "description": "可爱小猫" }
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | string | 是 | 非空，最长 64 字符 | 贴纸包名称 |
| `description` | string | 否 | 最长 255 字符 | 贴纸包描述 |

#### 成功响应

状态码：`201 Created`，返回 `pack` 对象（同列表字段）。

### `POST /api/v1/stickers/packs/favorites` / `DELETE /api/v1/stickers/packs/favorites`

整包收藏/取消收藏（幂等）。请求体：`{ "pack_id": 3 }`。收藏后**包内容更新自动同步**；被封禁的贴纸包不能收藏（`403`）。

### `GET /api/v1/stickers/packs/:id`

贴纸包详情（含包内贴纸列表，按包内排序升序）。被封禁的贴纸包返回 `403`。

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "pack": { "id": 3, "sticker_count": 1, "is_favorite": true },
  "stickers": [
    { "uuid": "...", "file_uuid": "...", "pack_id": 3, "name": "喵", "source": 0, "top": 0, "created_at": "..." }
  ]
}
```

### `POST /api/v1/stickers/packs/:id/stickers`

把**本人上传**的贴纸加入自己的贴纸包。请求体：`{ "sticker_uuid": "..." }`。单包上限 100 张（超限返回 `400`，"一个贴纸包最多添加 100 张贴纸"）；已在其他贴纸包的贴纸需先移出。

### `DELETE /api/v1/stickers/packs/:id/stickers/:sticker_uuid`

把贴纸移出自己的贴纸包（贴纸本身保留）。非包主/贴纸不在该包内按 `404` 处理。

## 评论中的贴纸

- 评论/回复的 `content` 中可携带内嵌标记 `[sticker:<uuid>]`（由前端贴纸选择器在光标处插入），标记随文字在任意位置穿插展示。
- **一条评论最多使用 25 张贴纸**（可重复使用同一张贴纸，按标记出现次数展示）；超过 25 个标记返回 `400`（"贴纸使用错误，一条评论最多使用 25 张贴纸"）。
- 评论正文长度限制（1000 字符）按去除贴纸标记后的文字计算，贴纸标记不占正文字数。
- 使用的贴纸必须在使用者的可用集合内（本人上传/收藏夹/已收藏贴纸包），否则 `403`；所在贴纸包被封禁的贴纸不可用于新评论（`403`）。
- 评论/回复响应中的 `stickers` 数组按 `content` 中标记出现顺序给出各张贴纸的展示信息：`hidden=true` 表示该贴纸无法显示（所在贴纸包被封禁或贴纸已删除），前端隐藏该贴纸并在评论下灰字提示「部分贴纸未显示」。

详见 `api-moments.md` 评论相关章节。
