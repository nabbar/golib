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

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	cpttls "github.com/nabbar/golib/config/components/tls"
	liberr "github.com/nabbar/golib/errors"
	libsmtp "github.com/nabbar/golib/smtp"
	libsts "github.com/nabbar/golib/status"
)

type componentSmtp struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	m sync.Mutex
	t string
	s libsmtp.SMTP
}

func (c *componentSmtp) _CheckDep() bool {
	return c != nil && c.t != ""
}

func (c *componentSmtp) _GetTLS() (libtls.TLSConfig, liberr.Error) {
	if !c._CheckDep() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if i := cpttls.Load(c.get, c.t); i == nil {
		return nil, ErrorDependencyTLSDefault.Error(nil)
	} else if tls := i.GetTLS(); tls == nil {
		return nil, ErrorDependencyTLSDefault.Error(nil)
	} else {
		return tls, nil
	}
}

func (c *componentSmtp) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	if !c._CheckDep() {
		return ErrorComponentNotInitialized.Error(nil)
	}

	c.m.Lock()
	defer c.m.Unlock()

	var (
		err liberr.Error
		cli libsmtp.SMTP
		cfg libsmtp.ConfigModel

		isReload = c.s != nil
	)

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
	}

	if cli, err = cfg.GetSMTP(); err != nil {
		if isReload {
			return ErrorReloadComponent.Error(err)
		}
		return ErrorStartComponent.Error(err)
	}

	if isReload {
		_ = c.s.Close
	}

	c.s = cli

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

func (c *componentSmtp) Type() string {
	return ComponentType
}

func (c *componentSmtp) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentSmtp) RegisterFuncStart(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentSmtp) RegisterFuncReload(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentSmtp) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c._CheckDep() && c.s != nil
}

func (c *componentSmtp) IsRunning(atLeast bool) bool {
	if !c.IsStarted() {
		return false
	}

	c.m.Lock()
	defer c.m.Unlock()

	e := c.s.Check(c.ctx())
	return e == nil
}

func (c *componentSmtp) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentSmtp) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentSmtp) Stop() {
	if !c.IsStarted() {
		return
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.s.Close()
}

func (c *componentSmtp) Dependencies() []string {
	c.m.Lock()
	defer c.m.Unlock()

	if !c._CheckDep() {
		return []string{cpttls.ComponentType}
	}

	return []string{c.t}
}

func (c *componentSmtp) SetTLSKey(tlsKey string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.t = tlsKey
}

func (c *componentSmtp) GetSMTP() (libsmtp.SMTP, liberr.Error) {
	if !c.IsStarted() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	c.m.Lock()
	defer c.m.Unlock()

	return c.s.Clone(), nil
}

func (c *componentSmtp) SetStatusRouter(sts libsts.RouteStatus, prefix string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.s.StatusRouter(sts, prefix)
}
