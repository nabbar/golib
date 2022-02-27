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

package mail

import (
	"sync"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libmail "github.com/nabbar/golib/mail"
)

type componentMail struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	m sync.Mutex
	e libmail.Mail
}

func (c *componentMail) _CheckDep() bool {
	return c != nil
}

func (c *componentMail) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	var (
		err liberr.Error
		mlr libmail.Mail
		cfg libmail.Config

		isReload = c.e != nil
	)

	if !c._CheckDep() {
		return ErrorComponentNotInitialized.Error(nil)
	}

	if !isReload && c.fsb != nil {
		if err = c.fsb(); err != nil {
			return err
		}
	} else if isReload && c.frb != nil {
		if err = c.frb(); err != nil {
			return err
		}
	}

	if cfg, err = c._getConfig(getCfg); err != nil {
		return err
	} else if mlr, err = cfg.NewMailer(); err != nil {
		return err
	} else {
		c.e = mlr
	}

	if !isReload && c.fsa != nil {
		if err = c.fsa(); err != nil {
			return err
		}
	} else if isReload && c.fra != nil {
		if err = c.fra(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentMail) Type() string {
	return ComponentType
}

func (c *componentMail) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentMail) RegisterFuncStart(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentMail) RegisterFuncReload(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentMail) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c._CheckDep() && c.e != nil
}

func (c *componentMail) IsRunning(atLeast bool) bool {
	return !c.IsStarted()
}

func (c *componentMail) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentMail) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentMail) Stop() {
	c.m.Lock()
	defer c.m.Unlock()

	c.e = nil
}

func (c *componentMail) Dependencies() []string {
	return make([]string, 0)
}

func (c *componentMail) GetMail() (libmail.Mail, liberr.Error) {
	if !c.IsStarted() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	c.m.Lock()
	defer c.m.Unlock()

	return c.e.Clone(), nil
}
