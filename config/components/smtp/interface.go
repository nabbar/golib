/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package smtp

import (
	"context"

	libatm "github.com/nabbar/golib/atomic"
	libcfg "github.com/nabbar/golib/config"
	cpttls "github.com/nabbar/golib/config/components/tls"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	lbsmtp "github.com/nabbar/golib/mail/smtp"
)

type CptSMTP interface {
	cfgtps.Component

	SetTLSKey(tlsKey string)
	GetSMTP() (lbsmtp.SMTP, error)
}

func New(ctx context.Context, tlsKey string) CptSMTP {
	if tlsKey == "" {
		tlsKey = cpttls.ComponentType
	}

	c := &mod{
		x: libctx.New[uint8](ctx),
		t: libatm.NewValue[string](),
		s: libatm.NewValue[lbsmtp.SMTP](),
	}

	c.t.Store(tlsKey)

	return c
}

func Register(cfg libcfg.Config, key string, cpt CptSMTP) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(ctx context.Context, cfg libcfg.Config, key, tlsKey string) {
	cfg.ComponentSet(key, New(ctx, tlsKey))
}

func Load(getCpt cfgtps.FuncCptGet, key string) CptSMTP {
	if getCpt == nil {
		return nil
	} else if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(CptSMTP); !ok {
		return nil
	} else {
		return h
	}
}
