/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package njs_smtp

import (
	"bytes"
	"fmt"
	"net/smtp"
	"time"

	njs_version "github.com/nabbar/golib/njs-version"
)

type sendmail struct {
	to      ListMailAddress
	cc      ListMailAddress
	bcc     ListMailAddress
	from    MailAddress
	replyTo MailAddress

	subject    string
	msgHtml    MailTemplate
	msgText    *bytes.Buffer
	attachment []Attachment

	messageId string
	mailer    string

	testMode bool
}

func (s *sendmail) SetTo(listMail ListMailAddress) {
	s.to = listMail
}

func (s *sendmail) AddTo(mail ...MailAddress) {
	s.to.Add(mail...)
}

func (s *sendmail) SetCc(listMail ListMailAddress) {
	s.cc = listMail
}

func (s *sendmail) AddCc(mail ...MailAddress) {
	s.cc.Add(mail...)
}

func (s *sendmail) SetBcc(listMail ListMailAddress) {
	s.bcc = listMail
}

func (s *sendmail) AddBcc(mail ...MailAddress) {
	s.bcc.Add(mail...)
}

func (s *sendmail) SetFrom(mail MailAddress) {
	s.from = mail
}

func (s *sendmail) SetReplyTo(mail MailAddress) {
	s.replyTo = mail
}

func (s *sendmail) SetSubject(subject string) {
	s.subject = subject
}

func (s *sendmail) SetHtml(m MailTemplate) {
	s.msgHtml = m
}

func (s *sendmail) SetBody(p *bytes.Buffer) {
	s.msgText = p
}

func (s *sendmail) AddAttachment(a ...Attachment) {
	s.attachment = append(s.attachment, a...)
}

func (s *sendmail) SetMessageId(id string) {
	s.messageId = id
}

func (s *sendmail) SetMailer(mailer string) {
	s.mailer = mailer
}

func (s *sendmail) NJSMailer(version njs_version.Version) {
	s.mailer = version.GetHeader()
}

func (s *sendmail) SetTestMode(enable bool) {
	s.testMode = enable
}

func (s sendmail) Clone() (SendMail, error) {
	var (
		la = make([]Attachment, 0)
	)

	for _, a := range s.attachment {
		if a == nil {
			continue
		}

		la = append(la, a.Clone())
	}

	var res = &sendmail{
		to:         nil,
		cc:         nil,
		bcc:        nil,
		from:       nil,
		replyTo:    nil,
		subject:    s.subject,
		msgHtml:    nil,
		msgText:    nil,
		attachment: la,
		messageId:  s.messageId,
		mailer:     s.mailer,
		testMode:   s.testMode,
	}

	if s.msgText != nil {
		res.msgText = bytes.NewBuffer(s.msgText.Bytes())
	}

	if s.msgHtml != nil {
		if tpl, err := s.msgHtml.Clone(); err != nil {
			return nil, err
		} else {
			res.msgHtml = tpl
		}
	}

	if s.to != nil {
		res.to = s.to.Clone()
	}

	if s.cc != nil {
		res.cc = s.cc.Clone()
	}

	if s.bcc != nil {
		res.bcc = s.bcc.Clone()
	}

	if s.from != nil {
		res.from = s.from.Clone()
	}

	if s.replyTo != nil {
		res.replyTo = s.replyTo.Clone()
	}

	return res, nil
}

func (s sendmail) SendSMTP(cli SMTP) (err error) {
	var c *smtp.Client

	defer func(cli *smtp.Client) {
		if cli != nil {
			cli.Quit()
			cli.Close()
		}
	}(c)

	if c, err = cli.Client(); err != nil {
		return
	} else if err = s.Send(c); err != nil {
		c.Reset()
		return err
	}

	return nil
}

