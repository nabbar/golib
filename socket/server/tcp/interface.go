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

package tcp

import (
	"sync/atomic"

	libsck "github.com/nabbar/golib/socket"
)

// ServerTcp defines the interface for a TCP server implementation.
// It extends the base github.com/nabbar/golib/socket.Server interface
// with TCP-specific functionality.
//
// The server supports:
//   - TLS/SSL encryption via SetTLS()
//   - Graceful shutdown with connection draining
//   - Callback registration for connection events, errors, and server info
//   - Atomic state management for thread-safe operations
//
// See github.com/nabbar/golib/socket.Server for inherited methods:
//   - Listen(context.Context) error - Start accepting connections
//   - Shutdown(context.Context) error - Graceful shutdown
//   - Close() error - Immediate shutdown
//   - IsRunning() bool - Check if server is accepting connections
//   - IsGone() bool - Check if all connections are closed
//   - OpenConnections() int64 - Get current connection count
//   - Done() <-chan struct{} - Channel closed when server stops listening
//   - SetTLS(bool, TLSConfig) error - Configure TLS
//   - RegisterFuncError(FuncError) - Register error callback
//   - RegisterFuncInfo(FuncInfo) - Register connection info callback
//   - RegisterFuncInfoServer(FuncInfoSrv) - Register server info callback
type ServerTcp interface {
	libsck.Server

	// RegisterServer sets the TCP address for the server to listen on.
	// The address must be in the format "host:port" or ":port" to bind to all interfaces.
	//
	// Example addresses:
	//   - "127.0.0.1:8080" - Listen on localhost port 8080
	//   - ":8080" - Listen on all interfaces port 8080
	//   - "0.0.0.0:8080" - Explicitly listen on all IPv4 interfaces
	//
	// This method must be called before Listen(). Returns ErrInvalidAddress
	// if the address is empty or malformed.
	RegisterServer(address string) error
}

// New creates a new TCP server instance.
//
// Parameters:
//   - u: Optional UpdateConn callback invoked when a new connection is accepted,
//     before the handler is called. Can be used to set connection options
//     (e.g., TCP keepalive, buffer sizes). Pass nil if not needed.
//   - h: Required HandlerFunc function that processes each connection.
//     Receives Reader and Writer interfaces for the connection.
//     The handler runs in its own goroutine per connection.
//
// The returned server must have RegisterServer() called to set the listen address
// before calling Listen().
//
// Example usage:
//
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    io.Copy(w, r) // Echo server
//	}
//
//	srv := tcp.New(nil, handler)
//	srv.RegisterServer(":8080")
//	srv.Listen(context.Background())
//
// The server is safe for concurrent use and manages connection lifecycle,
// including graceful shutdown and connection draining.
//
// See github.com/nabbar/golib/socket.HandlerFunc and socket.UpdateConn for
// callback function signatures.
func New(u libsck.UpdateConn, h libsck.HandlerFunc) ServerTcp {
	c := new(atomic.Value)
	c.Store(make(chan []byte))

	s := new(atomic.Value)
	s.Store(make(chan struct{}))

	r := new(atomic.Value)
	r.Store(make(chan struct{}))

	return &srv{
		ssl: new(atomic.Value),
		upd: u,
		hdl: h,
		msg: c,
		stp: s,
		rst: r,
		run: new(atomic.Bool),
		gon: new(atomic.Bool),
		fe:  new(atomic.Value),
		fi:  new(atomic.Value),
		fs:  new(atomic.Value),
		ad:  new(atomic.Value),
		nc:  new(atomic.Int64),
	}
}
