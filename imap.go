package mail

import (
	"bytes"
	"fmt"
	"runtime"
	"time"

	"apigo.cc/go/cast"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/jhillyerd/enmime"
)

// recv 内部接收邮件实现
func (m *Mailbox) recv(opt *RecvOption) (*RecvResult, error) {
	if m.config == nil || m.config.Username == "" {
		return nil, fmt.Errorf("mailbox not configured")
	}

	addr := fmt.Sprintf("%s:%d", m.config.IMAPHost, m.config.IMAPPort)
	c, err := imapclient.DialTLS(addr, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = c.Logout().Wait()
		_ = c.Close()
	}()

	// 发送客户端标识
	_, _ = c.ID(&imap.IDData{
		Name:    "Apigo Mail Client",
		Version: "1.0.1",
		Vendor:  "apigo.cc",
		OS:      runtime.GOOS,
	}).Wait()

	password := ""
	if m.config.Password != nil {
		p := m.config.Password.Open()
		password = p.String()
		defer p.Close()
	}

	if err := c.Login(m.config.Username, password).Wait(); err != nil {
		return nil, err
	}

	mailboxName := "INBOX"
	if opt != nil && opt.Mailbox != "" {
		mailboxName = opt.Mailbox
	}

	_, err = c.Select(mailboxName, nil).Wait()
	if err != nil {
		return nil, err
	}

	limit := uint32(20)
	if opt != nil && opt.Limit > 0 {
		limit = opt.Limit
	}

	startUID := uint32(1)
	if opt != nil && opt.Tag > 0 {
		startUID = opt.Tag + 1
	}

	criteria := &imap.SearchCriteria{}
	if startUID > 1 {
		criteria.UID = []imap.UIDSet{{imap.UIDRange{Start: imap.UID(startUID), Stop: 0}}}
	}

	if opt != nil {
		if opt.Unseen {
			criteria.NotFlag = append(criteria.NotFlag, imap.FlagSeen)
		}
		if opt.Read {
			criteria.Flag = append(criteria.Flag, imap.FlagSeen)
		}
		if opt.Answered {
			criteria.Flag = append(criteria.Flag, imap.FlagAnswered)
		}
		if opt.Flagged {
			criteria.Flag = append(criteria.Flag, imap.FlagFlagged)
		}
		if opt.Deleted {
			criteria.Flag = append(criteria.Flag, imap.FlagDeleted)
		}
		if opt.Since != "" {
			criteria.Since = cast.To[time.Time](opt.Since)
		}
		if opt.Before != "" {
			criteria.Before = cast.To[time.Time](opt.Before)
		}
		if len(opt.Search) > 0 {
			criteria.Text = opt.Search
		}
		if len(opt.Body) > 0 {
			criteria.Body = opt.Body
		}
		for _, word := range opt.Subject {
			criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "Subject", Value: word})
		}
		for _, word := range opt.From {
			criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "From", Value: word})
		}
		for _, word := range opt.To {
			criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "To", Value: word})
		}
		if len(opt.Not) > 0 {
			criteria.Not = append(criteria.Not, imap.SearchCriteria{Text: opt.Not})
		}
	}

	searchData, err := c.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, err
	}

	var uidSet imap.UIDSet
	var maxUID uint32
	count := uint32(0)
	if all, ok := searchData.All.(imap.UIDSet); ok {
		for _, set := range all {
			for uid := set.Start; uid <= set.Stop; uid++ {
				uidSet = append(uidSet, imap.UIDRange{Start: uid, Stop: uid})
				if uint32(uid) > maxUID {
					maxUID = uint32(uid)
				}
				count++
				if count >= limit {
					goto FETCH
				}
				if set.Stop == 0 { // Infinite range
					break
				}
			}
		}
	}

FETCH:
	rr := &RecvResult{List: []Mail{}, Tag: maxUID}
	if len(uidSet) == 0 {
		return rr, nil
	}

	fetchData, err := c.Fetch(uidSet, &imap.FetchOptions{Envelope: true, BodySection: []*imap.FetchItemBodySection{{}}}).Collect()
	if err != nil {
		return rr, err
	}

	for _, msg := range fetchData {
		mail := Mail{
			UID:     uint32(msg.UID),
			Subject: msg.Envelope.Subject,
			To:      formatAddressList(msg.Envelope.To),
			Cc:      formatAddressList(msg.Envelope.Cc),
			Bcc:     formatAddressList(msg.Envelope.Bcc),
			Date:    msg.Envelope.Date.Format(time.RFC3339),
			Size:    msg.RFC822Size,
			Flags:   formatFlags(msg.Flags),
		}
		if len(msg.Envelope.From) > 0 {
			mail.From = formatAddress(&msg.Envelope.From[0])
		}

		for _, section := range msg.BodySection {
			if len(section.Bytes) > 0 {
				env, err := enmime.ReadEnvelope(bytes.NewReader(section.Bytes))
				if err == nil {
					mail.Text = env.Text
					mail.HTML = env.HTML
					mail.Headers = env.Root.Header
					for _, attach := range env.Attachments {
						mail.Attachments = append(mail.Attachments, Attachment{
							Name:        attach.FileName,
							Data:        attach.Content,
							ContentType: attach.ContentType,
						})
					}
					for _, embed := range env.Inlines {
						mail.Embeds = append(mail.Embeds, Attachment{
							Name:        embed.FileName,
							Data:        embed.Content,
							ContentType: embed.ContentType,
							ContentID:   embed.ContentID,
						})
					}
				}
			}
		}
		rr.List = append(rr.List, mail)
	}

	if opt != nil && opt.MarkAsRead && len(uidSet) > 0 {
		_ = c.Store(uidSet, &imap.StoreFlags{
			Op:    imap.StoreFlagsAdd,
			Flags: []imap.Flag{imap.FlagSeen},
		}, nil).Close()
	}

	return rr, nil
}

func formatAddress(addr *imap.Address) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s>", addr.Name, addr.Addr())
	}
	return addr.Addr()
}

func formatAddressList(addrs []imap.Address) []string {
	res := make([]string, len(addrs))
	for i, addr := range addrs {
		res[i] = formatAddress(&addr)
	}
	return res
}

func formatFlags(flags []imap.Flag) []string {
	res := make([]string, len(flags))
	for i, f := range flags {
		res[i] = string(f)
	}
	return res
}
