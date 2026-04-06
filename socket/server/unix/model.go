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
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
	sckidl "github.com/nabbar/golib/socket/idlemgr"
)

// sckFile represents the configuration metadata for the Unix socket file.
// It stores the path, standard POSIX permissions, and Group ID (GID) for ownership.
type sckFile struct {
	File string      // Filesystem path (e.g., /var/run/my.sock).
	Perm libprm.Perm // Permissions (e.g., 0600).
	GID  int32       // Group ID for file ownership.
}

// srv is the internal concrete implementation of the ServerUnix interface.
// It employs advanced concurrency patterns to ensure extreme performance under load.
//
// # Design Patterns & Concurrency Model:
//   - Lock-Free State Management: Uses `atomic.Bool` and `atomic.Int64` for high-frequency checks.
//   - Resource Recycling: Utilizes a `sync.Pool` to store `sCtx` structures, achieving zero-allocation connections.
//   - Centralized Idle Control: Integrates with `sckidl.Manager` to avoid per-connection timers.
//   - Broadcast Signaling: Uses a `chan struct{}` (gnc) to instantaneously notify all handlers of a shutdown.
//
// # Key Differences from TCP Servers:
//   - Transport: Uses local Unix domain sockets (AF_UNIX) which bypass the network stack.
//   - Filesystem Integration: Manages file permissions and cleanup of the socket file.
//   - No TLS: Unix sockets are inherently local and secure via filesystem access control; TLS is a no-op.
//
// Fields:
//   - upd: Optional `UpdateConn` callback for socket-level tuning (e.g., buffer sizes).
//   - hdl: Required `HandlerFunc` that implements the business logic.
//   - idl: Duration for the idle connection timeout.
//   - run: Atomic flag indicating if the listener is currently active.
//   - gon: Atomic flag indicating if the server has entered the shutdown phase.
//   - fe: Atomic storage for the `FuncError` callback.
//   - fi: Atomic storage for the `FuncInfo` callback (connection events).
//   - fs: Atomic storage for the `FuncInfoSrv` callback (lifecycle events).
//   - ad: Atomic storage for the socket file configuration (`sckFile`).
//   - gnc: Atomic storage for the 'gone' channel used for shutdown broadcasting.
//   - id: The centralized `Idle Manager` instance.
//   - nc: Atomic counter of currently open client connections.
//   - pol: The `sync.Pool` used for recycling `sCtx` structures.
type srv struct {
	upd libsck.UpdateConn  // Connection update callback (optional)
	hdl libsck.HandlerFunc // Connection handler function (required)
	idl time.Duration      // idle connection timeout duration
	run *atomic.Bool       // Atomic: Listener status
	gon *atomic.Bool       // Atomic: Shutdown status

	fe libatm.Value[libsck.FuncError]   // Atomic: Error notification callback
	fi libatm.Value[libsck.FuncInfo]    // Atomic: Connection status callback
	fs libatm.Value[libsck.FuncInfoSrv] // Atomic: Server lifecycle callback

	ad  libatm.Value[sckFile]       // Atomic: Address configuration
	gnc libatm.Value[chan struct{}] // Atomic: Broadcast shutdown channel
	id  sckidl.Manager              // The centralized idle detection manager
	nc  *atomic.Int64               // Atomic: Active connection counter
	pol *sync.Pool                  // sync.Pool for connection structure recycling
}

// Listener returns the network protocol, the listener's address, and whether TLS is enabled.
//
// Returns:
//   - network: Always `libptc.NetworkUnix`.
//   - listener: The filesystem path of the socket file.
//   - tls: Always false (not supported for Unix sockets).
func (o *srv) Listener() (network libptc.NetworkProtocol, listener string, tls bool) {
	a := o.ad.Load()
	return libptc.NetworkUnix, a.File, false
}

// OpenConnections returns the current count of active client connections.
// This is safe to call concurrently and is backed by an atomic counter.
//
// Returns:
//   - int64: Total number of currently open connections.
func (o *srv) OpenConnections() int64 {
	return o.nc.Load()
}

// IsRunning reports whether the server is currently accepting new connections.
//
// Returns:
//   - bool: True if the listener loop is active.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone reports whether the server has initiated its shutdown phase.
// Once "gone", the server stops accepting new connections and broadcasts
// a termination signal to all active handlers via the 'gone' channel.
//
// Returns:
//   - bool: True if the server is in the shutdown/draining phase.
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

// setGone is an internal helper that triggers the shutdown sequence.
//
// Logic Workflow:
//  1. Atomic Swap: Sets the `gon` flag to true. If it was already true, returns early.
//  2. Broadcast: Closes the current 'gone' channel (`gnc`). This acts as an instant
//     broadcast signal to all goroutines listening on this channel.
func (o *srv) setGone() {
	if o == nil {
		return
	}

	// Use atomic swap to ensure we only close the channel once per listen cycle.
	if o.gon.Swap(true) {
		return
	}

	// Retrieve and close the current broadcast channel.
	if ch := o.gnc.Load(); ch != nil {
		close(ch)
	}
}

// getGoneChan retrieves the broadcast signaling channel for the current listen cycle.
// Connection goroutines use this channel in a `select` statement to react to server shutdowns.
//
// Returns:
//   - <-chan struct{}: A read-only channel that is closed when the server shuts down.
func (o *srv) getGoneChan() <-chan struct{} {
	if o == nil {
		return nil
	}
	return o.gnc.Load()
}

