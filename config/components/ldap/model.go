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

package ldap

import (
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
	lbldap "github.com/nabbar/golib/ldap"
)

type componentLDAP struct {
	a *atomic.Value // slice attributes []string
	c *atomic.Value // config
	l *atomic.Value // client LDAP
	x libctx.Config[uint8]
}

func (o *componentLDAP) GetConfig() *lbldap.Config {
	if i := o.c.Load(); i == nil {
		return nil
	} else if v, k := i.(*lbldap.Config); !k {
		return nil
	} else if len(v.Uri) < 1 {
		return nil
	} else if v.PortLdap < 1 && v.Portldaps < 1 {
		return nil
	} else {
		var cfg = lbldap.Config{}
		cfg = *v
		return &cfg
	}
}

func (o *componentLDAP) SetConfig(opt *lbldap.Config) {
	if opt == nil {
		opt = &lbldap.Config{}
	}

	o.c.Store(opt)
}

func (o *componentLDAP) GetLDAP() *lbldap.HelperLDAP {
	if i := o.l.Load(); i == nil {
		return nil
	} else if v, k := i.(*lbldap.HelperLDAP); !k {
		return nil
	} else if v == nil {
		return nil
	} else if n := v.Clone(); n == nil {
		return nil
	} else if n.Check() != nil {
		return nil
	} else {
		n.Attributes = o.GetAttributes()
		return n
	}
}

func (o *componentLDAP) SetLDAP(l *lbldap.HelperLDAP) {
	if l == nil {
		l = &lbldap.HelperLDAP{}
	}

	o.l.Store(l)
}

func (o *componentLDAP) GetAttributes() []string {
	if i := o.a.Load(); i == nil {
		return make([]string, 0)
	} else if v, k := i.([]string); !k {
		return make([]string, 0)
	} else if len(v) > 0 {
		return v
	} else {
		return make([]string, 0)
	}
}

func (o *componentLDAP) SetAttributes(att []string) {
	if att == nil {
		att = make([]string, 0)
	}

	o.a.Store(att)
}
