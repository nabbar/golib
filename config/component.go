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
	liberr "github.com/nabbar/golib/errors"
)

type Component interface {
	// Type return the component type.
	Type() string

	// RegisterContext is called by Config to register a function to get the main context.
	// This function can be used into start / reload function to use context interface.
	RegisterContext(fct FuncContext)

	// RegisterGet is called by Config to register a function to get a component by his key.
	// This function can be used for dependencies into start / reload function.
	RegisterGet(fct FuncComponentGet)

	// RegisterFuncStartBefore is called to register a function to be called before the start function.
	RegisterFuncStartBefore(fct func() liberr.Error)

	// RegisterFuncStartAfter is called to register a function to be called after the start function.
	RegisterFuncStartAfter(fct func() liberr.Error)

	// RegisterFuncReloadBefore is called to register a function to be called before the reload function.
	RegisterFuncReloadBefore(fct func() liberr.Error)

	// RegisterFuncReloadAfter is called to register a function to be called after the reload function.
	RegisterFuncReloadAfter(fct func() liberr.Error)

	// Start is called by the Config interface when the global configuration as been started
	// This function can be usefull to start server in go routine with a configuration stored
	// itself.
	Start(getCpt FuncComponentGet, getCfg FuncComponentConfigGet) liberr.Error

	// IsStarted is trigger by the Config interface with function ComponentIsStarted.
	// This function can be usefull to know if the start server function is still call.
	IsStarted() bool

	// Reload is called by the Config interface when the global configuration as been updated
	// It receives a func as param to grab a config model by sending a model structure.
	// It must configure itself, and stop / start his server if possible or return an error.
	Reload(getCpt FuncComponentGet, getCfg FuncComponentConfigGet) liberr.Error

	// Stop is called by the Config interface when global context is done.
	// The context done can arrive by stopping the application or by received a signal KILL/TERM.
	// This function must stop cleanly the component.
	Stop()

	// IsRunning is trigger by the Config interface with function ComponentIsRunning.
	// This function can be usefully to know if the component server function is still call.
	// The atLeast param is used to know if the function must return true on first server is running
	// or if all server must be running to return true.
	IsRunning(atLeast bool) bool

	// DefaultConfig is called by Config.GetDefault.
	// It must return a slice of byte containing the default json config for this component.
	DefaultConfig() []byte

	// Dependencies is called by Config to define if this component need other component.
	// Each other component can be call by calling Config.Get
	Dependencies() []string
}
