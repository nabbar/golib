//go:build linux || darwin

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

package unixgram

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

// ServerUnixGram defines the interface for a Unix domain datagram socket server implementation.
// It extends the base github.com/nabbar/golib/socket.Server interface
// with Unix datagram socket-specific functionality.
//
// The server operates in connectionless datagram mode (SOCK_DGRAM):
//   - Creates a Unix socket file in the filesystem
//   - No persistent connections are maintained
//   - Each datagram is processed independently
//   - Supports file permissions and group ownership
//   - Graceful shutdown supported
//   - Callback registration for events and errors
//
// Unlike connection-oriented Unix sockets (unix package):
//   - No per-client connections (similar to UDP vs TCP)
//   - Single handler processes all datagrams
//   - OpenConnections() returns 1 when running, 0 when stopped
//   - No connection draining needed on shutdown
//
// See github.com/nabbar/golib/socket.Server for inherited methods:
//   - Listen(context.Context) error - Start accepting datagrams
//   - Shutdown(context.Context) error - Graceful shutdown
//   - Close() error - Immediate shutdown
//   - IsRunning() bool - Check if server is accepting datagrams
//   - IsGone() bool - Check if server has stopped
//   - OpenConnections() int64 - Returns 1 if running, 0 if stopped
//   - Done() <-chan struct{} - Channel closed when listener stops
//   - Gone() <-chan struct{} - Always closed (no connections to drain)
//   - SetTLS(bool, TLSConfig) error - No-op for Unix sockets (always returns nil)
//   - RegisterFuncError(FuncError) - Register error callback
//   - RegisterFuncInfo(FuncInfo) - Register datagram event callback
//   - RegisterFuncInfoServer(FuncInfoSrv) - Register server event callback
//
// See github.com/nabbar/golib/socket/server/unix for connection-oriented Unix sockets.
// See github.com/nabbar/golib/socket/server/udp for UDP datagram sockets (IP-based).
type ServerUnixGram interface {
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

// New creates a new Unix domain datagram socket server instance.
//
// Parameters:
//   - u: Optional UpdateConn callback invoked when the socket is created.
//     Can be used to set socket options. Pass nil if not needed.
//     Note: For datagram sockets, this is called once on Listen(), not per datagram.
//   - h: Required HandlerFunc function that processes each datagram.
//     Receives Reader and Writer interfaces for the datagram.
//     The handler runs in its own goroutine for the server's lifetime.
//
// The returned server must have RegisterSocket() called to set the socket file path
// before calling Listen().
//
// Example usage:
//
//	handler := func(r socket.Reader, w socket.Writer) {
//	    defer r.Close()
//	    defer w.Close()
//	    buf := make([]byte, 65507) // Max datagram size
//	    for {
//	        n, err := r.Read(buf)
//	        if err != nil {
//	            break
//	        }
//	        w.Write(buf[:n]) // Echo back to sender
//	    }
//	}
//
//	srv := unixgram.New(nil, handler)
//	srv.RegisterSocket("/tmp/myapp.sock", 0600, -1)
//	srv.Listen(context.Background())
//
// The server is safe for concurrent use and manages lifecycle properly.
// Unlike connection-oriented Unix sockets, datagrams are stateless.
//
// Datagram behavior:
//   - Each Read() receives a complete datagram from any sender
//   - Each Write() sends to the last sender's address
//   - No connection state is maintained between datagrams
//   - Sender addresses are tracked internally for replies
//
// See github.com/nabbar/golib/socket.HandlerFunc and socket.UpdateConn for
// callback function signatures.
func New(upd libsck.UpdateConn, hdl libsck.HandlerFunc, cfg sckcfg.Server) (ServerUnixGram, error) {
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
		run: new(atomic.Bool),
		gon: new(atomic.Bool),
		fe:  libatm.NewValueDefault[libsck.FuncError](dfe, dfe),
		fi:  libatm.NewValueDefault[libsck.FuncInfo](dfi, dfi),
		fs:  libatm.NewValueDefault[libsck.FuncInfoSrv](dfs, dfs),
		ad:  libatm.NewValue[sckFile](),
	}

	if e := s.RegisterSocket(cfg.Address, cfg.PermFile, cfg.GroupPerm); e != nil {
		return nil, e
	}

	s.gon.Store(true)

	return s, nil
}
