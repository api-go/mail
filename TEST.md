# TEST Report

## v1.5.6 验证
- 独立模块解析、全量测试与基准测试均通过。

## 测试覆盖范围
- [x] **基础解析**: 地址解析 (`ParseAddr`, `FormatAddr`)。
- [x] **安全性**: `safe.SafeBuf` 密码加解密及内存擦除流程。
- [x] **IMAP 接收**: 基于 `imapclient` v2 的邮件搜索、UID 追踪及拉取。
- [x] **MIME 解析**: HTML、Text 分离，附件提取，内嵌资源处理。
- [x] **Service 模式**: 后台 `poll` 逻辑，`Start`/`Stop` 生命周期管理。
- [x] **事件回调**: `On` 规则匹配及处理器触发。

## 性能测试 (Benchmark)
*注：由于邮件服务涉及网络 IO，性能受网络环境影响较大。*

| 场景 | 耗时/操作 | 说明 |
| :--- | :--- | :--- |
| IMAP 搜索与解析 | ~800ms | 包含网络往返及 MIME 深度解析 |
| SMTP 发送 | ~1.2s | 取决于附件大小及服务器响应 |
| 内存占用 | 低 | 采用流式处理及即时擦除 |

## 验证结果
- **单元测试**: `go test -v ./...` 通过 (2026-06-21)。
- **并发安全性**: 经过 `sync.RWMutex` 保护，支持多协程下注册 Handler 及发送。
- **JS 错误调用栈**: JS 桥接层改用具名导出并使用 `jsmod.MakeError` 包裹错误，确保 JS 抛出异常时携带准确的 Go 运行时堆栈。

## v1.5.8 发布验证（2026-10-10）

- GOWORK=off go test -v ./... 与 go test -bench=. ./... 通过，依赖均为已发布 tag。
- API 以 GOWORK=off go doc 与导出源码交叉核对（getExports 工具不可用）。

## GitHub migration validation (pending release)

Module paths and internal dependency versions are migrated. Development verification uses GOWORK=off and explicit sibling replacements; remove these replacements before release. See migration validation logs for actual results. Existing unrelated working changes are preserved and require separate review.

API inventory verification: `GOWORK=off go doc` succeeded; output retained in the migration validation directory. Source exports were preserved by the path migration.

## GitHub release verification

GOWORK=off go test -v ./... and go test -bench=. ./... passed using published GitHub dependencies, without local replacements. go mod verify passed.
