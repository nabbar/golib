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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/smtp"

	. "github.com/nabbar/golib/njs-errors"
)

type ContentType uint8

const (
	CONTENTTYPE_MIXED ContentType = iota
	CONTENTTYPE_ALTERNATIVE
	CONTENTTYPE_HTML
	CONTENTTYPE_TEXT
)

func (c ContentType) String() string {
	switch c {
	case CONTENTTYPE_MIXED:
		return "multipart/mixed"
	case CONTENTTYPE_ALTERNATIVE:
		return "multipart/alternative"
	case CONTENTTYPE_HTML:
		return "text/html"
	case CONTENTTYPE_TEXT:
		return "text/plain"
	}

	return ""
}

type ioData struct {
	p *bytes.Buffer
	w io.WriteCloser
	b string
}

func (i *ioData) getBoundary() (string, Error) {
	if i.b == "" {
		var buf [30]byte

		_, err := io.ReadFull(rand.Reader, buf[:])

		if err != nil {
			return "", RAND_READER.ErrorParent(err)
		}

		bnd := fmt.Sprintf("%x", buf[:])

		i.b = "-----=" + bnd[:28]
	}

	return i.b, nil
}

func (i ioData) GetBuffer() *bytes.Buffer {
	return i.p
}

func (i *ioData) CRLF() Error {
	return i.String("\r\n")
}

func (i *ioData) ContentType(ct ContentType, charset string) Error {
	if charset != "" {
		return i.Header("Content-Type", fmt.Sprintf("\"%s\"; charset=%s", ct.String(), charset))
	} else {
		return i.Header("Content-Type", fmt.Sprintf("\"%s\"", ct.String()))
	}
}

func (i *ioData) BoundaryStart(ct ContentType) Error {
	if b, err := i.getBoundary(); err != nil {
		return err
	} else if err = i.Header("Content-Type", fmt.Sprintf("%s; boundary=\"%s\"", ct.String(), b)); err != nil {
		return err
	} else {
		return i.CRLF()
	}
}

func (i *ioData) BoundaryPart() Error {
	if i.b == "" {
		return nil
	}

	if b, err := i.getBoundary(); err != nil {
		return err
	} else if err = i.String(fmt.Sprintf("--%s", b)); err != nil {
		return err
	} else {
		return i.CRLF()
	}
}

func (i *ioData) BoundaryEnd() Error {
	if b, err := i.getBoundary(); err != nil {
		return err
	} else if err = i.CRLF(); err != nil {
		return err
	} else if err = i.String(fmt.Sprintf("--%s--", b)); err != nil {
		return err
	} else {
		return i.CRLF()
	}
}

func (i *ioData) Header(key, value string) Error {
	return i.String(fmt.Sprintf("%s: %s\r\n", key, value))
}

func (i *ioData) String(value string) Error {
	if i.p == nil {
		i.p = bytes.NewBuffer(make([]byte, 0))
	}

	if _, e := i.p.WriteString(value); e != nil {
		return BUFFER_WRITE_STRING.ErrorParent(e)
	}

	return nil
}

func (i *ioData) Bytes(value []byte) Error {
	if i.p == nil {
		i.p = bytes.NewBuffer(make([]byte, 0))
	}

	// write base64 content in lines of up to 76 chars
	tmp := make([]byte, 0)
	for n, l := 0, len(value); n < l; n++ {
		tmp = append(tmp, value[n])

		if (n+1)%76 == 0 {
			if _, e := i.p.Write(tmp); e != nil {
				return BUFFER_WRITE_BYTES.ErrorParent(e)
			} else if e := i.CRLF(); e != nil {
				return e
			}

			tmp = make([]byte, 0)
		}
	}

	if len(tmp) != 0 {
		if _, e := i.p.Write(tmp); e != nil {
			return BUFFER_WRITE_BYTES.ErrorParent(e)
		} else if e := i.CRLF(); e != nil {
			return e
		}
	}

	return nil
}

func (i *ioData) Send() Error {
	if i.w == nil {
		return IO_WRITER_MISSING.Error(nil)
	}
	if i.p == nil || i.p.Len() < 1 {
		return BUFFER_EMPTY.Error(nil)
	}

	if _, e := i.w.Write(i.p.Bytes()); e != nil {
		return IO_WRITER_ERROR.ErrorParent(e)
	}

	return nil
}

func (i *ioData) AttachmentStart(c ContentType) Error {
	return i.BoundaryStart(c)
}

func (i *ioData) AttachmentAddFile(contentType, attachmentName string, attachment *bytes.Buffer) Error {
	var (
		c = make([]byte, base64.StdEncoding.EncodedLen(attachment.Len()))
	)

	// convert attachment in base64
	base64.StdEncoding.Encode(c, attachment.Bytes())

	if len(c) < 1 {
		return BUFFER_EMPTY.Error(nil)
	}

	if e := i.BoundaryPart(); e != nil {
		return e
	} else if e = i.Header("Content-Type", contentType); e != nil {
		return e
	} else if e = i.Header("Content-Transfer-Encoding", "base64"); e != nil {
		return e
	} else if e = i.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachmentName)); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.Bytes(c); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	}

	return nil
}

func (i *ioData) AttachmentAddBody(m MailTemplate, ct ContentType) Error {
	var (
		e Error
		p *bytes.Buffer
	)

	if m.IsEmpty() {
		return EMPTY_HTML.Error(nil)
	}

	switch ct {
	case CONTENTTYPE_HTML:
		if p, e = m.GetBufferHtml(nil); e != nil {
			return e
		}

	case CONTENTTYPE_TEXT:
		if p, e = m.GetBufferText(nil); e != nil {
			return e
		}
	}

	b := make([]byte, base64.StdEncoding.EncodedLen(p.Len()))
	base64.StdEncoding.Encode(b, p.Bytes())
	p.Reset()
	p.Write(b)

	if e = i.BoundaryPart(); e != nil {
		return e
	} else if e = i.ContentType(ct, m.GetCharset()); e != nil {
		return e
	} else if e = i.Header("Content-Transfer-Encoding", "base64"); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.Bytes(p.Bytes()); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	}

	return nil
}

func (i *ioData) AttachmentEnd() Error {
	if e := i.BoundaryEnd(); e != nil {
		return e
	} else if e = i.BoundaryEnd(); e != nil {
		return e
	} else {
		return i.CRLF()
	}
}

type IOData interface {
	ContentType(ct ContentType, charset string) Error
	Header(key, value string) Error
	String(value string) Error
	Bytes(value []byte) Error
	CRLF() Error

	Send() Error
	GetBuffer() *bytes.Buffer

	AttachmentStart(c ContentType) Error
	AttachmentAddFile(contentType, attachmentName string, attachment *bytes.Buffer) Error
	AttachmentAddBody(m MailTemplate, ct ContentType) Error
	AttachmentEnd() Error
}

func NewIOData(cli *smtp.Client) (IOData, Error) {
	if w, e := cli.Data(); e != nil {
		return nil, SMTP_CLIENT_DATA.ErrorParent(e)
	} else {
		return &ioData{
			w: w,
			p: bytes.NewBuffer(make([]byte, 0)),
		}, nil
	}
}
