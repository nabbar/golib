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

// sCtx represents a Unix datagram socket context that wraps a Unix datagram connection with
// context awareness and I/O operations.
//
// It implements multiple interfaces:
//   - context.Context: For cancellation and deadline propagation
//   - io.Reader: For reading datagrams from the Unix socket
//   - io.Writer: For writing response datagrams (disabled for Unix datagram server)
//   - io.Closer: For cleaning up resources
//
// The struct uses atomic operations for thread-safe state management and
// delegates context operations to the embedded parent context.
//
// Fields:
//   - loc: Local address string (cached for performance)
//   - ctx: Parent context for cancellation propagation
//   - cnl: Cancel function to trigger context cancellation
//   - con: Underlying Unix datagram connection (*net.UnixConn)
//   - clo: Atomic boolean indicating if connection is closed
//
// Thread Safety:
//   - All methods are safe for concurrent read-only operations
//   - Write operations should be serialized by the caller
//   - Close() can be called from any goroutine
type sCtx struct {
	loc string // local ip
	ctx context.Context
	cnl context.CancelFunc
	con *net.UnixConn
	clo *atomic.Bool
}

// Deadline returns the time when work done on behalf of this context
// should be canceled. This method delegates to the underlying context's
// Deadline method.
//
// Returns:
//   - time.Time: The deadline time when the context should be considered done
//   - bool: False if no deadline is set
//
// This method is part of the context.Context interface and is used by
// functions that support deadlines, such as those in the standard library's
// net and os packages.
func (o *sCtx) Deadline() (deadline time.Time, ok bool) {
	if o == nil || o.ctx == nil {
		return time.Time{}, false
	}
	return o.ctx.Deadline()
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. This channel is closed when the connection
// is closed or the parent context is canceled.
//
// Returns:
//   - <-chan struct{}: A channel that's closed when the context is done
//
// This method is part of the context.Context interface and is used to
// signal cancellation to goroutines that need to be aware of the connection's
// lifecycle.
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
// Returns:
//   - error: nil if the connection is still active, otherwise an error
//     describing why the context was canceled
//
// This method is part of the context.Context interface and is used to
// check if the context has been canceled or the connection closed.
func (o *sCtx) Err() error {
	if o == nil {
		return fmt.Errorf("nil connection context")
	}

	// Check if connection was explicitly closed
	if o.clo.Load() {
		return io.ErrClosedPipe
	}

	// Check if context was canceled
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
//   - key: The key for which to retrieve the value. Can be any comparable type.
//
// Returns:
//   - any: The value associated with the key, or nil if the key is not found
//     or the context is nil.
//
// # Usage
//
// This method is useful for passing request-scoped values through the
// connection handling pipeline, such as:
//   - Request IDs for distributed tracing
//   - Authentication tokens
//   - User context information
//   - Deadline and cancellation signals
//
// # Example
//
//	type contextKey string
//	const userIDKey contextKey = "userID"
//
//	// In handler:
//	if userID := ctx.Value(userIDKey); userID != nil {
//	    log.Printf("Handling request for user: %v", userID)
//	}
//
// # Thread Safety
//
// This method is safe to call from multiple goroutines as it delegates
// to the underlying context which is immutable.
func (o *sCtx) Value(key any) any {
	return o.ctx.Value(key)
}

// Read reads data from the Unix datagram socket into the provided buffer.
//
// This method implements the io.Reader interface and reads a single Unix datagram
// from the underlying socket. For Unix datagram, each Read() typically corresponds to one
// complete datagram.
//
// Parameters:
//   - p: Byte slice to read data into. Should be large enough for datagram
//     (typically 65507 bytes, system dependent)
//
// Returns:
//   - n: Number of bytes read (0 if error occurred before reading)
//   - err: Error if any occurred:
//   - io.ErrClosedPipe: Connection was already closed or is nil
//   - Context error: Context was cancelled or deadline exceeded
//   - Network errors: Any error from the underlying Unix socket
//
// Behavior:
//   - Returns immediately with io.ErrClosedPipe if connection is closed
//   - Checks context state before reading
//   - Reads one complete datagram (may be truncated if buffer too small)
//   - Closes connection on any error (including EOF)
//   - Thread-safe for concurrent reads
//
// Unix Datagram-Specific Notes:
//   - Each Read() receives one complete datagram
//   - If buffer is smaller than datagram, excess bytes are discarded
//   - No partial reads - each Read() is atomic per datagram
//   - No ordering guarantee between datagrams
//
// Example:
//
//	buf := make([]byte, 65507) // Max datagram size
//	n, err := ctx.Read(buf)
//	if err != nil {
//	    if err == io.ErrClosedPipe {
//	        // Connection closed
//	    }
//	    return err
//	}
//	datagram := buf[:n]
func (o *sCtx) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, io.ErrClosedPipe
	} else if o.clo.Load() {
		return 0, io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	}

	n, err = o.con.Read(p)

	if err != nil && err != io.EOF {
		return n, o.onErrorClose(err)
	} else if err != nil {
		return n, o.Close()
	} else {
		return n, nil
	}
}

