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

package database

import (
	"sync"
	"time"

	libcfg "github.com/nabbar/golib/config"
	cptlog "github.com/nabbar/golib/config/components/log"
	libdbs "github.com/nabbar/golib/database"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

type componentDatabase struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func(cpt libcfg.Component) liberr.Error
	fsb func(cpt libcfg.Component) liberr.Error
	fra func(cpt libcfg.Component) liberr.Error
	frb func(cpt libcfg.Component) liberr.Error

	m  sync.Mutex
	l  string
	li bool
	ls time.Duration
	d  libdbs.Database
}

func (c *componentDatabase) _CheckDep() bool {
	return c != nil && c.l != ""
}

func (c *componentDatabase) _GetLogger() (liblog.Logger, liberr.Error) {
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

func (c *componentDatabase) _getFct() (func(cpt libcfg.Component) liberr.Error, func(cpt libcfg.Component) liberr.Error) {
	var isReload = c.IsStarted()

	c.m.Lock()
	defer c.m.Unlock()

	if isReload {
		return c.frb, c.fra
	} else {
		return c.fsb, c.fsa
	}
}

func (c *componentDatabase) _runFct(fct func(cpt libcfg.Component) liberr.Error) liberr.Error {
	if fct != nil {
		return fct(c)
	}

	return nil
}

func (c *componentDatabase) _runCli(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	var (
		err liberr.Error
		cnf libdbs.Config
	)

	if cnf, err = c._getConfig(getCfg); err != nil {
		return err
	}

	if c.d != nil {
		c.d.Close()
	}

	if c.d, err = libdbs.New(&cnf); err != nil {
		return ErrorStartDatabase.Error(err)
	}

	return nil
}

func (c *componentDatabase) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
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
func (c *componentDatabase) Type() string {
	return ComponentType
}

func (c *componentDatabase) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentDatabase) RegisterFuncStart(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentDatabase) RegisterFuncReload(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentDatabase) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	if c == nil || c.d == nil {
		return false
	}

	if db := c.GetDatabase(); db == nil {
		return false
	} else if e := db.CheckConn(); e != nil {
		return false
	}

	return true
}

func (c *componentDatabase) IsRunning(atLeast bool) bool {
	return c.IsStarted()
}

func (c *componentDatabase) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentDatabase) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentDatabase) Stop() {
	if db := c.GetDatabase(); db != nil {
		db.Close()
	}
}

func (c *componentDatabase) Dependencies() []string {
	if c == nil {
		return []string{cptlog.ComponentType}
	}

	c.m.Lock()
	defer c.m.Unlock()

	return []string{c.l}
}

func (c *componentDatabase) SetLOGKey(logKey string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.l = logKey
}

func (c *componentDatabase) SetLogOptions(ignoreRecordNotFoundError bool, slowThreshold time.Duration) {
	c.m.Lock()
	defer c.m.Unlock()

	c.li = ignoreRecordNotFoundError
	c.ls = slowThreshold
}

func (c *componentDatabase) SetDatabase(db libdbs.Database) {
	c.m.Lock()
	defer c.m.Unlock()

	c.d = db
}

func (c *componentDatabase) GetDatabase() libdbs.Database {
	c.m.Lock()
	defer c.m.Unlock()

	return c.d
}
