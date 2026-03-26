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

// fctMiddleWare defines a function signature for middleware components in the health check chain.
// Each middleware receives a reference to the chain (middleWare) and is responsible for
// executing logic before or after calling Next() to proceed to the next component in the stack.
//
// Middleware can be used for logging, status updates, metrics collection, or modifying
// the execution context.
type fctMiddleWare func(m middleWare) error

// middleWare is an interface that describes the contract for a health check execution pipeline.
// It manages a stack of middleware functions, providing them with context and configuration,
// and controlling the flow of execution from one middleware to the next.
type middleWare interface {
	// Context returns the execution context associated with the current health check run.
	// This context usually includes a timeout derived from the monitor's configuration (checkTimeout).
	Context() context.Context

	// Config returns the runtime configuration (runCfg) used for this specific health check execution.
	Config() *runCfg

	// Run starts the execution of the middleware stack using the provided base context.
	// It applies the configured timeout and initiates the traversal of the middleware chain.
	Run(ctx context.Context)

	// Next triggers the execution of the next middleware function in the stack.
	// It returns the error encountered during the execution of the subsequent chain components.
	Next() error

	// Add appends a new middleware function to the execution stack.
	// Middlewares are typically executed in reverse order of addition (LIFO).
	Add(fct fctMiddleWare)
}

// mdl is the private implementation of the middleWare interface.
// It keeps track of the execution state, including the context, configuration, and the stack position.
type mdl struct {
	// ctx is the active context for the current middleware execution, including timeout.
	ctx context.Context
	// cfg is the snapshot of the monitor's runtime configuration.
	cfg *runCfg
	// crs (cursor) tracks the current position in the middleware stack during execution.
	crs int
	// mdl is the slice of middleware functions that form the execution chain.
	mdl []fctMiddleWare
}

// newMiddleware initializes and returns a new middleware chain instance.
//
// Parameters:
//   - cfg: The runtime configuration containing thresholds and timeouts.
//   - fct: The primary health check function to be executed at the end of the chain.
//
// The health check function is automatically wrapped and added as the final element (base) of the chain.
// If the health check function is nil, a default error-returning function is used instead.
func newMiddleware(cfg *runCfg, fct montps.HealthCheck) middleWare {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/middleware/newMiddleware", r)
		}
	}()

	o := &mdl{
		ctx: nil,
		cfg: cfg,
		crs: 0,
		mdl: make([]fctMiddleWare, 0),
	}

	// Add the actual health check logic as the first item in the slice.
	// Note: Since the chain executes from the end of the slice (see Run/Next),
	// this base function will be the last to execute.
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

// Context retrieves the current execution context.
// This context is populated and valid only after the Run() method has been called.
func (m *mdl) Context() context.Context {
	return m.ctx
}

// Config retrieves the runtime configuration associated with the middleware chain.
func (m *mdl) Config() *runCfg {
	return m.cfg
}

// Run initiates the execution of the middleware pipeline.
//
// Workflow:
//  1. Creates a child context with a timeout derived from m.cfg.checkTimeout.
//  2. Initializes the cursor (m.crs) to the end of the middleware stack.
//  3. Triggers the first execution by calling Next().
//  4. Automatically cancels the context upon completion.
func (m *mdl) Run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/middleware/Run", r)
		}
	}()

	var cnl context.CancelFunc

	// Ensure the execution is bounded by the checkTimeout configuration.
	m.ctx, cnl = context.WithTimeout(ctx, m.cfg.checkTimeout)
	defer cnl()

	// Set the cursor to the top of the stack (last added item).
	m.crs = len(m.mdl)
	// Start the chain traversal.
	_ = m.Next()
}

// Next identifies and executes the next middleware function in the chain.
// It decrements the cursor and invokes the corresponding function from the stack.
// If no more functions are available (cursor < 0), it returns nil.
//
// Security:
// It includes a recovery mechanism to catch panics within middleware functions,
// converting them into standard errors to prevent the monitor runner from crashing.
func (m *mdl) Next() (err error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/middleware/Next", r)
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	// Move to the next element in the stack.
	m.crs--

	// Execute the function at the current cursor position.
	if m.crs >= 0 && m.crs < len(m.mdl) {
		return m.mdl[m.crs](m)
	}

	// End of the chain.
	return nil
}

// Add appends a new middleware function to the internal stack.
// Because the execution starts from the end of the slice, functions added later
// will be executed earlier in the pipeline (behaving like a LIFO stack).
func (m *mdl) Add(fct fctMiddleWare) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/middleware/Add", r)
		}
	}()

	m.mdl = append(m.mdl, fct)
}
