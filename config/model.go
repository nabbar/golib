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

package config

import (
	"context"
	"fmt"
	"os"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	libvpr "github.com/nabbar/golib/viper"
	spfcbr "github.com/spf13/cobra"
	spfvpr "github.com/spf13/viper"

	"io"
	"sync"
)

type configModel struct {
	m sync.Mutex

	ctx  libctx.Config
	fcnl func()

	cpt ComponentList

	fctGolibViper   func() libvpr.Viper
	fctStartBefore  func() liberr.Error
	fctStartAfter   func() liberr.Error
	fctReloadBefore func() liberr.Error
	fctReloadAfter  func() liberr.Error
	fctStopBefore   func()
	fctStopAfter    func()
}

func (c *configModel) _ComponentGetConfig(key string, model interface{}) liberr.Error {
	var (
		err error
		vpr libvpr.Viper
		vip *spfvpr.Viper
	)

	if c.cpt.ComponentHas(key) {
		if c.fctGolibViper == nil {
			return ErrorConfigMissingViper.Error(nil)
		} else if vpr = c.fctGolibViper(); vpr == nil {
			return ErrorConfigMissingViper.Error(nil)
		} else if vip = vpr.Viper(); vip == nil {
			return ErrorConfigMissingViper.Error(nil)
		}

		err = vip.UnmarshalKey(key, model)
	} else {
		return ErrorComponentNotFound.ErrorParent(fmt.Errorf("component '%s'", key))
	}

	return ErrorComponentConfigError.Iferror(err)
}

func (c *configModel) Context() context.Context {
	return c.ctx
}

func (c *configModel) ContextMerge(ctx libctx.Config) bool {
	return c.ctx.Merge(ctx)
}

func (c *configModel) ContextStore(key string, cfg interface{}) {
	c.ctx.Store(key, cfg)
}

func (c *configModel) ContextLoad(key string) interface{} {
	return c.ctx.Load(key)
}

func (c *configModel) ContextSetCancel(fct func()) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fcnl = fct
}

func (c *configModel) cancel() {
	c.cancelCustom()
	c.Stop()
}

func (c *configModel) cancelCustom() {
	c.m.Lock()
	defer c.m.Unlock()

	if c.fcnl != nil {
		c.fcnl()
	}
}

func (c *configModel) RegisterFuncViper(fct func() libvpr.Viper) {
	c.fctGolibViper = fct
}

func (c *configModel) Start() liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.fctStartBefore != nil {
		if err := c.fctStartBefore(); err != nil {
			return err
		}
	}

	if err := c.cpt.ComponentStart(c._ComponentGetConfig); err != nil {
		return err
	}

	if c.fctStartAfter != nil {
		if err := c.fctStartAfter(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configModel) RegisterFuncStartBefore(fct func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()
	c.fctStartBefore = fct
}

func (c *configModel) RegisterFuncStartAfter(fct func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()
	c.fctStartAfter = fct
}

func (c *configModel) Reload() liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	if c.fctReloadBefore != nil {
		if err := c.fctReloadBefore(); err != nil {
			return err
		}
	}

	if err := c.cpt.ComponentReload(c._ComponentGetConfig); err != nil {
		return err
	}

	if c.fctReloadAfter != nil {
		if err := c.fctReloadAfter(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configModel) RegisterFuncReloadBefore(fct func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()
	c.fctReloadBefore = fct
}

func (c *configModel) RegisterFuncReloadAfter(fct func() liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()
	c.fctReloadAfter = fct
}

func (c *configModel) Stop() {
	c.m.Lock()
	defer c.m.Unlock()

	if c.fctStopBefore != nil {
		c.fctStopBefore()
	}

	for _, k := range c.ComponentKeys() {
		cpt := c.ComponentGet(k)

		if cpt == nil {
			continue
		}

		cpt.Stop()
	}

	if c.fctStopAfter != nil {
		c.fctStopAfter()
	}
}

func (c *configModel) RegisterFuncStopBefore(fct func()) {
	c.m.Lock()
	defer c.m.Unlock()
	c.fctStopBefore = fct
}

func (c *configModel) RegisterFuncStopAfter(fct func()) {
	c.m.Lock()
	defer c.m.Unlock()
	c.fctStopAfter = fct
}

func (c *configModel) Shutdown(code int) {
	c.cancel()
	os.Exit(code)
}

func (c *configModel) ComponentHas(key string) bool {
	return c.cpt.ComponentHas(key)
}

func (c *configModel) ComponentType(key string) string {
	return c.cpt.ComponentType(key)
}

func (c *configModel) ComponentGet(key string) Component {
	return c.cpt.ComponentGet(key)
}

func (c *configModel) ComponentDel(key string) {
	c.cpt.ComponentDel(key)
}

func (c *configModel) ComponentSet(key string, cpt Component) {
	cpt.Init(key, c.Context, c.ComponentGet, func() *spfvpr.Viper {
		if c.fctGolibViper == nil {
			return nil
		} else if vpr := c.fctGolibViper(); vpr == nil {
			return nil
		} else {
			return vpr.Viper()
		}
	})

	c.cpt.ComponentSet(key, cpt)
}

func (c *configModel) ComponentList() map[string]Component {
	return c.cpt.ComponentList()
}

func (c *configModel) ComponentKeys() []string {
	return c.cpt.ComponentKeys()
}

func (c *configModel) ComponentStart(getCfg FuncComponentConfigGet) liberr.Error {
	return c.cpt.ComponentStart(getCfg)
}

func (c *configModel) ComponentIsStarted() bool {
	return c.cpt.ComponentIsStarted()
}

func (c *configModel) ComponentReload(getCfg FuncComponentConfigGet) liberr.Error {
	return c.cpt.ComponentReload(getCfg)
}

func (c *configModel) ComponentStop() {
	c.cpt.ComponentStop()
}

func (c *configModel) ComponentIsRunning(atLeast bool) bool {
	return c.cpt.ComponentIsRunning(atLeast)
}

func (c *configModel) DefaultConfig() io.Reader {
	return c.cpt.DefaultConfig()
}

func (c *configModel) RegisterFlag(Command *spfcbr.Command, Viper *spfvpr.Viper) error {
	return c.cpt.RegisterFlag(Command, Viper)
}
