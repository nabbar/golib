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
 * OUT of OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
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
// This is used for dynamically loading monitor names from component configurations.
type FuncGetCfgCpt func(key string) cfgtps.ComponentMonitor

// Route defines the HTTP routing interface for exposing the status endpoint.
type Route interface {
	// Expose provides a generic handler for status requests that can be used
	// with any framework that uses `context.Context`.
	Expose(ctx context.Context)

	// MiddleWare is a Gin middleware that processes status requests, handling
	// content negotiation for format and verbosity.
	MiddleWare(c *ginsdk.Context)

	// SetErrorReturn registers a custom factory function for creating error
	// formatters, allowing for customized error responses.
	SetErrorReturn(f func() liberr.ReturnGin)
}

// Info defines the interface for setting the application's version and build information.
type Info interface {
	// SetInfo manually sets the application name, release version, and build hash.
	SetInfo(name, release, hash string)

	// SetVersion sets the application information from a `version.Version` object,
	// which is the recommended approach.
	SetVersion(vers libver.Version)
}

// Pool defines the interface for managing the monitor components that contribute
// to the overall health status.
type Pool interface {
	montps.PoolStatus

	// RegisterPool registers a function that provides the monitor pool. This
	// dependency injection allows the pool to be managed externally.
	RegisterPool(fct montps.FuncPool)
}

// Status is the main interface that combines all health status functionality.
// It provides a complete solution for application health monitoring and reporting.
type Status interface {
	encoding.TextMarshaler
	json.Marshaler

	Route
	Info
	Pool

	// RegisterGetConfigCpt registers a function to retrieve component configurations
	// by key, enabling dynamic resolution of monitor names.
	RegisterGetConfigCpt(fct FuncGetCfgCpt)

	// SetConfig applies a configuration for status computation and HTTP response codes.
	SetConfig(cfg Config)

	// GetConfig returns the current configuration used for status computation.
	GetConfig() Config

	// IsHealthy checks if the overall system or specific components are healthy
	// (i.e., their status is `OK` or `WARN`).
	IsHealthy(name ...string) bool

	// IsStrictlyHealthy checks if the overall system or specific components are
	// strictly healthy (i.e., their status is `OK`).
	IsStrictlyHealthy(name ...string) bool

	// IsCacheHealthy checks if the cached overall status is healthy (`OK` or `WARN`).
	// This is a high-performance check suitable for frequent calls.
	IsCacheHealthy() bool

	// IsCacheStrictlyHealthy checks if the cached overall status is strictly `OK`.
	// This is a high-performance check suitable for frequent calls.
	IsCacheStrictlyHealthy() bool
}

// New creates a new `Status` instance.
//
// The returned instance must be configured before use:
//   - Call `SetInfo` or `SetVersion` to set application information.
//   - Call `RegisterPool` to connect a monitor pool.
//   - Optionally, call `SetConfig` to customize behavior and `SetErrorReturn` to
//     customize error formatting.
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
