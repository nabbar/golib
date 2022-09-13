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

package natsServer

import (
	"sync"

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	cpttls "github.com/nabbar/golib/config/components/tls"
	liberr "github.com/nabbar/golib/errors"
	libnat "github.com/nabbar/golib/nats"
	libsts "github.com/nabbar/golib/status"
	natsrv "github.com/nats-io/nats-server/v2/server"
)

type componentNats struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func(cpt libcfg.Component) liberr.Error
	fsb func(cpt libcfg.Component) liberr.Error
	fra func(cpt libcfg.Component) liberr.Error
	frb func(cpt libcfg.Component) liberr.Error

	m sync.Mutex
	t string
	n libnat.Server
}

func (c *componentNats) _CheckDep() bool {
	return c != nil && c.t != ""
}

func (c *componentNats) _GetTLS() (libtls.TLSConfig, liberr.Error) {
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

func (c *componentNats) _getFct() (func(cpt libcfg.Component) liberr.Error, func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.n != nil {
		return c.frb, c.fra
	} else {
		return c.fsb, c.fsa
	}
}

func (c *componentNats) _runFct(fct func(cpt libcfg.Component) liberr.Error) liberr.Error {
	if fct != nil {
		return fct(c)
	}

	return nil
}

func (c *componentNats) _runCli(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	var (
		tls libtls.TLSConfig
		err liberr.Error
		cfg libnat.Config
		opt *natsrv.Options
	)

	if cfg, err = c._getConfig(getCfg); err != nil {
		return err
	}

	if tls, err = c._GetTLS(); err != nil {
		return err
	}

	if opt, err = cfg.NatsOption(tls); err != nil {
		return ErrorStartComponent.Error(err)
	}

	if c.n != nil {
		c.n.SetOptions(opt)
		if err = c.n.Restart(c.ctx()); err != nil {
			return ErrorReloadComponent.Error(err)
		}
	} else {
		c.n = libnat.NewServer(opt, cfg.Status)
		if err = c.n.Listen(c.ctx()); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentNats) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	if !c._CheckDep() {
		return ErrorComponentNotInitialized.Error(nil)
	}

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

func (c *componentNats) Type() string {
	return ComponentType
}

func (c *componentNats) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper, sts libcfg.FuncRouteStatus) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentNats) RegisterFuncStart(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentNats) RegisterFuncReload(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentNats) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c != nil && c.n != nil
}

func (c *componentNats) IsRunning(atLeast bool) bool {
	if c.IsStarted() {
		c.m.Lock()
		defer c.m.Unlock()

		return c.n.IsRunning()
	}

	return false
}

func (c *componentNats) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentNats) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentNats) Stop() {
	if c.IsRunning(true) {
		c.m.Lock()
		defer c.m.Unlock()

		c.n.Shutdown()
	}
}

func (c *componentNats) Dependencies() []string {
	c.m.Lock()
	defer c.m.Unlock()

	if !c._CheckDep() {
		return []string{cpttls.ComponentType}
	}

	return []string{c.t}
}

func (c *componentNats) SetTLSKey(tlsKey string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.t = tlsKey
}

func (c *componentNats) GetServer() (libnat.Server, liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c == nil {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	return c.n, nil
}

func (c *componentNats) SetStatusRouter(sts libsts.RouteStatus, prefix string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.n.StatusRouter(sts, prefix)
}
