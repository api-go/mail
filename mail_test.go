package mail_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/api-go/cast"
	"github.com/api-go/file"
	"github.com/api-go/mail"
	"github.com/api-go/safe"
)

func TestMail(t *testing.T) {
	// 1. 加载配置
	var rawConf map[string]any
	err := file.UnmarshalFile("env.yml", &rawConf)
	if err != nil {
		t.Skip("Failed to load env.yml, skipping real mail test", err)
		return
	}

	mailMap := cast.To[map[string]any](rawConf["mail"])
	mboxMap := cast.To[map[string]any](mailMap["mailbox"])
	defaultMbox := cast.To[map[string]any](mboxMap["default"])
	if defaultMbox == nil {
		t.Skip("Default mailbox not configured in env.yml, skipping real mail test")
		return
	}

	// 准备真实的 MailboxConfig
	var mboxCfg mail.MailboxConfig
	cast.Convert(&mboxCfg, defaultMbox)

	// 手动处理 Password，因为 YAML 中是字符串
	password := cast.String(defaultMbox["password"])
	if password != "" {
		mboxCfg.Password = safe.NewSafeBuf([]byte(password))
	}
	mboxCfg.PollInterval = 5 * time.Second

	fmt.Printf("Using mailbox: %s\n", mboxCfg.Username)

	// 2. 创建 Mailbox 实例
	mbox, err := mail.New(&mboxCfg)
	if err != nil {
		t.Fatalf("Failed to create mailbox: %v", err)
	}

	// 3. 测试解析地址
	name, address, err := mail.ParseAddr("测试工程师 <test@example.com>")
	if err != nil {
		t.Errorf("ParseAddr failed: %v", err)
	}
	if name != "测试工程师" || address != "test@example.com" {
		t.Errorf("Unexpected address parsing: name=%s, address=%s", name, address)
	}

	// 4. 测试 Service 模式与回调
	received := make(chan mail.Mail, 1)
	mbox.On(&mail.RecvOption{Limit: 1}, func(m mail.Mail) {
		fmt.Printf("Callback received mail: %s\n", m.Subject)
		select {
		case received <- m:
		default:
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = mbox.Start(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to start mail service: %v", err)
	}
	defer func() { _ = mbox.Stop(ctx) }()

	select {
	case m := <-received:
		fmt.Printf("Successfully received mail via service: %s\n", m.Subject)
	case <-time.After(20 * time.Second):
		t.Log("No mail received via service in 20s (normal if no new mail)")
	}

	// 5. 测试手动发送 (可选)
	/*
	err = mbox.Send([]string{mboxCfg.Username}, "Service Test", "Hello from service mode", nil)
	if err != nil {
		t.Errorf("Send mail failed: %v", err)
	}
	*/
}
