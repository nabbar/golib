/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package pool

import (
	"context"

	libtls "github.com/nabbar/golib/certificates"
	libhtp "github.com/nabbar/golib/httpserver"
	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
)

// Config is a slice of server configurations used to create a pool of servers.
// It provides convenience methods for bulk operations on multiple server configurations.
type Config []libhtp.Config

// FuncWalkConfig is a callback function for iterating over server configurations.
// Return true to continue iteration, false to stop.
type FuncWalkConfig func(cfg libhtp.Config) bool

// SetHandlerFunc registers the same handler function with all server configurations in the slice.
// This is useful for setting a shared handler across multiple servers before pool creation.
func (p Config) SetHandlerFunc(hdl srvtps.FuncHandler) {
	for i, c := range p {
		c.RegisterHandlerFunc(hdl)
		p[i] = c
	}
}

// SetDefaultTLS sets the default TLS configuration provider for all server configurations.
// This allows servers to inherit a shared TLS configuration when needed.
func (p Config) SetDefaultTLS(f libtls.FctTLSDefault) {
	for i, c := range p {
		c.SetDefaultTLS(f)
		p[i] = c
	}
}

// SetContext sets the context provider function for all server configurations.
// This provides a shared context source for all servers in the configuration.
func (p Config) SetContext(f context.Context) {
	for i, c := range p {
		c.SetContext(f)
		p[i] = c
	}
}

// Pool creates a new server pool from the configurations.
// All configurations are validated and instantiated as servers in the pool.
// Returns an error if any configuration is invalid or server creation fails.
//
// Parameters:
//   - ctx: Context provider for server operations (can be nil)
//   - hdl: Handler function for all servers (can be nil if already set on configs)
//   - defLog: Default logger function (can be nil)
//
// Returns:
//   - Pool: Initialized pool with all servers
//   - error: Aggregated errors from server creation, nil if all succeed
func (p Config) Pool(ctx context.Context, hdl srvtps.FuncHandler, defLog liblog.FuncLog) (Pool, error) {
	var (
		r = New(ctx, hdl)
		e = ErrorPoolAdd.Error(nil)
	)

	p.Walk(func(cfg libhtp.Config) bool {
		if err := r.StoreNew(cfg, defLog); err != nil {
			e.Add(err)
		}
		return true
	})

	if !e.HasParent() {
		e = nil
	}

	return r, e
}

// Walk iterates over all configurations, calling the provided function for each.
// Iteration stops if the callback returns false or all configurations have been processed.
// Does nothing if the callback function is nil.
func (p Config) Walk(fct FuncWalkConfig) {
	if fct == nil {
		return
	}

	for _, c := range p {
		if !fct(c) {
			return
		}
	}
}

// Validate validates all server configurations in the slice.
// Returns an aggregated error containing all validation failures, or nil if all are valid.
//
// Returns:
//   - error: Aggregated validation errors, nil if all configurations are valid
func (p Config) Validate() error {
	var e = ErrorPoolValidate.Error(nil)

	p.Walk(func(cfg libhtp.Config) bool {
		var err error

		if err = cfg.Validate(); err != nil {
			e.Add(err)
		}

		return true
	})

	if !e.HasParent() {
		e = nil
	}

	return e
}
