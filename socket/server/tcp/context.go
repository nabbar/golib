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
	"context"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"
)

// sCtx implements a high-performance, context-aware wrapper for TCP connections.
// It fulfills the following interfaces:
//   - context.Context: Allows propagation of cancellation and deadlines to handlers.
//   - io.Reader: Provides thread-safe, activity-tracking reads from the socket.
//   - io.Writer: Provides thread-safe, activity-tracking writes to the socket.
//   - io.Closer: Ensures idempotent and safe resource release.
//   - idlemgr.Identifier: Integrates with a centralized idle connection manager.
//
// # Design Pattern: Object Pooling
//
// To minimize Garbage Collector (GC) pressure in high-concurrency environments
// (e.g., thousands of connections per second), sCtx is designed to be recycled
// using sync.Pool. The reset() method is used to reinitialize the structure
// for a new connection without reallocating memory.
//
// # Activity Tracking Flow
//
// Any I/O operation (Read or Write) resets the internal activity counter (cnt).
// This counter is periodically incremented by the idle manager. If it reaches
// a threshold defined in the server configuration, the connection is closed.
//
//	[User Handler] --(Read/Write)--> [sCtx] --(Reset cnt)--> [Socket]
//	     |                               ^
//	     v                               |
//	[Idle Manager] --(Periodical Inc)----+
//	     |
//	     +-- (If threshold reached) --> [Close()]
//
// # Lifecycle Diagram
//
//	Alloc/Pool Get -> reset() -> Register in IdleMgr -> User Handler -> Unregister -> reset() -> Pool Put
type sCtx struct {
	ptc string // Protocol designation (e.g., "tcp" or "tcp+tls")
	rem string // Cached remote address string
	loc string // Cached local address string

	ctx context.Context    // Parent context for cancellation propagation
	cnl context.CancelFunc // Cancellation function for this specific connection
	con io.ReadWriteCloser // Underlying network connection (net.TCPConn or tls.Conn)
	clo atomic.Bool        // Idempotent closure flag
	cnt atomic.Uint32      // Inactivity counter for idle management
}

// Deadline returns the time when work done on behalf of this context
// should be canceled. This method delegates to the underlying context's
// Deadline method.
//
// Part of the context.Context interface.
func (o *sCtx) Deadline() (deadline time.Time, ok bool) {
	if o == nil || o.ctx == nil {
		return time.Time{}, false
	}
	return o.ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. This channel is closed when the connection
// is closed (via Close()) or the parent context is canceled.
//
// Part of the context.Context interface.
func (o *sCtx) Done() <-chan struct{} {
	if o == nil || o.ctx == nil {
		// Return a closed channel for nil receiver
		c := make(chan struct{})
		close(c)
		return c
	}
	return o.ctx.Done()
}

// Err returns a non-nil error value after the connection is closed or
// the context is canceled. It returns io.ErrClosedPipe if the connection
// was explicitly closed, or the context's error if it was canceled.
//
// If the context is canceled, it automatically triggers a Close() to
// ensure resources are released.
//
// Part of the context.Context interface.
func (o *sCtx) Err() error {
	if o == nil {
		return fmt.Errorf("nil connection context")
	}

	// Check if connection was explicitly closed via atomic flag
	if o.clo.Load() {
		return io.ErrClosedPipe
	}

	// Check if context was canceled or exceeded deadline
	if e := o.ctx.Err(); e != nil {
		// Close the connection and combine errors if needed
		if err := o.Close(); err != nil {
			return fmt.Errorf("%v (close error: %v)", e, err)
		}
		return e
	}

	return nil
}

// Value retrieves the value associated with the given key from the context.
// This method implements the context.Context interface and provides access
// to values stored in the parent context.
//
// Parameters:
//   - key: The key for which to retrieve the value.
//
// Returns:
//   - any: The value associated with the key, or nil if the key is not found.
func (o *sCtx) Value(key any) any {
	if o == nil || o.ctx == nil {
		return nil
	}
	return o.ctx.Value(key)
}

// Read reads data from the TCP connection into the provided buffer.
// It automatically resets the idle activity counter on successful read.
//
// # Error Handling
//
//   - io.EOF: Triggers a graceful Close() and returns io.EOF.
//   - Any other error: Triggers an onErrorClose() which ensures the connection
//     is terminated and returns the error.
//   - Context cancellation: Detected before reading, returns the context error.
//
// Part of the io.Reader interface.
func (o *sCtx) Read(p []byte) (n int, err error) {
	if o == nil || o.clo.Load() || o.ctx == nil || o.con == nil {
		return 0, io.ErrClosedPipe
	}

	if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	}

	n, err = o.con.Read(p)

	// Activity detected: Reset the idle manager counter
	o.cnt.Store(0)

	if err != nil && err != io.EOF {
		// Unexpected I/O error
		return n, o.onErrorClose(err)
	} else if err != nil {
		// io.EOF: Peer closed the connection
		return n, o.Close()
	}

	return n, nil
}

