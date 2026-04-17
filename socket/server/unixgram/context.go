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
	"context"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
)

// sCtx represents a sophisticated wrapper around a Unix domain datagram socket,
// providing deep integration with the standard Go library's Context pattern and I/O interfaces.
//
// # Architectural Context and Object Pooling
//
// In a high-performance Unixgram server, datagrams can arrive at a rate of tens of thousands per
// second. To accommodate this without overwhelming the Go Garbage Collector (GC), sCtx instances
// are managed by a sync.Pool.
//
//   - Reuse: Each sCtx is reset and returned to the pool once its lifecycle (e.g., the handler execution)
//     is complete.
//   - Reset: The reset() method re-initializes the context, cancellation function, and connection
//     pointers, ensuring no state leaks between processing cycles.
//
// # Interface Implementation and Behavioral Contracts
//
// sCtx satisfies multiple standard library interfaces, each with specific behaviors for Unixgram:
//
// 1. context.Context:
//    Allows for propagation of deadlines and cancellation signals. When the server is shut down,
//    the context is cancelled, which immediately notifies the handler to stop processing.
//
// 2. io.Reader:
//    Reads from the underlying SOCK_DGRAM socket. In datagram mode, each call to Read()
//    returns exactly one complete datagram. If the provided buffer is smaller than the datagram,
//    the excess bytes are discarded by the kernel.
//
// 3. io.Writer:
//    Intentionally returns io.ErrClosedPipe. This is a deliberate design choice. Since Unixgram
//    is connectionless, a naked Write() lacks a destination address. For sending responses,
//    one should use the net.UnixConn's WriteTo() method with the sender's address.
//
// 4. io.Closer:
//    Cleans up the context, cancels the cancellation function (cnl), and closes the underlying
//    net.UnixConn. It is safe and idempotent.
//
// # Thread Safety and State Management
//
// The 'clo' field (atomic.Bool) ensures that the context state is managed safely across multiple
// goroutines. For instance, a reader goroutine and a monitoring goroutine can safely check
// or modify the connection state simultaneously without locks.
type sCtx struct {
	loc string             // The local filesystem path where the socket is bound.
	ctx context.Context    // The parent context, used for propagation of signals.
	cnl context.CancelFunc // The cancellation function for this specific context instance.
	con *net.UnixConn      // The actual network connection pointer.
	clo atomic.Bool        // An atomic flag indicating if the context has been closed.
}

// Deadline returns the point in time after which work done on behalf of this context
// should be canceled. This method delegates directly to the underlying context's Deadline.
//
// Returns:
//   - time.Time: The exact deadline timestamp.
//   - bool: True if a deadline is set, false otherwise.
//
// Use Case:
// A handler processing a large batch of datagrams can check its deadline to decide
// whether to continue processing or to abort and clean up before the server shuts down.
func (o *sCtx) Deadline() (deadline time.Time, ok bool) {
	if o == nil || o.ctx == nil {
		return time.Time{}, false
	}
	return o.ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this context
// should be canceled. This signal is critical for implementing responsive,
// low-latency shutdown logic in your handler.
//
// Returns:
//   - <-chan struct{}: A channel that is closed when the context is cancelled or the connection is closed.
//
// Example Pattern:
//
//	for {
//	    select {
//	    case <-ctx.Done():
//	        return // Stop processing
//	    default:
//	        n, err := ctx.Read(buffer)
//	        // ... process ...
//	    }
//	}
func (o *sCtx) Done() <-chan struct{} {
	if o == nil || o.ctx == nil {
		// If the receiver or context is nil, we return an already closed channel
		// to signal that no work should be performed.
		c := make(chan struct{})
		close(c)
		return c
	}
	return o.ctx.Done()
}

// Err returns a non-nil error value after Done is closed. It clarifies the reason
// for the context's termination.
//
// Returns:
//   - io.ErrClosedPipe: If the connection was explicitly closed by the application.
//   - context.Canceled: If the server or parent context initiated a shutdown.
//   - context.DeadlineExceeded: If the context's deadline was surpassed.
//   - nil: If the context is still active.
func (o *sCtx) Err() error {
	if o == nil {
		return fmt.Errorf("nil connection context")
	}

	// First, check the atomic flag to see if we manually closed the pipe.
	if o.clo.Load() {
		return io.ErrClosedPipe
	}

	// Then, check the underlying context's error state.
	if o.ctx == nil {
		return io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		// If the context is erroring (cancelled), we ensure the connection is also closed.
		if err := o.Close(); err != nil {
			return fmt.Errorf("%v (additional close error: %v)", e, err)
		}
		return e
	}

	return nil
}

// Value retrieves the data associated with a key from the context's internal store.
// This is used for passing request-scoped metadata through the server.
//
// Parameters:
//   - key: The unique key for the value (usually a private type to avoid collisions).
//
// Returns:
//   - any: The stored value or nil if the key is not found.
//
// Use Case:
// Storing tracing IDs (e.g., OpenTelemetry span contexts) within the sCtx so they
// are available to all processing functions without being explicitly passed as arguments.
func (o *sCtx) Value(key any) any {
	if o == nil || o.ctx == nil {
		return nil
	}
	return o.ctx.Value(key)
}

// Read reads a single Unix datagram from the socket into the provided byte slice 'p'.
//
// # Technical Behavior and Constraints
//
// - Connectionless Semantics: In SOCK_DGRAM mode, the kernel treats each datagram as a
//   discrete message. Read() blocks until at least one datagram is available.
//
// - Message Boundaries: Unlike TCP (which is a stream), Unixgram preserves boundaries.
//   Each call to Read() will return exactly one datagram, regardless of how many are
//   queued in the kernel buffer.
//
// - Truncation: If the provided slice 'p' is smaller than the incoming datagram,
//   the datagram is truncated to len(p) and the remaining bytes are discarded by
//   the operating system. It is recommended to use a buffer of at least 65535 bytes
//   to avoid accidental data loss.
//
// - Error Handling: Any error during the read (except EOF) will trigger an internal
//   Close() of the sCtx to maintain state consistency.
//
// Returns:
//   - n: The number of bytes successfully read.
//   - err: An error object, or nil if the read was successful.
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

	// In Unixgram, EOF might occur if the other side closes their socket file,
	// but since it's connectionless, it's more often an indication that our own listener is stopping.
	if err != nil && err != io.EOF {
		return n, o.onErrorClose(err)
	} else if err != nil {
		return n, o.Close()
	} else {
		return n, nil
	}
}

