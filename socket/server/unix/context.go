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
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
)

// sCtx (Server Context) is a core internal component that bridges several Go interfaces:
//   - context.Context: For cancellation propagation and metadata passing.
//   - io.Reader: For stream-oriented data reception from the Unix socket.
//   - io.Writer: For stream-oriented data transmission to the Unix socket.
//   - io.Closer: For deterministic resource cleanup.
//
// # Design Goals
//
// 1. Resource Recycling: To minimize Garbage Collection (GC) pressure, `sCtx` is designed to be
// reused via a `sync.Pool`. This is why it includes a `reset()` method. By recycling these
// structures, the server achieves zero allocations on the critical connection handling path.
//
// 2. High Performance I/O: It wraps the underlying `net.UnixConn` and provides context-aware
// Read and Write operations. It uses atomic primitives for state tracking (closure flag,
// activity counter) to avoid expensive mutex locking.
//
// 3. Centralized Idle Detection: Instead of using per-connection tickers, it exposes an atomic
// counter (`cnt`) that is incremented by an external `Idle Manager`. Every successful Read
// or Write operation resets this counter to zero.
//
// # Internal State (Fields)
//
//   - rem: The string representation of the remote peer's socket path.
//   - loc: The string representation of the local server's socket path.
//   - ctx: The active `context.Context` for the current connection lifecycle.
//   - cnl: The `CancelFunc` associated with the active context.
//   - con: The underlying `io.ReadWriteCloser` (typically a `*net.UnixConn`).
//   - clo: An `atomic.Bool` flag indicating if the connection has been closed.
//   - cnt: An `atomic.Uint32` counter used for centralized idle timeout management.
//
// # Thread Safety & Concurrency
//
// `sCtx` is safe for concurrent access by multiple goroutines:
//   - Metadata (rem, loc) is immutable after the `reset()` call.
//   - State flags (clo, cnt) use lock-free atomic operations.
//   - The underlying context is inherently thread-safe.
// Note: As with all `io.Reader/Writer` implementations, concurrent Read or concurrent Write calls
// on the same `sCtx` from different goroutines will lead to interleaved data and undefined behavior.
type sCtx struct {
	rem string // remote peer's socket path (e.g., /tmp/client.sock)
	loc string // local server's socket path (e.g., /tmp/server.sock)

	ctx context.Context
	cnl context.CancelFunc
	con io.ReadWriteCloser
	clo atomic.Bool
	cnt atomic.Uint32
}

// Deadline returns the time when work done on behalf of this connection should be canceled.
// This is part of the context.Context interface implementation.
//
// It delegates directly to the underlying context provided during initialization/reset.
//
// Returns:
//   - deadline: The specific time when the context expires.
//   - ok: True if a deadline is set, false otherwise.
func (o *sCtx) Deadline() (deadline time.Time, ok bool) {
	if o == nil || o.ctx == nil {
		return time.Time{}, false
	}
	return o.ctx.Deadline()
}

// Done returns a channel that is closed when work done on behalf of this connection
// should be canceled. This is part of the context.Context interface implementation.
//
// The channel is closed when:
//  1. The underlying connection is explicitly closed via `Close()`.
//  2. The parent context (from the server's `Listen` call) is cancelled.
//  3. An I/O error occurs that triggers an automatic closure.
//  4. The connection is recycled or reset.
//
// Returns:
//   - <-chan struct{}: A channel that signals the end of the connection's lifecycle.
func (o *sCtx) Done() <-chan struct{} {
	if o == nil || o.ctx == nil {
		// Return a pre-closed channel for nil receivers to prevent hanging.
		c := make(chan struct{})
		close(c)
		return c
	}
	return o.ctx.Done()
}

// Err returns a non-nil error value after the connection is closed or the context
// is canceled. This is part of the context.Context interface implementation.
//
// # Error Priorities:
//  1. If `clo` is true: Returns `io.ErrClosedPipe` (explicit closure).
//  2. If the context has an error: Returns that error, potentially wrapped with close errors.
//
// Returns:
//   - error: nil if the connection is active, otherwise the reason for termination.
func (o *sCtx) Err() error {
	if o == nil {
		return fmt.Errorf("nil connection context")
	}

	// Priority 1: Check if the connection was explicitly marked as closed.
	if o.clo.Load() {
		return io.ErrClosedPipe
	}

	// Priority 2: Check for context-driven cancellation or timeout.
	if e := o.ctx.Err(); e != nil {
		// Attempt a cleanup and report any additional errors during closure.
		if err := o.Close(); err != nil {
			return fmt.Errorf("%v (additional close error: %v)", e, err)
		}
		return e
	}

	return nil
}

