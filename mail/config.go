/*
 * MIT License
 *
 * Copyright (c) 2021 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package mail

import (
	"fmt"
	"net/textproto"

	libval "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
)

type Config struct {
	Charset  string `json:"charset" yaml:"charset" toml:"charset" mapstructure:"charset" validate:"required"`
	Subject  string `json:"subject" yaml:"subject" toml:"subject" mapstructure:"subject" validate:"required"`
	Encoding string `json:"encoding" yaml:"encoding" toml:"encoding" mapstructure:"encoding" validate:"required"`
	Priority string `json:"priority" yaml:"priority" toml:"priority" mapstructure:"priority" validate:"required"`

	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty" mapstructure:"headers,omitempty"`

	From       string `json:"from" yaml:"from" toml:"from" mapstructure:"from" validate:"required,email"`
	Sender     string `json:"sender,omitempty" yaml:"sender,omitempty" toml:"sender,omitempty" mapstructure:"sender,omitempty" validate:"email"`
	ReplyTo    string `json:"replyTo,omitempty" yaml:"replyTo,omitempty" toml:"replyTo,omitempty" mapstructure:"replyTo,omitempty" validate:"email"`
	ReturnPath string `json:"returnPath,omitempty" yaml:"returnPath,omitempty" toml:"returnPath,omitempty" mapstructure:"returnPath,omitempty" validate:"email"`

	To  []string `json:"to,omitempty" yaml:"to,omitempty" toml:"to,omitempty" mapstructure:"to,omitempty" validate:"email"`
	Cc  []string `json:"cc,omitempty" yaml:"cc,omitempty" toml:"cc,omitempty" mapstructure:"cc,omitempty" validate:"email"`
	Bcc []string `json:"bcc,omitempty" yaml:"bcc,omitempty" toml:"bcc,omitempty" mapstructure:"bcc,omitempty" validate:"email"`

	Attach []ConfigFile `json:"attach,omitempty" yaml:"attach,omitempty" toml:"attach,omitempty" mapstructure:"attach,omitempty"`
	Inline []ConfigFile `json:"inline,omitempty" yaml:"inline,omitempty" toml:"inline,omitempty" mapstructure:"inline,omitempty"`
}

type ConfigFile struct {
	Name string `json:"name" yaml:"name" toml:"name" mapstructure:"name" validate:"required"`
	Mime string `json:"mime" yaml:"mime" toml:"mime" mapstructure:"mime" validate:"required"`
	Path string `json:"path" yaml:"path" toml:"path" mapstructure:"path" validate:"required,file"`
}

func (c Config) Validate() liberr.Error {
	val := libval.New()
	err := val.Struct(c)

	if e, ok := err.(*libval.InvalidValidationError); ok {
		return ErrorMailConfigInvalid.ErrorParent(e)
	}

	out := ErrorMailConfigInvalid.Error(nil)

	for _, e := range err.(libval.ValidationErrors) {
		//nolint goerr113
		out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
	}

	if out.HasParent() {
		return out
	}

	return nil
}

func (c Config) NewMailer() (Mail, liberr.Error) {
	m := &mail{
		headers:  make(textproto.MIMEHeader),
		charset:  "UTF-8",
		encoding: ParseEncoding(c.Encoding),
		priority: ParsePriority(c.Priority),
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
	if len(c.Headers) > 0 {
		for k, v := range c.Headers {
			m.headers.Set(k, v)
		}
	}

	if c.Charset != "" {
		m.charset = c.Charset
	}

	m.Email().SetFrom(c.From)

	if c.Sender != "" {
		m.Email().SetSender(c.Sender)
	}
	if c.Sender != "" {
		m.Email().SetReplyTo(c.ReplyTo)
	}
	if c.Sender != "" {
		m.Email().SetReturnPath(c.ReturnPath)
	}

	if len(c.To) > 0 {
		m.Email().AddRecipients(RecipientTo, c.To...)
	}

	if len(c.Cc) > 0 {
		m.Email().AddRecipients(RecipientCC, c.Cc...)
	}

	if len(c.Bcc) > 0 {
		m.Email().AddRecipients(RecipientBCC, c.Bcc...)
	}

	if len(c.Attach) > 0 {
		for _, f := range c.Attach {
			if h, e := libiot.NewFileProgressPathRead(f.Path, 0); e != nil {
				return nil, e
			} else {
				m.AddAttachment(f.Name, f.Mime, h, false)
			}
		}
	}

	if len(c.Inline) > 0 {
		for _, f := range c.Inline {
			if h, e := libiot.NewFileProgressPathRead(f.Path, 0); e != nil {
				return nil, e
			} else {
				m.AddAttachment(f.Name, f.Mime, h, true)
			}
		}
	}

	return m, nil
}
