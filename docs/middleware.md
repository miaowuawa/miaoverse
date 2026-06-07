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

UA 解析工具 `validate.ParseUA` 目前值为：

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
