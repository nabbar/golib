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
	"crypto/tls"
	"net"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// ServerTcp defines the interface for a TCP server implementation.
// It extends the base github.com/nabbar/golib/socket.Server interface
// with TCP-specific functionality.
//
// # Features
//
// ## Connection Handling
//   - Concurrent connection handling with goroutine per connection
//   - Configurable idle timeout for connections
//   - Graceful connection draining during shutdown
//   - Connection state tracking and monitoring
//
// ## Security
//   - TLS/SSL encryption with configurable cipher suites
//   - Support for mutual TLS (mTLS) with client certificate verification
//   - Secure defaults for TLS configuration
//
// ## Monitoring & Observability
//   - Connection lifecycle callbacks (new, read, write, close)
//   - Error reporting through configurable callbacks
//   - Server status notifications
//   - Atomic counters for active connections
//
// ## Thread Safety
//   - All exported methods are safe for concurrent use
//   - Atomic operations for state management
//   - No shared state between connections
//
// # Lifecycle
//
// A typical server lifecycle follows these steps:
//  1. Create: server := tcp.New(updateFunc, handler, config)
//  2. Configure: server.RegisterServer(":8080")
//  3. Start: server.Listen(ctx)
//  4. (Run until shutdown signal)
//  5. Shutdown: server.Shutdown(ctx)
//
// # Error Handling
//
// The server provides multiple ways to handle errors:
//   - Return values from methods (e.g., Listen(), Shutdown())
//   - Error callback function (RegisterFuncError)
//   - Context cancellation for timeouts
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

// New creates and initializes a new TCP server instance with the provided configuration.
//
// # Parameters
//
//   - upd: Optional UpdateConn callback that's invoked when a new connection is accepted,
//     before the handler is called. Use this to configure connection-specific settings:
//
//   - TCP keepalive
//
//   - Read/write timeouts
//
//   - Buffer sizes
//
//   - Other TCP options
//
//     Example:
//     upd := func(conn net.Conn) error {
//     if tcpConn, ok := conn.(*net.TCPConn); ok {
//     return tcpConn.SetKeepAlive(true)
//     }
//     return nil
//     }
//
//   - hdl: Required HandlerFunc that processes each client connection. This function
//     runs in its own goroutine per connection. The handler receives:
//
//   - r: A socket.Reader for reading from the client
//
//   - w: A socket.Writer for writing to the client
//
//     Example echo server handler:
//     hdl := func(r socket.Reader, w socket.Writer) {
//     defer r.Close()
//     defer w.Close()
//     if _, err := io.Copy(w, r); err != nil {
//     log.Printf("Error in handler: %v", err)
//     }
//     }
//
//   - cfg: Server configuration including address, timeouts, and TLS settings.
//     Use socket.DefaultServerConfig() for default values.
//
// # Return Value
//
// Returns a new ServerTcp instance that implements the Server interface.
// The server is not started until Listen() is called.
//
// # Example Usage
//
// Basic echo server:
//
//	func main() {
//	  // Create a simple echo handler
//	  handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    io.Copy(w, r) // Echo back received data
//	  }
//
//	  // Create server with default config
//	  cfg := socket.DefaultServerConfig(":8080")
//	  srv, err := tcp.New(nil, handler, cfg)
//	  if err != nil {
//	    log.Fatalf("Failed to create server: %v", err)
//	  }
//
//	  // Start the server
//	  ctx := context.Background()
//	  if err := srv.Listen(ctx); err != nil {
//	    log.Fatalf("Server error: %v", err)
//	  }
//	}
//
// # Error Handling
//
// The following errors may be returned:
//   - ErrInvalidHandler: if hdl is nil
//   - Any error returned by the configuration validation
//
// # Concurrency
//
// The returned server instance is safe for concurrent use by multiple goroutines.
// The handler function (hdl) may be called concurrently for different connections.
//
// # Memory Management
//
// The server manages the lifecycle of connections and associated resources.
// Ensure that all resources are properly closed by calling Shutdown() or Close()
// when the server is no longer needed.
//
// See also:
//   - github.com/nabbar/golib/socket.HandlerFunc
//   - github.com/nabbar/golib/socket.UpdateConn
//   - github.com/nabbar/golib/socket/config.ServerConfig
func New(upd libsck.UpdateConn, hdl libsck.HandlerFunc, cfg sckcfg.Server) (ServerTcp, error) {
	if e := cfg.Validate(); e != nil {
		return nil, e
	} else if hdl == nil {
		return nil, ErrInvalidHandler
	}

	var (
		ssl = &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		}
		dfe libsck.FuncError   = func(_ ...error) {}
		dfi libsck.FuncInfo    = func(_, _ net.Addr, _ libsck.ConnState) {}
		dfs libsck.FuncInfoSrv = func(_ string) {}
	)

	s := &srv{
		ssl: libatm.NewValueDefault[*tls.Config](ssl, ssl),
		upd: upd,
		hdl: hdl,
		idl: 0,
		run: new(atomic.Bool),
		gon: new(atomic.Bool),
		fe:  libatm.NewValueDefault[libsck.FuncError](dfe, dfe),
		fi:  libatm.NewValueDefault[libsck.FuncInfo](dfi, dfi),
		fs:  libatm.NewValueDefault[libsck.FuncInfoSrv](dfs, dfs),
		ad:  libatm.NewValue[string](),
		nc:  new(atomic.Int64),
	}

	if cfg.ConIdleTimeout > 0 {
		s.idl = cfg.ConIdleTimeout.Time()
	}

	if e := s.RegisterServer(cfg.Address); e != nil {
		return nil, e
	}

	k, t := cfg.GetTLS()
	if e := s.SetTLS(k, t); e != nil {
		return nil, e
	}

	s.gon.Store(true)

	return s, nil
}
