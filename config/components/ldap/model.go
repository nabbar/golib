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
	"context"
	"sync"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	lbldap "github.com/nabbar/golib/ldap"
)

type componentLDAP struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	m sync.Mutex
	l *lbldap.HelperLDAP
}

func (c *componentLDAP) _GetContext() context.Context {
	c.m.Lock()
	defer c.m.Unlock()

	if c.ctx != nil {
		if x := c.ctx(); x != nil {
			return x
		}
	}

	return context.Background()
}

func (c *componentLDAP) _CheckInit() bool {
	return c != nil && c.l != nil
}

func (c *componentLDAP) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	var isReload = c.l != nil

	if isReload && c.frb != nil {
		if err := c.frb(); err != nil {
			return err
		}
	} else if !isReload && c.fsb != nil {
		if err := c.fsb(); err != nil {
			return err
		}
	}

	cfg := lbldap.Config{}
	if err := getCfg(c.key, &cfg); err != nil {
		return ErrorParamsInvalid.Error(err)
	}

	if l, e := lbldap.NewLDAP(c._GetContext(), &cfg, nil); e != nil {
		return ErrorConfigInvalid.ErrorParent(e)
	} else {
		c.l = l
	}

	if isReload && c.fra != nil {
		if err := c.fra(); err != nil {
			return err
		}
	} else if !isReload && c.fsa != nil {
		if err := c.fsa(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentLDAP) Type() string {
	return ComponentType
}

func (c *componentLDAP) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentLDAP) RegisterFuncStart(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentLDAP) RegisterFuncReload(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentLDAP) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c != nil && c.l != nil
}

func (c *componentLDAP) IsRunning(atLeast bool) bool {
	return c.IsStarted()
}

func (c *componentLDAP) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentLDAP) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentLDAP) Stop() {

}

func (c *componentLDAP) Dependencies() []string {
	return make([]string, 0)
}

func (c *componentLDAP) LDAP() *lbldap.HelperLDAP {
	c.m.Lock()
	defer c.m.Unlock()

	return c.l
}
