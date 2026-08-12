# Miaoverse Content/User API 文档

## 用户api部分

## 路由分组

接口按业务域分组，统一挂在 `/api/v1` 下：

| 分组 | 前缀 | 是否需要登录 | 内容 |
| --- | --- | --- | --- |
| 认证 | `/api/v1/auth` | 否（账号列表/切换需登录） | 短信验证码、登录、注册、切换账号 |
| 动态 | `/api/v1/moment` | 是 | 发布动态、获取动态详情、评论/回复（楼中楼）、给动态点赞，见 `api-moments.md` |
| 用户 | `/api/v1/user` | 是 | 资料、文件、关注、拉黑/屏蔽/不想看、惩罚记录、查看他人资料/内容/关系 |

除认证组外，其余分组接口均要求登录（`mwu_sess_id` cookie）且账号状态正常；被权限封禁的账号按各接口说明返回 `40302`。

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

### `POST /api/v1/auth/sms/send`

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
curl -X POST http://localhost:3000/api/v1/auth/sms/send \
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
| `500` | 验证码写入 Redis 失败、短信服务发送失败（返回通用文案，不暴露短信服务内部错误） |

## 短信验证码登录

### `POST /api/v1/auth/login/sms`

使用手机号、验证码 UUID 和验证码登录。若手机号没有账号，会自动创建第一个账号并登录。若手机号绑定多个账号，会返回账号列表并进入待选择账号状态。已注销的账号不会出现在待选择列表中；若手机号下账号全部注销，视为新手机号自动创建新账号。

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
| `uuid` | string | 是 | UUID v4，小写十六进制格式 | `/auth/sms/send` 返回的 `code_uuid` |
| `code` | number | 是 | 数字 | 用户收到的 4 位短信验证码 |
| `a` | string | 是 | 只能包含数字，且时间校验通过 | 见上方 `a` 参数生成规则 |

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/auth/login/sms \
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
      "bio": "",
      "gender": 0,
      "status": 1,
      "created_at": "2026-06-01T12:00:00+08:00",
      "updated_at": "2026-06-01T12:00:00+08:00"
    }
  ]
}
```

注意：`users` 中的字段来自 Go 结构体 `model/dao/user.User`，JSON 字段名为小写下划线格式（`id`、`username`、`nickname`、`region`、`avatar`、`bio`、`gender`、`status`、`created_at`、`updated_at`）。

#### Session 行为

- 单账号登录或自动注册成功时，session 中会写入 `Phone`、`Region`、`UID`。
- 多账号待选择时，session 中会写入 `PendingLoginPhone`、`PendingLoginRegion`，并清理 `UID`。
- 后续调用 `/api/v1/auth/login/choose` 时必须携带同一个 `mwu_sess_id` cookie。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失、参数格式错误、`a` 超时或无效 |
| `403` | 验证码错误、验证码不存在或验证码已过期 |
| `500` | Redis、数据库或 session 写入异常 |

## 选择登录账号

### `POST /api/v1/auth/login/choose`

当 `/api/v1/auth/login/sms` 返回 `300 Multiple Choices` 时，调用该接口选择具体账号完成登录。

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
curl -i -X POST http://localhost:3000/api/v1/auth/login/choose \
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

### `POST /api/v1/auth/register/sms`

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

字段说明与 `/api/v1/auth/login/sms` 相同。

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/auth/register/sms \
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

## 获取可切换账号列表

### `GET /api/v1/auth/accounts`

返回当前会话登录手机号绑定的全部账号。仅要求已登录，不校验当前账号状态，因此当前账号被封禁时也能查看列表并切换。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/auth/accounts \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "current": 10001,
  "users": [
    {
      "id": 10001,
      "username": "user_xxx",
      "nickname": "nickname_xxx",
      "region": 86,
      "avatar": "",
      "bio": "",
      "gender": 0,
      "status": 1,
      "created_at": "2026-06-01T12:00:00+08:00",
      "updated_at": "2026-06-01T12:00:00+08:00"
    },
    {
      "id": 10002,
      "username": "user_yyy",
      "nickname": "nickname_yyy",
      "region": 86,
      "avatar": "",
      "bio": "",
      "gender": 0,
      "status": 1,
      "created_at": "2026-06-01T12:00:00+08:00",
      "updated_at": "2026-06-01T12:00:00+08:00"
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `current` | number | 当前登录账号的用户 ID |
| `users` | array | 该手机号绑定的可登录账号（字段与 `user` 表一致，含封禁账号，不含已注销账号） |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `401` | 未登录或 session 中没有 `UID`/手机号 |
| `500` | 数据库查询异常 |

## 切换账号

### `POST /api/v1/auth/switch`

将当前会话切换到同一手机号绑定的另一个账号。仅要求已登录，不校验当前账号状态；目标账号必须属于当前会话手机号且账号状态为正常（未被封禁、未注销）。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "uid": 10002
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `uid` | number | 是 | 要切换到的账号 ID，必须属于当前会话手机号 |

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/auth/switch \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -c cookie.txt \
  -d '{"uid":10002}'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "登录成功",
  "uid": 10002
}
```

