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
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// getSocketFile retrieves and validates the configured Unix socket file path.
// It ensures that the server's internal address configuration is well-formed.
//
// Internal Workflow:
//  1. Check if the server's internal address (`ad`) is set.
//  2. Verify that the file path is not empty.
//  3. Validate the path through the `checkFile()` helper.
//
// Returns:
//   - string: The validated socket path.
//   - error: `ErrInvalidUnixFile` if the path is invalid or not set.
func (o *srv) getSocketFile() (string, error) {
	if o == nil {
		return "", ErrInvalidUnixFile
	} else if a := o.ad.Load(); a == (sckFile{}) {
		return "", ErrInvalidUnixFile
	} else if len(a.File) < 1 {
		return "", ErrInvalidUnixFile
	} else {
		return o.checkFile(a.File)
	}
}

// getSocketPerm retrieves the configured filesystem permissions for the Unix socket.
// If no permissions are explicitly set, it defaults to `0770` (User/Group access).
//
// Returns:
//   - libprm.Perm: The current permission settings.
func (o *srv) getSocketPerm() libprm.Perm {
	if a := o.ad.Load(); a == (sckFile{}) {
		return libprm.Perm(0770)
	} else if a.Perm == 0 {
		return libprm.Perm(0770)
	} else {
		return a.Perm
	}
}

// getSocketGroup retrieves the Group ID (GID) for the Unix socket file.
// This is used to set the file ownership via `os.Chown`.
//
// Logic Hierarchy:
//  1. If GID is >= 0: Returns the configured GID.
//  2. If GID is -1 (default): Attempts to retrieve the current process's GID.
//  3. Fallback: Returns -1 if no ownership change should be performed.
//
// Returns:
//   - int: The GID to be used for the socket file.
func (o *srv) getSocketGroup() int {
	if a := o.ad.Load(); a == (sckFile{}) {
		return -1
	} else if a.GID >= 0 {
		return int(a.GID)
	} else if gid := syscall.Getgid(); gid <= MaxGID {
		return gid
	}

	return -1
}

// checkFile performs path normalization and ensures a clean state before listening.
//
// Cleanup Strategy:
//  1. Normalizes the path using `filepath.Join`.
//  2. Checks if a file already exists at the target path.
//  3. If it exists, it is automatically removed. This is necessary because
//     Unix socket files must be recreated by the `net.Listen` call.
//
// Returns:
//   - string: The normalized path.
//   - error: Any filesystem error encountered during validation or removal.
func (o *srv) checkFile(unixFile string) (string, error) {
	if len(unixFile) < 1 {
		return unixFile, ErrInvalidUnixFile
	} else {
		unixFile = filepath.Join(filepath.Dir(unixFile), filepath.Base(unixFile))
	}

	if _, e := os.Stat(unixFile); e != nil && !errors.Is(e, os.ErrNotExist) {
		return unixFile, e
	} else if e != nil {
		return unixFile, nil
	} else if e = os.Remove(unixFile); e != nil {
		return unixFile, e
	}

	return unixFile, nil
}

