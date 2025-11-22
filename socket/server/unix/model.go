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

package unix

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

var (
	// closedChanStruct is a pre-closed channel used as a sentinel value
	// to indicate that a channel has been closed or should be treated as closed.
	// This avoids repeated channel allocations and provides a consistent closed state.
	closedChanStruct chan struct{}
)

// init initializes the package-level closedChanStruct sentinel channel.
func init() {
	closedChanStruct = make(chan struct{})
	close(closedChanStruct)
}

// srv is the internal implementation of the ServerUnix interface.
// It uses atomic operations for thread-safe state management and operates
// in connection-oriented mode (SOCK_STREAM).
//
// Unlike UDP servers, Unix socket servers:
//   - Maintain per-client persistent connections
//   - Accept multiple concurrent connections
//   - Track connection count atomically
//   - Support graceful connection draining on shutdown
//   - Cannot use TLS (SetTLS is a no-op)
//
// All fields use atomic types or are immutable after construction to ensure
// thread safety without explicit locking.
type srv struct {
	upd libsck.UpdateConn  // Connection update callback (optional, called per accepted connection)
	hdl libsck.HandlerFunc // Connection handler function (required)
	msg *atomic.Value      // Message channel (chan []byte)
	stp *atomic.Value      // Stop listening channel (chan struct{})
	rst *atomic.Value      // Reset/Gone channel (chan struct{}) for connection draining
	run *atomic.Bool       // Server is accepting connections flag
	gon *atomic.Bool       // Server is draining/closing connections flag

	fe *atomic.Value // Error callback (FuncError)
	fi *atomic.Value // Connection info callback (FuncInfo)
	fs *atomic.Value // Server info callback (FuncInfoSrv)

	sf *atomic.Value // Unix socket file path (string)
	sp *atomic.Int64 // Socket file permissions (os.FileMode as int64)
	sg *atomic.Int32 // Socket file group ID (int32)

	nc *atomic.Int64 // Active connection count
}

// OpenConnections returns the current number of active connections.
// Returns the actual count of open client connections being handled.
//
// This is safe to call from multiple goroutines and provides real-time
// connection tracking.
func (o *srv) OpenConnections() int64 {
	return o.nc.Load()
}

// IsRunning returns true if the server is currently accepting new connections.
// Returns false if the server has not started, is shutting down, or has stopped.
//
// This is safe to call concurrently and provides the server's listener state.
func (o *srv) IsRunning() bool {
	return o.run.Load()
}

// IsGone returns true if the server has stopped accepting connections and
// is draining or has drained all existing connections.
//
// This state is set by StopGone() and indicates the server is in final
// shutdown phase. Unlike IsRunning(), this specifically tracks the
// connection draining state.
func (o *srv) IsGone() bool {
	return o.gon.Load()
}

// Done returns a channel that is closed when the server stops accepting connections.
// This channel is closed during StopListen().
//
// Use this to detect when Listen() has exited and the listener has stopped.
// Returns a pre-closed channel if the server is nil or not initialized.
//
// Note: This does not wait for existing connections to close. Use Gone() for that.
func (o *srv) Done() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	}

	if i := o.stp.Load(); i != nil {
		if c, k := i.(chan struct{}); k {
			return c
		}
	}

	return closedChanStruct
}

// Gone returns a channel that is closed when all connections have been closed
// and the server has fully stopped.
//
// This channel is closed by StopGone() after all connections have drained.
// Unlike Done(), which indicates listener stopped, Gone() indicates all
// connections are closed.
//
// Returns a pre-closed channel if the server is nil, not initialized, or
// already in the gone state.
func (o *srv) Gone() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	}
	if o.IsGone() {
		return closedChanStruct
	} else if i := o.rst.Load(); i != nil {
		if g, k := i.(chan struct{}); k {
			return g
		}
	}

	return closedChanStruct
}

// Close performs an immediate shutdown of the server using a background context.
// This is equivalent to calling Shutdown(context.Background()).
//
// For controlled shutdown with a custom timeout, use Shutdown() directly.
func (o *srv) Close() error {
	return o.Shutdown(context.Background())
}

