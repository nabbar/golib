package mail

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/nabbar/golib/smtp"

	"github.com/nabbar/golib/ioutils"

	liberr "github.com/nabbar/golib/errors"
	simple "github.com/xhit/go-simple-mail"
)

type Sender interface {
	Close() error
	Send(ctx context.Context, cli smtp.SMTP) liberr.Error
	SendClose(ctx context.Context, cli smtp.SMTP) liberr.Error
}

type sender struct {
	data ioutils.FileProgress
	from string
	rcpt []string
}

func (m *mail) Sender() (snd Sender, err liberr.Error) {
	e := simple.NewMSG()
	f := make([]ioutils.FileProgress, 0)

	switch m.GetPriority() {
	case PriorityHigh:
		e.SetPriority(simple.PriorityHigh)
	case PriorityLow:
		e.SetPriority(simple.PriorityLow)
	}

	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	switch m.GetEncoding() {
	case EncodingNone, EncodingBinary:
		e.Encoding = simple.EncodingNone
	case EncodingBase64:
		e.Encoding = simple.EncodingBase64
	case EncodingQuotedPrintable:
		e.Encoding = simple.EncodingQuotedPrintable
	}

	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	e.Charset = m.GetCharset()

	e.SetSubject(m.GetSubject())
	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	e.SetDate(m.date.Format("2006-01-02 15:04:05 MST"))
	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	if r := m.Email().GetFrom(); len(r) > 0 {
		e.SetFrom(r)
	}
	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	if r := m.Email().GetReplyTo(); len(r) > 0 {
		e.SetReplyTo(r)
	}
	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	if r := m.Email().GetReturnPath(); len(r) > 0 {
		e.SetReturnPath(r)
	}
	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	if r := m.Email().GetSender(); len(r) > 0 {
		e.SetSender(r)
	}
	if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	}

	if r := m.address.GetRecipients(RecipientTo); len(r) > 0 {
		e.AddTo(r...)
		if e.Error != nil {
			return nil, ErrorSenderInit.ErrorParent(e.Error)
		}
	}

	if r := m.address.GetRecipients(RecipientCC); len(r) > 0 {
		e.AddCc(r...)
		if e.Error != nil {
			return nil, ErrorSenderInit.ErrorParent(e.Error)
		}
	}

	if r := m.address.GetRecipients(RecipientBCC); len(r) > 0 {
		e.AddBcc(r...)
		if e.Error != nil {
			return nil, ErrorSenderInit.ErrorParent(e.Error)
		}
	}

	if len(m.attach) > 0 {
		for _, i := range m.attach {
			if t, er := ioutils.NewFileProgressTemp(); er != nil {
				return nil, er
			} else if _, er := t.ReadFrom(i.data); er != nil {
				return nil, ErrorIORead.ErrorParent(er)
			} else if e.AddAttachment(t.FilePath(), i.name); e.Error != nil {
				return nil, ErrorSenderInit.ErrorParent(e.Error)
			} else {
				f = append(f, t)
			}
		}
	}

	if len(m.inline) > 0 {
		for _, i := range m.inline {
			if t, er := ioutils.NewFileProgressTemp(); er != nil {
				return nil, er
			} else if _, er := t.ReadFrom(i.data); er != nil {
				return nil, ErrorIORead.ErrorParent(er)
			} else if e.AddInline(t.FilePath(), i.name); e.Error != nil {
				return nil, ErrorSenderInit.ErrorParent(e.Error)
			} else {
				f = append(f, t)
			}
		}
	}

	if len(m.body) > 0 {
		for i, b := range m.body {
			var (
				buf = bytes.NewBuffer(make([]byte, 0))
				enc = simple.TextPlain
			)

			if b.contentType == ContentHTML {
				enc = simple.TextHTML
			}

			if _, er := buf.ReadFrom(b.body); er != nil {
				return nil, ErrorIORead.ErrorParent(er)
			} else if i > 0 {
				e.AddAlternative(enc, buf.String())
			} else {
				e.SetBody(enc, buf.String())
			}
			if e.Error != nil {
				return nil, ErrorSenderInit.ErrorParent(e.Error)
			}
		}
	}

	s := &sender{}

	defer func() {
		if err != nil || snd == nil {
			_ = s.Close()
		}
	}()

	s.from = m.Email().GetFrom()
	s.rcpt = make([]string, 0)
	s.rcpt = append(s.rcpt, m.Email().GetRecipients(RecipientTo)...)
	s.rcpt = append(s.rcpt, m.Email().GetRecipients(RecipientCC)...)
	s.rcpt = append(s.rcpt, m.Email().GetRecipients(RecipientBCC)...)

	if tmp, err := ioutils.NewFileProgressTemp(); err != nil {
		return nil, err
	} else if _, er := tmp.WriteString(e.GetMessage()); er != nil {
		return nil, ErrorIOWrite.ErrorParent(er)
	} else if e.Error != nil {
		return nil, ErrorSenderInit.ErrorParent(e.Error)
	} else if _, er = tmp.Seek(0, io.SeekStart); er != nil {
		return nil, ErrorIOWrite.ErrorParent(er)
	} else {
		s.data = tmp
		snd = s
	}

	return
}

func (s *sender) SendClose(ctx context.Context, cli smtp.SMTP) liberr.Error {
	defer func() {
		_ = s.Close()
	}()

	if e := s.Send(ctx, cli); e != nil {
		return e
	}

	return nil
}

func (s *sender) Send(ctx context.Context, cli smtp.SMTP) liberr.Error {
	if e := cli.Check(ctx); e != nil {
		return ErrorSmtpClient.ErrorParent(e)
	}

	if len(s.from) < 4 {
		return ErrorParamsEmpty.ErrorParent(fmt.Errorf("parameters 'from' is not valid"))
	} else if len(s.rcpt) < 1 || len(s.rcpt[0]) < 4 {
		return ErrorParamsEmpty.ErrorParent(fmt.Errorf("parameters 'receipient' is not valid"))
	}

	e := cli.Send(ctx, s.from, s.rcpt, s.data)
	if e != nil {
		return e
	}

	if _, err := s.data.Seek(0, io.SeekStart); err != nil {
		return ErrorIOWrite.ErrorParent(err)
	}

	return nil
}

func (s *sender) Close() error {
	return s.data.Close()
}
