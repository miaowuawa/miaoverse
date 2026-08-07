# Miaoverse Content/User API 文档

## 用户api部分

## 通用约定

- 基础地址：`http://{host}:{port}`
- API 前缀：`/api/v1`
- 请求体格式：除健康检查和文件上传外，接口均使用 `Content-Type: application/json`
- 响应格式：JSON
- 请求 ID：服务会在响应头中写入 `X-Miaoverse-ReqID`
- Session Cookie：`mwu_sess_id`
- 登录相关接口会写入或读取服务端 session。调用需要保持同一个 cookie，尤其是多账号选择流程。

## `a` 参数生成规则

多个接口都要求请求体中携带字段 `a`，用于校验请求时间。代码中的校验逻辑是：

1. 获取当前毫秒时间戳，例如 `Date.now()`。
2. 对时间戳做平方。
3. 将平方后的十进制字符串反转。
4. 作为 `a` 字段传入。
5. 服务端允许客户端时间与服务端当前时间相差不超过 `1000ms`。

JavaScript 示例：

```js
function buildA() {
  const ts = BigInt(Date.now());
  return (ts * ts).toString().split("").reverse().join("");
}
```

## 通用响应

错误响应通常使用以下结构：

```json
{
  "code": 400,
  "msg": "请求错误，请检查参数"
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | number | HTTP 状态码 |
| `msg` | string | 响应说明 |

## 健康检查

### `GET /`

用于检查服务是否存活。

#### 请求示例

```bash
curl -i http://localhost:3000/
```

#### 成功响应

状态码：`200 OK`

响应体为纯文本：

```text
Miaoverse API Resp at 2026-06-01 12:00:00
```

## 发送短信验证码

### `POST /api/v1/sms/send`

向指定手机号发送短信验证码。验证码有效期在代码中按 `5 分钟` 发送给短信服务，实际校验依赖 Redis 中验证码的过期时间配置。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |

#### 请求体

```json
{
  "phone": "13800138000",
  "region": "86",
  "a": "..."
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `phone` | string | 是 | 只能包含数字 | 手机号 |
| `region` | string | 是 | 数字字符串 | 手机区号，例如中国大陆为 `"86"` |
| `a` | string | 是 | 只能包含数字，且时间校验通过 | 见上方 `a` 参数生成规则 |

#### 请求示例

```bash
curl -X POST http://localhost:3000/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","region":"86","a":"'"$(node -e 'const ts=BigInt(Date.now());process.stdout.write((ts*ts).toString().split("").reverse().join(""))')"'" }'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code_uuid": "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a",
  "msg": "发送成功"
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code_uuid` | string | 本次验证码 UUID。后续短信登录或注册时作为 `uuid` 传入 |
| `msg` | string | 响应说明 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失、参数格式错误、`a` 超时或无效 |
| `500` | 验证码写入 Redis 失败、短信服务发送失败 |

## 短信验证码登录

### `POST /api/v1/user/login/sms`

使用手机号、验证码 UUID 和验证码登录。若手机号没有账号，会自动创建第一个账号并登录。若手机号绑定多个账号，会返回账号列表并进入待选择账号状态。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |

#### 请求体

```json
{
  "phone": "13800138000",
  "region": 86,
  "uuid": "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a",
  "code": 1234,
  "a": "..."
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `phone` | string | 是 | 只能包含数字 | 手机号 |
| `region` | number | 是 | 数字 | 手机区号，例如 `86` |
| `uuid` | string | 是 | UUID v4，小写十六进制格式 | `/sms/send` 返回的 `code_uuid` |
| `code` | number | 是 | 数字 | 用户收到的 4 位短信验证码 |
| `a` | string | 是 | 只能包含数字，且时间校验通过 | 见上方 `a` 参数生成规则 |

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/user/login/sms \
  -H "Content-Type: application/json" \
  -c cookie.txt \
  -d '{
    "phone": "13800138000",
    "region": 86,
    "uuid": "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a",
    "code": 1234,
    "a": "..."
  }'
```

#### 成功响应：已有单账号登录

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "登录成功",
  "uid": 10001
}
```

#### 成功响应：自动注册并登录

状态码：`201 Created`

```json
{
  "code": 201,
  "msg": "注册并登录成功",
  "uid": 10001
}
```

#### 成功响应：多账号待选择

状态码：`300 Multiple Choices`

```json
{
  "code": 300,
  "msg": "请选择要登录的账号",
  "users": [
    {
      "id": 10001,
      "username": "user_xxx",
      "nickname": "nickname_xxx",
      "region": 86,
      "avatar": "",
      "gender": 0,
      "status": 1,
      "created_at": "2026-06-01T12:00:00+08:00",
      "updated_at": "2026-06-01T12:00:00+08:00"
    }
  ]
}
```

注意：`users` 中的字段来自 Go 结构体 `model/dao/user.User`，JSON 字段名为小写下划线格式（`id`、`username`、`nickname`、`region`、`avatar`、`gender`、`status`、`created_at`、`updated_at`）。

#### Session 行为

- 单账号登录或自动注册成功时，session 中会写入 `Phone`、`Region`、`UID`。
- 多账号待选择时，session 中会写入 `PendingLoginPhone`、`PendingLoginRegion`，并清理 `UID`。
- 后续调用 `/api/v1/user/login/choose` 时必须携带同一个 `mwu_sess_id` cookie。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失、参数格式错误、`a` 超时或无效 |
| `403` | 验证码错误、验证码不存在或验证码已过期 |
| `500` | Redis、数据库或 session 写入异常 |

## 选择登录账号

### `POST /api/v1/user/login/choose`

当 `/api/v1/user/login/sms` 返回 `300 Multiple Choices` 时，调用该接口选择具体账号完成登录。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 必须携带前一步返回的 `mwu_sess_id` |

#### 请求体

```json
{
  "uid": 10001
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `uid` | number | 是 | 要登录的用户 ID，必须属于前一步验证码验证通过的手机号 |

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/user/login/choose \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -c cookie.txt \
  -d '{"uid":10001}'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "登录成功",
  "uid": 10001
}
```

#### Session 行为

选择成功后，session 中会清理 `PendingLoginPhone`、`PendingLoginRegion`，并写入 `Phone`、`Region`、`UID`。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失、没有待选择账号状态 |
| `403` | 选择的 `uid` 不属于本次验证码验证通过的手机号 |
| `500` | 数据库或 session 写入异常 |

## 为手机号注册新账号

### `POST /api/v1/user/register/sms`

使用短信验证码为已有手机号注册一个新账号，并立即登录。代码逻辑要求该手机号已经至少存在一个账号；如果没有账号，会返回 `404`，提示应先使用短信登录自动创建首个账号。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |

#### 请求体

```json
{
  "phone": "13800138000",
  "region": 86,
  "uuid": "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a",
  "code": 1234,
  "a": "..."
}
```

字段说明与 `/api/v1/user/login/sms` 相同。

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/user/register/sms \
  -H "Content-Type: application/json" \
  -c cookie.txt \
  -d '{
    "phone": "13800138000",
    "region": 86,
    "uuid": "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a",
    "code": 1234,
    "a": "..."
  }'
```

