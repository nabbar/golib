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
	"net/smtp"

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

	testMode bool
	smtpcli  *smtp.Client
}

func (s *sendmail) SetTo(mail ...MailAddress) {
	panic("implement me")
}

func (s *sendmail) AddTo(mail ...MailAddress) {
	panic("implement me")
}

func (s *sendmail) SetCc(mail ...MailAddress) {
	panic("implement me")
}

func (s *sendmail) AddCc(mail ...MailAddress) {
	panic("implement me")
}

func (s *sendmail) SetBcc(mail ...MailAddress) {
	panic("implement me")
}

func (s *sendmail) AddBcc(mail ...MailAddress) {
	panic("implement me")
}

func (s *sendmail) SetFrom(mail MailAddress) {
	panic("implement me")
}

func (s *sendmail) SetReplyTo(mail MailAddress) {
	panic("implement me")
}

func (s *sendmail) SetSubject(subject string) {
	panic("implement me")
}

func (s *sendmail) SetHtml(p []byte, charset string) {
	panic("implement me")
}

func (s *sendmail) SetRich(p []byte, charset string) {
	panic("implement me")
}

func (s *sendmail) SetText(p []byte, charset string) {
	panic("implement me")
}

func (s *sendmail) AddAttachment(a Attachment) {
	panic("implement me")
}

func (s *sendmail) SetMessageId(id string) {
	panic("implement me")
}

func (s *sendmail) SetMailer(mailer string) {
	panic("implement me")
}

func (s *sendmail) NJSMailer(version njs_version.Version) {
	panic("implement me")
}

func (s *sendmail) SetTestMode(enable bool) {
	panic("implement me")
}

func (s sendmail) Clone(cli *smtp.Client) SendMail {
	var la = make([]Attachment, 0)

	for _, a := range s.attachment {
		la = append(la, a.Clone())
	}

	return &sendmail{
		to:         s.to.Clone(),
		cc:         s.cc.Clone(),
		bcc:        s.bcc.Clone(),
		from:       s.from.Clone(),
		replyTo:    s.replyTo.Clone(),
		subject:    s.subject,
		attachment: la,
		testMode:   s.testMode,
		smtpcli:    cli,
	}
}

type SendMail interface {
	SetTo(mail ...MailAddress)
	AddTo(mail ...MailAddress)

	SetCc(mail ...MailAddress)
	AddCc(mail ...MailAddress)

	SetBcc(mail ...MailAddress)
	AddBcc(mail ...MailAddress)

	SetFrom(mail MailAddress)
	SetReplyTo(mail MailAddress)

	SetSubject(subject string)

	SetHtml(p []byte, charset string)
	SetRich(p []byte, charset string)
	SetText(p []byte, charset string)

	AddAttachment(a Attachment)

	SetMessageId(id string)
	SetMailer(mailer string)
	NJSMailer(version njs_version.Version)

	SetTestMode(enable bool)

	Clone(cli *smtp.Client) SendMail
}

func NewSendMail(cli *smtp.Client) SendMail {
	return &sendmail{
		testMode: false,
		smtpcli:  cli,
	}
}
