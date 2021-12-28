/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package mail

import (
	"io"
	"net/textproto"
	"time"

	liberr "github.com/nabbar/golib/errors"
)

type Mail interface {
	Clone() Mail

	SetCharset(charset string)
	GetCharset() string

	SetPriority(p Priority)
	GetPriority() Priority

	SetSubject(subject string)
	GetSubject() string

	SetEncoding(enc Encoding)
	GetEncoding() Encoding

	SetDateTime(datetime time.Time)
	GetDateTime() time.Time
	SetDateString(layout, datetime string) liberr.Error
	GetDateString() string

	AddHeader(key string, values ...string)
	GetHeader(key string) []string
	GetHeaders() textproto.MIMEHeader

	SetBody(ct ContentType, body io.ReadCloser)
	AddBody(ct ContentType, body io.ReadCloser)
	GetBody() []Body

	SetAttachment(name string, mime string, data io.ReadCloser, inline bool)
	AddAttachment(name string, mime string, data io.ReadCloser, inline bool)
	AttachFile(filepath string, data io.ReadCloser, inline bool)
	GetAttachment(inline bool) []File

	Email() Email

	Sender() (Sender, liberr.Error)
}

func New() Mail {
	m := &mail{
		headers:  make(textproto.MIMEHeader),
		charset:  "UTF-8",
		encoding: EncodingNone,
		address: &email{
			from:       "",
			sender:     "",
			replyTo:    "",
			returnPath: "",
			to:         make([]string, 0),
			cc:         make([]string, 0),
			bcc:        make([]string, 0),
		},
		attach: make([]File, 0),
		inline: make([]File, 0),
		body:   make([]Body, 0),
	}

	m.headers.Set("MIME-Version", "1.0")

	return m
}

func (m *mail) Clone() Mail {
	return &mail{
		date:    m.date,
		attach:  m.attach,
		inline:  m.inline,
		body:    m.body,
		charset: m.charset,
		subject: m.subject,
		headers: m.headers,
		address: &email{
			from:       m.address.from,
			sender:     m.address.sender,
			replyTo:    m.address.replyTo,
			returnPath: m.address.returnPath,
			to:         m.address.to,
			cc:         m.address.cc,
			bcc:        m.address.bcc,
		},
		encoding: m.encoding,
		priority: m.priority,
	}
}

type Email interface {
	SetFrom(mail string)
	GetFrom() string

	SetSender(mail string)
	GetSender() string

	SetReplyTo(mail string)
	GetReplyTo() string

	SetReturnPath(mail string)
	GetReturnPath() string

	SetRecipients(rt recipientType, rcpt ...string)
	AddRecipients(rt recipientType, rcpt ...string)
	GetRecipients(rt recipientType) []string
}