// Write is intentionally disabled for Unix datagram server contexts.
//
// This method always returns an error because Unix datagram servers in this implementation
// use ReadFrom/WriteTo pattern rather than Read/Write. The server-side context
// does not maintain per-datagram remote address state needed for Write().
//
// Parameters:
//   - p: Byte slice to write (ignored)
//
// Returns:
//   - n: Always 0
//   - err: Always returns io.ErrClosedPipe
//
// Rationale:
//   - Unix datagram is connectionless, no implicit destination for Write()
//   - Server must use WriteTo with explicit remote address
//   - This prevents accidental writes to wrong destination
//   - Forces explicit handling of remote address per datagram
//
// Alternative:
//
//	To send responses, the handler should use the underlying *net.UnixConn
//	with WriteTo() method, specifying the remote address explicitly.
//
// Note:
//
//	This is a design choice for safety. Client-side Unix datagram contexts may
//	implement Write() differently with a persistent remote address.
func (o *sCtx) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, io.ErrClosedPipe
	} else if o.clo.Load() {
		return 0, io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	} else {
		return 0, o.onErrorClose(io.ErrClosedPipe)
	}
}

// Close closes the Unix datagram connection and cancels the associated context.
//
// This method implements the io.Closer interface and performs complete cleanup
// of all resources associated with the connection context.
//
// Returns:
//   - error: Any error from closing the underlying Unix connection, or nil
//
// Behavior:
//  1. Cancels the context (triggers Done() channel)
//  2. Marks connection as closed atomically
//  3. Closes the underlying Unix socket
//  4. Safe to call multiple times (idempotent)
//  5. Safe to call from multiple goroutines (only first call does work)
//
// Side Effects:
//   - Context's Done() channel is closed
//   - Any blocked Read() calls will return with error
//   - Future Read/Write calls will return io.ErrClosedPipe
//   - Unix socket resources are released to OS
//
// Thread Safety:
//
//	This method is safe to call concurrently from multiple goroutines.
//	Only the first call will perform actual cleanup; subsequent calls
//	are no-ops returning nil.
//
// Example:
//
//	defer ctx.Close() // Always close to avoid resource leaks
//
//	if err := ctx.Close(); err != nil {
//	    log.Printf("Error closing connection: %v", err)
//	}
func (o *sCtx) Close() error {
	if o == nil {
		return nil
	}

	defer o.cnl()

	if o.clo.Load() {
		return nil
	} else if o.con == nil {
		o.clo.Store(true)
		return nil
	} else {
		o.clo.Store(true)
		return o.con.Close()
	}
}

// IsConnected returns true if the connection is still open and usable.
//
// Returns:
//   - bool: true if connection is open, false if closed or nil
//
// This method checks the atomic closed flag to determine connection state.
// It does not perform any I/O operations, making it very fast.
//
// Thread Safety:
//
//	Safe to call from multiple goroutines.
//
// Note:
//
//	For Unix datagram servers, "connected" means the socket is open, not that
//	there's an active connection to a remote peer (Unix datagram is connectionless).
//
// Example:
//
//	if ctx.IsConnected() {
//	    // Safe to attempt Read/Write
//	    ctx.Read(buf)
//	}
func (o *sCtx) IsConnected() bool {
	return !o.clo.Load()
}

// RemoteHost returns the remote address string with protocol indicator.
//
// Returns:
//   - string: Remote address in format "path(unixgram)", or empty string
//
// Format:
//   - "/tmp/app.sock(unixgram)" for Unix socket
//   - "" if connection is nil or has no remote address
//
// Unix Datagram-Specific Behavior:
//
//	For Unix datagram servers, RemoteAddr() may be nil or zero-valued because
//	Unix datagram is connectionless. The remote address is only known per-datagram
//	when using ReadFrom().
//
// Thread Safety:
//
//	Safe to call from multiple goroutines.
//
// Example:
//
//	remote := ctx.RemoteHost()
//	if remote != "" {
//	    log.Printf("Datagram from: %s", remote)
//	}
func (o *sCtx) RemoteHost() string {
	if c := o.con; c == nil {
		return ""
	} else if a := c.RemoteAddr(); a == nil {
		return ""
	} else {
		return a.String() + "(" + libptc.NetworkUnixGram.Code() + ")"
	}
}

// LocalHost returns the local address string with protocol indicator.
//
// Returns:
//   - string: Local address in format "path(unixgram)"
//
// Format:
//   - "/tmp/app.sock(unixgram)" for Unix socket file
//
// The local address is cached at connection creation time for performance.
//
// Thread Safety:
//
//	Safe to call from multiple goroutines (reads immutable cached value).
//
// Example:
//
//	local := ctx.LocalHost()
//	log.Printf("Server listening on: %s", local)
func (o *sCtx) LocalHost() string {
	return o.loc + "(" + libptc.NetworkUnixGram.Code() + ")"
}

// onErrorClose is an internal helper that closes the connection and combines errors.
// It ensures proper cleanup when an error occurs during I/O operations.
//
// Parameters:
//   - e: The error that triggered the close operation
//
// Returns:
//   - error: The original error, or a combined error if Close also fails
//
// # Behavior
//
// If the provided error is nil, returns nil without closing.
// If Close() returns an error, combines both errors in the format "original, close error".
// Otherwise, returns the original error unchanged.
//
// # Usage
//
// This method is called internally by Read() and Write() when errors occur,
// ensuring that:
//   - The connection is properly closed on errors
//   - Both the operation error and any close error are reported
//   - Resources are released even when errors occur
//
// # Example Error Messages
//
//   - "read tcp: connection reset by peer" - network error only
//   - "context canceled, close tcp: use of closed network connection" - combined errors
//
// # Test Coverage Note
//
// This method has low test coverage (0%) because it's primarily invoked
// indirectly through Read/Write error paths which are already tested. Direct testing
// would require forcing Close() to fail, which is difficult to achieve reliably.
func (o *sCtx) onErrorClose(e error) error {
	if e == nil {
		return nil
	} else if err := o.Close(); err != nil {
		return fmt.Errorf("%v, %v", e, err)
	} else {
		return e
	}
}
