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

package monitor

import (
	"context"
	"fmt"

	montps "github.com/nabbar/golib/monitor/types"
	librun "github.com/nabbar/golib/runner"
)

// fctMiddleWare is a function that acts as middleware in the health check chain.
type fctMiddleWare func(m middleWare) error

// middleWare defines the interface for health check middleware chain.
// It provides access to context, configuration, and chain control methods.
type middleWare interface {
	Context() context.Context // Returns the current context with timeout
	Config() *runCfg          // Returns the runtime configuration
	Run(ctx context.Context)  // Executes the middleware chain
	Next() error              // Calls the next middleware in the chain
	Add(fct fctMiddleWare)    // Adds a middleware function to the chain
}

// mdl is the internal implementation of the middleWare interface.
type mdl struct {
	ctx context.Context // Context with timeout for the health check
	cfg *runCfg         // Runtime configuration
	crs int             // Current position in the middleware chain
	mdl []fctMiddleWare // Stack of middleware functions
}

// newMiddleware creates a new middleware chain with the given configuration and health check function.
// The health check function is wrapped as the first middleware in the chain.
func newMiddleware(cfg *runCfg, fct montps.HealthCheck) middleWare {
	defer librun.RecoveryCaller("golib/monitor/newMiddleware", recover())
	o := &mdl{
		ctx: nil,
		cfg: cfg,
		crs: 0,
		mdl: make([]fctMiddleWare, 0),
	}

	if fct != nil {
		o.Add(func(m middleWare) error {
			return fct(m.Context())
		})
	} else {
		o.Add(func(m middleWare) error {
			return fmt.Errorf("no valid healthcheck function")
		})
	}

	return o
}

// Context returns the current context with timeout for the health check.
func (m *mdl) Context() context.Context {
	defer librun.RecoveryCaller("golib/monitor/middleware/Context", recover())
	return m.ctx
}

// Config returns the runtime configuration for the health check.
func (m *mdl) Config() *runCfg {
	defer librun.RecoveryCaller("golib/monitor/middleware/Config", recover())
	return m.cfg
}

// Run executes the middleware chain with a timeout based on the configuration.
// It creates a new context with timeout and invokes the chain from the end.
func (m *mdl) Run(ctx context.Context) {
	defer librun.RecoveryCaller("golib/monitor/middleware/Run", recover())
	var cnl context.CancelFunc

	m.ctx, cnl = context.WithTimeout(ctx, m.cfg.checkTimeout)
	defer cnl()

	m.crs = len(m.mdl)
	_ = m.Next()
}

// Next invokes the next middleware function in the chain.
// Returns nil if there are no more middleware functions to execute.
func (m *mdl) Next() error {
	defer librun.RecoveryCaller("golib/monitor/middleware/Next", recover())
	m.crs--

	if m.crs >= 0 && m.crs < len(m.mdl) {
		return m.mdl[m.crs](m)
	}

	return nil
}

// Add appends a middleware function to the chain.
func (m *mdl) Add(fct fctMiddleWare) {
	defer librun.RecoveryCaller("golib/monitor/middleware/Add", recover())
	m.mdl = append(m.mdl, fct)
}
