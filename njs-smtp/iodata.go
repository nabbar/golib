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
	w io.WriteCloser
	b string
}

func (i *ioData) getBoundary() (string, error) {
	if i.b == "" {
		var buf [30]byte

		_, err := io.ReadFull(rand.Reader, buf[:])

		if err != nil {
			return "", err
		}

		bnd := fmt.Sprintf("%x", buf[:])

		i.b = bnd[:28]
	}

	return i.b, nil
}

func (i *ioData) CRLF() error {
	return i.String("\r\n")
}

func (i *ioData) Header(key, value string) error {
	return i.String(fmt.Sprintf("%s: %s\r\n", key, value))
}

func (i *ioData) String(value string) error {
	if i.w == nil {
		return fmt.Errorf("empty writer")
	}

	if _, e := i.w.Write([]byte(value)); e != nil {
		return e
	}

	return nil
}

func (i *ioData) Bytes(value []byte) error {
	if i.w == nil {
		return fmt.Errorf("empty writer")
	}

	if _, e := i.w.Write([]byte(value)); e != nil {
		return e
	}

	// write base64 content in lines of up to 76 chars
	tmp := make([]byte, 0)
	for n, l := 0, len(value); n < l; n++ {
		tmp = append(tmp, value[n])

		if (n+1)%76 == 0 {
			if _, e := i.w.Write(tmp); e != nil {
				return e
			} else if e := i.CRLF(); e != nil {
				return e
			}

			tmp = make([]byte, 0)
		}
	}

	if len(tmp) != 0 {
		if _, e := i.w.Write(tmp); e != nil {
			return e
		} else if e := i.CRLF(); e != nil {
			return e
		}
	}

	return nil
}

func (i *ioData) AttachmentStart(c ContentType) error {
	if b, e := i.getBoundary(); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.Header("Content-Type", fmt.Sprintf("%s; boundary=\"%s\"", c.String(), b)); e != nil {
		return e
	} else {
		return i.CRLF()
	}
}

func (i *ioData) AttachmentAddFile(contentType, attachmentName string, attachment *bytes.Buffer) error {
	var (
		b string
		e error
		c = make([]byte, base64.StdEncoding.EncodedLen(attachment.Len()))
	)

	// convert attachment in base64
	base64.StdEncoding.Encode(c, attachment.Bytes())

	if len(c) < 1 {
		return fmt.Errorf("encoded buffer is empty")
	}

	if b, e = i.getBoundary(); e != nil {
		return e
	} else if e = i.String(fmt.Sprintf("--%s", b)); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
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
	}

	return nil
}

func (i *ioData) AttachmentAddBody(m MailTemplate, ct ContentType) error {
	var (
		b string
		e error
		p *bytes.Buffer
	)

	if m.IsEmpty() {
		return fmt.Errorf("text/html content is empty")
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

	if b, e = i.getBoundary(); e != nil {
		return e
	} else if e = i.String(fmt.Sprintf("--%s", b)); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.Header("Content-Type", fmt.Sprintf("%s; charset=%s", ct.String(), m.GetCharset())); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.Bytes(p.Bytes()); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	}

	return nil
}

func (i *ioData) AttachmentEnd() error {
	if b, e := i.getBoundary(); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.String(fmt.Sprintf("--%s--", b)); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else if e = i.String(fmt.Sprintf("--%s--", b)); e != nil {
		return e
	} else if e = i.CRLF(); e != nil {
		return e
	} else {
		return i.CRLF()
	}
}

type IOData interface {
	Header(key, value string) error
	String(value string) error
	Bytes(value []byte) error
	CRLF() error

	AttachmentStart(c ContentType) error
	AttachmentAddFile(contentType, attachmentName string, attachment *bytes.Buffer) error
	AttachmentAddBody(m MailTemplate, ct ContentType) error
	AttachmentEnd() error
}

func NewIOData(cli *smtp.Client) (IOData, error) {
	if w, e := cli.Data(); e != nil {
		return nil, e
	} else {
		return &ioData{
			w: w,
		}, nil
	}
}
