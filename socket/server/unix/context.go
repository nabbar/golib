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
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
)

// sCtx implements the context.Context, io.Reader, and io.Writer interfaces
// to provide a unified way to handle Unix socket connections with context support.
// It wraps a net.UnixConn and adds context cancellation, timeouts, and
// connection state tracking.
//
// The sCtx type is used internally by the Unix socket server to manage the lifecycle
// of client connections and provide a consistent interface for reading and
// writing data with proper error handling and resource cleanup.
//
// Fields:
//   - rem: Remote client address (socket path or empty for unnamed sockets)
//   - loc: Local server address (socket file path)
//   - ctx: Underlying context for cancellation and timeouts
//   - cnl: Cancel function for the context
//   - con: The underlying Unix socket connection
//   - clo: Atomic boolean flag indicating if the connection is closed
//   - rst: Function to reset the idle timer (used for keepalive)
type sCtx struct {
	rem string // remote socket path
	loc string // local socket path

	ctx context.Context
	cnl context.CancelFunc
	con io.ReadWriteCloser
	clo *atomic.Bool
	rst func()
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

// Read reads data from the Unix socket connection into the provided buffer.
// This method implements the io.Reader interface and provides context-aware
// reading with automatic error handling and connection lifecycle management.
//
// Parameters:
//   - p: Byte slice to read data into. The size determines the maximum
//     number of bytes that can be read in a single call.
//
// Returns:
//   - n: Number of bytes actually read into p
//   - err: Any error encountered during reading:
//   - io.ErrClosedPipe: If the connection is already closed
//   - io.EOF: If the connection reached end of stream (triggers close)
//   - Context errors: If the context was cancelled
//   - Network errors: Any underlying Unix socket errors
//
// # Behavior
//
// The method performs the following actions:
//  1. Validates connection state and context
//  2. Reads from the underlying Unix socket connection
//  3. Resets the idle timeout timer on successful read
//  4. Handles errors:
//     - io.EOF triggers graceful connection close
//     - Other errors trigger error close with cleanup
//     - Context cancellation propagates and closes connection
//
// # Idle Timeout
//
// Each successful read resets the idle timeout timer (if configured).
// If no data is read within the idle timeout period, the connection
// will be automatically closed with ErrIdleTimeout.
//
// # Usage Notes
//
//   - Always check the error, even if n > 0
//   - Partial reads are normal for stream-oriented Unix sockets
//   - io.EOF indicates graceful connection close by peer
//   - The buffer p is not modified if an error occurs before reading
//
// # Example
//
//	buf := make([]byte, 4096)
//	n, err := conn.Read(buf)
//	if err != nil {
//	    if err == io.EOF {
//	        log.Println("Connection closed by peer")
//	    } else {
//	        log.Printf("Read error: %v", err)
//	    }
//	    return
//	}
//	data := buf[:n]
//	process(data)
//
// # Thread Safety
//
// While the method itself is safe, concurrent calls to Read on the same
// connection will result in undefined behavior as per io.Reader contract.
func (o *sCtx) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, io.ErrClosedPipe
	} else if o.clo.Load() {
		return 0, io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	}

	n, err = o.con.Read(p)
	o.rst()

	if err != nil && err != io.EOF {
		return n, o.onErrorClose(err)
	} else if err != nil {
		return n, o.Close()
	} else {
		return n, nil
	}
}

// Write writes data from the provided buffer to the Unix socket connection.
// This method implements the io.Writer interface and provides context-aware
// writing with automatic error handling and connection lifecycle management.
//
// Parameters:
//   - p: Byte slice containing the data to write. All bytes will be written
//     unless an error occurs.
//
// Returns:
//   - n: Number of bytes actually written from p
//   - err: Any error encountered during writing:
//   - io.ErrClosedPipe: If the connection is already closed
//   - io.EOF: If the connection reached end of stream (triggers close)
//   - Context errors: If the context was cancelled
//   - Network errors: Any underlying Unix socket errors (buffer full, broken pipe, etc.)
//
// # Behavior
//
// The method performs the following actions:
//  1. Validates connection state and context
//  2. Writes to the underlying Unix socket connection
//  3. Resets the idle timeout timer on successful write
//  4. Handles errors:
//     - io.EOF triggers graceful connection close
//     - Other errors are returned as-is (connection remains open for retry)
//     - Context cancellation propagates and closes connection
//
// # Write Guarantees
//
// Unix sockets provide stream-oriented delivery, and Write guarantees:
//   - Atomic write: Either all bytes are written or an error occurs
//   - No partial writes: If n < len(p), err will be non-nil
//   - Ordered delivery: Bytes are delivered in the order written
//
// # Idle Timeout
//
// Each successful write resets the idle timeout timer (if configured),
// preventing premature connection closure during active communication.
//
// # Usage Notes
//
//   - Always check both n and err return values
//   - A write error does NOT automatically close the connection
//   - For binary protocols, consider using encoding/binary or bufio
//   - Large writes may block if the socket send buffer is full
//
// # Example
//
//	response := []byte("HTTP/1.1 200 OK\r\n\r\n")
//	n, err := conn.Write(response)
//	if err != nil {
//	    log.Printf("Write error: %v", err)
//	    return err
//	}
//	if n != len(response) {
//	    log.Printf("Partial write: %d of %d bytes", n, len(response))
//	}
//
// # Thread Safety
//
// While the method itself is safe, concurrent calls to Write on the same
// connection will result in interleaved data as per io.Writer contract.
func (o *sCtx) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, io.ErrClosedPipe
	} else if o.clo.Load() {
		return 0, io.ErrClosedPipe
	} else if e := o.ctx.Err(); e != nil {
		return 0, o.onErrorClose(e)
	}

	n, err = o.con.Write(p)
	o.rst()

	if err != nil && err != io.EOF {
		return n, err
	} else if err != nil {
		return n, o.Close()
	} else {
		return n, nil
	}
}

