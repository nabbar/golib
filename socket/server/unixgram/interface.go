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
	"sync"

	libatm "github.com/nabbar/golib/atomic"
	libprm "github.com/nabbar/golib/file/perm"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
)

// MaxGID defines the maximum allowed Unix group ID value (32767).
// Group IDs must be within this range to be valid on Linux systems.
const MaxGID = 32767

// ServerUnixGram defines the interface for a Unix domain datagram socket server implementation.
// It extends the base github.com/nabbar/golib/socket.Server interface with
// Unix datagram socket-specific functionality.
//
// # Key Concepts: Connectionless (SOCK_DGRAM)
//
// Unlike Stream sockets (SOCK_STREAM), Unix datagram sockets do not establish a
// bi-directional, stateful connection between a client and the server. Instead:
//  1. The server listens on a single filesystem socket file.
//  2. Any local process can send independent datagrams to this file.
//  3. The server processes each datagram as a separate entity.
//  4. No "connection" is established; thus, OpenConnections() always returns 0.
//
// # Performance and Resource Management
//
// To handle high-throughput scenarios, this implementation uses:
//  - sync.Pool for context recycling (sCtx).
//  - Event-driven shutdown using Go channels for zero latency.
//  - Atomic state flags for lock-free concurrency.
//
// # Methods inherited from libsck.Server:
//   - Listen(context.Context) error: Starts the listener.
//   - Shutdown(context.Context) error: Gracefully stops the listener.
//   - Close() error: Immediately stops the listener.
//   - IsRunning() bool: True if accepting datagrams.
//   - IsGone() bool: True if the server has finished its lifecycle.
//   - OpenConnections() int64: Returns 0 for datagram servers.
//   - Done() <-chan struct{}: Channel closed when the listener exits.
//   - Gone() <-chan struct{}: Channel closed when the server is fully stopped.
type ServerUnixGram interface {
	libsck.Server

	// RegisterSocket sets the Unix socket file path, permissions, and group ownership.
	//
	// Parameters:
	//   - unixFile: Absolute or relative path to the socket file (e.g., "/tmp/app.sock").
	//   - perm: File permissions (e.g., 0600 for user-only, 0660 for user+group).
	//   - gid: Group ID for the socket file, or -1 to use the process's default group.
	//
	// Returns:
	//   - ErrInvalidGroup if gid exceeds MaxGID (32767).
	//   - nil on success.
	//
	// Note: The socket file is created only when Listen() is called and is
	// automatically removed on shutdown.
	RegisterSocket(unixFile string, perm libprm.Perm, gid int32) error
}

// New creates a new Unix domain datagram socket server instance.
//
// Parameters:
//   - upd: Optional UpdateConn callback invoked once the net.UnixConn is created.
//     Useful for setting low-level socket options (e.g., SO_RCVBUF, SO_SNDBUF).
//   - hdl: Required HandlerFunc that processes all incoming datagrams.
//     The handler is executed in a single, persistent goroutine for the
//     entire duration of the server's lifecycle.
//   - cfg: Configuration for the server, including initial address and permissions.
//
// # Handler Behavior (Unixgram)
//
// The handler receives an sCtx context. In datagram mode, this context represents
// the single local socket endpoint. You should use a loop within your handler:
//
//	handler := func(ctx libsck.Context) {
//	    buf := make([]byte, 65507) // Standard max datagram size
//	    for {
//	        n, err := ctx.Read(buf)
//	        if err != nil {
//	            return // Exit when the socket is closed or context cancelled
//	        }
//	        // Process your datagram here: buf[:n]
//	    }
//	}
//
// # Data Flow Schema
//
//	[User Request] --> [cfg] --> [New] --> [RegisterSocket] --> [Listen]
//	                                                              |
//	                                      [Listener Lifecycle] <--|
//	                                             |
//	                   [sync.Pool] <--- [Recycling Contexts] <--- [Handler Loop]
//
// Returns:
//   - ServerUnixGram instance if successful.
//   - Error if configuration validation fails or if the handler is nil.
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
		fe:  libatm.NewValueDefault[libsck.FuncError](dfe, dfe),
		fi:  libatm.NewValueDefault[libsck.FuncInfo](dfi, dfi),
		fs:  libatm.NewValueDefault[libsck.FuncInfoSrv](dfs, dfs),
		ad:  libatm.NewValue[sckFile](),
		gnc: libatm.NewValueDefault[chan struct{}](make(chan struct{}), make(chan struct{})),
		pol: &sync.Pool{
			New: func() interface{} {
				return &sCtx{}
			},
		},
	}

	if e := s.RegisterSocket(cfg.Address, cfg.PermFile, cfg.GroupPerm); e != nil {
		return nil, e
	}

	s.gon.Store(true)

	return s, nil
}
