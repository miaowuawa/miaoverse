# Miaoverse Content/User API 文档

## 用户api部分

## 路由分组

接口按业务域分组，统一挂在 `/api/v1` 下：

| 分组 | 前缀 | 是否需要登录 | 内容 |
| --- | --- | --- | --- |
| 认证 | `/api/v1/auth` | 否（账号列表/切换需登录） | 短信验证码、登录、注册、切换账号 |
| 动态 | `/api/v1/moment` | 是 | 发布动态、获取动态详情、评论/回复（楼中楼）、给动态/评论点赞，见 `api-moments.md` |
| 用户 | `/api/v1/user` | 是 | 资料、密码、文件、关注、拉黑/屏蔽/不想看、惩罚记录、查看他人资料/内容/关系 |
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
| `phone` | string | 是 | 5-15 位纯数字（仅 ASCII 0-9，不允许 `+`、空格、`-` 或全角数字） | 手机号 |
| `region` | string | 是 | 数字字符串 | 手机区号，例如中国大陆为 `"86"` |
| `action` | string | 否 | 只允许公开场景 `LOGIN`（大小写不敏感） | 验证码业务场景，缺省为 `LOGIN`。见下方「业务场景隔离」 |
| `a` | string | 是 | 只能包含数字（仅 ASCII 0-9），且时间校验通过 | 见上方 `a` 参数生成规则 |

#### 业务场景隔离

验证码在 Redis 中按「业务场景 + 区号 + 手机号」存储，不同场景的验证码互不通用：

| `action` | 用途 | 申请接口 | 校验接口 |
| --- | --- | --- | --- |
| `LOGIN`（缺省） | 短信登录、注册、为手机号追加账号 | `POST /api/v1/auth/sms/send`（公开） | `POST /api/v1/auth/login/sms`、`POST /api/v1/auth/register/sms` |
| `CHANGE_PASSWORD` | 设置/修改账号密码 | `POST /api/v1/user/password/sms`（需登录） | `PUT /api/v1/user/password` |

也就是说，为登录申请的验证码**不能**用于修改密码，反之亦然，避免验证码被跨场景重放。

`CHANGE_PASSWORD` 场景**不允许**通过公开接口申请：该场景的手机号由服务端从登录会话取，
攻击者无法用公开接口给任意手机号发送「修改密码」短信（短信轰炸 / 社工诱导）。

#### 频次限制

- 同一「业务场景 + 区号 + 手机号」每 `60` 秒只能成功申请一次验证码，冷却期内重复申请返回 `429`。
- 单个验证码最多允许输错 `5` 次，达到上限后验证码立即作废，必须重新申请（防止 4 位数字验证码被在线枚举）。

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
| `code_uuid` | string | 本次验证码 UUID。后续短信登录、注册或修改密码时作为 `uuid` 传入 |
| `msg` | string | 响应说明 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失、参数格式错误、`action` 不在公开场景白名单内、`a` 超时或无效 |
| `429` | 同一「场景 + 手机号」在 60 秒冷却期内重复申请（`msg` 为「验证码发送过于频繁，请稍后再试」） |
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
| `phone` | string | 是 | 5-15 位纯数字（仅 ASCII 0-9，不允许 `+`、空格、`-` 或全角数字） | 手机号 |
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

返回当前登录用户的基础信息，用于前端刷新/恢复登录态与渲染个人资料页。

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
  },
  "phone": "+86 138****8000"
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `user` | object | `user` 表字段，见上方「修改用户信息」 |
| `phone` | string | 打码后的绑定手机号。手机号保存在 `user_credentials` 表且当前不支持更改，这里只做只读展示；未绑定手机号（非短信登录会话）时字段省略 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `401` | 未登录或 session 中没有 `UID` |
| `500` | 数据库查询异常 |

## 修改用户信息

### `PUT /api/v1/user/info`

