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

package http

import (
	"fmt"
	"net/http"
	"sync"

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	cptlog "github.com/nabbar/golib/config/components/log"
	cpttls "github.com/nabbar/golib/config/components/tls"
	liberr "github.com/nabbar/golib/errors"
	libhts "github.com/nabbar/golib/httpserver"
	liblog "github.com/nabbar/golib/logger"
)

type componentHttp struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

	m sync.Mutex

	tls string
	log string

	run  bool
	hand map[string]http.Handler

	pool libhts.PoolServer
}

func (c *componentHttp) _CheckDep() bool {
	return c != nil && len(c.hand) > 0 && c.tls != "" && c.log != ""
}

func (c *componentHttp) _CheckInit() bool {
	return c != nil && c._CheckDep() && c.pool != nil
}

func (c *componentHttp) _GetTLS() (libtls.TLSConfig, liberr.Error) {
	if !c._CheckDep() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if i := cpttls.Load(c.get, c.tls); i == nil {
		return nil, ErrorDependencyTLSDefault.Error(nil)
	} else if tls := i.GetTLS(); tls == nil {
		return nil, ErrorDependencyTLSDefault.Error(nil)
	} else {
		return tls, nil
	}
}

func (c *componentHttp) _GetLogger() (liblog.Logger, liberr.Error) {
	if !c._CheckDep() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if i := cptlog.Load(c.get, c.log); i == nil {
		return nil, ErrorDependencyLogDefault.Error(nil)
	} else if log := i.Log(); log == nil {
		return nil, ErrorDependencyLogDefault.Error(nil)
	} else {
		return log, nil
	}
}

func (c *componentHttp) _getPoolServerConfig(getCfg libcfg.FuncComponentConfigGet) (libhts.PoolServerConfig, liberr.Error) {
	cnf := make(libhts.PoolServerConfig, 0)

	if !c._CheckDep() {
		return cnf, ErrorComponentNotInitialized.Error(nil)
	}

	if err := getCfg(c.key, &cnf); err != nil {
		return cnf, ErrorParamsInvalid.Error(err)
	}

	if tls, err := c._GetTLS(); err != nil {
		return cnf, err
	} else {
		cnf.MapUpdate(func(sCFG libhts.ServerConfig) libhts.ServerConfig {
			sCFG.SetDefaultTLS(func() libtls.TLSConfig {
				return tls
			})

			sCFG.SetParentContext(c.ctx)
			return sCFG
		})
	}

	if err := cnf.Validate(); err != nil {
		return cnf, ErrorConfigInvalid.Error(err)
	} else if len(c.hand) < 1 {
		return cnf, ErrorComponentNotInitialized.ErrorParent(fmt.Errorf("missing handler"))
	}

	return cnf, nil
}

func (c *componentHttp) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		err liberr.Error
		cnf libhts.PoolServerConfig

		isReload = c.IsStarted()
	)

	c.m.Lock()
	defer c.m.Unlock()

	if !isReload && c.fsb != nil {
		if err = c.fsb(); err != nil {
			return err
		}
	} else if isReload && c.frb != nil {
		if err = c.frb(); err != nil {
			return err
		}
	}

	if cnf, err = c._getPoolServerConfig(getCfg); err != nil {
		return err
	}

	if c.pool != nil {
		if p, e := cnf.UpdatePoolServer(c.pool); e != nil {
			return ErrorReloadComponent.Error(e)
		} else {
			c.pool = p
		}
	} else if p, e := cnf.PoolServer(); e != nil {
		return ErrorReloadComponent.Error(e)
	} else {
		c.pool = p
	}

	c.pool.SetLogger(func() liblog.Logger {
		if log, err := c._GetLogger(); err != nil {
			return liblog.GetDefault()
		} else {
			return log
		}
	})

	if err = c.pool.ListenMultiHandler(c.hand); err != nil {
		return ErrorStartComponent.Error(err)
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

func (c *componentHttp) Type() string {
	return ComponentType
}

func (c *componentHttp) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentHttp) RegisterFuncStart(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentHttp) RegisterFuncReload(before, after func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentHttp) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c._CheckInit() && c.pool.IsRunning(true)
}

func (c *componentHttp) IsRunning(atLeast bool) bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c._CheckInit() && c.pool.IsRunning(atLeast)
}

func (c *componentHttp) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentHttp) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentHttp) Stop() {
	c.m.Lock()
	defer c.m.Unlock()

	if c._CheckInit() {
		c.pool.Shutdown()
	}
}

func (c *componentHttp) Dependencies() []string {
	c.m.Lock()
	defer c.m.Unlock()

	if !c._CheckDep() {
		return []string{cpttls.ComponentType, cptlog.ComponentType}
	}

	return []string{c.tls, c.log}
}

func (c *componentHttp) SetTLSKey(tlsKey string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.tls = tlsKey
}

func (c *componentHttp) SetLOGKey(logKey string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.log = logKey
}

func (c *componentHttp) SetHandler(handler map[string]http.Handler) {
	c.m.Lock()
	defer c.m.Unlock()

	c.hand = handler
}

func (c *componentHttp) GetPool() libhts.PoolServer {
	c.m.Lock()
	defer c.m.Unlock()

	return c.pool
}

func (c *componentHttp) SetPool(pool libhts.PoolServer) {
	c.m.Lock()
	defer c.m.Unlock()

	c.pool = pool
}
