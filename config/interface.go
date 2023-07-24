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

	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	shlcmd "github.com/nabbar/golib/shell/command"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

type FuncEvent func() liberr.Error

type Config interface {
	/*
		// Section Context : github.com/nabbar/golib/context
	*/

	// Context return the config context instance
	Context() libctx.Config[string]

	// CancelAdd allow to register a slice of custom function called on cancel context.
	// On context cancel event or signal kill, term... this function will be called
	// before config stop and main context cancel function
	CancelAdd(fct ...func())

	// CancelClean allow clear the all Cancel func registered into slice
	CancelClean()

	/*
		// Section Manage : github.com/nabbar/golib/config
	*/

	// Start will trigger the start function of all registered component.
	// If any component return an error, this func will stop the start
	// process and return the error.
	Start() liberr.Error

	// Reload triggers the Reload function of each registered Component.
	Reload() liberr.Error

	// Stop will trigger the stop function of all registered component.
	// All component must stop cleanly.
	Stop()

	// Shutdown will trigger all stop function.
	// This function will call the Stop function and the private function cancel.
	// This will stop all process and do like a SIGTERM/SIGINT signal.
	// This will finish by an os.Exit with the given parameter code.
	Shutdown(code int)

	/*
		// Section Events : github.com/nabbar/golib/config
	*/

	// RegisterFuncViper is used to expose golib Viper instance to all config component.
	// With this function, the component can load his own config part and start or reload.
	RegisterFuncViper(fct libvpr.FuncViper)

	// RegisterFuncStartBefore allow to register a func to be call when the config Start
	// is trigger. This func is call before the start sequence.
	RegisterFuncStartBefore(fct FuncEvent)

	// RegisterFuncStartAfter allow to register a func to be call when the config Start
	// is trigger. This func is call after the start sequence.
	RegisterFuncStartAfter(fct FuncEvent)

	// RegisterFuncReloadBefore allow to register a func to be call when the config Reload
	// is trigger. This func is call before the reload sequence.
	RegisterFuncReloadBefore(fct FuncEvent)

	// RegisterFuncReloadAfter allow to register a func to be call when the config Reload
	// is trigger. This func is call after the reload sequence.
	RegisterFuncReloadAfter(fct FuncEvent)

	// RegisterFuncStopBefore allow to register a func to be call when the config Stop
	// is trigger. This func is call before the stop sequence.
	RegisterFuncStopBefore(fct FuncEvent)

	// RegisterFuncStopAfter allow to register a func to be call when the config Stop
	// is trigger. This func is call after the stop sequence.
	RegisterFuncStopAfter(fct FuncEvent)

	// RegisterDefaultLogger allow to register a func to return a default logger.
	// This logger can be used by component to extend config ot to log message.
	RegisterDefaultLogger(fct liblog.FuncLog)

	/*
		// Section Component : github.com/nabbar/golib/config
	*/
	cfgtps.ComponentList
	cfgtps.ComponentMonitor

	/*
		// Section Shell Command : github.com/nabbar/golib/shell
	*/
	GetShellCommand() []shlcmd.Command
}

var (
	ctx context.Context
	cnl context.CancelFunc
)

func init() {
	ctx, cnl = context.WithCancel(context.Background())
}

func Shutdown() {
	cnl()
}

func WaitNotify() {
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
}

func New(vrs libver.Version) Config {
	fct := func() context.Context {
		return ctx
	}

	c := &configModel{
		m:    sync.RWMutex{},
		ctx:  libctx.NewConfig[string](fct),
		cpt:  libctx.NewConfig[string](fct),
		fct:  libctx.NewConfig[uint8](fct),
		fcnl: nil,
	}

	c.RegisterVersion(vrs)

	go func() {
		select {
		case <-c.ctx.Done():
			c.cancel()
		}
	}()

	return c
}