#### 成功响应

状态码：`201 Created`

```json
{
  "code": 201,
  "msg": "新账号注册并登录成功",
  "uid": 10002
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失、参数格式错误、`a` 超时或无效 |
| `403` | 验证码错误、验证码不存在或验证码已过期 |
| `404` | 该手机号还没有任何账号 |
| `500` | Redis、数据库或 session 写入异常 |

## 修改用户信息

### `PUT /api/v1/user/info`

全量更新当前登录用户的基础信息。该接口只更新 `user` 表中的非凭据信息，不修改手机号、密码、第三方登录等个人凭据。

#### 请求头
| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体
```json
{
  "username": "miaoverse_user",
  "nickname": "Miaowu",
  "region": 86,
  "avatar": "https://example.com/avatar.png",
  "gender": 0
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 是 | 用户名，1-64 字符 |
| `nickname` | string | 是 | 昵称，1-64 字符 |
| `region` | number | 是 | 用户资料地区，必须大于 0 |
| `avatar` | string | 是 | 头像 URL，可为空字符串，最长 255 字符 |
| `gender` | number | 是 | `0` 未知，`1` 男，`2` 女，`3` 非二元性别 |

#### 成功响应
状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "用户信息修改成功",
  "user": {
    "id": 10001,
    "username": "miaoverse_user",
    "nickname": "Miaowu",
    "region": 86,
    "avatar": "https://example.com/avatar.png",
    "gender": 0,
    "status": 1,
    "created_at": "2026-06-01T12:00:00+08:00",
    "updated_at": "2026-06-01T12:00:00+08:00"
  }
}
```

### `PATCH /api/v1/user/info`

部分更新当前登录用户的基础信息。请求体至少包含一个可更新字段；允许字段与 `PUT /api/v1/user/info` 相同。

#### 请求示例
```bash
curl -i -X PATCH http://localhost:3000/api/v1/user/info \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{"nickname":"新的昵称","avatar":""}'
```

#### 可能的错误
| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、字段为空、字段长度/取值不合法，或 PATCH 未包含任何可更新字段 |
| `401` | 未登录或 session 中没有 `UID` |
| `404` | session 中的用户不存在 |
| `409` | `username` 或 `nickname` 等唯一字段与已有用户冲突 |
| `500` | 数据库异常 |

## 用户文件

以下接口都挂在 `/api/v1/user` 登录态路由下，必须携带有效 `mwu_sess_id` cookie。

### `POST /api/v1/user/files`

上传当前登录用户的文件。文件会写入 S3，并在数据库 `files` 表中创建记录。
如果已存在相同 SHA-256 hash 的 active 文件，服务会复用已有 S3 对象链接，只创建新的数据库记录，避免重复上传。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `multipart/form-data` |

#### 表单字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | file | 是 | 要上传的文件 |
| `file_type` | string | 否 | `image`、`video`、`audio`、`document`、`other`。不填时按 MIME 自动粗略识别 |

上传大小由配置项 `upload.max_file_size_bytes` 控制，默认 `20971520` 字节。

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/user/files \
  -b cookie.txt \
  -F "file=@/path/to/avatar.png" \
  -F "file_type=image"
```