// Close performs an immediate shutdown of the server.
// It is equivalent to calling `Shutdown()` with a background context.
//
// Returns:
//   - error: Any error encountered during listener closure or resource cleanup.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// Shutdown initiates a graceful shutdown of the server.
//
// Shutdown Lifecycle:
//  1. Signaling: Calls `setGone()` to stop accepting new connections and notify all active handlers.
//  2. Timeout: Applies a context with a timeout (defaulting to 1 second) to wait for connection draining.
//  3. Draining: Monitors the atomic connection counter (`nc`) and the `run` flag.
//  4. Exit: Returns `ErrShutdownTimeout` if the draining takes longer than the context's deadline.
//
// Parameters:
//   - ctx: The context that governs the shutdown's maximum duration.
//
// Returns:
//   - error: Nil on success, or `ErrShutdownTimeout` if connection draining fails within the timeout.
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	} else if !o.IsRunning() || o.IsGone() {
		return nil
	}

	o.setGone()

	var (
		tck = time.NewTicker(3 * time.Millisecond)
		cnl context.CancelFunc
	)

	// Create a sub-context for the draining phase.
	ctx, cnl = context.WithTimeout(ctx, time.Second) // #nosec
	defer func() {
		tck.Stop()
		cnl()
	}()

	// Wait for all connection goroutines to finish and for the listener to stop.
	for o.IsRunning() || o.OpenConnections() > 0 {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			// Periodic check of the atomic state flags.
			break // nolint
		}
	}

	return nil
}

// SetTLS is a no-op for Unix socket servers.
// Unix domain sockets reside in the filesystem and are inherently local.
// Security is managed via filesystem permissions, making TLS redundant.
// This method always returns nil to satisfy the interface.
func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	return nil
}

// RegisterFuncError registers a callback for error reporting.
// This callback is invoked for listener errors, I/O errors, and internal management issues.
func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.fe.Store(f)
}

// RegisterFuncInfo registers a callback for connection lifecycle events.
// It is triggered for every new connection, I/O activity, and closure.
func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.fi.Store(f)
}

// RegisterFuncInfoServer registers a callback for server-level informational logs.
// This is used to report server start, stop, and socket configuration changes.
func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}

	o.fs.Store(f)
}

// RegisterSocket defines the socket file's metadata before starting the server.
//
// # Validation:
//   - Path: Must be a non-empty, well-formed Unix address.
//   - Group: The GID must be within the valid range (MaxGID = 32767).
//
// Parameters:
//   - unixFile: The absolute or relative path to the socket.
//   - perm: The permissions to apply (e.g., 0600).
//   - gid: The Group ID to assign to the file, or -1 for default.
func (o *srv) RegisterSocket(unixFile string, perm libprm.Perm, gid int32) error {
	if len(unixFile) < 1 {
		return ErrInvalidUnixFile
	} else if _, err := net.ResolveUnixAddr(libptc.NetworkUnix.Code(), unixFile); err != nil {
		return err
	} else if gid > MaxGID {
		return ErrInvalidGroup
	}

	o.ad.Store(sckFile{File: unixFile, Perm: perm, GID: gid})

	return nil
}

// fctError is an internal helper that safely invokes the registered error callback.
func (o *srv) fctError(e ...error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unix/fctError", r)
		}
	}()

	if o == nil {
		return
	} else if len(e) < 1 {
		return
	}

	var ok = false
	for _, err := range e {
		if err != nil {
			ok = true
			break
		}
	}

	if !ok {
		return
	} else if f := o.fe.Load(); f != nil {
		f(e...)
	}
}

// fctInfo is an internal helper that safely invokes the registered connection info callback.
func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unix/fctInfo", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fi.Load(); f != nil {
		f(local, remote, state)
	}
}

// fctInfoSrv is an internal helper that safely invokes the registered server info callback.
func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unix/fctInfoSrv", r)
		}
	}()

	if o == nil {
		return
	} else if f := o.fs.Load(); f != nil {
		f(fmt.Sprintf(msg, args...))
	}
}

// idleTimeout returns the current duration after which an inactive connection is closed.
// This is used during connection setup to determine if it should be registered with the manager.
func (o *srv) idleTimeout() time.Duration {
	if o == nil {
		return 0
	} else if o.idl < time.Second {
		// Timeouts less than 1 second are considered disabled.
		return 0
	} else {
		return o.idl
	}
}

// getContext retrieves a recycled `sCtx` structure from the `sync.Pool`.
// If the pool is empty, it allocates a new one. It always calls `reset()`
// to ensure the structure is properly initialized for the new connection.
//
// Returns:
//   - *sCtx: A clean, initialized connection context.
func (o *srv) getContext(ctx context.Context, cnl context.CancelFunc, con io.ReadWriteCloser, l, r net.Addr) *sCtx {
	if o == nil || o.pol == nil {
		return &sCtx{}
	}

	if i := o.pol.Get(); i != nil {
		if c, ok := i.(*sCtx); ok {
			c.reset(ctx, cnl, con, l, r)
			return c
		}
	}

	return &sCtx{}
}

// putContext returns an `sCtx` structure to the `sync.Pool` for future reuse.
func (o *srv) putContext(c *sCtx) {
	if o == nil || o.pol == nil || c == nil {
		return
	}

	o.pol.Put(c)
}
