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
	mime2 "mime"
	"net/textproto"
	"path"
	"time"

	liberr "github.com/nabbar/golib/errors"
)

const (
	DateTimeLayout = time.RFC1123Z

	headerMimeVersion = "MIME-Version"
	headerDate        = "Date"
	headerSubject     = "Subject"
)

type mail struct {
	date     time.Time
	attach   []File
	inline   []File
	body     []Body
	charset  string
	subject  string
	headers  textproto.MIMEHeader
	address  *email
	encoding Encoding
	priority Priority
}

func (m *mail) Email() Email {
	return m.address
}

func (m *mail) SetCharset(charset string) {
	m.charset = charset
}

func (m *mail) GetCharset() string {
	return m.charset
}

func (m *mail) SetPriority(p Priority) {
	m.priority = p
}

func (m *mail) GetPriority() Priority {
	return m.priority
}

func (m *mail) SetSubject(subject string) {
	m.subject = subject
}

func (m *mail) GetSubject() string {
	return m.subject
}

func (m *mail) SetEncoding(enc Encoding) {
	m.encoding = enc
}

func (m *mail) GetEncoding() Encoding {
	return m.encoding
}

func (m *mail) SetDateTime(datetime time.Time) {
	m.date = datetime
}

func (m *mail) GetDateTime() time.Time {
	return m.date
}

func (m *mail) SetDateString(layout, datetime string) liberr.Error {
	if t, e := time.Parse(layout, datetime); e != nil {
		return ErrorMailDateParsing.ErrorParent(e)
	} else {
		m.date = t
	}

	return nil
}

func (m *mail) GetDateString() string {
	return m.date.Format(DateTimeLayout)
}

func (m *mail) AddHeader(key string, values ...string) {
	m.headers = m.addHeader(m.headers, key, values...)
}

func (m *mail) addHeader(h textproto.MIMEHeader, key string, values ...string) textproto.MIMEHeader {
	for _, v := range values {
		if v == "" {
			continue
		}

		if len(h.Values(key)) > 0 {
			h.Add(key, v)
		} else {
			h.Set(key, v)
		}
	}
	return h
}

func (m *mail) GetHeader(key string) []string {
	switch key {
	case headerMimeVersion:
		return []string{"1.0"}
	case headerDate:
		return []string{m.GetDateString()}
	case headerSubject:
		return []string{m.GetSubject()}
	case headerPriority:
		return []string{m.priority.headerPriority()}
	case headerMSMailPriority:
		return []string{m.priority.headerMSMailPriority()}
	case headerImportance:
		return []string{m.priority.headerImportance()}
	case headerFrom:
		return []string{m.address.GetFrom()}
	case headerSender:
		return []string{m.address.GetSender()}
	case headerReplyTo:
		return []string{m.address.GetReplyTo()}
	case headerReturnPath:
		return []string{m.address.GetReturnPath()}
	case headerTo:
		return m.address.GetRecipients(RecipientTo)
	case headerCc:
		return m.address.GetRecipients(RecipientCC)
	case headerBcc:
		return m.address.GetRecipients(RecipientBCC)

	}

	return m.headers.Values(key)
}

func (m *mail) GetHeaders() textproto.MIMEHeader {
	h := make(textproto.MIMEHeader)
	h.Set(headerMimeVersion, "1.0")
	h.Set(headerDate, m.GetDateString())
	h.Set(headerSubject, m.GetSubject())

	m.priority.getHeader(func(key string, values ...string) {
		h = m.addHeader(h, key, values...)
	})

	m.address.getHeader(func(key string, values ...string) {
		h = m.addHeader(h, key, values...)
	})

	for k := range m.headers {
		h = m.addHeader(h, k, m.headers.Values(k)...)
	}

	return h
}

func (m *mail) SetBody(ct ContentType, body io.ReadCloser) {
	m.body = make([]Body, 0)
	m.body = append(m.body, NewBody(ct, body))
}

func (m *mail) AddBody(ct ContentType, body io.ReadCloser) {
	for i, b := range m.body {
		if b.contentType == ct {
			m.body[i] = NewBody(ct, body)
			return
		}
	}

	m.body = append(m.body, NewBody(ct, body))
}

func (m *mail) GetBody() []Body {
	return m.body
}

func (m *mail) SetAttachment(name string, mime string, data io.ReadCloser, inline bool) {
	if inline {
		m.inline = make([]File, 0)
		m.inline = append(m.inline, NewFile(name, mime, data))
	} else {
		m.attach = make([]File, 0)
		m.attach = append(m.attach, NewFile(name, mime, data))
	}
}

func (m *mail) AddAttachment(name string, mime string, data io.ReadCloser, inline bool) {
	if inline {

		for i, f := range m.attach {
			if name == f.name {
				m.inline[i] = NewFile(name, mime, data)
				return
			}
		}

		m.inline = append(m.inline, NewFile(name, mime, data))

	} else {

		for i, f := range m.attach {
			if name == f.name {
				m.attach[i] = NewFile(name, mime, data)
				return
			}
		}

		m.attach = append(m.attach, NewFile(name, mime, data))
	}
}

func (m *mail) AttachFile(filepath string, data io.ReadCloser, inline bool) {
	mime := mime2.TypeByExtension(path.Ext(filepath))

	if mime == "" {
		mime = "application/octet-stream"
	}

	m.AddAttachment(path.Base(filepath), mime, data, inline)
}

func (m *mail) GetAttachment(inline bool) []File {
	if inline {
		return m.inline
	}

	return m.attach
}