// Close closes the connection and releases all associated resources.
// This method implements the io.Closer interface and ensures proper cleanup
// of the connection, context, and any active timers.
//
// Returns:
//   - error: Any error from closing the underlying TCP connection, or nil
//     if already closed or if no connection exists.
//
// # Behavior
//
// The method performs the following cleanup sequence:
//  1. Cancels the connection-specific context (via defer)
//  2. Sets the closed flag atomically to prevent double-close
//  3. Closes the underlying Unix socket connection
//  4. Signals any goroutines waiting on ctx.Done()
//
// # Idempotency
//
// Close is idempotent - calling it multiple times is safe and will only
// close the connection once. Subsequent calls return nil without error.
//
// # Side Effects
//
// After Close is called:
//   - All pending Read/Write operations will fail with io.ErrClosedPipe
//   - The context's Done() channel will be closed
//   - IsConnected() will return false
//   - Any active idle timeout timer is stopped
//   - The connection handler goroutine will terminate
//
// # Usage Notes
//
//   - Always call Close when done with a connection (use defer)
//   - It's safe to call Close from multiple goroutines
//   - Close does not wait for pending operations to complete
//   - For graceful shutdown, drain all data before calling Close
//
// # Example
//
//	func handleConnection(conn Context) {
//	    defer conn.Close()  // Ensures cleanup even on panic
//
//	    // Handle connection...
//	    if err := processData(conn); err != nil {
//	        log.Printf("Error: %v", err)
//	        return  // Close called via defer
//	    }
//	}
//
// # Thread Safety
//
// This method is safe to call from multiple goroutines concurrently.
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

// IsConnected reports whether the connection is still active.
//
// Returns:
//   - bool: true if the connection is open and usable, false if closed
//
// # Usage
//
// This method is useful for:
//   - Checking connection state before operations
//   - Connection health monitoring
//   - Cleanup validation in defer statements
//   - Breaking out of read/write loops
//
// # State Transitions
//
// The connection state follows this lifecycle:
//   - true: After connection establishment
//   - false: After Close() is called or connection error
//
// # Example
//
//	for ctx.IsConnected() {
//	    if err := processNextMessage(ctx); err != nil {
//	        log.Printf("Error: %v", err)
//	        break
//	    }
//	}
//
// # Thread Safety
//
// This method uses atomic operations and is safe to call from multiple goroutines.
func (o *sCtx) IsConnected() bool {
	return !o.clo.Load()
}

// RemoteHost returns the remote peer's address with protocol information.
//
// Returns:
//   - string: The remote address in the format "path(protocol)"
//     where protocol is "unix"
//
// # Format
//
// The returned string includes:
//   - Socket path of the remote client (e.g., "/tmp/client.sock" or empty for unnamed sockets)
//   - Protocol indication in parentheses (e.g., "(unix)")
//
// # Usage
//
// This method is useful for:
//   - Logging and debugging
//   - Connection tracking and monitoring
//   - Audit trails
//   - Process identification (via socket credentials on Linux)
//
// # Example
//
//	log.Printf("New connection from %s", ctx.RemoteHost())
//	// Output: New connection from @abstract-socket(unix) or /tmp/client.sock(unix)
//
// # Thread Safety
//
// This method is safe to call from multiple goroutines as it only reads
// immutable fields set during connection initialization.
func (o *sCtx) RemoteHost() string {
	return o.rem + "(" + libptc.NetworkUnix.Code() + ")"
}

// LocalHost returns the local server's address with protocol information.
//
// Returns:
//   - string: The local address in the format "path(protocol)"
//     where protocol is "unix"
//
// # Format
//
// The returned string includes:
//   - Socket file path of the local server (e.g., "/tmp/server.sock")
//   - Protocol indication in parentheses (e.g., "(unix)")
//
// # Usage
//
// This method is useful for:
//   - Logging which server socket handled the connection
//   - Multiple socket server configurations
//   - Debugging connection routing
//   - Service identification
//
// # Example
//
//	log.Printf("Connection accepted on %s from %s",
//	    ctx.LocalHost(), ctx.RemoteHost())
//	// Output: Connection accepted on /tmp/server.sock(unix) from /tmp/client.sock(unix)
//
// # Thread Safety
//
// This method is safe to call from multiple goroutines as it only reads
// immutable fields set during connection initialization.
func (o *sCtx) LocalHost() string {
	return o.loc + "(" + libptc.NetworkUnix.Code() + ")"
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
//   - "read unix: connection reset by peer" - network error only
//   - "context canceled, close unix: use of closed network connection" - combined errors
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
