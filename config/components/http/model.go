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
	fsa func() liberr.Error
	fsb func() liberr.Error
	fra func() liberr.Error
	frb func() liberr.Error

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

func (c *componentHttp) Type() string {
	return ComponentType
}

func (c *componentHttp) RegisterContext(fct libcfg.FuncContext) {
	c.ctx = fct
}

func (c *componentHttp) RegisterGet(fct libcfg.FuncComponentGet) {
	c.get = fct
}

func (c *componentHttp) RegisterFuncStartBefore(fct func() liberr.Error) {
	c.fsb = fct
}

func (c *componentHttp) RegisterFuncStartAfter(fct func() liberr.Error) {
	c.fsa = fct
}

func (c *componentHttp) RegisterFuncReloadBefore(fct func() liberr.Error) {
	c.frb = fct
}

func (c *componentHttp) RegisterFuncReloadAfter(fct func() liberr.Error) {
	c.fra = fct
}

func (c *componentHttp) getTLS(getCpt libcfg.FuncComponentGet) (libtls.TLSConfig, liberr.Error) {
	if !c._CheckDep() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if i := cpttls.Load(getCpt, c.tls); i == nil {
		return nil, ErrorDependencyTLSDefault.Error(nil)
	} else if tls := i.GetTLS(); tls == nil {
		return nil, ErrorDependencyTLSDefault.Error(nil)
	} else {
		return tls, nil
	}
}

func (c *componentHttp) getLogger(getCpt libcfg.FuncComponentGet) (liblog.Logger, liberr.Error) {
	if !c._CheckDep() {
		return nil, ErrorComponentNotInitialized.Error(nil)
	}

	if i := cptlog.Load(getCpt, c.log); i == nil {
		return nil, ErrorDependencyLogDefault.Error(nil)
	} else if log := i.Log(); log == nil {
		return nil, ErrorDependencyLogDefault.Error(nil)
	} else {
		return log, nil
	}
}

func (c *componentHttp) _getPoolServerConfig(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) (libhts.PoolServerConfig, liberr.Error) {
	cnf := make(libhts.PoolServerConfig, 0)

	if !c._CheckDep() {
		return cnf, ErrorComponentNotInitialized.Error(nil)
	}

	if err := getCfg(&cnf); err != nil {
		return cnf, ErrorParamsInvalid.Error(err)
	}

	if tls, err := c.getTLS(getCpt); err != nil {
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

func (c *componentHttp) Start(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		err liberr.Error
		cnf libhts.PoolServerConfig
	)

	if c.fsb != nil {
		if err = c.fsb(); err != nil {
			return err
		}
	}

	if cnf, err = c._getPoolServerConfig(getCpt, getCfg); err != nil {
		return err
	}

	if p, e := cnf.PoolServer(); e != nil {
		return ErrorStartPoolServer.Error(e)
	} else {
		c.pool = p
	}

	c.pool.SetLogger(func() liblog.Logger {
		if log, err := c.getLogger(getCpt); err != nil {
			return liblog.GetDefault()
		} else {
			return log
		}
	})

	if err = c.pool.ListenMultiHandler(c.hand); err != nil {
		return ErrorStartPoolServer.Error(err)
	}

	if c.fsa != nil {
		if err = c.fsa(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentHttp) Reload(getCpt libcfg.FuncComponentGet, getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	var (
		err liberr.Error
		cnf libhts.PoolServerConfig
	)

	if c.frb != nil {
		if err = c.frb(); err != nil {
			return err
		}
	}

	if cnf, err = c._getPoolServerConfig(getCpt, getCfg); err != nil {
		return err
	}

	if c.pool != nil {
		if p, e := cnf.UpdatePoolServer(c.pool); e != nil {
			return ErrorReloadPoolServer.Error(e)
		} else {
			c.pool = p
		}
	} else if p, e := cnf.PoolServer(); e != nil {
		return ErrorReloadPoolServer.Error(e)
	} else {
		c.pool = p
	}

	c.pool.SetLogger(func() liblog.Logger {
		if log, err := c.getLogger(getCpt); err != nil {
			return liblog.GetDefault()
		} else {
			return log
		}
	})

	c.pool.Restart()

	if c.fra != nil {
		if err = c.fra(); err != nil {
			return err
		}
	}

	return nil
}

func (c *componentHttp) Stop() {
	c.run = false

	if !c._CheckInit() {
		return
	}

	c.pool.Shutdown()
}

func (c *componentHttp) IsStarted() bool {
	return c._CheckInit() && c.pool.IsRunning(true)
}

func (c *componentHttp) IsRunning(atLeast bool) bool {
	return c._CheckInit() && c.pool.IsRunning(atLeast)
}

func (c *componentHttp) Dependencies() []string {
	return []string{cpttls.ComponentType, cptlog.ComponentType}
}

func (c *componentHttp) SetTLSKey(tlsKey string) {
	if c == nil {
		return
	}

	c.tls = tlsKey
}

func (c *componentHttp) SetLOGKey(logKey string) {
	if c == nil {
		return
	}

	c.log = logKey
}

func (c *componentHttp) SetHandler(handler map[string]http.Handler) {
	if c == nil {
		return
	}

	c.hand = handler
}

func (c *componentHttp) GetPool() libhts.PoolServer {
	if c == nil || c.pool == nil {
		return nil
	}

	return c.pool
}

func (c *componentHttp) SetPool(pool libhts.PoolServer) {
	if c == nil {
		return
	}

	c.pool = pool
}
