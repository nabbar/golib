//go:build linux || darwin

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

package unix

import (
	"net"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libprm "github.com/nabbar/golib/file/perm"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// MaxGID defines the maximum allowed Unix group ID value (32767).
// Group IDs must be within this range to be valid on Linux systems.
const MaxGID = 32767

// ServerUnix defines the interface for a Unix domain socket server implementation.
// It extends the base github.com/nabbar/golib/socket.Server interface
// with Unix socket-specific functionality.
//
// The server operates in connection-oriented mode (SOCK_STREAM):
//   - Creates a Unix socket file in the filesystem
//   - Accepts persistent connections from clients
//   - Each connection is handled independently in a separate goroutine
//   - Supports file permissions and group ownership
//   - Graceful shutdown with connection draining
//   - Configurable idle timeouts
//   - Comprehensive event callbacks
//
// # Thread Safety
//
// All exported methods are safe for concurrent use by multiple goroutines.
// The server maintains internal synchronization to handle concurrent connections.
//
// # Error Handling
//
// The server provides multiple ways to handle errors:
//   - Return values from methods
//   - Error callback function
//   - Context cancellation
//
// # Lifecycle
//
// 1. Create server with New()
// 2. Configure with RegisterSocket()
// 3. Start with Listen()
// 4. Handle connections in handler function
// 5. Shut down with Shutdown() or Close()
//
// See github.com/nabbar/golib/socket.Server for inherited methods:
//   - Listen(context.Context) error - Start accepting connections
//   - Shutdown(context.Context) error - Graceful shutdown
//   - Close() error - Immediate shutdown
//   - IsRunning() bool - Check if server is accepting connections
//   - IsGone() bool - Check if all connections are closed
//   - OpenConnections() int64 - Get active connection count
//   - Done() <-chan struct{} - Channel closed when listener stops
//   - Gone() <-chan struct{} - Channel closed when all connections close
//   - SetTLS(bool, TLSConfig) error - No-op for Unix sockets (always returns nil)
//   - RegisterFuncError(FuncError) - Register error callback
//   - RegisterFuncInfo(FuncInfo) - Register connection event callback
//   - RegisterFuncInfoServer(FuncInfoSrv) - Register server event callback
type ServerUnix interface {
	libsck.Server

	// RegisterSocket sets the Unix socket file path, permissions, and group ownership.
	//
	// Parameters:
	//   - unixFile: Absolute or relative path to the socket file (e.g., "/tmp/app.sock")
	//   - perm: File permissions (e.g., 0600 for user-only, 0660 for user+group)
	//   - gid: Group ID for the socket file, or -1 to use default group
	//
	// The socket file will be created when Listen() is called and removed on shutdown.
	// If the file exists, it will be deleted before creating the new socket.
	//
	// Returns ErrInvalidGroup if gid exceeds MaxGID (32767).
	//
	// This method must be called before Listen().
	RegisterSocket(unixFile string, perm libprm.Perm, gid int32) error
}

// New creates a new Unix domain socket server instance with the specified
// connection handler and optional connection updater.
//
// Parameters:
//   - u: Optional UpdateConn callback invoked for each accepted connection.
//     Can be used to set socket options, configure timeouts, or track connections.
//     If nil, no per-connection configuration is performed.
//   - h: Required HandlerFunc that processes each connection. The handler receives
//     io.Reader and io.Writer interfaces for the connection and runs in its own
//     goroutine. The handler should handle the entire lifecycle of the connection.
//
// The returned server must be configured with RegisterSocket() before calling Listen().
//
// # Example
//
//	srv := unix.New(
//	    // Optional connection updater
//	    func(conn net.Conn) {
//	        if tcpConn, ok := conn.(*net.UnixConn); ok {
//	            _ = tcpConn.SetReadBuffer(8192)
//	        }
//	    },
//
//	    // Required connection handler
//	    func(r io.Reader, w io.Writer) {
//	        // Handle connection
//	        _, _ = io.Copy(w, r) // Echo server example
//	    },
//	)
//
// # Performance Considerations
//
// The handler function should be designed to handle multiple concurrent connections
// efficiently. For CPU-bound work, consider using a worker pool pattern.
//
// # Error Handling
//
// Any panics in the handler will be recovered and logged, but the connection
// will be closed. For better error handling, use recover() in your handler.
//
// Example usage:
//
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    buf := make([]byte, 4096)
//	    for {
//	        n, err := r.Read(buf)
//	        if err != nil {
//	            break
//	        }
//	        w.Write(buf[:n])
//	    }
//	}
//
//	srv := unix.New(nil, handler)
//	srv.RegisterSocket("/tmp/myapp.sock", 0600, -1)
//	srv.Listen(context.Background())
//
// The server is safe for concurrent use and manages connection lifecycle properly.
// Connections persist until explicitly closed, unlike UDP.
//
// See github.com/nabbar/golib/socket.HandlerFunc and socket.UpdateConn for
// callback function signatures.
func New(upd libsck.UpdateConn, hdl libsck.HandlerFunc, cfg sckcfg.Server) (ServerUnix, error) {
	if e := cfg.Validate(); e != nil {
		return nil, e
	} else if hdl == nil {
		return nil, ErrInvalidHandler
	}

	var (
		dfe libsck.FuncError   = func(_ ...error) {}
		dfi libsck.FuncInfo    = func(_, _ net.Addr, _ libsck.ConnState) {}
		dfs libsck.FuncInfoSrv = func(_ string) {}
	)

	s := &srv{
		upd: upd,
		hdl: hdl,
		idl: 0,
		run: new(atomic.Bool),
		gon: new(atomic.Bool),
		fe:  libatm.NewValueDefault[libsck.FuncError](dfe, dfe),
		fi:  libatm.NewValueDefault[libsck.FuncInfo](dfi, dfi),
		fs:  libatm.NewValueDefault[libsck.FuncInfoSrv](dfs, dfs),
		ad:  libatm.NewValue[sckFile](),
		nc:  new(atomic.Int64),
	}

	if cfg.ConIdleTimeout > 0 {
		s.idl = cfg.ConIdleTimeout.Time()
	}

	if e := s.RegisterSocket(cfg.Address, cfg.PermFile, cfg.GroupPerm); e != nil {
		return nil, e
	}

	s.gon.Store(true)

	return s, nil
}
