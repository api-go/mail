# @go/mail

极致精简的高性能邮件客户端，支持 SMTP 发送与 IMAP 接收，集成 `starter.Service` 支持后台自动监听，并提供多实例管理。

## 设计哲学
*   **统一性**: 与 `github.com/api-go/*` 基础设施深度集成（Log, Config, Cast, Safe, Starter）。
*   **服务化**: 支持作为 `starter.Service` 运行，通过规则过滤实现多实例后台自动处理。
*   **多实例**: 支持通过 `Get(name)` 管理多个邮箱账号。
*   **安全性**: 强制使用 `safe.SafeBuf` 保护内存中的密码。
*   **高性能**: 利用 `imapclient` v2 实现流式处理。

## 配置示例 (env.yml)
```yaml
mail:
  mailbox:
    default:
      smtpHost: smtp.126.com
      smtpPort: 465
      imapHost: imap.126.com
      imapPort: 993
      username: user@126.com
      password: your_password
      senderName: 测试工程师
      pollInterval: 1m
```

## API 参考

### 多实例管理
```go
import "github.com/api-go/mail"

// 获取默认实例并注册回调
mbox := mail.GetMailbox("default") 
mbox.On(&mail.RecvOption{Unseen: true}, func(m mail.Mail) {
    fmt.Println("收到新邮件:", m.Subject)
})
```

// 注册为后台服务
starter.Register("mail-default", mbox, 1, 0, 0)
```

### 发送富媒体邮件 (HTML + 附件 + 内嵌图片)
```go
mbox.Send([]string{"target@example.com"}, "报表", "<h1>见附件</h1><img src='cid:chart'>", &mail.SendOption{
    HTML: true,
    Attachments: []mail.Attachment{
        {Name: "data.xlsx", Path: "./data.xlsx"},
    },
    Embeds: []mail.Attachment{
        {Name: "chart.png", Path: "./chart.png", ContentID: "chart"},
    },
})
```

### 接收邮件与解析
```go
// 手动接收
res, _ := mbox.Recv(&mail.RecvOption{Limit: 1})
for _, m := range res.List {
    fmt.Println(m.Subject)
    fmt.Println(m.Text)      // 纯文本正文
    fmt.Println(m.HTML)      // HTML 正文
    fmt.Println(m.Headers)   // 原始请求头 map[string][]string

    // 遍历附件
    for _, a := range m.Attachments {
        fmt.Println("附件:", a.Name, len(a.Data))
    }
}
```

### 规则过滤说明
`RecvOption` 支持丰富的过滤条件：
- `Unseen`: 仅未读
- `Subject`, `From`, `To`, `Body`: 关键字包含过滤
- `Since`, `Before`: 日期过滤
- `Not`: 排除关键字
- `MarkAsRead`: 处理后标记为已读


## v1.5.8 依赖基线

本版本对齐 Log v1.5.17 与发布时模块索引中的上游已发布依赖；公开 API 保持兼容。

### 实时压缩日志（v1.5.10）

正式依赖已按顺序发布，go.mod 不保留本地开发 replace。文件日志默认采用 `.gz` + 同文件级 `.dict`，两份文件须一起管理；通过 logv 或内置 log/logv 读取，NoCompress=true 可保留明文。


### Release v1.5.10 — 2026-11-12

Released GitHub dependencies; no local development replacements. Full tests and benchmarks validated with GOWORK=off.
