# Referer 检测中间件

`middleware.RequireReferrer` 用于只允许可信页面来源调用某些敏感接口。它不会自动挂到全局或任何现有接口，需要在具体路由上按需添加。

HTTP 标准请求头实际名称是 `Referer`，虽然常见英文会写作 `Referrer`。代码中沿用功能名 `RequireReferrer`，读取请求头时使用 `Referer`。

## 放行规则

默认行为：

- 没有 `Referer`：放行。浏览器隐私设置、代理或客户端策略可能会移除这个请求头。
- WAP/H5 移动浏览器 UA：放行。移动端浏览器或 WebView 的来源信息可能不稳定；后续原生 App platform 接入后再单独处理。
- 有 `Referer` 且来源 Host 与当前请求 Host 一致：放行。
- 有 `Referer` 且命中配置的 `AllowedHosts` 或 `AllowedOrigins`：放行。
- 其他情况：返回 `403` 和 `error.invalid_referrer` 文案。

## 可用中间件

```go
middleware.RequireReferrer()
```

默认只接受同 Host 来源，同时允许无 `Referer` 和 WAP/H5 移动浏览器 UA。

```go
middleware.RequireReferrer(middleware.ReferrerConfig{
    AllowedHosts: []string{
        "www.miaoverse.com",
        "admin.miaoverse.com",
    },
})
```

额外允许指定 Host。Host 可以带端口，例如 `localhost:5173`。

```go
middleware.RequireReferrer(middleware.ReferrerConfig{
    AllowedOrigins: []string{
        "https://www.miaoverse.com",
        "http://localhost:5173",
    },
})
```

额外允许指定 Origin。Origin 会同时校验协议和 Host，适合区分 `http` / `https` 或本地开发端口。

```go
middleware.RequireReferrer(middleware.ReferrerConfig{
    RejectSameHost:   false,
    RejectNoReferrer: false,
    RejectMobileUA:   false,
    AllowedOrigins: []string{
        "https://www.miaoverse.com",
    },
})
```

完整配置示例。`Reject*` 开关默认都是 `false`，也就是默认放过同 Host、无 `Referer` 和 WAP/H5 移动浏览器 UA。

## 路由用法

只对需要的接口使用，不要挂到全局：

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

也可以和用户检查中间件叠加：

```go
userGroup.Post(
    "/some/sensitive/action",
    middleware.RequireUser(servants, UserCheck.AccountActive()),
    middleware.RequireReferrer(middleware.ReferrerConfig{
        AllowedOrigins: []string{"https://www.miaoverse.com"},
    }),
    func(c fiber.Ctx) error {
        return someHandler(c, servants)
    },
)
```

当前版本暂时没有把它添加到任何实际接口。
