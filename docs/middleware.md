# 中间件说明

本文档记录当前可用的中间件、适用场景和路由用法。除 `middleware.Initial` 外，其他中间件都应按接口需要显式挂载。

## 全局基础中间件

```go
middleware.Initial(app, servants)
```

启动 API 时调用一次，当前包含：

- `requestid`：为每个请求生成 `X-Miaoverse-ReqID`。
- `session`：从 `mwu_sess_id` Cookie 读取服务端 session。

当前不再对已登录 session 做 `User-Agent` 强绑定。浏览器版本升级、插件、代理或系统 WebView 都可能导致 UA 变化，强绑定容易误伤正常用户。

UA 解析工具 `util.Validate.ParseUA` 目前值为：

- `pc`：桌面浏览器或无法识别的普通 UA。
- `wap`：手机/平板/H5 浏览器 UA。
- `bot`：爬虫或脚本类 UA。

后续接入原生 App 时，再单独加入 App platform 识别和策略。

## 用户检查中间件

```go
middleware.RequireUser(servants, checks...)
```

用于需要登录用户身份的接口。中间件负责：

- 从 session 读取 `UID`。
- 查询用户信息。
- 执行 `service/UserCheck` 中的检查函数。
- 将检查结果统一转换为 HTTP 状态码和 i18n 文案。
- 将用户上下文写入 `ctx.Locals`，业务可通过 `middleware.CurrentUID(ctx)` 或 `middleware.CurrentUser(ctx)` 获取。

可用检查函数：

```go
UserCheck.AccountActive()
UserCheck.PhoneBound()
UserCheck.PasswordSet()
UserCheck.Certified()
UserCheck.CredentialBound(credType, failure)
```

示例：

```go
userGroup.Post(
    "/some/action",
    middleware.RequireUser(
        servants,
        UserCheck.AccountActive(),
        UserCheck.PhoneBound(),
    ),
    func(c fiber.Ctx) error {
        uid, ok := middleware.CurrentUID(c)
        if !ok {
            return c.SendStatus(fiber.StatusUnauthorized)
        }
        return someHandler(c, servants, uid)
    },
)
```

职责边界：

- `service/UserCheck` 只做检查，返回 `Result`。
- `middleware.RequireUser` 负责返回信息、状态码和 JSON 响应。

## 拉黑校验中间件

```go
middleware.RequireNoBlock(servants, middleware.BlockGuardConfig{...})
```

用于校验当前登录用户与目标用户之间不存在任意一方的拉黑关系（拉黑或被拉黑均拒绝，body `code` 为 `40301`）。

目标用户 ID 的解析方式（三选一）：

| 配置字段 | 说明 |
| --- | --- |
| `PathParam` | 从路径参数解析，如 `/users/:uid/...` |
| `BodyField` | 从请求体字段解析，如 `target` |
| `Resolver` | 自定义解析器，支持先查业务对象再取作者（如点赞/评论/回复场景） |

其他配置：

- `CheckMuteUnwatch`：额外校验目标被自己屏蔽/不想看（关注场景）。
- `AllowSelf`：允许目标为自己（如回复自己的评论、评论自己的动态）；默认 `false`。
- `AllowAnonymous`：允许未登录访问。未登录请求跳过拉黑校验直接放行（Resolver 仍会执行并缓存业务对象），用于公开内容详情等场景；默认 `false`。

Resolver 返回错误时：

- `errBlockTargetBlocked`：目标内容被屏蔽（动态/评论等 `*StatusBlocked` 状态），返回 HTTP `451`，body `code` 为 `45101`（见 `API-errors.md`）。
- `errBlockTargetNotFound`：目标不存在或已删除，返回 `404`。
- 其余错误：返回 `400`。

内置 Resolver：

- `ResolveMomentAuthor`：按 body 中 `moment_id` 查动态，返回作者 ID 并缓存动态对象（`BlockMoment`）。
- `ResolveMomentPathAuthor`：按路径参数 `:id` 查动态，返回作者 ID 并缓存动态对象（`BlockMoment`）。用于动态详情等场景。
- `ResolveCommentAuthor`：按路径参数 `:id` 查评论，返回评论作者 ID 并缓存评论对象（`BlockComment`）。用于「禁止回复拉黑/被拉黑的人的评论」。
- `ResolveCommentMomentAuthor`：按路径参数 `:id` 查评论，沿 target 链上溯到所属动态，返回动态作者 ID，并缓存被回复评论、楼中楼首条评论（`BlockCommentRoot`）与动态对象（`BlockMoment`）。用于回复评论/楼中楼对话场景。

多个 `RequireNoBlock` 可叠加，例如回复评论先校验与动态作者的拉黑关系，再校验与被回复评论作者的拉黑关系。

## Referer 检测中间件

```go
middleware.RequireReferrer(config ...middleware.ReferrerConfig)
```

用于只允许可信页面来源调用某些敏感接口。当前没有挂到任何实际接口，需要按需添加。

默认放行：

- 没有 `Referer` 的请求。
- WAP/H5 移动浏览器 UA。
- `Referer` Host 与当前请求 Host 一致的请求。

示例：

```go
v1.Post(
    "/some/sensitive/action",
    middleware.RequireReferrer(middleware.ReferrerConfig{
        AllowedOrigins: []string{"https://www.miaoverse.com"},
    }),
    func(c fiber.Ctx) error {
        return someHandler(c, servants)
    },
)
```

更详细配置见 [referrer-middleware.md](./referrer-middleware.md)。
