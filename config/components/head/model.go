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

package head

import (
	"sync"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	librtr "github.com/nabbar/golib/router"
)

type componentHead struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func(cpt libcfg.Component) liberr.Error
	fsb func(cpt libcfg.Component) liberr.Error
	fra func(cpt libcfg.Component) liberr.Error
	frb func(cpt libcfg.Component) liberr.Error

	m sync.Mutex
	h librtr.Headers
}

func (c *componentHead) _getFct() (func(cpt libcfg.Component) liberr.Error, func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.h != nil {
		return c.frb, c.fra
	} else {
		return c.fsb, c.fsa
	}
}

func (c *componentHead) _runFct(fct func(cpt libcfg.Component) liberr.Error) liberr.Error {
	if fct != nil {
		return fct(c)
	}

	return nil
}

func (c *componentHead) _runCli(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	cnf := librtr.HeadersConfig{}
	if err := getCfg(c.key, &cnf); err != nil {
		return ErrorParamsInvalid.Error(err)
	}

	c.h = cnf.New()
	return nil
}

func (c *componentHead) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
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

func (c *componentHead) Type() string {
	return ComponentType
}

func (c *componentHead) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentHead) RegisterFuncStart(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentHead) RegisterFuncReload(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentHead) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c.h != nil
}

func (c *componentHead) IsRunning(atLeast bool) bool {
	return c.IsStarted()
}

func (c *componentHead) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentHead) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentHead) Stop() {

}

func (c *componentHead) Dependencies() []string {
	return []string{}
}

func (c *componentHead) GetHeaders() librtr.Headers {
	c.m.Lock()
	defer c.m.Unlock()

	return c.h
}

func (c *componentHead) SetHeaders(head librtr.Headers) {
	c.m.Lock()
	defer c.m.Unlock()

	c.h = head
}
