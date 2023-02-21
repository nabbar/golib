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
	"sync"

	libctx "github.com/nabbar/golib/context"
	lbldap "github.com/nabbar/golib/ldap"
)

type componentLDAP struct {
	m sync.RWMutex
	x libctx.Config[uint8]

	a []string
	c *lbldap.Config
	l *lbldap.HelperLDAP
}

func (o *componentLDAP) GetConfig() *lbldap.Config {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.c
}

func (o *componentLDAP) SetConfig(opt *lbldap.Config) {
	o.m.Lock()
	defer o.m.Unlock()

	o.c = opt
}

func (o *componentLDAP) GetLDAP() *lbldap.HelperLDAP {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.l
}

func (o *componentLDAP) SetLDAP(l *lbldap.HelperLDAP) {
	o.m.Lock()
	defer o.m.Unlock()

	o.l = l
}

func (o *componentLDAP) SetAttributes(att []string) {
	o.m.Lock()
	defer o.m.Unlock()

	o.a = att
	if o.l != nil {
		o.l.Attributes = att
	}
}