// StopGone signals all connections to close and waits for them to drain.
// This sets the IsGone() state and closes the Gone() channel.
//
// The method uses a 10-second timeout (overriding the provided context) and polls
// every 5ms until OpenConnections() returns 0. Returns ErrGoneTimeout if
// connections don't close within the timeout.
//
// This is typically called during Shutdown() to ensure graceful connection draining.
//
// The method is safe against double-close panics using defer/recover.
func (o *srv) StopGone(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	o.gon.Store(true)

	if i := o.rst.Load(); i != nil {
		if c, k := i.(chan struct{}); k && c != closedChanStruct {
			// Use defer recover to handle potential double close
			func() {
				defer func() {
					_ = recover() // Ignore panic from closing already closed channel
				}()
				close(c)
			}()
		}
	}
	o.rst.Store(closedChanStruct)

	var (
		tck = time.NewTicker(5 * time.Millisecond)
		cnl context.CancelFunc
	)

	ctx, cnl = context.WithTimeout(ctx, 10*time.Second)

	defer func() {
		tck.Stop()
		cnl()
	}()

	for {
		select {
		case <-ctx.Done():
			return ErrGoneTimeout
		case <-tck.C:
			if o.OpenConnections() > 0 {
				continue
			}
			return nil
		}
	}

}

// StopListen signals the server to stop accepting new connections and waits
// for the listener to exit. The Done() channel is closed when this completes.
//
// The method uses a 10-second timeout (overriding the provided context) and polls
// every 5ms until IsRunning() returns false. Returns ErrShutdownTimeout if the
// listener doesn't stop within the timeout.
//
// Existing connections remain active after StopListen(). Use StopGone() to
// drain connections, or Shutdown() to do both.
//
// The method is safe against double-close panics using defer/recover.
func (o *srv) StopListen(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	if i := o.stp.Load(); i != nil {
		if c, k := i.(chan struct{}); k && c != closedChanStruct {
			// Use defer recover to handle potential double close
			func() {
				defer func() {
					_ = recover() // Ignore panic from closing already closed channel
				}()
				close(c)
			}()
		}
	}
	o.stp.Store(closedChanStruct)

	var (
		tck = time.NewTicker(5 * time.Millisecond)
		cnl context.CancelFunc
	)

	ctx, cnl = context.WithTimeout(ctx, 10*time.Second)

	defer func() {
		tck.Stop()
		cnl()
	}()

	for {
		select {
		case <-ctx.Done():
			return ErrShutdownTimeout
		case <-tck.C:
			if o.IsRunning() {
				continue
			}
			return nil
		}
	}

}

// Shutdown performs a graceful server shutdown by stopping the listener
// and draining all connections.
//
// The method:
//  1. Applies a 25-second timeout to the provided context
//  2. Calls StopGone() to signal connections to close and wait for draining
//  3. Calls StopListen() to stop accepting new connections
//  4. Returns any error from either operation
//
// For Unix sockets, this ensures:
//   - No new connections are accepted
//   - Existing connections are closed gracefully
//   - The socket file is cleaned up
//
// Returns ErrShutdownTimeout or ErrGoneTimeout if operations exceed their timeouts.
func (o *srv) Shutdown(ctx context.Context) error {
	if o == nil {
		return ErrInvalidInstance
	}

	var cnl context.CancelFunc
	ctx, cnl = context.WithTimeout(ctx, 25*time.Second)
	defer cnl()

	e := o.StopGone(ctx)
	if err := o.StopListen(ctx); err != nil {
		return err
	} else {
		return e
	}
}

// SetTLS is a no-op for Unix socket servers.
// Unix domain sockets do not support TLS at the transport layer.
// Always returns nil regardless of parameters.
//
// For secure Unix socket communication, consider using file permissions
// to restrict access or application-level encryption.
func (o *srv) SetTLS(enable bool, config libtls.TLSConfig) error {
	return nil
}

