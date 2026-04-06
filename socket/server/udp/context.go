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

package udp

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
)

// sCtx represents a UDP socket context that wraps a UDP connection with
// context awareness and I/O operations.
//
// # Design Pattern
//
// This struct follows the "Contextual Wrapper" pattern, providing a unified
// interface for both lifetime management (via context.Context) and data
// exchange (via io.ReadCloser).
//
// # Interface Implementation
//
// It implements multiple interfaces to be compatible with standard Go idioms:
//   - context.Context: For cancellation and deadline propagation through the handler.
//   - io.Reader: For reading datagrams from the underlying UDP socket.
//   - io.Writer: Formally implemented but disabled to enforce UDP-specific semantics.
//   - io.Closer: For resource cleanup and signaling termination.
//
// # Lifecycle and State Flow
//
//	  [Start]
//	     │
//	     ▼
//	[Listen] ────▶ [New sCtx created] ────▶ [Handler Started]
//	                                           │
//	     ┌─────────────────────────────────────┴──────────┐
//	     │                                                │
//	     ▼                                                ▼
//	[Read Loop]                                   [External Stop]
//	     │                                                │
//	     ├────▶ [Context Cancelled] ◀─────────────────────┤
//	     │             OR                                 │
//	     ├────▶ [Server Shutdown] ◀───────────────────────┤
//	     │             OR                                 │
//	     └────▶ [Explicit Close()] ◀──────────────────────┘
//	                   │
//	                   ▼
//	             [Mark clo=true] (Atomic)
//	                   │
//	                   ▼
//	            [Close UDP Socket]
//	                   │
//	                   ▼
//	            [Cancel Context]
//	                   │
//	                   ▼
//	                 [End]
//
// # Thread Safety
//
// The struct uses sync/atomic.Bool for the 'clo' field to ensure that Close()
// is idempotent and safe to call from multiple goroutines simultaneously.
type sCtx struct {
	loc string             // Cached local address string to avoid repeated allocation/syscalls
	ctx context.Context    // Embedded context for cancellation propagation
	cnl context.CancelFunc // Function to cancel the embedded context
	con *net.UDPConn       // The underlying raw UDP connection
	clo atomic.Bool        // Atomic flag to track closed state (true = closed)
}

// Deadline returns the time when work done on behalf of this context
// should be canceled. This method delegates to the underlying context's
// Deadline method.
//
// This is essential for handlers that need to set per-request timeouts or
// respect global deadlines.
//
// Returns:
//   - time.Time: The deadline time when the context should be considered done.
//   - bool: False if no deadline is set.
func (o *sCtx) Deadline() (deadline time.Time, ok bool) {
	if o == nil || o.ctx == nil {
		return time.Time{}, false
	}
	return o.ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled.
//
// # Use Case
//
// Use this in a select block within your handler to react to server shutdown
// or connection termination:
//
//	select {
//	case <-ctx.Done():
//	    return // Clean exit
//	default:
//	    // Continue processing
//	}
func (o *sCtx) Done() <-chan struct{} {
	if o == nil || o.ctx == nil {
		// Return a closed channel for nil receiver to avoid blocking
		c := make(chan struct{})
		close(c)
		return c
	}
	return o.ctx.Done()
}

// Err returns a non-nil error value after the connection is closed or
// the context is canceled.
//
// # Priority Logic
//
// 1. If 'clo' is true, it returns io.ErrClosedPipe.
// 2. If 'ctx' is nil, it returns io.ErrClosedPipe.
// 3. If 'ctx.Err()' is non-nil, it attempts to Close() the connection
//    (if not already done) and returns the context error.
//
// This ensures that the caller always knows why the context was terminated.
func (o *sCtx) Err() error {
	if o == nil {
		return fmt.Errorf("nil connection context")
	}

	// Check if connection was explicitly closed via Close()
	if o.clo.Load() {
		return io.ErrClosedPipe
	}

	// Check if context was canceled by the parent (e.g. Server.Shutdown)
	if o.ctx == nil {
		return io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		// Ensure resources are released if context is cancelled externally
		if err := o.Close(); err != nil {
			return fmt.Errorf("%v (close error: %v)", e, err)
		}
		return e
	}

	return nil
}

// Value retrieves the value associated with the given key from the context.
//
// This allows passing metadata (like request IDs, trace IDs, or user info)
// through the socket handling pipeline without modifying method signatures.
//
// Parameters:
//   - key: The key to look up (must be comparable).
//
// Returns:
//   - any: The value or nil if not found.
func (o *sCtx) Value(key any) any {
	if o == nil || o.ctx == nil {
		return nil
	}
	return o.ctx.Value(key)
}

// Read reads data from the UDP socket into the provided buffer.
//
// # UDP Behavior
//
// Unlike TCP, which is a stream, UDP is datagram-oriented.
//   - Each Read call consumes exactly ONE datagram from the OS queue.
//   - If the buffer 'p' is smaller than the datagram, the excess data is DISCARDED by the OS.
//   - If no datagram is available, Read blocks until one arrives or the socket is closed.
//
// # Implementation Details
//
// Before calling the underlying Read, it checks:
//   - If the instance is nil.
//   - If the connection is marked as closed ('clo').
//   - If the context has been canceled.
//
// If any of these are true, it returns io.ErrClosedPipe immediately to avoid
// unnecessary blocking syscalls.
//
// Parameters:
//   - p: Destination buffer. For UDP, 65535 bytes is the theoretical max, but 1500 (MTU)
//        is a common practical limit for internet traffic.
//
// Returns:
//   - n: Bytes read.
//   - err: Error, or io.EOF if the connection was closed.
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

	n, err = o.con.Read(p)

	if err != nil && err != io.EOF {
		// Network error occurred, trigger cleanup
		return n, o.onErrorClose(err)
	} else if err != nil {
		// io.EOF or other termination
		return n, o.Close()
	} else {
		return n, nil
	}
}

