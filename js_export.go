package mail

import (
	"apigo.cc/go/cast"
	"apigo.cc/go/jsmod"
)

func init() {
	jsmod.Register("mail", map[string]any{
		// 默认邮箱操作 (对齐 PascalCase)
		"Send": func(to any, subject, content string, option *SendOption) error {
			var toList []string
			cast.Convert(&toList, to)
			return Send(toList, subject, content, option)
		},
		"Recv": func(opt *RecvOption) (*RecvResult, error) {
			return Recv(opt)
		},

		// 多实例获取 (对齐 PascalCase，支持可选参数)
		"Mailbox": func(name *string) *jsMailbox {
			target := "default"
			if name != nil {
				target = *name
			}
			m := GetMailbox(target)
			if m == nil {
				return nil
			}
			return &jsMailbox{m: m}
		},

		// 辅助工具
		"FormatAddr": FormatAddr,
		"ParseAddr":  ParseAddr,
	})
}

// jsMailbox 包装 Mailbox 实例
type jsMailbox struct {
	m *Mailbox
}

func (j *jsMailbox) Send(to any, subject, content string, option *SendOption) error {
	var toList []string
	cast.Convert(&toList, to)
	return j.m.Send(toList, subject, content, option)
}

func (j *jsMailbox) Recv(opt *RecvOption) (*RecvResult, error) {
	return j.m.Recv(opt)
}

func (j *jsMailbox) Health() error {
	return j.m.Health()
}
