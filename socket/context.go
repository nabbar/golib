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

package socket

import (
	"context"
	"io"
)

// Context extends context.Context with I/O operations and connection state for socket communication.
// It combines the standard context interface with Reader, Writer, and Closer interfaces, plus
// connection-specific methods for state checking and address retrieval.
//
// The Context interface is passed to HandlerFunc in server implementations and provides all
// necessary operations for handling a connection within a single abstraction. This design
// allows handlers to access the underlying connection context, perform I/O operations, and
// query connection state without needing multiple parameters.
//
// # Usage
//
// Server-side handler example:
//
//	func handleConnection(ctx socket.Context) {
//	    // Check connection state
//	    if !ctx.IsConnected() {
//	        return
//	    }
//
//	    // Log connection information
//	    log.Printf("Handling connection from %s to %s", ctx.RemoteHost(), ctx.LocalHost())
//
//	    // Read from connection
//	    buf := make([]byte, 1024)
//	    n, err := ctx.Read(buf)
//	    if err != nil {
//	        return
//	    }
//
//	    // Write response
//	    response := []byte("Response: " + string(buf[:n]))
//	    ctx.Write(response)
//
//	    // Check context cancellation
//	    select {
//	    case <-ctx.Done():
//	        log.Println("Context canceled")
//	        return
//	    default:
//	    }
//	}
//
// # Interface Composition
//
// The Context interface embeds multiple standard interfaces:
//
//   - context.Context: Provides deadline, cancellation, and value propagation
//   - io.Reader: Enables reading data from the connection
//   - io.Writer: Enables writing data to the connection
//   - io.Closer: Provides resource cleanup
//
// Plus custom methods for connection-specific operations:
//
//   - IsConnected(): Check if connection is still active
//   - RemoteHost(): Get remote endpoint address
//   - LocalHost(): Get local endpoint address
//
// # Thread Safety
//
// Thread safety depends on the specific implementation:
//
//   - Read/Write: Follow io.Reader/Writer contracts (not safe for concurrent calls on same connection)
//   - IsConnected: Safe for concurrent calls (uses atomic operations)
//   - RemoteHost/LocalHost: Safe for concurrent calls (immutable after creation)
//   - context.Context methods: Safe for concurrent calls (immutable)
//
// # Connection Lifecycle
//
//	New Connection → IsConnected() = true
//	                      ↓
//	             Read/Write Operations
//	                      ↓
//	                  Close()
//	                      ↓
//	              IsConnected() = false
//
// # Implementation Notes
//
// Implementations are provided in protocol-specific sub-packages:
//   - TCP: github.com/nabbar/golib/socket/server/tcp.sCtx
//   - UDP: github.com/nabbar/golib/socket/server/udp.sCtx
//   - Unix: github.com/nabbar/golib/socket/server/unix.sCtx
//   - UnixGram: github.com/nabbar/golib/socket/server/unixgram.sCtx
//
// Each implementation adapts the base connection type (net.Conn, net.PacketConn) to
// this unified interface while preserving protocol-specific characteristics.
type Context interface {
	// context.Context provides deadline, cancellation signal propagation, and request-scoped values.
	// This allows handlers to respect timeouts and cancellation from the parent context.
	context.Context

	// io.Reader enables reading data from the connection.
	// Read blocks until data is available or an error occurs.
	//
	// For stream protocols (TCP, Unix): Reads bytes from the stream.
	// For datagram protocols (UDP, UnixGram): Reads one complete datagram.
	//
	// Returns:
	//   - n: Number of bytes read
	//   - err: io.EOF on connection close, or other I/O errors
	io.Reader

	// io.Writer enables writing data to the connection.
	// Write blocks until data is written or an error occurs.
	//
	// For stream protocols (TCP, Unix): Writes bytes to the stream.
	// For datagram protocols (UDP, UnixGram): Sends one complete datagram.
	//
	// Returns:
	//   - n: Number of bytes written
	//   - err: io.ErrClosedPipe if connection closed, or other I/O errors
	io.Writer

	// io.Closer enables closing the connection and releasing resources.
	// After Close is called, subsequent Read/Write operations will fail.
	//
	// Close is idempotent and safe to call multiple times.
	//
	// Returns:
	//   - err: Error if cleanup fails, nil otherwise
	io.Closer

	// IsConnected returns true if the connection is active and usable.
	// Returns false after Close() is called or if the connection is broken.
	//
	// This method is safe for concurrent calls and uses atomic operations.
	//
	// Example:
	//
	//	if ctx.IsConnected() {
	//	    ctx.Write(data)
	//	}
	IsConnected() bool

	// RemoteHost returns the remote endpoint address as a string.
	// The format depends on the protocol:
	//   - TCP/UDP: "host:port" (e.g., "192.168.1.10:54321")
	//   - Unix: "socket-path" (e.g., "/tmp/app.sock")
	//
	// The returned string includes the protocol code for disambiguation.
	//
	// This method is safe for concurrent calls as it returns an immutable value
	// set during connection initialization.
	//
	// Example:
	//
	//	remote := ctx.RemoteHost()
	//	log.Printf("Connection from: %s", remote)
	RemoteHost() string

	// LocalHost returns the local endpoint address as a string.
	// The format depends on the protocol:
	//   - TCP/UDP: "host:port" (e.g., "0.0.0.0:8080")
	//   - Unix: "socket-path" (e.g., "/tmp/app.sock")
	//
	// The returned string includes the protocol code for disambiguation.
	//
	// This method is safe for concurrent calls as it returns an immutable value
	// set during connection initialization.
	//
	// Example:
	//
	//	local := ctx.LocalHost()
	//	log.Printf("Server listening on: %s", local)
	LocalHost() string
}
