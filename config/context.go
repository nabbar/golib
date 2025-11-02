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

	libctx "github.com/nabbar/golib/context"
)

// Context returns the shared application context for all components.
// This context is used for:
//   - Storing shared application state (key-value pairs)
//   - Coordinating cancellation across components
//   - Providing context to component operations
//
// The context is thread-safe and can be accessed concurrently by multiple components.
func (o *model) Context() libctx.Config[string] {
	return o.ctx
}

// CancelAdd registers custom functions to execute on context cancellation.
// These functions are called before Stop() when:
//   - Application receives termination signals (SIGINT, SIGTERM, SIGQUIT)
//   - Shutdown() is called explicitly
//   - The shared context is cancelled
//
// Parameters:
//   - fct: Variadic list of functions to register
//
// Use cases:
//   - Flush buffers before shutdown
//   - Close external connections
//   - Save state to persistent storage
//   - Send shutdown notifications
//
// Thread-safe: Uses mutex protection for concurrent access.
func (o *model) CancelAdd(fct ...func()) {
	for _, f := range fct {
		if fct == nil {
			continue
		}

		o.seq.Add(1)
		o.cnl.Store(o.seq.Load(), f)
	}
}

// CancelClean removes all registered cancel functions.
// This resets the cancellation handler list to empty.
//
// Typically used in testing or when reinitializing the configuration.
// Does not affect components; only removes custom cancel handlers.
//
// Thread-safe: Uses mutex protection for concurrent access.
func (o *model) CancelClean() {
	o.cnl.Range(func(k uint64, _ context.CancelFunc) bool {
		o.cnl.Delete(k)
		return true
	})
}

// cancel is the internal cancellation handler.
// It executes all registered cancel functions and then stops all components.
// Called automatically when the context is cancelled or Shutdown() is invoked.
//
// Execution order:
//  1. Execute all registered cancel functions (from CancelAdd)
//  2. Call Stop() to shutdown all components
func (o *model) cancel() {
	o.cnl.Range(func(k uint64, f context.CancelFunc) bool {
		o.cnl.Delete(k)
		f()
		return true
	})

	o.Stop()
}
