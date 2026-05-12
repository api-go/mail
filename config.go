package mail

import (
	"time"

	"apigo.cc/go/safe"
)

// Config 邮件模块配置
type Config struct {
	Mailbox map[string]*MailboxConfig
}

// MailboxConfig 单个邮箱配置
type MailboxConfig struct {
	SMTPHost     string
	SMTPPort     int
	IMAPHost     string
	IMAPPort     int
	Username     string
	Password     *safe.SafeBuf
	SenderName   string
	SenderMail   string
	PollInterval time.Duration // 轮询间隔，用于 Service 模式
	Rules        []RecvOption  // 默认监听规则
}
