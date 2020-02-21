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
	"encoding/base64"
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
	forceType  ContentType
	attachment []Attachment

	messageId string
	mailer    string

	testMode bool
}

func (s sendmail) GetToString() string {
	return s.to.String()
}

func (s *sendmail) SetListTo(listMail ListMailAddress) {
	s.to = listMail
}

func (s *sendmail) SetTo(mail ...MailAddress) {
	s.to = NewListMailAddress()
	s.to.Add(mail...)
}

func (s *sendmail) AddTo(mail ...MailAddress) {
	s.to.Add(mail...)
}

func (s *sendmail) SetListCc(listMail ListMailAddress) {
	s.cc = listMail
}

func (s *sendmail) SetCc(mail ...MailAddress) {
	s.cc = NewListMailAddress()
	s.cc.Add(mail...)
}

func (s *sendmail) AddCc(mail ...MailAddress) {
	s.cc.Add(mail...)
}

func (s *sendmail) SetListBcc(listMail ListMailAddress) {
	s.bcc = listMail
}

func (s *sendmail) SetBcc(mail ...MailAddress) {
	s.bcc = NewListMailAddress()
	s.bcc.Add(mail...)
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

func (s *sendmail) HtmlRegisterData(data interface{}) {
	s.msgHtml.RegisterData(data)
}

func (s *sendmail) SetHtml(m MailTemplate) {
	s.msgHtml = m
}

func (s *sendmail) SetBody(p *bytes.Buffer) {
	s.msgText = p
}

func (s *sendmail) SetForceOnly(ct ContentType) {
	s.forceType = ct
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
		forceType:  s.forceType,
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

func (s sendmail) SendSMTP(cli SMTP) (err error, buff *bytes.Buffer) {
	var c *smtp.Client

	defer func(cli *smtp.Client) {
		if cli != nil {
			cli.Quit()
			cli.Close()
		}
	}(c)

	if c, err = cli.Client(); err != nil {
		return
	} else if err, buff = s.Send(c); err != nil {
		c.Reset()
		return err, buff
	} else {
		return
	}
}

func (s sendmail) Send(cli *smtp.Client) (err error, buff *bytes.Buffer) {
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
	} else if s.forceType == CONTENTTYPE_TEXT {
		ctBody = CONTENTTYPE_TEXT
	} else if s.forceType == CONTENTTYPE_HTML && !s.msgHtml.IsEmpty() {
		ctBody = CONTENTTYPE_HTML
	} else if !s.msgHtml.IsEmpty() {
		ctBody = CONTENTTYPE_ALTERNATIVE
	} else if s.msgText.Len() > 0 {
		ctBody = CONTENTTYPE_TEXT
	} else {
		return fmt.Errorf("no attachment & no contents"), nil
	}

	if len(s.from.AddressOnly()) < 7 {
		return fmt.Errorf("from address is empty"), nil
	}

	if err = cli.Noop(); err != nil {
		return
	}

	if err = cli.Mail(s.from.String()); err != nil {
		return
	}

	if s.testMode {
		if err = cli.Rcpt(s.from.String()); err != nil {
			return
		}
	} else {
		for _, a := range s.to.Slice() {
			if err = cli.Rcpt(a.String()); err != nil {
				return
			}
		}
		for _, a := range s.cc.Slice() {
			if err = cli.Rcpt(a.String()); err != nil {
				return
			}
		}
		for _, a := range s.bcc.Slice() {
			if err = cli.Rcpt(a.String()); err != nil {
				return
			}
		}
	}

	if iod, err = NewIOData(cli); err != nil {
		return
	}

	if err = iod.Header("From", s.from.String()); err != nil {
		return
	}

	if s.to.IsEmpty() {
		return fmt.Errorf("to address is empty"), nil
	} else if err = iod.Header("To", s.to.String()); err != nil {
		return
	}

	if !s.cc.IsEmpty() {
		if err = iod.Header("Cc", s.cc.String()); err != nil {
			return
		}
	}

	if s.replyTo != nil && s.replyTo.AddressOnly() != "" {
		if err = iod.Header("Reply-To", s.replyTo.String()); err != nil {
			return
		}
		if err = iod.Header("Return-Path", s.replyTo.String()); err != nil {
			return
		}
	} else {
		if err = iod.Header("Reply-To", s.from.String()); err != nil {
			return
		}
		if err = iod.Header("Return-Path", s.from.String()); err != nil {
			return
		}
	}

	if len(s.subject) < 1 {
		return fmt.Errorf("subjetc is empty"), nil
	} else {
		var (
			b = []byte(s.subject)
			c = make([]byte, base64.StdEncoding.EncodedLen(len(b)))
		)

		// convert subjet in base64 for utf8 char
		base64.StdEncoding.Encode(c, b)
		if err = iod.Header("Subject", fmt.Sprintf("=?utf-8?B?%s?=", string(c))); err != nil {
			return
		}
	}

	if len(s.mailer) < 1 {
		return fmt.Errorf("mailer is empty"), nil
	} else {
		if err = iod.Header("X-Mailer", s.mailer); err != nil {
			return
		}
	}

	if len(s.messageId) > 0 {
		if err = iod.Header("Message-ID", s.messageId); err != nil {
			return
		}
	}

	if err = iod.Header("Date", time.Now().Format(time.RFC1123Z)); err != nil {
		return
	}

	if err = iod.Header("MIME-Version", "1.0"); err != nil {
		return
	}

	if ctBody == CONTENTTYPE_TEXT {
		if s.msgText.Len() > 0 {
			if err = iod.String(s.msgText.String()); err != nil {
				return
			}
		} else if s.msgHtml.IsEmpty() {
			return fmt.Errorf("empty content mail"), nil
		} else if err = iod.AttachmentAddBody(s.msgHtml, CONTENTTYPE_TEXT); err != nil {
			return
		}
	} else if ctBody == CONTENTTYPE_HTML {
		if s.msgHtml.IsEmpty() {
			return fmt.Errorf("empty content mail"), nil
		} else if err = iod.AttachmentAddBody(s.msgHtml, CONTENTTYPE_HTML); err != nil {
			return
		}
	} else {
		if err = iod.AttachmentStart(ctBody); err != nil {
			return
		}

		if !s.msgHtml.IsEmpty() {
			if err = iod.AttachmentAddBody(s.msgHtml, CONTENTTYPE_HTML); err != nil {
				return
			}
			if err = iod.AttachmentAddBody(s.msgHtml, CONTENTTYPE_TEXT); err != nil {
				return
			}
		}

		for _, a := range s.attachment {
			if err = iod.AttachmentAddFile(a.GetContentType(), a.GetName(), a.GetBuffer()); err != nil {
				return
			}
		}

		if err = iod.AttachmentEnd(); err != nil {
			return
		}
	}

	err = iod.Send()
	buff = iod.GetBuffer()
	return
}

type SendMail interface {
	SetListTo(listMail ListMailAddress)
	GetToString() string
	SetTo(mail ...MailAddress)
	AddTo(mail ...MailAddress)

	SetListCc(listMail ListMailAddress)
	SetCc(mail ...MailAddress)
	AddCc(mail ...MailAddress)

	SetListBcc(listMail ListMailAddress)
	SetBcc(mail ...MailAddress)
	AddBcc(mail ...MailAddress)

	SetFrom(mail MailAddress)
	SetReplyTo(mail MailAddress)

	SetSubject(subject string)
	SetHtml(m MailTemplate)
	HtmlRegisterData(data interface{})
	SetBody(p *bytes.Buffer)
	SetForceOnly(ct ContentType)
	AddAttachment(a ...Attachment)

	SetMessageId(id string)
	SetMailer(mailer string)
	NJSMailer(version njs_version.Version)

	SetTestMode(enable bool)

	Clone() (SendMail, error)
	Send(cli *smtp.Client) (err error, buff *bytes.Buffer)
	SendSMTP(cli SMTP) (err error, buff *bytes.Buffer)
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
		forceType:  CONTENTTYPE_ALTERNATIVE,
		attachment: make([]Attachment, 0),
		messageId:  "",
		mailer:     "",
		testMode:   false,
	}
}
