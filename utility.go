package mail

import (
	"net/mail"
	"runtime"
)

func runtimeOS() string {
	return runtime.GOOS
}

// ParseAddr 解析邮件地址
func ParseAddr(addr string) (name, address string, err error) {
	r, err := mail.ParseAddress(addr)
	if err != nil {
		return "", "", err
	}
	return r.Name, r.Address, nil
}

// FormatAddr 格式化邮件地址
func FormatAddr(name, address string) string {
	addr := &mail.Address{
		Name:    name,
		Address: address,
	}
	return addr.String()
}