// Write writes data from the provided buffer to the TCP connection.
// It automatically resets the idle activity counter on successful write.
//
// # Performance Note
//
// In the current implementation, any write error (even partial) triggers
// a connection closure via onErrorClose() to maintain state consistency.
//
// Part of the io.Writer interface.
func (o *sCtx) Write(p []byte) (n int, err error) {
	if o == nil || o.clo.Load() || o.ctx == nil || o.con == nil {
		return 0, io.ErrClosedPipe
	}

	if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	}

	n, err = o.con.Write(p)

	// Activity detected: Reset the idle manager counter
	o.cnt.Store(0)

	if err != nil && err != io.EOF {
		return n, o.onErrorClose(err)
	} else if err != nil {
		return n, o.Close()
	}

	return n, nil
}

// Close closes the connection and releases all associated resources.
// It is idempotent and safe for concurrent use.
//
// # Cleanup Sequence
//
//  1. Atomic check: If already closed, return nil.
//  2. Cancel context: Triggers o.cnl() to signal all dependent goroutines.
//  3. Close socket: Terminates the underlying net.Conn or tls.Conn.
//
// Part of the io.Closer interface.
func (o *sCtx) Close() error {
	if o == nil {
		return nil
	}

	// Always ensure the cancel function is called to release context resources.
	if o.cnl != nil {
		defer o.cnl()
	}

	// Swap returns the old value. If it was already true, we do nothing.
	if o.clo.Swap(true) {
		return nil
	}

	if o.con == nil {
		return nil
	}

	return o.con.Close()
}

// IsConnected reports whether the connection is still active and not closed.
func (o *sCtx) IsConnected() bool {
	return !o.clo.Load()
}

// RemoteHost returns the remote peer's address with protocol information.
// Format: "1.2.3.4:1234(tcp)" or "1.2.3.4:1234(tcp+tls)".
func (o *sCtx) RemoteHost() string {
	return o.rem + "(" + o.ptc + ")"
}

// LocalHost returns the local server's address with protocol information.
// Format: "0.0.0.0:8080(tcp)" or "0.0.0.0:8080(tcp+tls)".
func (o *sCtx) LocalHost() string {
	return o.loc + "(" + o.ptc + ")"
}

// Ref returns a unique reference for the connection, used by the idle manager.
// Currently, returns the RemoteHost string.
//
// Implements idlemgr.Identifier.
func (o *sCtx) Ref() string {
	return o.RemoteHost()
}

// Inc increments the idle activity counter. This is typically called by
// the idle manager during its periodic scan.
//
// Implements idlemgr.Identifier.
func (o *sCtx) Inc() {
	o.cnt.Add(1)
}

// Get returns the current value of the idle activity counter.
//
// Implements idlemgr.Identifier.
func (o *sCtx) Get() uint32 {
	return o.cnt.Load()
}

// onErrorClose is an internal helper that closes the connection and combines errors.
// It ensures proper cleanup when an error occurs during I/O operations.
func (o *sCtx) onErrorClose(e error) error {
	if e == nil {
		return nil
	}
	if err := o.Close(); err != nil {
		return fmt.Errorf("%v, %v", e, err)
	}
	return e
}

// reset reinitializes the sCtx structure for reuse in a sync.Pool.
// It wipes all connection-specific data and sets the new context and connection.
func (o *sCtx) reset(ctx context.Context, cnl context.CancelFunc, con io.ReadWriteCloser, l, r net.Addr, t bool) {
	if l != nil {
		o.loc = l.String()
		if t {
			o.ptc = l.Network() + "+tls"
		} else {
			o.ptc = l.Network()
		}
	} else {
		o.loc = ""
		o.ptc = ""
	}

	if r != nil {
		o.rem = r.String()
	} else {
		o.rem = ""
	}

	o.ctx = ctx
	o.cnl = cnl
	o.con = con

	// Reset atomic flags and counters
	o.clo.Store(false)
	o.cnt.Store(0)
}
