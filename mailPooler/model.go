/*
 *  MIT License
 *
 *  Copyright (c) 2021 Nicolas JUHEL
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

package mailPooler

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/smtp"

	libsmtp "github.com/nabbar/golib/smtp"
	smtpcf "github.com/nabbar/golib/smtp/config"
)

type pooler struct {
	s libsmtp.SMTP
	c Counter
}

func (p *pooler) Reset() error {
	if p.s == nil {
		return ErrorParamEmpty.Error(errors.New("smtp client is not define"))
	}

	if err := p.c.Reset(); err != nil {
		return err
	}

	return nil
}

func (p *pooler) NewPooler() Pooler {
	if p.s == nil {
		return &pooler{
			s: nil,
			c: p.c.Clone(),
		}
	} else {
		return &pooler{
			s: p.s.Clone(),
			c: p.c.Clone(),
		}
	}
}

func (p *pooler) Send(ctx context.Context, from string, to []string, data io.WriterTo) error {
	if p.s == nil {
		return ErrorParamEmpty.Error(errors.New("smtp client is not define"))
	}

	if err := p.c.Pool(ctx); err != nil {
		return err
	}

	return p.s.Send(ctx, from, to, data)
}

func (p *pooler) Client(ctx context.Context) (*smtp.Client, error) {
	if p.s == nil {
		return nil, ErrorParamEmpty.Error(errors.New("smtp client is not define"))
	}

	return p.s.Client(ctx)
}

func (p *pooler) Close() {
	if p.s != nil {
		p.s.Close()
	}
}

func (p *pooler) Check(ctx context.Context) error {
	if p.s == nil {
		return ErrorParamEmpty.Error(errors.New("smtp client is not define"))
	}

	return p.s.Check(ctx)
}

func (p *pooler) Clone() libsmtp.SMTP {
	return p.NewPooler()
}

func (p *pooler) UpdConfig(cfg smtpcf.SMTP, tslConfig *tls.Config) {
	if p.s != nil {
		p.s.UpdConfig(cfg, tslConfig)
	} else {
		p.s, _ = libsmtp.New(cfg, tslConfig)
	}
}
