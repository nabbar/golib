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
	attachment []Attachment

	messageId string
	mailer    string

	msgText *bytes.Buffer
	msgHtml *bytes.Buffer
	msgRich *bytes.Buffer

	charText string
	charHtml string
	charRich string

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

func (s *sendmail) SetHtml(p []byte, charset string) {
	s.msgHtml = bytes.NewBuffer(p)
	s.charHtml = charset
}

func (s *sendmail) SetRich(p []byte, charset string) {
	s.msgRich = bytes.NewBuffer(p)
	s.charRich = charset
}

func (s *sendmail) SetText(p []byte, charset string) {
	s.msgText = bytes.NewBuffer(p)
	s.charText = charset
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

func (s sendmail) Clone() SendMail {
	var la = make([]Attachment, 0)

	for _, a := range s.attachment {
		la = append(la, a.Clone())
	}

	return &sendmail{
		to:  s.to.Clone(),
		cc:  s.cc.Clone(),
		bcc: s.bcc.Clone(),

		from:    s.from.Clone(),
		replyTo: s.replyTo.Clone(),

		subject:    s.subject,
		attachment: la,

		messageId: s.messageId,
		mailer:    s.mailer,

		msgText: bytes.NewBuffer(s.msgText.Bytes()),
		msgHtml: bytes.NewBuffer(s.msgHtml.Bytes()),
		msgRich: bytes.NewBuffer(s.msgRich.Bytes()),

		charText: s.charText,
		charHtml: s.charHtml,
		charRich: s.charRich,

		testMode: s.testMode,
	}
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
			if e := cli.Close(); err != nil && e != nil {
				err = fmt.Errorf("%v, %v", err, e)
			} else if e != nil {
				err = fmt.Errorf("%v", e)
			}
		}
	}()

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

	/*
	 @TODO : send contents
	 - check html
	 - check rich
	 - check text
	 - check attachment
	 - if no one => error
	 - if 2 types => multipart
	 - if attachment => mixed
	 - if only one => direct
	*/

	return nil
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

	SetHtml(p []byte, charset string)
	SetRich(p []byte, charset string)
	SetText(p []byte, charset string)

	AddAttachment(a ...Attachment)

	SetMessageId(id string)
	SetMailer(mailer string)
	NJSMailer(version njs_version.Version)

	SetTestMode(enable bool)

	Clone() SendMail
	Send(cli *smtp.Client) error
}

func NewSendMail() SendMail {
	return &sendmail{
		testMode: false,
	}
}
