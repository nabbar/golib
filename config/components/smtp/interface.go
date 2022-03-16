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
	"sync"

	libsts "github.com/nabbar/golib/status"

	cpttls "github.com/nabbar/golib/config/components/tls"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libsmtp "github.com/nabbar/golib/smtp"
)

const (
	ComponentType = "smtp"
)

type ComponentSMTP interface {
	libcfg.Component

	SetTLSKey(tlsKey string)
	GetSMTP() (libsmtp.SMTP, liberr.Error)
	SetStatusRouter(sts libsts.RouteStatus, prefix string)
}

func New(tlsKey string) ComponentSMTP {
	if tlsKey == "" {
		tlsKey = cpttls.ComponentType
	}

	return &componentSmtp{
		ctx: nil,
		get: nil,
		fsa: nil,
		fsb: nil,
		fra: nil,
		frb: nil,
		m:   sync.Mutex{},
		t:   tlsKey,
		s:   nil,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentSMTP) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, key, tlsKey string) {
	cfg.ComponentSet(key, New(tlsKey))
}

func Load(getCpt libcfg.FuncComponentGet, key string) ComponentSMTP {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentSMTP); !ok {
		return nil
	} else {
		return h
	}
}
