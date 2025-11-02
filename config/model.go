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
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
)

// Internal constants for function registry keys.
// These are used to store registered functions in the fct context map.
const (
	fctViper        uint8 = iota + 1 // Viper configuration provider function
	fctStartBefore                   // Before-start hook function
	fctStartAfter                    // After-start hook function
	fctReloadBefore                  // Before-reload hook function
	fctReloadAfter                   // After-reload hook function
	fctStopBefore                    // Before-stop hook function
	fctStopAfter                     // After-stop hook function
	fctVersion                       // Application version information
	fctLoggerDef                     // Default logger provider function
	fctMonitorPool                   // Monitor pool provider function
)

// model is the internal implementation of the Config interface.
// It provides thread-safe component orchestration and lifecycle management.
//
// Fields:
//   - ctx: Shared application context for component communication
//   - cpt: Component registry (thread-safe map of components)
//   - fct: Function registry for hooks and providers (thread-safe map)
//   - cnl: Slice of custom cancel functions (mutex-protected)
//
// Thread Safety:
//   - Component operations are thread-safe via context synchronization
//   - Cancel function list is protected by mutex m
//   - Hook registrations are thread-safe via context storage
type model struct {
	ctx libctx.Config[string]                       // Shared application context
	cpt libatm.MapTyped[string, cfgtps.Component]   // Component registry
	fct libatm.MapTyped[uint8, any]                 // Function and hook registry
	cnl libatm.MapTyped[uint64, context.CancelFunc] // Custom cancel functions (mutex-protected)
	seq *atomic.Uint64                              // sequence for cancel function
}