// Write is explicitly disabled for the server-side sCtx.
//
// # Rationale for Disabling Write()
//
// Unix Datagram sockets (SOCK_DGRAM) are connectionless. A server socket typically
// acts as a "mailbox" receiving messages from many different senders. A standard
// Write() call does not include a destination address; it requires the socket to
// have been "connected" to a specific peer using the Connect() system call.
//
// Since our server context handles multiple potential peers, an unqualified Write()
// would be ambiguous.
//
// # How to Respond to a Datagram:
//
// To send a reply, you should use the WriteTo() method of the underlying connection
// which accepts an explicit remote address. You can obtain this address by using
// the ReadFrom() method of the net.UnixConn instead of the standard Read().
//
// Returns:
//   - 0, io.ErrClosedPipe (Always).
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
		// We explicitly prevent Write() to avoid developer confusion about the connectionless nature of Unixgram.
		return 0, o.onErrorClose(io.ErrClosedPipe)
	}
}

// Close performs a comprehensive cleanup of the context's resources.
//
// Lifecycle Actions:
//  1. Atomic Flag: Sets 'clo' to true using an atomic Swap, preventing race conditions.
//  2. Context Signal: Executes 'cnl()' to close the Done() channel and notify all listeners.
//  3. Network Cleanup: Closes the underlying *net.UnixConn to release the file descriptor.
//
// This method is idempotent and safe for concurrent calls from multiple goroutines.
//
// Returns:
//   - error: Any error encountered while closing the network connection.
func (o *sCtx) Close() error {
	if o == nil {
		return nil
	}

	// Ensure we only perform the close logic once.
	if o.clo.Swap(true) {
		return nil
	}

	// Trigger context cancellation.
	if o.cnl != nil {
		o.cnl()
	}

	// Release network resources.
	if o.con == nil {
		return nil
	} else {
		return o.con.Close()
	}
}

// IsConnected returns a boolean indicating whether the context and its underlying
// connection are currently active and capable of performing I/O.
//
// Note:
// In the context of Unixgram, "Connected" means the local socket file descriptor is open
// and the context has not been cancelled. It does not imply an active handshake or
// session with a remote peer.
func (o *sCtx) IsConnected() bool {
	if o == nil {
		return false
	}
	return !o.clo.Load()
}

// RemoteHost returns a string representing the remote peer's address.
//
// Format: "path/to/peer.sock(unixgram)"
//
// Returns:
//   - An empty string if the socket is not connected to a specific remote peer.
//
// Use Case:
// If you have used Connect() on the underlying socket to lock it to a specific
// peer, this method will return that peer's path. Otherwise, for a generic
// server socket, this will likely be empty.
func (o *sCtx) RemoteHost() string {
	if o == nil {
		return ""
	} else if c := o.con; c == nil {
		return ""
	} else if a := c.RemoteAddr(); a == nil {
		return ""
	} else {
		return a.String() + "(" + libptc.NetworkUnixGram.Code() + ")"
	}
}

// LocalHost returns the filesystem path of the local socket bound to this context.
//
// Format: "/path/to/socket.sock(unixgram)"
//
// This value is cached during the context's initialization (reset) for high performance.
func (o *sCtx) LocalHost() string {
	if o == nil {
		return ""
	}
	return o.loc + "(" + libptc.NetworkUnixGram.Code() + ")"
}

// onErrorClose is an internal utility that facilitates the "fail-fast and clean-up"
// philosophy. It closes the context upon encountering an error and ensures that
// both the original error and any error during closure are reported.
//
// Parameters:
//   - e: The original error that occurred during an I/O operation.
//
// Returns:
//   - error: A formatted error combining the original and the close error, or just the original.
func (o *sCtx) onErrorClose(e error) error {
	if e == nil {
		return nil
	} else if err := o.Close(); err != nil {
		// We combine errors to provide maximum visibility into the failure.
		return fmt.Errorf("%v, %v", e, err)
	} else {
		return e
	}
}

// reset prepares the sCtx instance for a new processing cycle. This is called
// by the srv when retrieving an sCtx from the sync.Pool.
//
// Internal State Transition:
// - Atomic 'clo' is set to false.
// - All pointers (ctx, cnl, con) are updated to the new request's values.
// - The local address string is cached.
func (o *sCtx) reset(ctx context.Context, cnl context.CancelFunc, con *net.UnixConn, loc string) {
	o.ctx = ctx
	o.cnl = cnl
	o.con = con
	o.loc = loc
	o.clo.Store(false)
}
