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

package request

import (
	"context"
	"sync"

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	cpttls "github.com/nabbar/golib/config/components/tls"
	liberr "github.com/nabbar/golib/errors"
	libreq "github.com/nabbar/golib/request"
)

type componentRequest struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	sts libcfg.FuncRouteStatus
	key string

	fsa func(cpt libcfg.Component) liberr.Error
	fsb func(cpt libcfg.Component) liberr.Error
	fra func(cpt libcfg.Component) liberr.Error
	frb func(cpt libcfg.Component) liberr.Error

	m sync.Mutex
	r libreq.Request
	t string
	c libreq.FctHttpClient
}

func (c *componentRequest) _GetContext() context.Context {
	c.m.Lock()
	defer c.m.Unlock()

	if c.ctx != nil {
		if x := c.ctx(); x != nil {
			return x
		}
	}

	return context.Background()
}

func (c *componentRequest) _CheckInit() bool {
	return c != nil && c.r != nil
}

func (c *componentRequest) _GetTLS() libtls.TLSConfig {
	if c.t == "" {
		return nil
	}

	if i := cpttls.Load(c.get, c.t); i == nil {
		return nil
	} else if tls := i.GetTLS(); tls == nil {
		return nil
	} else {
		return tls
	}
}

func (c *componentRequest) _getFct() (func(cpt libcfg.Component) liberr.Error, func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.r != nil {
		return c.frb, c.fra
	} else {
		return c.fsb, c.fsa
	}
}

func (c *componentRequest) _runFct(fct func(cpt libcfg.Component) liberr.Error) liberr.Error {
	if fct != nil {
		return fct(c)
	}

	return nil
}

func (c *componentRequest) _runCli(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	cfg := libreq.Options{}
	cfg.SetDefaultTLS(c._GetTLS)

	if err := getCfg(c.key, &cfg); err != nil {
		return ErrorParamInvalid.Error(err)
	}

	if c.r == nil {
		if r, e := cfg.New(c._GetContext, c.c, c._GetTLS); e != nil {
			return ErrorConfigInvalid.ErrorParent(e)
		} else {
			c.r = r
		}
	} else {
		if r, e := cfg.Update(c.r, c._GetContext, c.c, c._GetTLS); e != nil {
			return ErrorConfigInvalid.ErrorParent(e)
		} else {
			c.r = r
		}
	}

	if c.sts != nil {
		if s := c.sts(); s != nil {
			c.r.StatusRegister(s, c.key)
		}
	}

	return nil
}

func (c *componentRequest) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	fb, fa := c._getFct()

	if err := c._runFct(fb); err != nil {
		return err
	} else if err = c._runCli(getCfg); err != nil {
		return err
	} else if err = c._runFct(fa); err != nil {
		return err
	}

	return nil
}

func (c *componentRequest) Type() string {
	return ComponentType
}

func (c *componentRequest) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper, sts libcfg.FuncRouteStatus) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
	c.sts = sts
}

func (c *componentRequest) RegisterFuncStart(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentRequest) RegisterFuncReload(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentRequest) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c != nil && c.r != nil
}

func (c *componentRequest) IsRunning(atLeast bool) bool {
	return c.IsStarted()
}

func (c *componentRequest) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentRequest) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentRequest) Stop() {

}

func (c *componentRequest) Dependencies() []string {
	c.m.Lock()
	defer c.m.Unlock()

	if c == nil || c.t == "" {
		return []string{cpttls.ComponentType}
	}

	return []string{c.t}
}

func (c *componentRequest) SetHTTPClient(fct libreq.FctHttpClient) {
	c.m.Lock()
	defer c.m.Unlock()

	c.c = fct
}

func (c *componentRequest) SetDefaultTLS(key string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.t = key
}

func (c *componentRequest) Request() (libreq.Request, error) {
	c.m.Lock()
	defer c.m.Unlock()

	return c.r.Clone()
}