全量更新当前登录用户的基础信息。该接口只更新 `user` 表中的非凭据信息，不修改手机号、区号、密码、头像文件、第三方登录等个人凭据。

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
  "avatar": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d",
  "bio": "Hello, Miaoverse!",
  "gender": 0
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 是 | 账号名，2-64 **字符**，只允许 ASCII 字母/数字/`_`/`-`/`.`，且必须以字母或数字开头；全局唯一。修改要求未被封禁 `PermNickname`（16），否则返回 `403`，body 中 `code` 为 `40302` |
| `nickname` | string | 是 | 昵称，1-64 **字符**；全局唯一。允许中英文、emoji 等可见字符，但不允许控制字符与 Unicode 格式字符（如零宽字符、双向覆盖符）。修改要求未被封禁 `PermNickname`（16），否则返回 `403`，body 中 `code` 为 `40302` |
| `avatar` | string | 否 | 头像文件 UUID（由 `POST /api/v1/user/files` 上传得到）。传值才修改头像，不传或为 `null` 表示不改动。校验规则与 `PUT /api/v1/user/avatar` **完全一致**：文件必须属于当前用户、为 active 图片文件、且为公开（`permission=0`），并要求未被封禁 `PermAvatar`（8），否则返回 `400` / `403`（`code` 为 `40302`） |
| `bio` | string | 是 | 个性签名，可为空字符串，最长 255 **字符**；允许多行（`\r\n`、`\r` 会统一为 `\n`）。修改要求未被封禁 `PermSignature`（32），否则返回 `403`，body 中 `code` 为 `40302` |
| `gender` | number | 是 | `0` 未知，`1` 男，`2` 女，`3` 非二元性别 |

#### 关于手机号与区号

- 请求体中的 `phone`、`region` **不会生效**：服务端只接收上表列出的字段。手机号与区号属于账号身份凭据（保存在 `user_credentials` 表），当前不支持更改；区号变更会使 `user.region` 与手机号凭据不同步，因此一并冻结。
- 头像既可以通过本接口的 `avatar` 字段（可与昵称等资料一起提交）修改，也可以单独调用 `PUT /api/v1/user/avatar`；两条路径共用同一套校验（`service/UserProfile`），不存在绕过文件归属检查的可能。
- 修改密码请使用 `PUT /api/v1/user/password`。

#### 头像修改示例

```bash
# 1. 上传头像文件，拿到文件 uuid（permission=0 表示公开，头像必须公开）
curl -s -X POST http://localhost:3000/api/v1/user/files \
  -b cookie.txt \
  -F "file=@/path/to/avatar.png" \
  -F "file_type=image" \
  -F "permission=0"

# 2. 与其它资料一起提交
curl -i -X PATCH http://localhost:3000/api/v1/user/info \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{"nickname":"新的昵称","avatar":"15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"}'
```

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
    "avatar": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d",
    "bio": "Hello, Miaoverse!",
    "gender": 0,
    "status": 1,
    "created_at": "2026-06-01T12:00:00+08:00",
    "updated_at": "2026-06-01T12:00:00+08:00"
  },
  "phone": "+86 138****8000"
}
```

响应中的 `phone` 为打码后的绑定手机号（与 `GET /api/v1/user/me` 一致）；会话未绑定手机号时省略。
`PATCH /api/v1/user/info` 的响应结构与之相同。

### `PATCH /api/v1/user/info`

部分更新当前登录用户的基础信息。请求体至少包含一个可更新字段；允许字段与 `PUT /api/v1/user/info` 相同，未传字段（`null` 或缺失）保持不变。

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
| `400` | 请求体不是 JSON、字段为空、字段长度/字符集/取值不合法、`avatar` 文件不存在/非本人/非图片/非公开，或 PATCH 未包含任何可更新字段 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 修改 `username`/`nickname` 时 `PermNickname` 被封禁、修改 `avatar` 时 `PermAvatar` 被封禁，或修改 `bio` 时 `PermSignature` 被封禁（`code` 为 `40302`） |
| `404` | session 中的用户不存在 |
| `409` | `username` 或 `nickname` 等唯一字段与已有用户冲突 |
| `500` | 数据库异常 |

## 账号密码

密码用于「密码登录」等场景，是可选的第二凭据：新账号默认只有手机号凭据，未设置密码。
密码通过手机验证码设置或修改，**手机号本身不支持更改**，因此改密不会影响手机号绑定关系。

### `POST /api/v1/user/password/sms`

向**当前登录账号绑定的手机号**发送「修改密码」场景的验证码（`action=CHANGE_PASSWORD`）。
设置密码与修改密码共用该接口。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json`（请求体可为空对象 `{}`） |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

