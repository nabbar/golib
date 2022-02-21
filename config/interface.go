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
	"os"
	"os/signal"
	"sync"
	"syscall"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	libvpr "github.com/nabbar/golib/viper"
)

type FuncContext func() context.Context
type FuncComponentGet func(key string) Component
type FuncComponentConfigGet func(model interface{}) liberr.Error

type Config interface {
	/*
		// Section Context : github.com/nabbar/golib/context
	*/

	// Context return the current context pointer
	Context() context.Context

	// ContextMerge trigger the golib/context/config interface
	// and will merge the stored context value into current context
	ContextMerge(ctx libctx.Config) bool

	// ContextStore trigger the golib/context/config interface
	// and will store a context value into current context
	ContextStore(key string, cfg interface{})

	// ContextLoad trigger the golib/context/config interface
	// and will restore a context value or nil
	ContextLoad(key string) interface{}

	// ContextSetCancel allow to register a custom function called on cancel context.
	// On context cancel event or signal kill, term... this function will be called
	// before config stop and main context cancel function
	ContextSetCancel(fct func())

	/*
		// Section Event : github.com/nabbar/golib/config
	*/

	RegisterFuncViper(fct func() libvpr.Viper)

	// Start will trigger the start function of all registered component.
	// If any component return an error, this func will stop the start
	// process and return the error.
	Start() liberr.Error

	// RegisterFuncStartBefore allow to register a func to be call when the config Start
	// is trigger. This func is call before the start sequence.
	RegisterFuncStartBefore(fct func() liberr.Error)

	// RegisterFuncStartAfter allow to register a func to be call when the config Start
	// is trigger. This func is call after the start sequence.
	RegisterFuncStartAfter(fct func() liberr.Error)

	// Reload triggers the Reload function of each registered Component.
	Reload() liberr.Error

	// RegisterFuncReloadBefore allow to register a func to be call when the config Reload
	// is trigger. This func is call before the reload sequence.
	RegisterFuncReloadBefore(fct func() liberr.Error)

	// RegisterFuncReloadAfter allow to register a func to be call when the config Reload
	// is trigger. This func is call after the reload sequence.
	RegisterFuncReloadAfter(fct func() liberr.Error)

	// Stop will trigger the stop function of all registered component.
	// All component must stop cleanly.
	Stop()

	// RegisterFuncStopBefore allow to register a func to be call when the config Stop
	// is trigger. This func is call before the stop sequence.
	RegisterFuncStopBefore(fct func())

	// RegisterFuncStopAfter allow to register a func to be call when the config Stop
	// is trigger. This func is call after the stop sequence.
	RegisterFuncStopAfter(fct func())

	/*
		// Section Component : github.com/nabbar/golib/config
	*/
	ComponentList
}

var (
	ctx context.Context
	cnl context.CancelFunc
)

func init() {
	ctx, cnl = context.WithCancel(context.Background())

	go func() {
		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT)
		signal.Notify(quit, syscall.SIGTERM)
		signal.Notify(quit, syscall.SIGQUIT)

		select {
		case <-quit:
			cnl()
		case <-ctx.Done():
			cnl()
		}
	}()
}

func New() Config {
	c := &configModel{
		m:   sync.Mutex{},
		ctx: libctx.NewConfig(ctx),
		cpt: newComponentList(),
	}

	go func() {
		select {
		case <-c.ctx.Done():
			c.cancel()
		}
	}()

	return c
}