// Listen starts the Unix domain socket server and blocks until the server stops.
//
// # Server Lifecycle Workflow:
//  1. Validation: Verifies that a handler and socket path are correctly configured.
//  2. Idle Management: Starts the centralized `Idle Manager` (if a timeout > 0 is set).
//  3. Preparation: Swaps the `gon` (gone) flag to false and initializes a new broadcast channel `gnc`.
//  4. Resource Setup: Creates the Unix socket file, sets its permissions and ownership.
//  5. Watchdog: Spawns a dedicated goroutine that unblocks the `Accept()` call if the context is cancelled.
//  6. Accept Loop: Enters a direct, blocking `net.Listener.Accept()` loop.
//  7. Accept Logic: For every successful connection, it spawns a `Conn` goroutine.
//  8. Cleanup: When the loop exits (error, shutdown, or context cancellation):
//     - Closes the listener.
//     - Stops the Idle Manager.
//     - Removes the socket file from the filesystem.
//     - Sets the `run` and `gon` flags.
//     - Closes the `gnc` channel to signal all handlers to exit.
//
// # Concurrency Management:
// The server uses a single goroutine for the accept loop and one goroutine per connection.
// It leverages atomic types and channels for thread-safe state synchronization.
//
// Parameters:
//   - ctx: The context that governs the server's overall lifetime.
//
// Returns:
//   - error: Any terminal error encountered during listener creation or during the accept loop.
func (o *srv) Listen(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unix/listen", r)
		}
	}()

	var (
		e error        // terminal error
		l net.Listener // actual unix socket listener
		a string       // socket path
	)

	// Phase 1: Pre-launch validation.
	if a, e = o.getSocketFile(); e != nil {
		o.fctError(e)
		return e
	} else if o.hdl == nil {
		o.fctError(ErrInvalidHandler)
		return ErrInvalidHandler
	} else if l, e = o.getListen(a); e != nil {
		o.fctError(e)
		return e
	}

	// Phase 2: Start background managers.
	if o.id != nil {
		if e = o.id.Start(ctx); e != nil {
			o.fctError(e)
		}
	}

	// Phase 3: Prepare the state flags and signaling channels for this cycle.
	o.gon.Store(false)
	o.gnc.Store(make(chan struct{}))

	// Channel used to signal that the loop has physically exited.
	done := make(chan struct{})

	// Phase 4: Deferred cleanup logic.
	defer func() {
		o.fctInfoSrv("closing listen socket '%s %s'", libptc.NetworkUnix.String(), a)

		if l != nil {
			_ = l.Close()
		}

		if o.id != nil {
			_ = o.id.Stop(context.Background())
		}

		// Always clean up the filesystem socket file.
		if a != "" {
			_ = os.Remove(a)
		}

		// Update server state flags.
		o.run.Store(false)
		o.setGone() // Set the 'gone' flag and close the broadcast channel.
		close(done)
	}()

	// Phase 5: Shutdown Watchdog.
	// This goroutine waits for a stop signal (ctx or gone) and closes the listener
	// to unblock the current blocking `Accept()` call.
	go func() {
		select {
		case <-ctx.Done():
		case <-o.getGoneChan():
		case <-done: // Prevent leakage if the main loop exits for other reasons.
			return
		}

		if l != nil {
			_ = l.Close()
		}
	}()

	o.run.Store(true)

	// Phase 6: Core Accept Loop.
	for {
		conn, err := l.Accept()
		if err != nil {
			// Check if the cancellation came from the context.
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Check if we are in a graceful shutdown phase.
			if o.IsGone() {
				return nil
			}

			// Handle standard "closed connection" errors.
			if errors.Is(err, net.ErrClosed) {
				return nil
			} else if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}

			// Report unexpected errors and continue accepting.
			o.fctError(err)
			continue
		}

		if conn == nil {
			continue
		}

		// Phase 7: Per-Connection Handling.
		go func(c net.Conn) {
			lc := c.LocalAddr()
			rc := c.RemoteAddr()

			// Report connection lifecycle events.
			defer o.fctInfo(lc, rc, libsck.ConnectionClose)
			o.fctInfo(lc, rc, libsck.ConnectionNew)

			o.Conn(ctx, c)
		}(conn)
	}
}

// Conn processes a single client connection in its own goroutine.
//
// # Workflow:
//  1. Resource Tracking: Increments the active connection count (`nc`).
//  2. Callback Notification: Invokes `UpdateConn` for per-socket tuning.
//  3. Pooling: Fetches an `sCtx` structure from the `sync.Pool` and `reset()` it.
//  4. Timeout Registration: If idle timeouts are enabled, registers the connection with the `Idle Manager`.
//  5. Logic Execution: Launches the user's `HandlerFunc` in a sub-goroutine.
//  6. Synchronization: Blocks, waiting for one of three signals:
//     - Context cancellation (the server is shutting down).
//     - `sCtx` closure (the handler finished or an I/O error occurred).
//     - 'Gone' signal (broadcast shutdown).
//  7. Final Cleanup: Decrements the connection counter, closes the connection, unregisters from the manager,
//     and returns the `sCtx` to the pool.
//
// Parameters:
//   - ctx: The server's main lifecycle context.
//   - con: The raw network connection from `Accept()`.
func (o *srv) Conn(ctx context.Context, con net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/server/unix/conn", r)
		}
	}()

	var (
		cnl context.CancelFunc
		dur = o.idleTimeout()

		sx *sCtx
	)

	// Phase 1: Connection initialization.
	o.nc.Add(1)

	defer func() {
		// Phase 3: Final cleanup and pooling.
		o.nc.Add(-1)

		if o.id != nil && dur > 0 {
			_ = o.id.Unregister(sx)
		}

		if con != nil {
			_ = con.Close()
		}

		if sx != nil {
			_ = sx.Close()
			o.putContext(sx) // Recycle the context structure.
		}

		if cnl != nil {
			cnl()
		}
	}()

	// Phase 2: Configuration and pooling.
	if o.upd != nil {
		o.upd(con)
	}

	if c, k := con.(io.ReadWriteCloser); k {
		// Create a specific context for this connection's lifecycle.
		ctx, cnl = context.WithCancel(ctx)
		// Fetch a recycled structure from the sync.Pool.
		sx = o.getContext(ctx, cnl, c, con.LocalAddr(), con.RemoteAddr())
	} else {
		return
	}

	if dur > 0 {
		// Enable centralized idle detection.
		_ = o.id.Register(sx)
	}

	// Phase 3: Start the business logic.
	if o.hdl == nil {
		return
	} else {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					librun.RecoveryCaller("golib/socket/server/unix/handler", r)
				}
			}()

			// Invoke the user's logic handler.
			o.hdl(sx)
		}()
	}

	// Phase 4: Lifecycle management.
	// Block here until the connection or server signals it's time to close.
	select {
	case <-ctx.Done():
		// The parent server context has been cancelled.
		return

	case <-sx.Done():
		// The connection itself has been closed (handler finished or I/O error).
		return

	case <-o.getGoneChan():
		// The server broadcast a shutdown signal via the 'gone' channel.
		return
	}
}