#### Session 行为

切换成功后 session 会重新生成（新 `mwu_sess_id`），写入 `Phone`、`Region`、`UID`，后续请求必须携带响应中的新 cookie。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失或格式错误 |
| `401` | 未登录或 session 中没有手机号 |
| `403` | 目标账号不属于当前会话手机号（普通 `403`）；目标账号已被封禁（`code` 为 `40303`） |
| `500` | 数据库或 session 写入异常 |

## 退出登录

### `POST /api/v1/auth/logout`销毁当前 session 并清除登录态 cookie。未登录时调用也返回成功（幂等）。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -c cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "退出登录成功"
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `500` | session 销毁异常 |

## 获取当前登录用户信息

### `GET /api/v1/user/me`

返回当前登录用户的基础信息，用于前端刷新/恢复登录态。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/user/me \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "user": {
    "id": 10001,
    "username": "user_xxx",
    "nickname": "nickname_xxx",
    "region": 86,
    "avatar": "",
    "bio": "",
    "gender": 0,
    "status": 1,
    "created_at": "2026-06-01T12:00:00+08:00",
    "updated_at": "2026-06-01T12:00:00+08:00"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `401` | 未登录或 session 中没有 `UID` |
| `500` | 数据库查询异常 |

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
  "bio": "Hello, Miaoverse!",
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
| `bio` | string | 是 | 个性签名，可为空字符串，最长 255 字符。修改时要求未被封禁签名权限位，否则返回 `403`，body 中 `code` 为 `40302` |
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
    "bio": "Hello, Miaoverse!",
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
  -d '{"nickname":"新的昵称","bio":"新的签名"}'
```

#### 可能的错误
| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、字段为空、字段长度/取值不合法，或 PATCH 未包含任何可更新字段 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 修改 `bio` 时签名权限被封禁（`code` 为 `40302`） |
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
| `permission` | number | 否 | 分享权限：`0` 给全部人公开，`1` 给好友公开，`2` 不给任何人公开（默认），`3` 给粉丝公开 |

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

访问控制规则：

- 文件所有者拉黑了当前查看者时，无论文件是否公开，都返回 `403`，提示「由于对方权限设置，无法查看此文件」。
- 文件分享权限为 `0`（全部人公开）时，任何登录用户可访问。
- 文件分享权限为 `1`（好友公开）时，仅当前查看者关注了文件所有者才可访问。
- 文件分享权限为 `3`（粉丝公开）时，仅文件所有者关注了当前查看者才可访问。
- 文件分享权限为 `2`（不公开）或不符合上述条件时，返回 `403`，提示「此文件并未公开分享，请检查登录账号」。

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
| `403` | 文件未公开分享（「此文件并未公开分享，请检查登录账号」），或文件所有者拉黑了查看者（「由于对方权限设置，无法查看此文件」） |
| `404` | 文件不存在或已删除（非 active 状态） |
| `503` | S3 未启用或文件存储服务不可用 |
| `500` | S3 临时链接生成或数据库查询异常 |

## 拉黑/屏蔽/不想看

### `POST /api/v1/user/blocks`

对目标用户执行拉黑、屏蔽或不想看操作，或取消对应操作。每个用户每种关系类型对应一个 Redis Bitmap（RoaringBitmap 序列化存储），首次操作自动初始化。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "target": 20002,
  "type": 1,
  "action": "add"
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `target` | number | 是 | 大于 0，且不能等于当前登录用户 ID | 目标用户 ID |
| `type` | number | 是 | `1` 拉黑，`2` 屏蔽，`3` 不想看 | 关系类型 |
| `action` | string | 是 | `add` 或 `remove` | 操作：`add` 添加，`remove` 取消 |

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/user/blocks \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{"target":20002,"type":1,"action":"add"}'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "操作成功",
  "target": 20002,
  "type": 1,
  "action": "add"
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、`target` 为 0 或等于当前用户、`type` 非法、`action` 不是 `add`/`remove` |
| `401` | 未登录或 session 中没有 `UID` |
| `500` | Redis 读写异常 |

## 获取其他用户内容列表

以下接口用于获取其他用户发布的内容（当前为动态）。先请求数量，再分页获取列表。

### `GET /api/v1/user/users/:uid/contents/count`

获取目标用户对当前登录用户可见的内容数量。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/user/users/20002/contents/count \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "count": 12
}
```

### `GET /api/v1/user/users/:uid/contents`

分页获取目标用户对当前登录用户可见的内容列表。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页数量，默认 `20`，最大 `100` |

#### 请求示例

```bash
curl -i "http://localhost:3000/api/v1/user/users/20002/contents?offset=0&limit=20" \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "contents": [
    {
      "id": 1,
      "type": "moment",
      "comment": 3,
      "like": 5
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | number | 内容 ID |
| `type` | string | 内容类型，当前为 `moment` |
| `comment` | number | 评论数 |
| `like` | number | 点赞数 |

#### 可见性规则

- 公开（`permission=0`）动态对所有人可见。
- 仅好友（`permission=1`）动态仅对互相关注的查看者可见。
- 仅粉丝（`permission=3`）动态仅对目标用户关注了的查看者可见。
- 仅自己（`permission=2`）动态仅本人可见。
- 草稿、已删除等非正常状态动态不返回。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `uid` 非法、`offset`/`limit` 取值非法 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 查看者与目标用户存在任意一方的拉黑关系，body 中 `code` 为 `40301`（见 `API-errors.md`） |
| `500` | 数据库查询异常 |

## 获取其他用户信息

### `GET /api/v1/user/users/:uid/info`

获取目标用户的公开资料，并附带当前登录用户对目标用户的拉黑/屏蔽/不想看关系状态。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/user/users/20002/info \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "user": {
    "id": 20002,
    "username": "user_xxx",
    "nickname": "nickname_xxx",
    "region": 86,
    "avatar": "",
    "bio": "",
    "gender": 0,
    "status": 1,
    "created_at": "2026-06-01T12:00:00+08:00",
    "updated_at": "2026-06-01T12:00:00+08:00",
    "block_status": 1,
    "punishment_mask": 2
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `user` | object | 目标用户资料，字段与 `user` 表一致。若目标用户已注销（`status=3`），`username` 固定返回「已注销的账号」、`bio` 固定返回站长留言文案 |
| `user.block_status` | number | 当前登录用户对目标用户的关系状态（位组合）：`0` 无关系，`1` 拉黑，`2` 屏蔽，`4` 不想看；可组合，如 `3` 表示拉黑+屏蔽 |
| `user.punishment_mask` | number | 目标用户当前生效中的权限封禁位掩码（十进制，按位或合并）：bit0(1) 评论，bit1(2) 发布动态，bit2(4) 私信，bit3(8) 头像，bit4(16) 昵称，bit5(32) 签名，bit6(64) 社交互动，bit7(128) 注销/注册，bit8(256) 上传文件。`0` 表示无生效封禁。前端可据此展示「该用户被禁止发送评论」等提示 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `uid` 非法 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 目标用户账号被封禁（`user.status = 2`），body 中 `code` 为 `40304`（见 `API-errors.md`） |
| `404` | 目标用户不存在 |
| `500` | 数据库或 Redis 查询异常 |

## 查询本人惩罚记录

### `GET /api/v1/user/punishments`

查询当前登录用户本人的全部惩罚记录。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/user/punishments \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "punishments": [
    {
      "id": 1,
      "user_id": 10001,
      "punishment_type": 2,
      "punishment_status": 1,
      "punishment_time": "2026-06-07T12:00:00+08:00",
      "punishment_end_time": "2026-06-14T12:00:00+08:00",
      "punishment_reason": "发布违规内容",
      "punishment_operator": 0,
      "punishment_remark": ""
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `punishment_type` | number | 被封禁权限的十进制位掩码，需自行解析二进制：bit0(1) 评论，bit1(2) 发布动态，bit2(4) 私信，bit3(8) 头像，bit4(16) 昵称，bit5(32) 签名，bit6(64) 社交互动，bit7(128) 注销/注册，bit8(256) 上传文件 |
| `punishment_status` | number | `1` 生效中，`2` 已到期，`3` 已撤销 |
| `punishment_end_time` | string/null | 封禁结束时间，`null` 表示永久 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `401` | 未登录或 session 中没有 `UID` |
| `500` | 数据库查询异常 |

## 获取用户关注/粉丝列表

### `GET /api/v1/user/users/:uid/following`

分页获取目标用户关注的用户列表。

### `GET /api/v1/user/users/:uid/followers`

分页获取目标用户的粉丝（关注了目标用户的人）列表。

查看者与目标用户之间存在任意一方拉黑关系时不可查询。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页数量，默认 `20`，最大 `100` |

#### 请求示例

```bash
curl -i "http://localhost:3000/api/v1/user/users/20002/following?offset=0&limit=20" \
  -b cookie.txt
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "count": 12,
  "users": [
    {
      "id": 20003,
      "username": "user_xxx",
      "nickname": "nickname_xxx",
      "region": 86,
      "avatar": "",
      "bio": "",
      "gender": 0,
      "status": 1,
      "created_at": "2026-06-01T12:00:00+08:00",
      "updated_at": "2026-06-01T12:00:00+08:00",
      "block_status": 0
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `count` | number | 总数 |
| `users` | array | 用户列表，字段与 `user` 表一致。列表中已注销账号（`status=3`）的 `username` 固定返回「已注销的账号」、`bio` 固定返回站长留言文案 |
| `users[].block_status` | number | 当前登录用户对列表中每个用户的关系状态（位组合）：`0` 无关系，`1` 拉黑，`2` 屏蔽，`4` 不想看 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `uid` 非法、`offset`/`limit` 取值非法 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 查看者与目标用户存在任意一方的拉黑关系，body 中 `code` 为 `40301`（见 `API-errors.md`） |
| `500` | 数据库或 Redis 查询异常 |

## 关注用户

### `POST /api/v1/user/follows`

关注目标用户。不能关注自己；被拉黑/拉黑对方、或目标用户被自己屏蔽/不想看时不能关注。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "target": 20002
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `target` | number | 是 | 大于 0，且不能等于当前登录用户 ID | 要关注的用户 ID |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "操作成功",
  "target": 20002,
  "action": "follow"
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、`target` 为 0 或等于当前用户 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 社交互动权限封禁（`code` 为 `40302`）；拉黑/被拉黑关系（`code` 为 `40301`）；目标用户被自己屏蔽或不想看（`code` 为 `40301`） |
| `500` | 数据库异常 |

## 推荐调用流程

### 首次短信登录或自动注册

1. 调用 `POST /api/v1/auth/sms/send` 获取 `code_uuid`。
2. 用户输入短信验证码。
3. 调用 `POST /api/v1/auth/login/sms`。
4. 如果返回 `200` 或 `201`，登录完成。
5. 如果返回 `300`，保存同一个 cookie，并调用 `POST /api/v1/auth/login/choose` 选择账号。

### 为同一手机号追加新账号

1. 调用 `POST /api/v1/auth/sms/send` 获取 `code_uuid`。
2. 用户输入短信验证码。
3. 调用 `POST /api/v1/auth/register/sms`。
4. 返回 `201` 时，新账号创建并登录完成。

### 已登录状态下切换账号

1. 调用 `GET /api/v1/auth/accounts` 获取当前手机号绑定的账号列表。
2. 用户选择目标账号。
3. 调用 `POST /api/v1/auth/switch` 切换，之后使用响应中新的 `mwu_sess_id`。

## 状态码速查

| 状态码 | 含义 |
| --- | --- |
| `200` | 请求成功 |
| `201` | 创建成功、注册成功或上传成功 |
| `300` | 需要用户选择一个账号继续登录 |
| `400` | 请求格式或参数错误 |
| `403` | 验证失败、账号不匹配、切换目标账号被封禁或账号状态不允许操作 |
| `404` | 注册新账号时，该手机号还没有首个账号 |
| `413` | 上传文件过大 |
| `503` | 文件存储服务不可用 |
| `500` | 服务端内部错误 |
