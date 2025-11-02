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
	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	htpool "github.com/nabbar/golib/httpserver/pool"
	srvtps "github.com/nabbar/golib/httpserver/types"
)

// mod is the internal implementation of the CptHttp interface.
// It uses atomic values for thread-safe access to configuration and state.
type mod struct {
	x libctx.Config[uint8]             // Context configuration storage for component state
	t libatm.Value[string]             // TLS component key
	h libatm.Value[srvtps.FuncHandler] // HTTP handler function
	s libatm.Value[htpool.Pool]        // Server pool instance
}

// SetTLSKey updates the TLS component key used for TLS configuration.
// This method is thread-safe.
//
// Parameters:
//   - tlsKey: The new TLS component key
func (o *mod) SetTLSKey(tlsKey string) {
	o.t.Store(tlsKey)
}

// SetHandler updates the HTTP handler function.
// The handler function is called when building or updating the server pool.
// This method is thread-safe.
//
// Parameters:
//   - fct: Function that returns a map of route keys to HTTP handlers
func (o *mod) SetHandler(fct srvtps.FuncHandler) {
	o.h.Store(fct)
}

// GetPool returns the current HTTP server pool.
// This method is thread-safe.
//
// Returns:
//   - The current server pool, or nil if not initialized
func (o *mod) GetPool() htpool.Pool {
	return o.s.Load()
}

// SetPool sets the HTTP server pool.
// If nil is passed, a new empty pool is created with the current handler.
// This method is thread-safe.
//
// Parameters:
//   - pool: The new server pool, or nil to create a new one
//
// Note: When setting a non-nil pool, it is used directly.
// When setting nil, a new pool is created but it will be empty until
// the component is started or reloaded with configuration.
func (o *mod) SetPool(pool htpool.Pool) {
	if pool == nil {
		pool = htpool.New(o.x, o.h.Load())
	}

	o.s.Store(pool)
}
