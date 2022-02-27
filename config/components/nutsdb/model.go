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

package nutsdb

import (
	"sync"
	"time"

	libcfg "github.com/nabbar/golib/config"
	cptlog "github.com/nabbar/golib/config/components/log"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	libndb "github.com/nabbar/golib/nutsdb"
	libsts "github.com/nabbar/golib/status"
)

type componentNutsDB struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	m sync.Mutex
	l string
	n libndb.NutsDB
}

func (c *componentNutsDB) _CheckDep() bool {
	return c != nil && c.l != ""
}

func (c *componentNutsDB) _GetLogger() (liblog.Logger, liberr.Error) {
	if !c._CheckDep() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if i := cptlog.Load(c.get, c.l); i == nil {
		return nil, ErrorDependencyLogDefault.Error(nil)
	} else if log := i.Log(); log == nil {
		return nil, ErrorDependencyLogDefault.Error(nil)
	} else {
		return log, nil
	}
}

func (c *componentNutsDB) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	if !c._CheckDep() {
		return ErrorComponentNotInitialized.Error(nil)
	}

	c.m.Lock()
	defer c.m.Unlock()

	var (
		err liberr.Error
		cfg libndb.Config

		isReload = c.n != nil
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

	srv := libndb.New(cfg)
	srv.SetLogger(func() liblog.Logger {
		if l, e := c._GetLogger(); e != nil {
			return liblog.GetDefault()
		} else {
			return l
		}
	})

	if err = srv.Listen(); err != nil {
		return err
	}

	if c.n != nil {
		_ = c.n.Shutdown()
	}

	c.n = srv

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

func (c *componentNutsDB) Type() string {
	return ComponentType
}

func (c *componentNutsDB) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentNutsDB) RegisterFuncStart(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentNutsDB) RegisterFuncReload(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentNutsDB) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c._CheckDep() && c.n != nil
}

func (c *componentNutsDB) IsRunning(atLeast bool) bool {
	if c.IsStarted() {
		c.m.Lock()
		defer c.m.Unlock()

		return c.n.IsRunning()
	}

	return false
}

func (c *componentNutsDB) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentNutsDB) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentNutsDB) Stop() {
	if c.IsRunning(true) {
		c.m.Lock()
		defer c.m.Unlock()

		_ = c.n.Shutdown()
	}
}

func (c *componentNutsDB) Dependencies() []string {
	c.m.Lock()
	defer c.m.Unlock()

	if !c._CheckDep() {
		return []string{cptlog.ComponentType}
	}

	return []string{c.l}
}

func (c *componentNutsDB) GetServer() (libndb.NutsDB, liberr.Error) {
	if c.IsStarted() {
		c.m.Lock()
		defer c.m.Unlock()

		return c.n, nil
	}

	return nil, ErrorComponentNotInitialized.Error(nil)
}

func (c *componentNutsDB) SetLogger(key string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.l = key
}

func (c *componentNutsDB) GetClient(tickSync time.Duration) (libndb.Client, liberr.Error) {
	if c != nil && c.IsStarted() {
		c.m.Lock()
		defer c.m.Unlock()

		return c.n.Client(c.ctx(), tickSync), nil
	}

	return nil, ErrorComponentNotInitialized.Error(nil)
}

func (c *componentNutsDB) SetStatusRouter(sts libsts.RouteStatus, prefix string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.n.StatusRouter(sts, prefix)
}