请求体不接受任何手机号参数：手机号与区号取自登录会话，因此该接口无法被用来给任意号码发短信。
接口本身不要求 `a` 参数（申请验证码不改变账号状态，真正的校验在 `PUT /api/v1/user/password`）。

#### 请求示例

```bash
curl -i -X POST http://localhost:3000/api/v1/user/password/sms \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{}'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code_uuid": "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a",
  "msg": "发送成功"
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 当前会话没有绑定手机号（`msg` 为「先绑定手机号再操作哦～」） |
| `429` | 同一手机号的「修改密码」验证码在 60 秒冷却期内重复申请 |
| `500` | 验证码写入 Redis 失败、短信网关发送失败（返回通用文案，不暴露网关内部信息） |

### `GET /api/v1/user/password`

查询当前登录账号是否已设置密码，供客户端展示「设置密码 / 修改密码」。
只能查询本人（`uid` 取自登录会话），不接受任何用户标识参数。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "has_password": true
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `has_password` | boolean | 当前账号是否已设置密码。响应中不包含任何密码哈希或盐值 |

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `401` | 未登录或 session 中没有 `UID` |
| `500` | 数据库查询异常 |

### `PUT /api/v1/user/password`

通过手机验证码设置或修改当前登录账号的密码。已设置密码时是覆盖（改密），未设置时是首次设置，两者流程一致。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "uuid": "9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a",
  "code": 1234,
  "password": "miaoverse2026",
  "a": "..."
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `uuid` | string | 是 | 申请验证码时返回的 `code_uuid`。必须使用 `action=CHANGE_PASSWORD` 申请，见「发送短信验证码」 |
| `code` | number | 是 | 短信验证码（纯数字） |
| `password` | string | 是 | 新密码明文。8-64 **字符**，且至少包含「字母 / 数字 / 符号」中的两类；不允许控制字符或首尾空白 |
| `a` | string | 是 | 见上方 `a` 参数生成规则 |

请求体**不包含手机号与区号**：两者一律取自当前登录会话，客户端无法指定，因此不存在把密码改绑到其他手机号或修改他人账号密码的路径。

#### 请求示例

```bash
curl -i -X PUT http://localhost:3000/api/v1/user/password \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{"uuid":"9b7846fe-a58b-4d10-8f0e-b7f37d5a2a9a","code":1234,"password":"miaoverse2026","a":"'"$(node -e 'const ts=BigInt(Date.now());process.stdout.write((ts*ts).toString().split("").reverse().join(""))')"'" }'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "密码设置成功",
  "has_password": true
}
```

#### 安全行为

- 密码使用 bcrypt 哈希后写入 `user_credentials`（`credential_type=1`、`credential_key=bcrypt`），明文不落库、不进日志；已存在密码凭证时按唯一键 `(user_id, credential_type, credential_key)` 覆盖，单条 SQL 完成，避免并发下产生重复凭证。
- 验证码必须来自 `CHANGE_PASSWORD` 场景，登录验证码无法复用；验证码一次性，校验通过后立即销毁。
- 同一个验证码最多允许输错 5 次，超过后验证码作废，必须重新申请。
- 密码强度校验先于验证码校验：弱密码直接返回 `400`，不会消耗一次性验证码。
- 修改成功后服务端会重新生成 session ID（保留会话内容），防御会话固定攻击。
- 本接口不做 `PermDeleteRegister` 等权限位封禁判断：封禁期间仍允许用户加固自己的账号密码。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、参数缺失或格式错误、`a` 超时、密码强度不足（`error.password_too_weak`） |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 当前会话没有绑定手机号（未通过短信登录，`msg` 为「先绑定手机号再操作哦～」），或验证码错误/过期/场景不匹配/已被使用 |
| `500` | Redis 或数据库异常 |

## 用户文件

以下接口都挂在 `/api/v1/user` 登录态路由下，必须携带有效 `mwu_sess_id` cookie。### `POST /api/v1/user/files`

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

**图片安全校验（防「图片藏 JS」）**：当文件被标记为图片（`file_type=image` 或声明 MIME 为 `image/*`）时，服务端会按文件头魔数嗅探真实类型，**仅接受 jpg/png/gif/webp 安全栅格图片**：

- SVG/HTML/JS/XML 等可内嵌脚本的格式一律拒绝（`400`）；
- 「声明 `image/png` 实际为 SVG」等伪装上传同样拒绝；
- 存储 `mime_type` 与 S3 `Content-Type` 使用嗅探出的真实类型，避免浏览器按伪造类型解析。

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
| `400` | 没有上传 `file` 字段或文件名无效；图片未通过安全图片校验（非 jpg/png/gif/webp 或伪装上传，msg：`图片仅支持 jpg/png/gif/webp 格式，不支持 SVG 等可携带脚本的图片`） |
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

通过文件 UUID 获取任意用户 active 文件的 S3 临时访问链接，用于帖子等场景查看/下载其他用户的文件或媒体。**未登录用户可访问公开（`permission=0`）文件**（临时链接使用固定匿名身份标识签名，不绑定任何用户）；登录用户获取的临时链接绑定当前登录用户身份，且只对 active 状态的文件生效。

访问控制规则：

- 文件所有者拉黑了当前查看者时，无论文件是否公开，都返回 `403`，提示「由于对方权限设置，无法查看此文件」。
- 文件分享权限为 `0`（全部人公开）时，任何登录用户可访问，未登录用户也可访问。
- 文件分享权限为 `1`（好友公开）时，仅当前查看者关注了文件所有者才可访问。
- 文件分享权限为 `3`（粉丝公开）时，仅文件所有者关注了当前查看者才可访问。
- 文件分享权限为 `2`（不公开）或不符合上述条件时，返回 `403`，提示「此文件并未公开分享，请检查登录账号」。
- 未登录用户访问非公开（`permission != 0`）文件时，按不存在（`404`）处理，避免泄露文件存在性。

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
| `403` | 文件未公开分享（「此文件并未公开分享，请检查登录账号」），或文件所有者拉黑了查看者（「由于对方权限设置，无法查看此文件」） |
| `404` | 文件不存在或已删除（非 active 状态）；未登录访问非公开文件 |
| `503` | S3 未启用或文件存储服务不可用 |
| `500` | S3 临时链接生成或数据库查询异常 |

## 用户头像

头像使用文件 UUID 表示（`user.avatar` 字段存文件 UUID，不再存 URL）。设置头像复用文件上传接口：先上传图片（`permission=0` 公开），再用返回的 UUID 调用设置头像接口。头像为公开可见文件，获取头像不受拉黑/屏蔽/账号封禁影响。

### `PUT /api/v1/user/avatar`

设置当前登录用户的头像。头像文件必须是当前用户自己的 active 图片文件，且必须公开（`permission=0`），否则返回 `400`。修改头像需要未被封禁头像权限位（`PermAvatar`，bit3），否则返回 `403`，body 中 `code` 为 `40302`。

本接口与 `PUT`/`PATCH /api/v1/user/info` 的 `avatar` 字段共用同一套校验（`service/UserProfile`）：
两者只差在交互方式——本接口适合「上传后立即生效」，资料接口适合「与昵称等字段一起保存」。

#### 请求头

| 名称 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `Cookie` | 是 | 已登录 session 的 `mwu_sess_id` |

#### 请求体

```json
{
  "avatar_uuid": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"
}
```

字段说明：

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
| --- | --- | --- | --- | --- |
| `avatar_uuid` | string | 是 | UUID v4 格式 | 头像文件 UUID，必须是本人 active 图片文件且公开（`permission=0`） |

#### 请求示例

```bash
curl -i -X PUT http://localhost:3000/api/v1/user/avatar \
  -H "Content-Type: application/json" \
  -b cookie.txt \
  -d '{"avatar_uuid":"15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"}'
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "头像设置成功",
  "avatar": {
    "avatar_uuid": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"
  }
}
```

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体不是 JSON、`avatar_uuid` 缺失/格式非法、文件不存在/非 active/不属于当前用户/非图片类型/未公开分享 |
| `401` | 未登录或 session 中没有 `UID` |
| `403` | 头像权限被封禁（`code` 为 `40302`） |
| `500` | 数据库异常 |

### `GET /api/v1/user/users/:uid/avatar`

获取任意用户当前头像的文件 UUID。头像为公开可见文件，不受拉黑/屏蔽/账号封禁影响，无需登录。

#### 请求示例

```bash
curl -i http://localhost:3000/api/v1/user/users/20002/avatar
```

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "获取成功",
  "avatar": {
    "avatar_uuid": "15b3d25d-66cc-4ddc-9949-33c9e84d8c5d"
  }
}
```

`avatar_uuid` 为空字符串表示用户未设置头像。拿到 UUID 后，通过 `GET /api/v1/user/files/:uuid/shared-link` 换取临时访问 URL 展示。

#### 可能的错误

| 状态码 | 场景 |
| --- | --- |
| `400` | `uid` 非法 |
| `404` | 目标用户不存在 |
| `500` | 数据库查询异常 |

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

分页获取目标用户对当前登录用户可见的内容列表。支持按分类拉取动态/文章/小说。

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `category` | string | 否 | 内容分类：`moment` 动态（默认）、`article` 文章、`novel` 小说 |
| `offset` | number | 否 | 偏移量，默认 `0` |
| `limit` | number | 否 | 每页数量，默认 `20`，最大 `100` |

#### 请求示例

```bash
curl -i "http://localhost:3000/api/v1/user/users/20002/contents?category=moment&offset=0&limit=20" \
  -b cookie.txt

curl -i "http://localhost:3000/api/v1/user/users/20002/contents?category=novel&offset=0&limit=20" \
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
      "like": 5,
      "chapter_count": 0
    },
    {
      "id": 100,
      "type": "article",
      "comment": 2,
      "like": 8,
      "chapter_count": 12
    }
  ]
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | number | 内容 ID |
| `type` | string | 内容类型：`moment` 动态 / `article` 文章（`category=novel` 时指小说根文章） |
| `comment` | number | 评论数 |
| `like` | number | 点赞数 |
| `chapter_count` | number | 小说（`category=novel`）已发布章节数；其余分类恒为 `0` |

#### 分类规则

- `moment`：可见性规则与下文一致。
- `article`：仅返回普通文章（`type=0`，非小说、非章节记录）。
- `novel`：仅返回小说根文章（`type=2` 且非章节记录），附已发布章节数；章节内容通过文章详情/分段接口获取。

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

`uid` 等于当前登录用户时（即查询本人资料），响应会额外返回 `user.phone` —— **打码后的**绑定手机号（如 `+86 138****8000`）；查询他人资料时该字段不返回。手机号当前不支持更改，只做只读展示。

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
    "punishment_mask": 2,
    "phone": "+86 138****8000"
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `user` | object | 目标用户资料，字段与 `user` 表一致。若目标用户已注销（`status=3`），`username` 固定返回「已注销的账号」、`bio` 固定返回站长留言文案 |
| `user.block_status` | number | 当前登录用户对目标用户的关系状态（位组合）：`0` 无关系，`1` 拉黑，`2` 屏蔽，`4` 不想看；可组合，如 `3` 表示拉黑+屏蔽 |
| `user.punishment_mask` | number | 目标用户当前生效中的权限封禁位掩码（十进制，按位或合并）：bit0(1) 评论，bit1(2) 发布动态，bit2(4) 私信，bit3(8) 头像，bit4(16) 昵称，bit5(32) 签名，bit6(64) 社交互动，bit7(128) 注销/注册，bit8(256) 上传文件。`0` 表示无生效封禁。前端可据此展示「该用户被禁止发送评论」等提示 |
| `user.phone` | string | **仅当 `:uid` 为本人时返回**，为打码后的绑定手机号；查询他人资料、或会话未绑定手机号时字段省略。响应中不会出现完整手机号 |

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
      "block_status": 0,
      "follow_status": 1
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
| `users[].follow_status` | number | 当前登录用户与列表中每个用户的关注关系（以当前登录用户为视角）：`0` 未关注，`1` 已关注（当前用户关注了对方），`2` 被关注（对方关注了当前用户），`3` 互相关注 |

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

### `DELETE /api/v1/user/follows`

取消关注目标用户（幂等：未关注时直接返回成功）。校验与关注一致。

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
| `target` | number | 是 | 大于 0，且不能等于当前登录用户 ID | 要取消关注的用户 ID |

#### 成功响应

状态码：`200 OK`

```json
{
  "code": 200,
  "msg": "操作成功",
  "target": 20002,
  "action": "unfollow"
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

### 修改个人资料

1. 调用 `GET /api/v1/user/me` 获取当前资料（`username`、`nickname`、`bio`、`gender`、`avatar`、`region`、打码手机号）。
2. 如需更换头像：先 `POST /api/v1/user/files`（`file_type=image`、`permission=0`）拿到文件 `uuid`。
3. 调用 `PATCH /api/v1/user/info` 提交发生变化的字段（未变化的字段不必传），
   可把 `avatar` 与 `nickname` 等字段放在同一次请求里一起提交。
4. 也可以单独调用 `PUT /api/v1/user/avatar` 立即更换头像（例如上传后即时生效的交互）。

### 设置 / 修改密码

1. 调用 `GET /api/v1/user/password` 判断当前是否已设置密码，决定页面文案是「设置密码」还是「修改密码」。
2. 调用 `POST /api/v1/user/password/sms`，服务端把「修改密码」验证码发到当前账号绑定的手机号，返回 `code_uuid`（60 秒冷却）。
3. 用户输入收到的验证码与新密码（两次输入需一致，客户端先自校验）。
4. 调用 `PUT /api/v1/user/password`，成功后服务端会重新生成 session ID，客户端需继续使用新的 `mwu_sess_id`。

## 状态码速查

| 状态码 | 含义 |
| --- | --- |
| `200` | 请求成功 |
| `201` | 创建成功、注册成功或上传成功 |
| `300` | 需要用户选择一个账号继续登录 |
| `400` | 请求格式或参数错误 |
| `403` | 验证失败、账号不匹配、切换目标账号被封禁或账号状态不允许操作 |
| `404` | 注册新账号时，该手机号还没有首个账号 |
| `409` | 唯一字段（账号名/昵称）与已有用户冲突 |
| `413` | 上传文件过大 |
| `429` | 验证码申请过于频繁（同一场景+手机号 60 秒冷却期内重复申请） |
| `503` | 文件存储服务不可用 |
| `500` | 服务端内部错误 |
