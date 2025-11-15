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
	"os"
)

// Start initiates the startup sequence for all registered components.
// The execution order is:
//  1. Execute all before-start hooks
//  2. Start all components in dependency order
//  3. Execute all after-start hooks
//
// If any hook or component returns an error, the sequence stops and the error is returned.
// Components are started sequentially according to their dependencies.
func (o *model) Start() error {
	if err := o.runFuncStartBefore(); err != nil {
		return err
	}

	if err := o.ComponentStart(); err != nil {
		return err
	}

	if err := o.runFuncStartAfter(); err != nil {
		return err
	}

	return nil
}

// Reload reloads all registered components without stopping them.
// The execution order is:
//  1. Execute all before-reload hooks
//  2. Reload all components in dependency order
//  3. Execute all after-reload hooks
//
// This allows components to refresh their configuration without a full restart.
// If any hook or component returns an error, the sequence stops and the error is returned.
func (o *model) Reload() error {
	if err := o.runFuncReloadBefore(); err != nil {
		return err
	}

	if err := o.ComponentReload(); err != nil {
		return err
	}

	if err := o.runFuncReloadAfter(); err != nil {
		return err
	}

	return nil
}

// Stop gracefully stops all registered components.
// The execution order is:
//  1. Execute all before-stop hooks
//  2. Stop all components in reverse dependency order
//  3. Execute all after-stop hooks
//
// This method does not return errors - it performs best-effort cleanup.
// Hook errors are ignored to ensure all components are stopped.
func (o *model) Stop() {
	_ = o.runFuncStopBefore()
	o.ComponentStop()
	_ = o.runFuncStopAfter()
}

// Shutdown performs complete application termination.
// It executes all registered cancel functions, stops all components,
// and then exits the process with the specified exit code.
//
// This method does not return - it terminates the process via os.Exit().
func (o *model) Shutdown(code int) {
	o.cancel()
	os.Exit(code)
}
