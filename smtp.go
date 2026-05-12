package mail

import (
	"crypto/tls"
	"fmt"
	"io"

	"apigo.cc/go/file"
	"github.com/jaytaylor/html2text"
	"gopkg.in/gomail.v2"
)

// SendOption 发送选项
type SendOption struct {
	Cc          []string
	Bcc         []string
	HTML        bool
	Attachments []Attachment
	Embeds      []Attachment
}

// Send 发送邮件
func (m *Mailbox) Send(to []string, subject, content string, option *SendOption) error {
	if m.config == nil || m.config.Username == "" {
		return fmt.Errorf("mailbox not configured")
	}

	msg := gomail.NewMessage()
	if m.config.SenderName != "" {
		msg.SetHeader("From", msg.FormatAddress(m.config.Username, m.config.SenderName))
	} else {
		msg.SetHeader("From", m.config.Username)
	}

	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)

	useHTML := false
	if option != nil {
		if len(option.Cc) > 0 {
			msg.SetHeader("Cc", option.Cc...)
		}
		if len(option.Bcc) > 0 {
			msg.SetHeader("Bcc", option.Bcc...)
		}
		for _, attachment := range option.Attachments {
			if attachment.Path != "" {
				if file.Exists(attachment.Path) {
					msg.Attach(attachment.Path)
				} else {
					m.logger.Error("Mail attachment file not found", "path", attachment.Path)
				}
			} else if len(attachment.Data) > 0 {
				name := attachment.Name
				if name == "" {
					name = "attachment"
				}
				msg.Attach(name, gomail.SetCopyFunc(func(w io.Writer) error {
					_, err := w.Write(attachment.Data)
					return err
				}))
			}
		}
		for _, embed := range option.Embeds {
			if embed.Path != "" {
				if file.Exists(embed.Path) {
					msg.Embed(embed.Path)
				} else {
					m.logger.Error("Mail embed file not found", "path", embed.Path)
				}
			} else if len(embed.Data) > 0 {
				name := embed.Name
				if name == "" {
					name = "image"
				}
				msg.Embed(name, gomail.SetCopyFunc(func(w io.Writer) error {
					_, err := w.Write(embed.Data)
					return err
				}))
			}
		}
		useHTML = option.HTML
	}

	if useHTML {
		msg.SetBody("text/html", content)
		if textContent, err := html2text.FromString(content, html2text.Options{TextOnly: true, PrettyTables: true}); err == nil {
			msg.AddAlternative("text/plain", textContent)
		} else {
			m.logger.Warning("Convert HTML to text failed", "err", err)
			msg.AddAlternative("text/plain", content)
		}
	} else {
		msg.SetBody("text/plain", content)
	}

	password := ""
	if m.config.Password != nil {
		p := m.config.Password.Open()
		password = p.String()
		defer p.Close()
	}

	dialer := gomail.NewDialer(m.config.SMTPHost, m.config.SMTPPort, m.config.Username, password)
	if m.config.SMTPPort == 465 {
		dialer.SSL = true
	} else {
		dialer.TLSConfig = &tls.Config{
			ServerName:         m.config.SMTPHost,
			InsecureSkipVerify: false,
		}
	}

	err := dialer.DialAndSend(msg)
	if err != nil {
		m.logger.Error("Send mail failed", "err", err, "to", to, "subject", subject)
	} else {
		m.logger.Info("Send mail success", "to", to, "subject", subject)
	}
	return err
}