// Value retrieves a value from the connection's context for the given key.
// This is part of the context.Context interface implementation.
//
// # Common Use Cases:
//   - Retrieving request-specific metadata (Request IDs).
//   - Accessing credentials passed through the context.
//   - Propagating tracing information across boundaries.
//
// Parameters:
//   - key: The unique key for the metadata.
//
// Returns:
//   - any: The value associated with the key, or nil if not found.
func (o *sCtx) Value(key any) any {
	if o == nil || o.ctx == nil {
		return nil
	}
	return o.ctx.Value(key)
}

// Read implements the io.Reader interface for the Unix socket connection.
// It provides context-aware reading and automatic integration with the Idle Manager.
//
// # Detailed Behavior:
//  1. State Check: Verifies that the receiver is not nil and the connection is not closed (`clo`).
//  2. Context Check: Verifies that the associated context has not been cancelled.
//  3. I/O Operation: Performs the actual read from the underlying Unix socket.
//  4. Activity Update: On success, resets the atomic counter `cnt` to zero.
//  5. Error Handling:
//     - `io.EOF`: Triggers a graceful `Close()` and returns the EOF.
//     - Other errors: Triggers `onErrorClose()` to clean up and wrap the error.
//
// # Idle Manager Integration:
// Every successful read is considered "activity". Resetting `cnt` to 0 prevents the
// centralized Idle Manager from timing out this connection prematurely.
//
// Parameters:
//   - p: The buffer to read data into.
//
// Returns:
//   - n: Number of bytes read.
//   - err: Any error encountered (nil on success).
func (o *sCtx) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, io.ErrClosedPipe
	} else if o.clo.Load() {
		return 0, io.ErrClosedPipe
	} else if o.ctx == nil || o.con == nil {
		return 0, io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	}

	// Perform the actual network I/O.
	n, err = o.con.Read(p)

	// Reset the idle activity counter.
	o.cnt.Store(0)

	if err != nil && err != io.EOF {
		// Handle unexpected network errors.
		return n, o.onErrorClose(err)
	} else if err != nil {
		// Handle graceful closure (EOF).
		return n, o.Close()
	} else {
		return n, nil
	}
}

// Write implements the io.Writer interface for the Unix socket connection.
// It provides context-aware writing and automatic integration with the Idle Manager.
//
// # Detailed Behavior:
//  1. State Check: Verifies that the receiver is not nil and the connection is not closed (`clo`).
//  2. Context Check: Verifies that the associated context has not been cancelled.
//  3. I/O Operation: Performs the actual write to the underlying Unix socket.
//  4. Activity Update: On success, resets the atomic counter `cnt` to zero.
//  5. Error Handling:
//     - `io.EOF`: Triggers a graceful `Close()`.
//     - Other errors: Triggers `onErrorClose()`.
//
// # Write Performance:
// Unix sockets are typically faster than TCP because they bypass the network stack. However,
// large writes may still block if the kernel's socket buffers are full.
//
// Parameters:
//   - p: The data buffer to be sent.
//
// Returns:
//   - n: Number of bytes written.
//   - err: Any error encountered (nil on success).
func (o *sCtx) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, io.ErrClosedPipe
	} else if o.clo.Load() {
		return 0, io.ErrClosedPipe
	} else if o.ctx == nil || o.con == nil {
		return 0, io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	}

	// Perform the actual network I/O.
	n, err = o.con.Write(p)

	// Reset the idle activity counter.
	o.cnt.Store(0)

	if err != nil && err != io.EOF {
		// Handle unexpected network errors.
		return n, o.onErrorClose(err)
	} else if err != nil {
		// Handle graceful closure (EOF).
		return n, o.Close()
	} else {
		return n, nil
	}
}

