package mail

import (
	"github.com/api-go/cast"
	"github.com/api-go/jsmod"
)

func init() {
	jsmod.Register("mail", map[string]any{
		// 默认邮箱操作 (对齐 PascalCase)
		"Send":       jsSend,
		"Recv":       jsRecv,

		// 多实例获取 (对齐 PascalCase，支持可选参数)
		"Mailbox":    jsMailboxGetter,

		// 辅助工具
		"FormatAddr": jsFormatAddr,
		"ParseAddr":  jsParseAddr,
	})
}

func jsSend(to any, subject, content string, option *SendOption) error {
	var toList []string
	cast.Convert(&toList, to)
	if err := Send(toList, subject, content, option); err != nil {
		return jsmod.MakeError(err)
	}
	return nil
}

func jsRecv(opt *RecvOption) (*RecvResult, error) {
	res, err := Recv(opt)
	if err != nil {
		return nil, jsmod.MakeError(err)
	}
	return res, nil
}

func jsMailboxGetter(name *string) *jsMailbox {
	target := "default"
	if name != nil {
		target = *name
	}
	m := GetMailbox(target)
	if m == nil {
		return nil
	}
	return &jsMailbox{m: m}
}

func jsFormatAddr(name, address string) string {
	return FormatAddr(name, address)
}

func jsParseAddr(addr string) (string, string, error) {
	name, address, err := ParseAddr(addr)
	if err != nil {
		return "", "", jsmod.MakeError(err)
	}
	return name, address, nil
}

// jsMailbox 包装 Mailbox 实例
type jsMailbox struct {
	m *Mailbox
}

func (j *jsMailbox) Send(to any, subject, content string, option *SendOption) error {
	var toList []string
	cast.Convert(&toList, to)
	if err := j.m.Send(toList, subject, content, option); err != nil {
		return jsmod.MakeError(err)
	}
	return nil
}

func (j *jsMailbox) Recv(opt *RecvOption) (*RecvResult, error) {
	res, err := j.m.Recv(opt)
	if err != nil {
		return nil, jsmod.MakeError(err)
	}
	return res, nil
}

func (j *jsMailbox) Health() error {
	if err := j.m.Health(); err != nil {
		return jsmod.MakeError(err)
	}
	return nil
}
