package mail

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"apigo.cc/go/log"
)

var (
	instances = make(map[string]*Mailbox)
	instMu    sync.RWMutex
)

// Mailbox 邮件客户端实例，实现 starter.Service 接口
type Mailbox struct {
	name     string
	config   *MailboxConfig
	handlers []mailHandler
	logger   *log.Logger
	lastUID  uint32
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

type mailHandler struct {
	opt *RecvOption
	fn  func(Mail)
}

// Attachment 附件或内嵌资源信息
type Attachment struct {
	Name        string // 文件名
	Path        string // 本地路径（发送时使用）
	Data        []byte // 文件数据
	ContentType string // 内容类型
	ContentID   string // 内嵌资源 ID (CID)
}

// Mail 邮件内容，支持富媒体
type Mail struct {
	UID         uint32
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Date        string
	Text        string              // 纯文本正文
	HTML        string              // HTML 正文
	Attachments []Attachment        // 普通附件
	Embeds      []Attachment        // 内嵌资源（如图片）
	Size        int64               // 邮件大小
	Flags       []string            // 邮件标记
	Headers     map[string][]string // 原始请求头
}

// RecvOption 接收选项，支持复杂的过滤规则
type RecvOption struct {
	Tag        uint32
	Mailbox    string
	Unseen     bool
	Read       bool
	Answered   bool
	Flagged    bool
	Deleted    bool
	Since      string
	Before     string
	Limit      uint32
	MarkAsRead bool

	Subject []string
	Search  []string
	Body    []string
	From    []string
	To      []string
	Cc      []string
	Bcc     []string
	Not     []string
}

// RecvResult 接收结果
type RecvResult struct {
	List []Mail
	Tag  uint32
}

// GetMailbox 获取或创建一个指定名称的邮件客户端实例
func GetMailbox(name string, configs ...*MailboxConfig) *Mailbox {
	instMu.RLock()
	m, ok := instances[name]
	instMu.RUnlock()
	if ok {
		return m
	}

	if len(configs) == 0 {
		return nil
	}

	m, _ = New(configs[0])
	if m != nil {
		m.name = name
		instMu.Lock()
		instances[name] = m
		instMu.Unlock()
	}
	return m
}

// New 创建一个邮件客户端实例
func New(cfg *MailboxConfig) (*Mailbox, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	mCfg := *cfg
	// 自动补全逻辑保持不变...
	if mCfg.SMTPHost == "" || mCfg.IMAPHost == "" {
		a := strings.SplitN(mCfg.Username, "@", 2)
		if len(a) == 2 {
			if mCfg.SMTPHost == "" {
				mCfg.SMTPHost = "smtp." + a[1]
			}
			if mCfg.SMTPPort == 0 {
				mCfg.SMTPPort = 465
			}
			if mCfg.IMAPHost == "" {
				mCfg.IMAPHost = "imap." + a[1]
			}
			if mCfg.IMAPPort == 0 {
				mCfg.IMAPPort = 993
			}
		}
	}

	if mCfg.SenderMail == "" {
		mCfg.SenderMail = mCfg.Username
	}

	m := &Mailbox{
		config: &mCfg,
		logger: log.DefaultLogger,
	}

	for _, rule := range cfg.Rules {
		r := rule
		m.On(&r, nil)
	}

	return m, nil
}

// On 在指定实例上注册邮件到达回调
func (m *Mailbox) On(opt *RecvOption, handler func(Mail)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, mailHandler{opt: opt, fn: handler})
}

// Recv 接收邮件
func (m *Mailbox) Recv(opt *RecvOption) (*RecvResult, error) {
	return m.recv(opt)
}

// --- starter.Service Implementation ---

func (m *Mailbox) Start(ctx context.Context, logger *log.Logger) error {
	m.mu.Lock()
	if m.cancel != nil {
		m.mu.Unlock()
		return nil
	}
	if logger != nil {
		m.logger = logger
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	interval := m.config.PollInterval
	if interval <= 0 {
		interval = 1 * time.Minute
	}

	m.wg.Add(1)
	go m.watch(interval)

	m.logger.Info("Mail service started", "name", m.name, "username", m.config.Username)
	return nil
}

func (m *Mailbox) Stop(ctx context.Context) error {
	m.mu.Lock()
	if m.cancel == nil {
		m.mu.Unlock()
		return nil
	}
	m.cancel()
	m.mu.Unlock()
	m.wg.Wait()
	return nil
}

func (m *Mailbox) Health() error {
	if m.config == nil {
		return fmt.Errorf("mailbox not configured")
	}
	return nil
}

// --- Internal Logic ---

func (m *Mailbox) watch(interval time.Duration) {
	defer m.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.poll()
		}
	}
}

func (m *Mailbox) poll() {
	m.mu.RLock()
	handlers := make([]mailHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	for i := range handlers {
		h := &handlers[i]
		opt := *h.opt
		if opt.Tag == 0 {
			opt.Tag = m.lastUID
		}

		res, err := m.Recv(&opt)
		if err != nil {
			m.logger.Error("Mail poll failed", "err", err, "name", m.name)
			continue
		}

		if res.Tag > m.lastUID {
			m.lastUID = res.Tag
		}
		if h.opt.Tag < res.Tag {
			h.opt.Tag = res.Tag
		}

		for _, mail := range res.List {
			if h.fn != nil {
				h.fn(mail)
			}
		}
	}
}
