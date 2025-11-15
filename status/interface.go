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

// Package status provides a comprehensive health check and status monitoring system for HTTP APIs.
//
// This package integrates with the Gin web framework to expose status endpoints that report
// the health of an application and its components. It supports multiple output formats (JSON/text),
// verbosity levels (short/full), and configurable health check strategies.
//
// Key features:
//   - HTTP endpoint exposure via Gin middleware
//   - Multiple output formats (JSON and plain text)
//   - Configurable verbosity (short status or detailed component information)
//   - Component health monitoring with flexible control modes
//   - Cached status computation to reduce overhead
//   - Application version and build information tracking
//
// The package works in conjunction with:
//   - github.com/nabbar/golib/monitor for component monitoring
//   - github.com/nabbar/golib/errors for error handling
//   - github.com/nabbar/golib/version for version information
//   - github.com/gin-gonic/gin for HTTP routing
//
// Basic usage:
//
//	import (
//	    "github.com/nabbar/golib/status"
//	    "github.com/nabbar/golib/context"
//	)
//
//	// Create a new status instance
//	sts := status.New(ctx)
//
//	// Set application information
//	sts.SetInfo("MyApp", "v1.0.0", "abc123")
//
//	// Register monitor pool
//	sts.RegisterPool(func() montps.Pool {
//	    return myMonitorPool
//	})
//
//	// Expose as Gin middleware
//	router.GET("/status", func(c *gin.Context) {
//	    sts.MiddleWare(c)
//	})
package status

import (
	"context"
	"sync"
	"sync/atomic"

	ginsdk "github.com/gin-gonic/gin"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

// Route defines the HTTP routing interface for exposing status endpoints.
// It provides methods to integrate status checks into Gin web applications.
type Route interface {
	// Expose handles the status endpoint request from a generic context.
	// If the context is a Gin context, it delegates to MiddleWare.
	// This method is typically used as a Gin route handler.
	Expose(ctx context.Context)

	// MiddleWare is the Gin middleware handler function that processes status requests.
	// It supports query parameters and headers for controlling output format:
	//   - Query param "short" or header "X-Verbose": controls verbosity (true = short output)
	//   - Query param "format" or header "Accept": controls output format (text/plain or application/json)
	//
	// The response includes application info, overall status, and optionally component details.
	MiddleWare(c *ginsdk.Context)

	// SetErrorReturn registers a custom error return model factory for formatting errors.
	// The function should return a new instance of liberr.ReturnGin for each call.
	// If not set, a default return model is used.
	// See github.com/nabbar/golib/errors for ReturnGin interface details.
	SetErrorReturn(f func() liberr.ReturnGin)
}

// Info defines the interface for setting application version and build information.
// This information is included in status endpoint responses.
type Info interface {
	// SetInfo manually sets the application name, release version, and build hash.
	// These values are displayed in status responses to identify the running application.
	//
	// Parameters:
	//   - name: application or service name
	//   - release: version string (e.g., "v1.2.3")
	//   - hash: build commit hash or identifier
	SetInfo(name, release, hash string)

	// SetVersion sets application information from a Version object.
	// This is the preferred method when using github.com/nabbar/golib/version.
	// It automatically extracts name, release, build hash, and build time.
	SetVersion(vers libver.Version)
}

// Pool defines the interface for managing monitor components.
// It extends montps.PoolStatus with pool registration capabilities.
// See github.com/nabbar/golib/monitor/types for PoolStatus interface details.
type Pool interface {
	montps.PoolStatus

	// RegisterPool registers a function that returns the monitor pool.
	// The pool contains all monitored components whose health contributes to overall status.
	// The function is called each time the status is computed, allowing dynamic pool updates.
	//
	// See github.com/nabbar/golib/monitor/types for FuncPool and Pool details.
	RegisterPool(fct montps.FuncPool)
}

// Status is the main interface combining all status functionality.
// It provides a complete solution for application health monitoring and status reporting.
type Status interface {
	Route
	Info
	Pool

	// SetConfig applies configuration for status computation and HTTP response codes.
	// The configuration controls:
	//   - HTTP status codes returned for each health state (OK, Warn, KO)
	//   - Mandatory component definitions and control modes
	//
	// See Config type for configuration details.
	SetConfig(cfg Config)

	// IsHealthy checks if the overall status or specific components are healthy.
	// Returns true if status is OK or Warn (>= Warn threshold).
	//
	// Parameters:
	//   - name: optional component names to check; if empty, checks overall status
	//
	// This is useful for basic health checks that tolerate warnings.
	IsHealthy(name ...string) bool

	// IsStrictlyHealthy checks if the overall status or specific components are strictly healthy.
	// Returns true only if status is OK (no warnings or errors).
	//
	// Parameters:
	//   - name: optional component names to check; if empty, checks overall status
	//
	// This is useful for strict health checks that require perfect health.
	IsStrictlyHealthy(name ...string) bool

	// IsCacheHealthy checks if the cached overall status is healthy (>= Warn).
	// Uses cached status value if within cache duration, otherwise recomputes.
	// This method is more efficient than IsHealthy for frequent checks.
	IsCacheHealthy() bool

	// IsCacheStrictlyHealthy checks if the cached overall status is strictly healthy (== OK).
	// Uses cached status value if within cache duration, otherwise recomputes.
	// This method is more efficient than IsStrictlyHealthy for frequent checks.
	IsCacheStrictlyHealthy() bool
}

// New creates a new Status instance with the given context function.
// The context function is used to obtain the current context for operations.
//
// Parameters:
//   - ctx: function that returns the current context, typically from github.com/nabbar/golib/context
//
// Returns a Status instance that must be configured before use:
//   - Call SetInfo or SetVersion to set application information
//   - Call RegisterPool to connect a monitor pool
//   - Optionally call SetConfig to customize behavior
//   - Optionally call SetErrorReturn to customize error formatting
func New(ctx context.Context) Status {
	s := &sts{
		m: sync.RWMutex{},
		p: nil,
		r: nil,
		x: libctx.New[string](ctx),
		c: ch{
			m: new(atomic.Int32),
			t: new(atomic.Value),
			c: new(atomic.Int64),
			f: nil,
		},
		fn: nil,
		fr: nil,
		fh: nil,
		fd: nil,
	}

	s.c.f = func() monsts.Status {
		r, _ := s.getStatus()
		return r
	}

	return s
}
