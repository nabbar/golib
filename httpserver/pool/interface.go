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

package pool

import (
	"context"
	"sync"

	liblog "github.com/nabbar/golib/logger"
	libsrv "github.com/nabbar/golib/runner"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	libhtp "github.com/nabbar/golib/httpserver"
	srvtps "github.com/nabbar/golib/httpserver/types"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

// FuncWalk is a callback function used when iterating over servers in the pool.
// The function receives the bind address and server instance for each iteration.
// Return true to continue iteration, false to stop.
type FuncWalk func(bindAddress string, srv libhtp.Server) bool

// Manage provides server management operations for a pool.
// All operations are thread-safe and can be called concurrently.
type Manage interface {
	// Walk iterates over all servers in the pool, calling the provided function for each.
	// Iteration stops if the callback returns false. Returns true if all servers were visited.
	Walk(fct FuncWalk)

	// WalkLimit iterates over specific servers identified by their bind addresses.
	// If no addresses are provided, behaves like Walk. Returns true if iteration completed.
	WalkLimit(fct FuncWalk, onlyBindAddress ...string)

	// Clean removes all servers from the pool.
	Clean()

	// Load retrieves a server by its bind address. Returns nil if not found.
	Load(bindAddress string) libhtp.Server

	// Store adds or updates a server in the pool, using its bind address as the key.
	Store(srv libhtp.Server)

	// Delete removes a server from the pool by its bind address.
	Delete(bindAddress string)

	// StoreNew creates a new server from configuration and adds it to the pool.
	// Returns an error if server creation or validation fails.
	StoreNew(cfg libhtp.Config, defLog liblog.FuncLog) error

	// LoadAndDelete atomically retrieves and removes a server.
	// Returns the server and true if found, nil and false otherwise.
	LoadAndDelete(bindAddress string) (val libhtp.Server, loaded bool)

	// MonitorNames returns a list of all monitoring identifiers for servers in the pool.
	MonitorNames() []string
}

// Filter provides filtering and querying operations for servers in the pool.
type Filter interface {
	// Has checks if a server with the given bind address exists in the pool.
	Has(bindAddress string) bool

	// Len returns the number of servers in the pool.
	Len() int

	// List returns a list of server field values matching the filter criteria.
	// fieldFilter specifies which field to match against, fieldReturn specifies which field to return.
	// Pattern uses glob-style matching (* wildcards), regex uses regular expressions.
	List(fieldFilter, fieldReturn srvtps.FieldType, pattern, regex string) []string

	// Filter creates a new pool containing only servers matching the criteria.
	// field specifies which field to filter on, pattern uses globs, regex uses regular expressions.
	Filter(field srvtps.FieldType, pattern, regex string) Pool
}

// Pool represents a collection of HTTP servers managed as a unified group.
// It combines server lifecycle management (Start/Stop/Restart) with advanced
// filtering, monitoring, and configuration operations. All methods are thread-safe.
type Pool interface {
	// Runner embeds base server interface for lifecycle management
	libsrv.Runner

	// Manage embeds server management operations
	Manage

	// Filter embeds filtering and query operations
	Filter

	// Clone creates a deep copy of the pool with an optional new context.
	// The cloned pool contains independent copies of all servers.
	Clone(ctx context.Context) Pool

	// Merge combines servers from another pool into this one.
	// Servers with conflicting bind addresses will be updated.
	Merge(p Pool, def liblog.FuncLog) error

	// Handler registers a handler function for all servers in the pool.
	Handler(fct srvtps.FuncHandler)

	// Monitor retrieves monitoring data for all servers in the pool.
	// Returns a slice of Monitor instances, one per server.
	Monitor(vrs libver.Version) ([]montps.Monitor, liberr.Error)
}

// New creates a new server pool with optional initial servers.
// The pool manages server lifecycle and provides unified operations across all servers.
//
// Parameters:
//   - ctx: Context provider function for server operations (can be nil)
//   - hdl: Handler function to register with all servers (can be nil)
//   - srv: Optional initial servers to add to the pool
//
// Returns:
//   - Pool: Initialized pool ready for use
//
// Example:
//
//	pool := pool.New(nil, handlerFunc)
//	pool.StoreNew(config1, nil)
//	pool.StoreNew(config2, nil)
//	pool.Start(context.Background())
func New(ctx context.Context, hdl srvtps.FuncHandler, srv ...libhtp.Server) Pool {
	p := &pool{
		m: sync.RWMutex{},
		p: libctx.New[string](ctx),
		h: hdl,
	}

	for _, s := range srv {
		if s != nil {
			p.Store(s)
		}
	}

	return p
}
