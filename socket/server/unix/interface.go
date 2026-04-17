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
	"context"
	"net"
	"sync"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	durbig "github.com/nabbar/golib/duration/big"
	libprm "github.com/nabbar/golib/file/perm"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	sckidl "github.com/nabbar/golib/socket/idlemgr"
)

// MaxGID defines the maximum allowed Unix group ID value (32767).
// This limit is common on standard 16-bit Unix systems and is used here for safety during
// file ownership operations (`os.Chown`).
const MaxGID = 32767

// ServerUnix defines the comprehensive interface for a high-performance Unix domain socket server.
// It extends the base `github.com/nabbar/golib/socket.Server` interface with Unix-specific
// functionality, such as socket file permission and ownership management.
//
// # Server Operation:
//   - Mode: Connection-oriented (SOCK_STREAM).
//   - Transport: Unix domain sockets (AF_UNIX), communicating via local files.
//   - Concurrency: Goroutine-per-connection model with optimized resource pooling.
//   - Lifecycle Management: Atomic state tracking for running/shutdown phases.
//
// # Key Responsibilities:
//   - Socket Creation: Handles the creation, validation, and cleanup of the Unix socket file.
//   - Permissions: Manages file access control via `chmod` and `chown`.
//   - Resource Optimization: Recycles connection contexts (`sCtx`) to reduce GC overhead.
//   - Idle Control: Integrates with a centralized Idle Manager for efficient timeout scanning.
//
// # Lifecycle Methods Inherited from libsck.Server:
//   - Listen(context.Context) error: Starts the accept loop. Blocks until shutdown or error.
//   - Shutdown(context.Context) error: Initiates a graceful shutdown with draining.
//   - Close() error: Triggers an immediate shutdown.
//   - IsRunning() bool: Returns true if the server is accepting connections.
//   - IsGone() bool: Returns true if the server has finished its shutdown cycle.
//   - OpenConnections() int64: Returns the number of currently active connections.
//
// # Thread Safety:
// All methods are safe for concurrent use by multiple goroutines. Internal synchronization
// is achieved via lock-free atomic operations and a `sync.Pool`.
type ServerUnix interface {
	libsck.Server

	// RegisterSocket configures the metadata for the Unix socket file.
	//
	// # Important:
	// This method MUST be called before `Listen()`.
	//
	// Parameters:
	//   - unixFile: The filesystem path for the socket file (e.g., "/var/run/app.sock").
	//   - perm: The permissions to apply to the socket file (e.g., 0600, 0660).
	//   - gid: The Group ID to assign to the socket file, or -1 to use the default group.
	//
	// Returns:
	//   - ErrInvalidUnixFile: If the path is empty or invalid.
	//   - ErrInvalidGroup: If the GID exceeds MaxGID.
	//
	// # Behavior:
	//   - The socket file is created only when `Listen()` is called.
	//   - Existing files at the same path are automatically removed during startup.
	//   - The socket file is removed from the filesystem during shutdown.
	RegisterSocket(unixFile string, perm libprm.Perm, gid int32) error
}

// New creates and initializes a new Unix domain socket server instance.
// It sets up the internal context pool, configuration, and monitoring callbacks.
//
// # Initialization Details:
//   - Callback Hooks: Initializes default, non-blocking callbacks for error and info reporting.
//   - Context Pool: Creates a `sync.Pool` to recycle `sCtx` structures, achieving zero-allocation connections.
//   - Idle Manager: If `cfg.ConIdleTimeout` is set, initializes a centralized `sckidl.Manager`
//     to perform efficient periodic scans of inactive connections.
//   - Gone Channel: Initializes the first signaling channel for the server lifecycle.
//
// # Performance Considerations:
//   - Under high load, the use of `sync.Pool` drastically reduces GC pressure compared to allocating a new
//     context for every connection.
//   - The centralized Idle Manager is more efficient than individual timers for thousands of connections.
//
// Parameters:
//   - upd: Optional callback for per-connection tuning (e.g., setting buffer sizes).
//   - hdl: Required callback that implements the server's business logic for each connection.
//   - cfg: Server configuration structure containing address, timeout, and permission settings.
//
// Returns:
//   - ServerUnix: The initialized server instance.
//   - error: If the configuration is invalid or if the handler is nil.
//
// # Example:
//
//	handler := func(r io.Reader, w io.Writer) {
//	    _, _ = io.WriteString(w, "Hello from Unix Server!\n")
//	}
//
//	srv, _ := unix.New(nil, handler, config.Server{
//	    Address: "/tmp/my.sock",
//	    PermFile: perm.New(0666),
//	})
//	_ = srv.Listen(context.Background())
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
		gnc: libatm.NewValueDefault[chan struct{}](make(chan struct{}), make(chan struct{})),
		id:  nil,
		nc:  new(atomic.Int64),
		pol: &sync.Pool{
			New: func() interface{} {
				return &sCtx{}
			},
		},
	}

	if c := cfg.ConIdleTimeout.Seconds(); c > 0 {
		s.idl = cfg.ConIdleTimeout.Time()

		i, e := sckidl.New(context.Background(), durbig.Seconds(c), durbig.Seconds(1))
		if e != nil {
			return nil, e
		}
		s.id = i
	}

	if e := s.RegisterSocket(cfg.Address, cfg.PermFile, cfg.GroupPerm); e != nil {
		return nil, e
	}

	s.gon.Store(true)

	return s, nil
}
