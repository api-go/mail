# CHANGELOG

## v1.5.5 (2026-09-05)
- **依赖对齐**: 升级全部 `apigo.cc/go/*` 依赖至当前已发布版本。

## v1.5.4 (2026-07-18)
- **依赖校验修复**: 更新 `go.sum` 为当前可验证的依赖 checksum。

## v1.5.3 (2026-06-21)
- **JS 对齐**: 重构 JS 导出为具名函数，并引入 `jsmod.MakeError` 动态包装错误以获取调用栈。
- **依赖更新**: 升级依赖 `jsmod` 至 `v1.5.3`，`cast` 至 `v1.5.3`，`rand` 至 `v1.5.3`，`encoding` 至 `v1.5.4`，`shell` 至 `v1.5.3`，`safe` 至 `v1.5.2`，`id` 至 `v1.5.4`，`file` 至 `v1.5.5`，`config` 至 `v1.5.3`，`log` 至 `v1.5.8`。

## v1.0.4 (2026-05-31)
- **新增**: 注册到 `jsmod`。支持低代码环境下的邮件发送与接收。
  - 提供 `send`, `recv` 默认快捷操作。
  - 支持通过 `get(name)` 进行多账户切换。
- **摩擦消除**: 使用 `go/cast` 自动处理收件人列表等参数转换，支持 JS 侧传入单字符串或数组。

## v1.0.1 (2026-05-13)
### 重大重构 (Breaking Changes)
- **移除 Must 系列函数**: 移除了 `MustSend`, `MustRecv`。建议结合 `apigo.cc/go/cast` 的 `cast.As` 实现无摩擦调用。
- **配置结构升级**: `MailboxConfig.Password` 更改为 `*safe.SafeBuf` 以提升安全性。

### 新特性 (New Features)
- **多实例管理**: 新增 `GetMailbox(name)` 方法，支持在一个进程内管理多个邮箱账号。
- **Service 模式**: 实现了 `starter.Service` 接口，支持后台自动巡检与定时拉取。
- **事件驱动**: 新增 `On(opt, handler)` 接口，支持基于规则过滤的邮件到达回调。
- **富媒体支持**: 增强了对 HTML 正文、附件以及内嵌资源 (Embeds/CID) 的解析与发送支持。
- **Header 暴露**: 邮件结构现在包含原始 Headers 映射。

### 优化与修复 (Optimizations)
- **安全性**: 密码仅在连接瞬间解密，使用后立即从内存擦除。
- **对齐**: 统一 Receiver 命名为 `m`。
- **IMAP**: 升级至 `imapclient` v2 流式处理，提升解析效率。