func (s sendmail) Send(cli *smtp.Client) (err error) {
	var (
		iod IOData
	)

	defer func() {
		if r := recover(); r != nil && err != nil {
			err = fmt.Errorf("%v, %v", err, r)
		} else if r != nil {
			err = fmt.Errorf("%v", r)
		}

		if cli != nil {
			cli.Reset()
			cli.Quit()
			cli.Close()
		}
	}()

	var ctBody ContentType

	if len(s.attachment) > 0 {
		ctBody = CONTENTTYPE_MIXED
	} else if !s.msgHtml.IsEmpty() {
		ctBody = CONTENTTYPE_MIXED
		//ctBody = CONTENTTYPE_ALTERNATIVE
	} else if s.msgText.Len() > 0 {
		ctBody = CONTENTTYPE_TEXT
	} else {
		return fmt.Errorf("no attachment & no contents")
	}

	if len(s.from.AddressOnly()) < 7 {
		return fmt.Errorf("from address is empty")
	}

	if err = cli.Noop(); err != nil {
		return
	}

	if err = cli.Mail(s.from.String()); err != nil {
		return
	}

	if s.testMode {
		cli.Rcpt(s.from.String())
	} else {
		for _, a := range s.to.Slice() {
			cli.Rcpt(a.String())
		}
		for _, a := range s.cc.Slice() {
			cli.Rcpt(a.String())
		}
		for _, a := range s.bcc.Slice() {
			cli.Rcpt(a.String())
		}
	}

	if iod, err = NewIOData(cli); err != nil {
		return
	}

	iod.Header("From", s.from.String())

	if s.to.IsEmpty() {
		return fmt.Errorf("to address is empty")
	} else {
		iod.Header("To", s.to.String())
	}

	if !s.cc.IsEmpty() {
		iod.Header("Cc", s.cc.String())
	}

	if !s.bcc.IsEmpty() {
		iod.Header("Reply-To", s.replyTo.String())
	}

	if len(s.subject) < 1 {
		return fmt.Errorf("subjetc is empty")
	} else {
		iod.Header("Subject", s.subject)
	}

	if len(s.mailer) < 1 {
		return fmt.Errorf("mailer is empty")
	} else {
		iod.Header("X-Mailer", s.mailer)
	}

	if len(s.messageId) > 0 {
		iod.Header("Message-ID", s.messageId)
	}

	iod.Header("Date", time.Now().Format(time.RFC1123Z))
	iod.Header("Auto-Submitted", "auto-generated")
	iod.Header("MIME-Version", "1.0")

	if ctBody != CONTENTTYPE_TEXT {
		if err = iod.AttachmentStart(ctBody); err != nil {
			return
		}

		for _, a := range s.attachment {
			if err = iod.AttachmentAddFile(a.GetContentType(), a.GetName(), a.GetBuffer()); err != nil {
				return
			}
		}

		if !s.msgHtml.IsEmpty() {
			if err = iod.AttachmentAddBody(s.msgHtml, CONTENTTYPE_HTML); err != nil {
				return
			}
			if err = iod.AttachmentAddBody(s.msgHtml, CONTENTTYPE_TEXT); err != nil {
				return
			}
		}

		return iod.AttachmentEnd()
	} else {
		if err = iod.ContentType(CONTENTTYPE_TEXT, "utf-8"); err != nil {
			return
		} else if err = iod.Bytes(s.msgText.Bytes()); err != nil {
			return
		} else if err = iod.CRLF(); err != nil {
			return
		} else {
			return iod.CRLF()
		}
	}
}

type SendMail interface {
	SetTo(listMail ListMailAddress)
	AddTo(mail ...MailAddress)

	SetCc(listMail ListMailAddress)
	AddCc(mail ...MailAddress)

	SetBcc(listMail ListMailAddress)
	AddBcc(mail ...MailAddress)

	SetFrom(mail MailAddress)
	SetReplyTo(mail MailAddress)

	SetSubject(subject string)
	SetHtml(m MailTemplate)
	SetBody(p *bytes.Buffer)
	AddAttachment(a ...Attachment)

	SetMessageId(id string)
	SetMailer(mailer string)
	NJSMailer(version njs_version.Version)

	SetTestMode(enable bool)

	Clone() (SendMail, error)
	Send(cli *smtp.Client) (err error)
	SendSMTP(cli SMTP) (err error)
}

func NewSendMail() SendMail {
	return &sendmail{
		to:         NewListMailAddress(),
		cc:         NewListMailAddress(),
		bcc:        NewListMailAddress(),
		from:       nil,
		replyTo:    nil,
		subject:    "",
		msgHtml:    nil,
		msgText:    bytes.NewBuffer(make([]byte, 0)),
		attachment: make([]Attachment, 0),
		messageId:  "",
		mailer:     "",
		testMode:   false,
	}
}
