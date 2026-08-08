# AGENTS.md

本文档是本项目的代码更新规范。后续所有自动化编码助手和协作者在修改代码前，都应先阅读并遵守这些规则。

## 总原则

- 先读现有代码结构，再动手修改。
- 保持改动边界清晰，不把 API、service、DAO、DTO、util 的职责混在一个文件里。
- 能复用已有模块就复用，不为一次性逻辑随手在 handler 里堆函数。
- 新增功能必须同时考虑配置、数据库、DTO、服务逻辑、路由、文档和测试。
- 不回滚或覆盖他人已有改动；遇到脏工作区时，只改本任务相关文件。
- 修改 Go 代码后必须运行 `gofmt`。
- 完成后至少运行 `go test ./...`，若无法运行，要说明原因。

## 目录职责

### `api/`

只放 HTTP 路由和 handler 编排逻辑。

handler 可以做：

- 读取 path/query/body/form 参数。
- 调用 middleware 获取登录态。
- 调用 service/DAO 完成业务。
- 根据结果返回响应。

handler 不应该做：

- 拼接对象 key、清洗文件名等业务 helper。
- MIME 类型归类等可复用工具逻辑。
- 数据库查询细节。
- 手写一堆重复的 `badRequest`、`serverError` 响应函数。
- 构造复杂 DTO 转换逻辑。

### `service/`

放业务领域逻辑。和具体业务对象强相关的 helper 应放在这里。

示例：

- 用户文件对象 key 构造。
- 用户文件名清洗。
- DAO model 到响应 DTO 的业务转换。
- 需要组合多个 DAO、S3、Redis 操作的业务流程。

命名建议：

- 用户文件相关逻辑放 `service/UserFile`。
- 用户 session 相关逻辑放 `service/UserSession`。
- 用户检查相关逻辑放 `service/UserCheck`。
- S3 存储底层能力放 `service/s3`。

### `dao/`

只放数据库访问逻辑。

DAO 可以做：

- Create、Query、Update、Delete。
- 按业务条件封装查询，例如按用户和 UUID 查询文件。
- 事务。

DAO 不应该做：

- HTTP 响应。
- S3 上传。
- 参数绑定。
- i18n 文案。
- 复杂业务决策，除非它本质上是数据库事务一致性的一部分。

### `model/dao/`

只放数据库模型结构体、表名。

不要在 model 里写数据库访问方法，也不要写 HTTP 响应逻辑。

### `consts/`

所有业务/领域常量集中放在 `consts` 包，按域拆文件：

- `user.go`：账号状态、权限位（Perm*）、惩罚状态、资料长度限制。
- `moment.go` / `comment.go` / `interact.go` / `file.go` / `notify.go`：各内容/互动域的状态、类型、权限、限制常量。
- `key.go`：中间件 Locals key、Session key、Redis key 前缀。
- `s3.go` / `misc.go`：S3 时长、时间格式、UA 类型、默认值等杂项。

常量命名带域前缀（如 `MomentStatusNormal`、`CommentTargetComment`、`UserStatusBanned`），后续统一修改限制或状态值时只改 `consts` 即可。

例外：与包内自定义类型强绑定的常量（如 `i18n.MessageKey`、`UserCheck.Failure`、`UserBlock.BlockType*`）保留在类型定义所在包，避免 `consts` 反向依赖业务包。

### `model/dto/`

放请求和响应 DTO。

- 请求 DTO 按业务域分目录，例如 `model/dto/file/uploadreq`。
- 响应结构体放 `model/dto/resp`。
- 面向用户返回的统一响应 helper 也放在 `model/dto/resp`，避免每个 handler 自己写一套。

例如：

- `resp.BadRequest(ctx)`
- `resp.Unauthorized(ctx)`
- `resp.ServerError(ctx)`
- `resp.FileUploaded(ctx, file)`
- `resp.FileTempLink(ctx, uuid, url, expiresAt)`

### `util/`

只放和业务实体无关、可复用的纯工具。

适合放 util 的逻辑：

- MIME 类型归类。
- UUID、数字、User-Agent 等通用校验。
- 通用字符串构造。
- 通用加密/哈希工具。

不适合放 util 的逻辑：

- 用户文件 object key 构造。
- 用户资料转换。
- 和某个业务表强绑定的逻辑。

### `config` 和 `model/server/conf`

新增配置项时必须同步修改：

- `model/server/conf/conf.go`
- 如需默认值，新增或修改 `model/server/conf/*.go`
- `config.template.yaml`
- 本地 `config` 模板文件，如仓库正在维护它
- `docs/配置文件指南.md`

配置项应写清楚单位，例如秒、分钟、字节。

### `scripts/db.sql`

新增数据库字段或表时必须同步更新 SQL。

如果是已有表新增字段，最终说明里要提醒需要执行迁移或 `ALTER TABLE`。

## 响应规范

- 通用 JSON 响应通过 `model/dto/resp` 集中返回。
- 不要在多个 handler 中重复定义 `badRequest`、`unauthorized`、`serverError`。
- 新增用户可见错误时，要同步：
  - `service/i18n/messages.go`
  - `locales/zh-CN.toml`
- 有业务含义的成功响应，也应尽量在 `model/dto/resp` 中集中构造。

## 文件上传规范

文件上传相关代码遵守以下边界：

- handler 只负责读取 multipart 文件、登录态、调用 DAO/S3/service。
- 文件名清洗、object key 构造、文件响应 DTO 转换放 `service/UserFile`。
- MIME 类型归类放 `util/filetype`。
- 数据库创建和查询放 `dao/user/file.go`。
- 文件数据库模型放 `model/dao/user/file.go`。
- 上传大小限制从配置读取，不要写死在 handler。

相同 hash 文件处理：

- 上传前先计算文件 SHA-256 hash。
- 如果已有 active 文件 hash 相同，复用已有 `object_key` 和 `file_url`，只创建新的文件记录。
- 不要重复上传相同内容到 S3。
- 若需要彻底解决并发重复上传，应增加数据库唯一约束、对象表或分布式锁，而不是只靠 handler 内查询。

## S3 规范

- 底层 S3 操作放 `service/s3`。
- 用户侧临时签名和临时链接能力放在 S3 service 内，API 不直接拼签名。
- 临时链接必须绑定登录用户身份和文件归属校验。
- 不要把 S3 object key 当作公开用户参数暴露。

## API 文档规范

新增或修改接口时，必须更新相关文档：

- `docs/API-xxx.md`
- 如涉及配置，同步更新 `docs/配置文件指南.md`

文档至少说明：

- 路由和方法。
- 是否需要登录。
- Content-Type。
- 请求参数。
- 成功响应。
- 主要错误状态码。

## 测试和验证

每次 Go 代码改动后：

```bash
gofmt -w <changed-go-files>
go test ./...
```

如果只改文档，可以不跑测试，但最终说明中要写明没有跑测试的原因。

## 禁止事项

- 不要为了省事把所有 helper 都塞进 handler。
- 不要新增重复的局部响应函数。遵守DRY原则。
- 不要把业务逻辑放进 DTO。
- 不要把 HTTP 状态码和 i18n 文案写进 DAO。
- 不要在未确认的情况下删除或回滚用户已有改动。
- 不要引入新依赖后忘记更新 `go.mod` 的 direct/indirect 状态。
- 不要只改代码不改文档。
