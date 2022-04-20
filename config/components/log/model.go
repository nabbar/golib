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

package log

import (
	"sync"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

type componentLog struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func(cpt libcfg.Component) liberr.Error
	fsb func(cpt libcfg.Component) liberr.Error
	fra func(cpt libcfg.Component) liberr.Error
	frb func(cpt libcfg.Component) liberr.Error

	d func() liblog.Logger

	m sync.Mutex
	l liblog.Logger
	v liblog.Level
}

func (c *componentLog) _getFct() (func(cpt libcfg.Component) liberr.Error, func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.l != nil {
		return c.frb, c.fra
	} else {
		return c.fsb, c.fsa
	}
}

func (c *componentLog) _runFct(fct func(cpt libcfg.Component) liberr.Error) liberr.Error {
	if fct != nil {
		return fct(c)
	}

	return nil
}

func (c *componentLog) _runCli(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.l == nil {
		if c.ctx == nil {
			return ErrorComponentNotInitialized.Error(nil)
		}

		c.l = liblog.New(c.ctx())
		c.l.SetLevel(c.v)
	}

	return nil
}

func (c *componentLog) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		cnf *liblog.Options
	)

	fb, fa := c._getFct()

	if err := c._runFct(fb); err != nil {
		return err
	} else if err = c._runCli(getCfg); err != nil {
		return err
	} else if cnf, err = c._GetOptions(getCfg); err != nil {
		return err
	} else if cnf == nil {
		return ErrorConfigInvalid.Error(nil)
	} else if e := c.l.SetOptions(cnf); e != nil {
		return ErrorReloadLog.Error(err)
	} else if err = c._runFct(fa); err != nil {
		return err
	}

	return nil
}

func (c *componentLog) Type() string {
	return ComponentType
}

func (c *componentLog) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentLog) RegisterFuncStart(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentLog) RegisterFuncReload(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentLog) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c.l != nil
}

func (c *componentLog) IsRunning(atLeast bool) bool {
	return c.IsStarted()
}

func (c *componentLog) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentLog) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentLog) Stop() {
	return
}

func (c *componentLog) Dependencies() []string {
	return make([]string, 0)
}

func (c *componentLog) Log() liblog.Logger {
	c.m.Lock()
	defer c.m.Unlock()

	if c.l != nil {
		if n, e := c.l.Clone(); e != nil {
			return c.d()
		} else {
			return n
		}
	}

	return c.d()
}

func (c *componentLog) SetLevel(lvl liblog.Level) {
	c.m.Lock()
	defer c.m.Unlock()

	c.v = lvl

	if c.l == nil {
		return
	}

	c.l.SetLevel(lvl)
}

func (c *componentLog) SetField(fields liblog.Fields) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.l == nil {
		return
	}

	c.l.SetFields(fields)
}

func (c *componentLog) SetOptions(opt *liblog.Options) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.l == nil {
		return ErrorComponentNotInitialized.Error(nil)
	}

	if e := c.l.SetOptions(opt); e != nil {
		return ErrorConfigInvalid.ErrorParent(e)
	}

	return nil
}