// RegisterFuncError registers a callback function for error notifications.
// The callback is invoked whenever an error occurs during server operation,
// including connection errors, I/O errors, and listener errors.
//
// The function receives variadic errors and should not block as it's called
// from various goroutines. Pass nil to clear the callback.
//
// Thread-safe and can be called at any time, even while the server is running.
//
// See github.com/nabbar/golib/socket.FuncError for the callback signature.
func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.fe.Store(f)
}

// RegisterFuncInfo registers a callback function for connection events.
// The callback is invoked for each connection event:
//   - ConnectionNew: New connection accepted
//   - ConnectionRead: Data read from connection
//   - ConnectionWrite: Data written to connection
//   - ConnectionClose: Connection closed
//   - ConnectionCloseRead: Read side closed (half-close)
//   - ConnectionCloseWrite: Write side closed (half-close)
//
// The function receives local and remote addresses and the event state.
// Should not block as it's called from connection handler goroutines.
// Pass nil to clear the callback.
//
// See github.com/nabbar/golib/socket.FuncInfo and ConnState for details.
func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.fi.Store(f)
}

// RegisterFuncInfoServer registers a callback function for server informational messages.
// The callback receives formatted string messages about server lifecycle events:
//   - Server starting/stopping
//   - Listener creation/closure
//   - Socket file creation/removal
//   - Configuration changes
//
// Should not block as it's called from the server's main goroutines.
// Pass nil to clear the callback.
//
// See github.com/nabbar/golib/socket.FuncInfoSrv for the callback signature.
func (o *srv) RegisterFuncInfoServer(f libsck.FuncInfoSrv) {
	if o == nil {
		return
	}

	o.fs.Store(f)
}

// RegisterSocket sets the Unix socket file path, permissions, and group ownership.
// Must be called before Listen().
//
// Parameters:
//   - unixFile: Path to the socket file (e.g., "/tmp/app.sock", "./app.sock")
//   - perm: File permissions (e.g., 0600 for owner-only, 0660 for owner+group)
//   - gid: Group ID for the socket file, or -1 to use the process's default group
//
// The socket file:
//   - Will be created when Listen() is called
//   - Will be removed on server shutdown
//   - Will be deleted if it exists before creating the new socket
//
// File permissions control who can connect to the socket:
//   - 0600: Only the socket owner can connect
//   - 0660: Owner and group members can connect
//   - 0666: Anyone can connect (use with caution)
//
// The address is validated using net.ResolveUnixAddr to ensure it's well-formed.
//
// Returns ErrInvalidGroup if gid exceeds maxGID (32767).
func (o *srv) RegisterSocket(unixFile string, perm os.FileMode, gid int32) error {
	if _, err := net.ResolveUnixAddr(libptc.NetworkUnix.Code(), unixFile); err != nil {
		return err
	} else if gid > maxGID {
		return ErrInvalidGroup
	}

	o.sf.Store(unixFile)
	o.sp.Store(int64(perm))
	o.sg.Store(gid)

	return nil
}

// fctError invokes the registered error callback if one exists.
// Safely handles nil server instances and nil errors.
// This is an internal helper used throughout the server for error reporting.
func (o *srv) fctError(e error) {
	if o == nil {
		return
	}

	if e == nil {
		return
	}

	v := o.fe.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

// fctInfo invokes the registered connection info callback if one exists.
// Reports connection events with local and remote addresses.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper called from connection handling.
func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	v := o.fi.Load()
	if v != nil {
		if fn, ok := v.(libsck.FuncInfo); ok && fn != nil {
			fn(local, remote, state)
		}
	}
}

// fctInfoSrv invokes the registered server info callback if one exists.
// Formats the message with fmt.Sprintf before passing to the callback.
// Safely handles nil callbacks to prevent panics.
// This is an internal helper for server lifecycle logging.
func (o *srv) fctInfoSrv(msg string, args ...interface{}) {
	if o == nil {
		return
	}

	v := o.fs.Load()
	if v != nil {
		if fn, ok := v.(libsck.FuncInfoSrv); ok && fn != nil {
			fn(fmt.Sprintf(msg, args...))
		}
	}
}