#### 成功响应

状态码：`201 Created`

```json
{
  "code": 201,
  "msg": "上传成功",
  "file": {
    "uuid": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d",
    "file_name": "avatar.png",
    "file_url": "https://cdn.example.com/uploads/10001/15b3d25d-66cc-4ddc-9949-33c9e84d8c5d/avatar.png",
    "file_type": "image",
    "file_ext": "png",
    "mime_type": "image/png",
    "file_size": 12345,
    "hash": "sha256hex...",
    "created_at": "2026-06-07 12:00:00"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 没有上传 `file` 字段或文件名无效 |
| `401` | 未登录或 session 中没有 `UID` |
| `413` | 文件超过 `upload.max_file_size_bytes` |
| `503` | S3 未启用或文件存储服务不可用 |
| `500` | S3 上传或数据库写入异常 |

### `GET /api/v1/user/files/:uuid/temp-link`

通过文件 UUID 获取当前登录用户自己文件的 S3 临时访问链接。不能获取其他用户文件。

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/user/files/15b3d25d-66cc-4ddc-9949-33c9e84d8c5d/temp-link \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "link": {
    "uuid": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d",
    "url": "https://s3.example.com/...",
    "expires_at": "2026-06-07T12:05:00Z"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | UUID 参数为空 |
| `401` | 未登录或 session 中没有 `UID` |
| `404` | 文件不存在、已删除或不属于当前用户 |
| `503` | S3 未启用或文件存储服务不可用 |
| `500` | S3 临时链接生成或数据库查询异常 |

### `GET /api/v1/user/files/:uuid/shared-link`

通过文件 UUID 获取任意用户 active 文件的 S3 临时访问链接，用于帖子等场景查看/下载其他用户的文件或媒体。临时链接绑定当前登录用户身份，且只对 active 状态的文件生效。

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/user/files/15b3d25d-66cc-4ddc-9949-33c9e84d8c5d/shared-link \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "link": {
    "uuid": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d",
    "url": "https://s3.example.com/...",
    "expires_at": "2026-06-07T12:05:00Z"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | UUID 参数为空或格式非法 |
| `401` | 未登录或 session 中没有 `UID` |
| `404` | 文件不存在或已删除（非 active 状态） |
| `503` | S3 未启用或文件存储服务不可用 |
| `500` | S3 临时链接生成或数据库查询异常 |

## 发布动态

### `POST /api/v1/user/moments`

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
curl -i -X POST http://localhost:3000/api/v1/user/moments \
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

## 推荐调用流程

### 首次短信登录或自动注册

1. 调用 `POST /api/v1/sms/send` 获取 `code_uuid`。
2. 用户输入短信验证码。
3. 调用 `POST /api/v1/user/login/sms`。
4. 如果返回 `200` 或 `201`，登录完成。
5. 如果返回 `300`，保存同一个 cookie，并调用 `POST /api/v1/user/login/choose` 选择账号。

### 为同一手机号追加新账号

1. 调用 `POST /api/v1/sms/send` 获取 `code_uuid`。
2. 用户输入短信验证码。
3. 调用 `POST /api/v1/user/register/sms`。
4. 返回 `201` 时，新账号创建并登录完成。

## 状态码速查

| 状态码 | 含义 |
| --- | --- |
| `200` | 请求成功 |
| `201` | 创建成功、注册成功或上传成功 |
| `300` | 需要用户选择一个账号继续登录 |
| `400` | 请求格式或参数错误 |
| `403` | 验证失败、账号不匹配或账号状态不允许操作 |
| `404` | 注册新账号时，该手机号还没有首个账号 |
| `413` | 上传文件过大 |
| `503` | 文件存储服务不可用 |
| `500` | 服务端内部错误 |