// Write is intentionally disabled for UDP server contexts.
//
// # Rationale
//
// A UDP server socket is "unconnected" in the network sense. While a TCP socket
// has a fixed destination, a UDP socket can receive datagrams from any source.
//
// In this library's design:
//   - sCtx.Read reads from the shared listener socket.
//   - To reply, you MUST use WriteTo with the specific remote address obtained
//     from the network layer (or the underlying net.UDPConn if exposed).
//
// Using Write() on an unconnected UDP socket would result in "destination address required".
//
// Returns:
//   - n: Always 0.
//   - err: Always io.ErrClosedPipe (to signal that this "pipe" doesn't support writing).
func (o *sCtx) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, io.ErrClosedPipe
	} else if o.clo.Load() {
		return 0, io.ErrClosedPipe
	} else if o.ctx == nil || o.con == nil {
		return 0, io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	} else {
		// Enforce write disablement
		return 0, o.onErrorClose(io.ErrClosedPipe)
	}
}

// Close performs a graceful resource cleanup and state transition.
//
// # Internals
//
// 1. Checks if already closed using o.clo.Swap(true). This is an atomic "Check-and-Set"
//    operation that ensures only the first caller proceeds with the actual closing logic.
// 2. Closes the underlying *net.UDPConn socket. This will unblock any pending Read() calls.
// 3. Cancels the internal context (o.cnl()). This notifies any goroutines watching o.Done().
//
// # Returns
//   - error: The error from net.UDPConn.Close(), if any.
func (o *sCtx) Close() error {
	if o == nil {
		return nil
	}

	// Ensure the context is cancelled last to allow cleanup logic in other goroutines
	// to see the connection as closed before the signal propagates.
	if o.cnl != nil {
		defer o.cnl()
	}

	// Atomic gatekeeper: only one Close() execution allowed.
	if o.clo.Swap(true) {
		return nil
	} else if o.con == nil {
		return nil
	} else {
		return o.con.Close()
	}
}

// IsConnected returns the readiness state of the connection context.
//
// Note: In UDP, "Connected" refers to the local socket state (is it open?),
// as UDP is inherently connectionless at the protocol level.
//
// Returns:
//   - bool: true if the socket is open and the context is active.
func (o *sCtx) IsConnected() bool {
	if o == nil {
		return false
	}
	return !o.clo.Load()
}

// RemoteHost returns the remote peer's identity.
//
// # Format
//
// Returns "address:port(udp)".
//
// # UDP Nuance
//
// For a server listener socket, RemoteAddr() might return nil or a generic value
// because the socket is not "connected" to a single peer. The remote address
// is typically extracted per-datagram during ReadFrom.
func (o *sCtx) RemoteHost() string {
	if o == nil {
		return ""
	} else if c := o.con; c == nil {
		return ""
	} else if a := c.RemoteAddr(); a == nil {
		return ""
	} else {
		return a.String() + "(" + libptc.NetworkUDP.Code() + ")"
	}
}

// LocalHost returns the local identity of the socket.
//
// # Format
//
// Returns "address:port(udp)".
func (o *sCtx) LocalHost() string {
	if o == nil {
		return ""
	}
	return o.loc + "(" + libptc.NetworkUDP.Code() + ")"
}

// onErrorClose is an internal helper that orchestrates the Close() and error reporting.
//
// If an I/O error occurs, it's vital to close the context immediately to prevent
// resource leaks and notify other parts of the system.
func (o *sCtx) onErrorClose(e error) error {
	if e == nil {
		return nil
	} else if err := o.Close(); err != nil {
		// Return a wrapped error combining the original failure and the cleanup failure.
		return fmt.Errorf("%v, %v", e, err)
	} else {
		return e
	}
}
