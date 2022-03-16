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

package aws

import (
	"net/http"
	"sync"

	libaws "github.com/nabbar/golib/aws"
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
)

type componentAws struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	m sync.Mutex
	d ConfigDriver
	c func() *http.Client
	a libaws.AWS
}

func (c *componentAws) _getHttpClient() *http.Client {
	if c.c == nil {
		return nil
	}

	return c.c()
}

func (c *componentAws) _getClient(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		isReload = c.a != nil
	)

	if !isReload && c.fsb != nil {
		if err := c.fsb(); err != nil {
			return err
		}
	} else if isReload && c.frb != nil {
		if err := c.frb(); err != nil {
			return err
		}
	}

	if cfg, err := c._getConfig(getCfg); err != nil {
		return err
	} else if cli, er := libaws.New(c.ctx(), cfg, c._getHttpClient()); er != nil {
		return er
	} else {
		c.a = cli
	}

	if !isReload && c.fsa != nil {
		if err := c.fsa(); err != nil {
			return err
		}
	} else if isReload && c.fra != nil {
		if err := c.fra(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentAws) Type() string {
	return ComponentType
}

func (c *componentAws) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentAws) RegisterFuncStart(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentAws) RegisterFuncReload(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentAws) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c.a != nil
}

func (c *componentAws) IsRunning(atLeast bool) bool {
	return c.IsStarted()
}

func (c *componentAws) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._getClient(getCfg)
}

func (c *componentAws) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._getClient(getCfg)
}

func (c *componentAws) Stop() {
	c.m.Lock()
	defer c.m.Unlock()

	c.a = nil
	return
}

func (c *componentAws) Dependencies() []string {
	return make([]string, 0)
}

func (c *componentAws) RegisterHTTPClient(fct func() *http.Client) {
	c.m.Lock()
	defer c.m.Unlock()

	c.c = fct
}

func (c *componentAws) GetAws() (libaws.AWS, liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.a == nil {
		return nil, ErrorComponentNotInitialized.Error(nil)
	} else {
		return c.a.Clone(c.ctx())
	}
}

func (c *componentAws) SetAws(a libaws.AWS) {
	c.m.Lock()
	defer c.m.Unlock()

	c.a = a
}
