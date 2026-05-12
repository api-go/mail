# CHANGELOG

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