// Close implements the io.Closer interface and manages the connection's teardown.
// This method is idempotent and thread-safe.
//
// # Cleanup Lifecycle:
//  1. Context Cancellation: Triggers the cancellation of the connection's context (`cnl`).
//  2. Atomic Swap: Uses `clo.Swap(true)` to ensure the closure logic only runs once.
//  3. Socket Closure: Closes the underlying `net.UnixConn`.
//
// # Importance in Pooling:
// After `Close()` is called, the `sCtx` structure remains in memory until it is
// returned to the `sync.Pool`. The `reset()` method must be called before it
// can be used for another connection.
//
// Returns:
//   - error: Any error encountered while closing the socket.
func (o *sCtx) Close() error {
	if o == nil {
		return nil
	}

	// Always trigger the context cancellation first to unblock any waiting goroutines.
	if o.cnl != nil {
		defer o.cnl()
	}

	// Use atomic swap to ensure we only close the underlying connection once.
	if o.clo.Swap(true) {
		// Already closed.
		return nil
	} else if o.con == nil {
		// No connection to close.
		return nil
	} else {
		// Perform the actual socket closure.
		return o.con.Close()
	}
}

// IsConnected provides a simple way to check the connection status.
//
// Returns:
//   - bool: True if the connection is active and not yet closed, false otherwise.
func (o *sCtx) IsConnected() bool {
	return !o.clo.Load()
}

// RemoteHost returns a formatted string of the remote client's socket path.
//
// Format: "path(protocol)" (e.g., "/tmp/client.sock(unix)")
//
// Returns:
//   - string: The formatted remote address.
func (o *sCtx) RemoteHost() string {
	return o.rem + "(" + libptc.NetworkUnix.Code() + ")"
}

// LocalHost returns a formatted string of the local server's socket path.
//
// Format: "path(protocol)" (e.g., "/tmp/server.sock(unix)")
//
// Returns:
//   - string: The formatted local address.
func (o *sCtx) LocalHost() string {
	return o.loc + "(" + libptc.NetworkUnix.Code() + ")"
}

// Ref returns the reference string for the Idle Manager.
// This implements the interface required by the centralized idle detection system.
//
// Returns:
//   - string: The RemoteHost string used as a unique identifier for logging.
func (o *sCtx) Ref() string {
	return o.RemoteHost()
}

// Inc increments the idle activity counter.
// This is called periodically by the external Idle Manager scanning loop.
func (o *sCtx) Inc() {
	o.cnt.Add(1)
}

// Get retrieves the current value of the idle activity counter.
// Used by the Idle Manager to determine if a connection has been inactive too long.
//
// Returns:
//   - uint32: The current idle count.
func (o *sCtx) Get() uint32 {
	return o.cnt.Load()
}

// onErrorClose is an internal helper that closes the connection and combines errors.
// It ensures proper cleanup when an error occurs during I/O operations.
//
// Parameters:
//   - e: The error that triggered the close operation.
//
// Returns:
//   - error: The original error, or a combined error if Close also fails.
func (o *sCtx) onErrorClose(e error) error {
	if e == nil {
		return nil
	} else if err := o.Close(); err != nil {
		return fmt.Errorf("%v, %v", e, err)
	} else {
		return e
	}
}

// reset re-initializes an `sCtx` structure for a new connection.
// This is a critical component of the `sync.Pool` implementation, ensuring that
// structures are sanitized before reuse.
//
// # Reset Actions:
//  1. Address Localization: Sets the local and remote socket paths.
//  2. Context Injection: Injects the new connection-specific context and cancel function.
//  3. Connection Injection: Injects the new `io.ReadWriteCloser`.
//  4. Atomic Reset: Resets the `clo` (closed) and `cnt` (activity) flags to their initial states.
//
// Parameters:
//   - ctx: The new context for the connection.
//   - cnl: The cancel function for the new context.
//   - con: The underlying Unix socket connection.
//   - l: The local address.
//   - r: The remote address.
func (o *sCtx) reset(ctx context.Context, cnl context.CancelFunc, con io.ReadWriteCloser, l, r net.Addr) {
	// Update addresses.
	if l != nil {
		o.loc = l.String()
	} else {
		o.loc = ""
	}

	if r != nil {
		o.rem = r.String()
	} else {
		o.rem = ""
	}

	// Update dependencies.
	o.ctx = ctx
	o.cnl = cnl
	o.con = con

	// Reset atomic states.
	o.clo.Store(false)
	o.cnt.Store(0)
}
