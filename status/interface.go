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

package status

import (
	"context"
	"encoding"
	"encoding/json"
	"sync"
	"sync/atomic"

	ginsdk "github.com/gin-gonic/gin"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	monsts "github.com/nabbar/golib/monitor/status"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

// FuncGetCfgCpt defines a function type for retrieving a component by its key.
// This function acts as a bridge between the status package and an external
// component management system. It is used to dynamically load monitor names
// from component configurations when `ConfigKeys` is used in the `Mandatory`
// section of the status configuration.
//
// The returned `cfgtps.ComponentMonitor` must provide the monitor names
// via its `GetMonitorNames()` method.
type FuncGetCfgCpt func(key string) cfgtps.ComponentMonitor

// Route defines the HTTP routing interface for exposing the status endpoint.
// This interface groups all methods related to handling HTTP requests, primarily
// for integration with the Gin web framework.
type Route interface {
	// Expose provides a generic handler for status requests that can be used
	// with any framework that uses `context.Context`. It acts as a wrapper
	// around the `MiddleWare` method, allowing for broader framework compatibility.
	// If the provided context is a `*gin.Context`, it will be handled; otherwise,
	// the call is a no-op.
	//
	// Parameters:
	//   - ctx: The context of the request, which may be a `*gin.Context`.
	Expose(ctx context.Context)

	// MiddleWare is a Gin middleware that processes status requests. It handles
	// content negotiation for format (JSON/text) and verbosity (full/short)
	// based on query parameters and HTTP headers. It calculates the current
	// health status and renders the appropriate response.
	//
	// Parameters:
	//   - c: The `*gin.Context` for the incoming HTTP request.
	MiddleWare(c *ginsdk.Context)

	// SetErrorReturn registers a custom factory function for creating error
	// formatters. This allows you to define how errors generated within the
	// status middleware are rendered in the HTTP response, ensuring consistent
	// error formatting across your application.
	//
	// Parameters:
	//   - f: A function that returns an instance of `liberr.ReturnGin`.
	SetErrorReturn(f func() liberr.ReturnGin)
}

// Info defines the interface for setting the application's version and build information.
// This information is included in the status response to help identify the exact
// version of the running service.
type Info interface {
	// SetInfo manually sets the application name, release version, and build hash.
	// This is a straightforward way to provide static version information.
	//
	// Parameters:
	//   - name: The name of the application (e.g., "my-api").
	//   - release: The release version (e.g., "v1.2.3").
	//   - hash: The build commit hash (e.g., "abcdef1").
	SetInfo(name, release, hash string)

	// SetVersion sets the application information from a `version.Version` object.
	// This is the recommended approach as it allows the version information to be
	// managed centrally and updated dynamically if needed.
	//
	// Parameters:
	//   - vers: An object that implements the `libver.Version` interface.
	SetVersion(vers libver.Version)
}

// Pool defines the interface for managing the monitor components that contribute
// to the overall health status. It embeds `montps.PoolStatus` to provide
// basic monitor management (Add, Del, Get, etc.) and adds a method for
// registering the pool itself.
type Pool interface {
	montps.PoolStatus

	// RegisterPool registers a function that provides the monitor pool. This
	// dependency injection pattern is crucial, as it decouples the status
	// package from the monitor pool's lifecycle. The status package can then
	// retrieve the most up-to-date pool instance whenever it needs to perform
	// a health check.
	//
	// Parameters:
	//   - fct: A function that returns an instance of `montps.Pool`.
	RegisterPool(fct montps.FuncPool)
}

// Status is the main interface that combines all health status functionality.
// It provides a complete solution for application health monitoring and reporting,
// from configuration to HTTP response rendering.
type Status interface {
	encoding.TextMarshaler
	json.Marshaler

	Route
	Info
	Pool

	// RegisterGetConfigCpt registers a function to retrieve component configurations
	// by key. This is essential for the dynamic `ConfigKeys` feature, allowing
	// the status system to query an external configuration manager for components
	// and their associated monitors.
	//
	// Parameters:
	//   - fct: The function that will be called to resolve a component key.
	RegisterGetConfigCpt(fct FuncGetCfgCpt)

	// SetConfig applies a configuration for status computation and HTTP response codes.
	// This is the primary method for defining your health check policies, including
	// which components are critical and how their failures should be handled.
	//
	// Parameters:
	//   - cfg: The `Config` object containing the desired settings.
	SetConfig(cfg Config)

	// GetConfig returns the current configuration used for status computation.
	// This can be useful for debugging or for services that need to inspect
	// the current health check policy.
	//
	// Returns:
	//   The current `Config` object.
	GetConfig() Config

	// IsHealthy performs a live (non-cached) health check to determine if the
	// overall system or a specific set of components are "healthy," meaning their
	// status is either `OK` or `WARN`. This is a "tolerant" check.
	//
	// Parameters:
	//   - name: An optional list of component names to check. If empty, checks all components.
	//
	// Returns:
	//   `true` if the aggregated status is `OK` or `WARN`, `false` otherwise.
	IsHealthy(name ...string) bool

	// IsStrictlyHealthy performs a live (non-cached) health check to determine if
	// the overall system or specific components are "strictly healthy," meaning
	// their status is `OK`. This is a "strict" check.
	//
	// Parameters:
	//   - name: An optional list of component names to check. If empty, checks all components.
	//
	// Returns:
	//   `true` only if the aggregated status is `OK`, `false` otherwise.
	IsStrictlyHealthy(name ...string) bool

	// IsCacheHealthy performs a "tolerant" health check using the cached status.
	// It returns `true` if the cached status is `OK` or `WARN`. This is a
	// high-performance check suitable for frequent calls (e.g., in a middleware).
	//
	// Returns:
	//   `true` if the cached status is `OK` or `WARN`, `false` otherwise.
	IsCacheHealthy() bool

	// IsCacheStrictlyHealthy performs a "strict" health check using the cached status.
	// It returns `true` only if the cached status is `OK`. This is also a
	// high-performance check.
	//
	// Returns:
	//   `true` only if the cached status is `OK`, `false` otherwise.
	IsCacheStrictlyHealthy() bool
}

// New creates a new, fully initialized `Status` instance.
//
// The returned instance is thread-safe but requires further configuration before it
// can be used effectively. At a minimum, you must:
//   1. Call `SetInfo` or `SetVersion` to provide application identity.
//   2. Call `RegisterPool` to link a monitor pool for health checks.
//
// You can also optionally call `SetConfig` to define custom health policies and
// `SetErrorReturn` to customize error formatting.
//
// Parameters:
//   - ctx: The root `context.Context` for the application.
//
// Returns:
//   A new `Status` instance.
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

	// Set the cache's refresh function. When the cache is stale, it will call
	// getStatus() to re-compute the health.
	s.c.f = func() monsts.Status {
		r, _ := s.getStatus()
		return r
	}

	return s
}
