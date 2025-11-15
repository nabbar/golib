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
	"errors"
	"io"
	"net"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

// Internal atomic map keys for storing client state.
// Using uint8 keys for efficient memory usage and fast lookups.
const (
	keyNetAddr uint8 = iota // UNIX socket file path (string)
	keyFctErr               // Error callback function (libsck.FuncError)
	keyFctInfo              // Info callback function (libsck.FuncInfo)
	keyNetConn              // Active UNIX socket connection (net.Conn)
)

// cli is the internal implementation of ClientUnix interface.
// It uses an atomic map to store all client state in a thread-safe manner,
// matching the architecture of the TCP/UDP clients for consistency.
//
// State management:
//   - All state is stored in a thread-safe atomic.Map[uint8]
//   - Keys can be safely deleted to represent "no value" without storing nil
//   - Multiple goroutines can safely call methods concurrently
//   - State changes trigger registered callbacks asynchronously
//
// UNIX socket-specific notes:
//   - Connection-oriented like TCP
//   - Uses filesystem path instead of network address
//   - No network overhead - kernel-space only
type cli struct {
	m libatm.Map[uint8] // Atomic map storing all client state by key
}

// SetTLS is a no-op for UNIX socket clients as UNIX sockets don't support TLS.
// UNIX domain sockets are local-only and use filesystem permissions for access control.
// For secure communication over UNIX sockets, consider:
//   - Filesystem permissions (chmod, chown)
//   - SELinux/AppArmor policies
//   - Application-level encryption
//
// This method always returns nil to maintain interface compatibility with
// github.com/nabbar/golib/socket.Client.
//
// Parameters are ignored:
//   - enable: Ignored
//   - config: Ignored
//   - serverName: Ignored
//
// Returns:
//   - nil: Always, as UNIX sockets don't support TLS
func (o *cli) SetTLS(enable bool, config libtls.TLSConfig, serverName string) error {
	return nil
}

// RegisterFuncError registers a callback function for error notifications.
//
// The callback is invoked asynchronously (in a separate goroutine) whenever
// an error occurs during client operations, including:
//   - Connection failures (socket not found, permission denied)
//   - Read/Write errors (connection closed, broken pipe)
//   - Socket path errors (invalid path, too long)
//
// Pass nil to unregister the current error callback. Only one error callback
// can be registered at a time; calling this method replaces any existing callback.
//
// The callback is executed asynchronously to avoid blocking socket I/O operations.
// Ensure your callback handles errors appropriately and returns quickly.
//
// Example:
//
//	client := unix.New("/tmp/app.sock")
//	client.RegisterFuncError(func(errs ...error) {
//	    for _, err := range errs {
//	        log.Printf("UNIX socket error: %v", err)
//	    }
//	})
//
// See github.com/nabbar/golib/socket.FuncError for callback signature details.
func (o *cli) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	if f == nil {
		o.m.Delete(keyFctErr)
	} else {
		o.m.Store(keyFctErr, f)
	}
}

// RegisterFuncInfo registers a callback function for connection operation notifications.
//
// The callback is invoked asynchronously (in a separate goroutine) for various
// connection operations:
//   - ConnectionDial: When socket connection starts
//   - ConnectionNew: When socket is connected and ready
//   - ConnectionRead: Before each read operation
//   - ConnectionWrite: Before each write operation
//   - ConnectionClose: When socket is closed
//
// The callback receives:
//   - local: Local socket address (typically empty addr for UNIX sockets)
//   - remote: Remote socket path as UnixAddr
//   - state: Operation state from github.com/nabbar/golib/socket.ConnState
//
// Pass nil to unregister the current info callback. Only one info callback
// can be registered at a time; calling this method replaces any existing callback.
//
// The callback is executed asynchronously to avoid blocking socket I/O.
// Keep callback execution time minimal to avoid impacting performance.
//
// Example:
//
//	client := unix.New("/tmp/app.sock")
//	client.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
//	    log.Printf("UNIX socket state: %v (socket: %v)", state, remote)
//	})
//
// See github.com/nabbar/golib/socket.FuncInfo and ConnState for details.
func (o *cli) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	if f == nil {
		o.m.Delete(keyFctInfo)
	} else {
		o.m.Store(keyFctInfo, f)
	}
}

// fctError invokes the registered error callback if present.
// This internal method is called whenever an error occurs during client operations.
// The callback is executed in a separate goroutine to avoid blocking.
// If no callback is registered or the error is nil, this method does nothing.
func (o *cli) fctError(e error) {
	if o == nil || e == nil {
		return
	}

	if v, k := o.m.Load(keyFctErr); k && v != nil {
		if fn, ok := v.(libsck.FuncError); ok && fn != nil {
			go fn(e)
		}
	}
}

// fctInfo invokes the registered info callback if present.
// This internal method is called for connection operation state changes.
// The callback is executed in a separate goroutine to avoid blocking I/O operations.
// If no callback is registered, this method does nothing.
func (o *cli) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	if v, k := o.m.Load(keyFctInfo); k && v != nil {
		if fn, ok := v.(libsck.FuncInfo); ok && fn != nil {
			go fn(local, remote, state)
		}
	}
}

// dial creates the UNIX socket connection to the server.
// This internal method uses net.Dialer to establish the connection.
//
// The socket path must exist and be accessible. Common errors:
//   - "no such file or directory": Socket file doesn't exist (server not running)
//   - "permission denied": Insufficient permissions to access socket
//   - "connection refused": Server not listening (rare for UNIX sockets)
//
// Returns:
//   - net.Conn: The UNIX socket connection (actually *net.UnixConn)
//   - error: ErrInstance if client is nil, ErrAddress if path is invalid,
//     or a network error if connection fails
func (o *cli) dial(ctx context.Context) (net.Conn, error) {
	if o == nil {
		return nil, ErrInstance
	}

	if v, k := o.m.Load(keyNetAddr); !k || v == nil {
		return nil, ErrAddress
	} else if adr, ok := v.(string); !ok {
		return nil, ErrAddress
	} else {
		d := net.Dialer{}
		return d.DialContext(ctx, libptc.NetworkUnix.Code(), adr)
	}
}

// IsConnected checks if the client has an active UNIX socket connection.
//
// This method checks the connection state by:
//  1. Verifying a connection object exists
//  2. Attempting a zero-byte write to test connection liveness
//
// Important notes:
//   - Returns true if connected and socket is alive
//   - Returns false if never connected or after Close() was called
//   - Also returns false if connection was broken (server closed, network issue)
//   - The liveness check (Write(nil)) is lightweight and doesn't send data
//   - Thread-safe and can be called from multiple goroutines concurrently
//
// Unlike network sockets, UNIX sockets are local and typically very reliable.
// However, the server may still close the connection unexpectedly.
//
// Example:
//
//	if client.IsConnected() {
//	    // Socket is ready for I/O operations
//	    _, err := client.Write([]byte("data"))
//	    if err != nil {
//	        log.Printf("Write failed: %v", err)
//	    }
//	}
func (o *cli) IsConnected() bool {
	if o == nil {
		return false
	}

	if i, k := o.m.Load(keyNetConn); !k || i == nil {
		return false
	} else if c, k := i.(net.Conn); !k || c == nil {
		return false
	} else if _, e := c.Write(nil); e != nil {
		return false
	}

	return true
}

// Connect establishes a connection to the UNIX domain socket.
//
// This method creates a connection to the socket file specified in New().
// The socket file must exist and be accessible (created by the server).
//
// The context parameter controls timeouts and cancellation:
//   - Use context.WithTimeout() to set a connection timeout
//   - Use context.WithCancel() to allow cancelling the operation
//   - The context is only used during connection establishment
//
// Operation states:
//   - Triggers ConnectionDial callback when connection attempt starts
//   - Triggers ConnectionNew callback when connection is established
//   - Triggers error callback if connection fails
//
// If a connection already exists, it is replaced. The old connection is closed
// automatically.
//
// Common errors:
//   - "no such file or directory": Server not running or socket not created
//   - "permission denied": Check file permissions on socket
//   - "connection refused": Server not accepting connections (rare)
//   - Context timeout: Server took too long to accept connection
//
// Returns:
//   - nil: Connection established successfully
//   - ErrInstance: If client is nil
//   - network error: If connection fails
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err := client.Connect(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
// Thread-safe: Can be called concurrently, but only one connection is active at a time.
func (o *cli) Connect(ctx context.Context) error {
	if o == nil {
		return ErrInstance
	}

	var (
		err error
		con net.Conn
	)

	o.fctInfo(&net.UnixAddr{}, &net.UnixAddr{}, libsck.ConnectionDial)
	if con, err = o.dial(ctx); err != nil {
		o.fctError(err)
		return err
	}

	o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionNew)

	// Replace old connection if exists
	if i, k := o.m.Swap(keyNetConn, con); k && i != nil {
		if c, k := i.(net.Conn); k && c != nil {
			_ = c.Close() // Close old connection
		}
	}

	return nil
}

// Read reads data from the UNIX socket into the provided buffer.
//
// This method implements io.Reader interface and reads data from the underlying
// UNIX socket connection. It blocks until data is available or an error occurs.
//
// Parameters:
//   - p: Buffer to read data into. The buffer should be large enough to
//     hold the expected data. Common sizes: 4096 or 8192 bytes.
//
// Returns:
//   - n: Number of bytes read (0 to len(p))
//   - err: nil on success, or an error:
//   - ErrInstance if client is nil
//   - ErrConnection if not connected (call Connect() first)
//   - io.EOF if connection closed by remote peer
//   - network error for other failures
//
// Behavior:
//   - Blocks until data arrives or error occurs
//   - May return fewer bytes than len(p) if less data is available
//   - Connection-oriented: preserves byte stream order
//   - Thread-safe but don't call concurrently on same client
//     (underlying socket is not safe for concurrent reads)
//
// UNIX socket advantages:
//   - Lower latency than TCP (no network stack)
//   - Higher throughput for local communication
//   - No packet fragmentation issues
//
// Example:
//
//	buf := make([]byte, 4096)
//	n, err := client.Read(buf)
//	if err != nil {
//	    if err == io.EOF {
//	        log.Println("Connection closed")
//	    } else {
//	        log.Printf("Read error: %v", err)
//	    }
//	    return
//	}
//	log.Printf("Received %d bytes: %s", n, buf[:n])
func (o *cli) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInstance
	}

	if i, k := o.m.Load(keyNetConn); !k || i == nil {
		err = ErrConnection
		o.fctError(err)
		return 0, err
	} else if c, k := i.(net.Conn); !k || c == nil {
		err = ErrConnection
		o.fctError(err)
		return 0, err
	} else {
		o.fctInfo(c.LocalAddr(), c.RemoteAddr(), libsck.ConnectionRead)
		n, err = c.Read(p)
		if err != nil {
			o.fctError(err)
		}
		return n, err
	}
}

// Write sends data to the UNIX socket.
//
// This method implements io.Writer interface and sends data to the remote
// peer through the UNIX socket connection.
//
// Parameters:
//   - p: Buffer containing data to send. All len(p) bytes will be sent.
//
// Returns:
//   - n: Number of bytes sent (0 to len(p))
//   - err: nil if all data sent successfully, or an error:
//   - ErrInstance if client is nil
//   - ErrConnection if not connected (call Connect() first)
//   - EPIPE (broken pipe) if connection closed by remote peer
//   - network error for other failures
//
// Behavior:
//   - Blocks until all data is sent or error occurs
//   - Atomic: either all data is sent or an error is returned
//   - Connection-oriented: preserves byte stream order
//   - Thread-safe but don't call concurrently on same client
//     (underlying socket is not safe for concurrent writes)
//
// UNIX socket advantages:
//   - No size limits like UDP (no 64KB datagram limit)
//   - No MTU fragmentation issues
//   - Kernel buffering for high throughput
//   - Local communication optimizations
//
// Example:
//
//	data := []byte("Hello, UNIX socket!")
//	n, err := client.Write(data)
//	if err != nil {
//	    log.Printf("Write error: %v", err)
//	    return
//	}
//	if n != len(data) {
//	    log.Printf("Partial write: %d of %d bytes", n, len(data))
//	}
func (o *cli) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInstance
	}

	if i, k := o.m.Load(keyNetConn); !k || i == nil {
		err = ErrConnection
		o.fctError(err)
		return 0, err
	} else if c, k := i.(net.Conn); !k || c == nil {
		err = ErrConnection
		o.fctError(err)
		return 0, err
	} else {
		o.fctInfo(c.LocalAddr(), c.RemoteAddr(), libsck.ConnectionWrite)
		n, err = c.Write(p)
		if err != nil {
			o.fctError(err)
		}
		return n, err
	}
}

// Close closes the UNIX socket connection and releases associated resources.
//
// This method closes the underlying socket and removes it from the client's
// state. After calling Close(), IsConnected() will return false and any
// subsequent Read() or Write() calls will return ErrConnection.
//
// Behavior:
//   - Triggers ConnectionClose callback before closing
//   - Closes the underlying socket
//   - Removes connection from client state atomically
//   - Safe to call multiple times (subsequent calls return ErrConnection)
//   - Thread-safe: Can be called concurrently with other operations
//   - Does NOT remove the socket file (server's responsibility)
//
// Returns:
//   - nil: Socket closed successfully
//   - ErrInstance: If client is nil
//   - ErrConnection: If no connection exists (already closed or never connected)
//   - network error: If underlying Close() fails (rare)
//
// Best practice: Always defer Close() after successful Connect() to ensure
// proper cleanup even if errors occur.
//
// Example:
//
//	client := unix.New("/tmp/app.sock")
//	if client == nil {
//	    log.Fatal("Invalid socket path")
//	}
//
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close() // Ensure cleanup
//
//	// ... use client ...
func (o *cli) Close() error {
	if o == nil {
		return ErrInstance
	}

	// Use LoadAndDelete to atomically remove the connection
	if i, k := o.m.LoadAndDelete(keyNetConn); !k || i == nil {
		return ErrConnection
	} else if c, k := i.(net.Conn); !k || c == nil {
		return ErrConnection
	} else {
		o.fctInfo(c.LocalAddr(), c.RemoteAddr(), libsck.ConnectionClose)
		return c.Close()
	}
}

// Once performs a one-shot request/response operation.
//
// This method is a convenience function that:
//  1. Connects to the UNIX socket
//  2. Sends all data from the request reader
//  3. Invokes the response callback to handle received data (if provided)
//  4. Closes the socket automatically
//
// This is useful for simple request/response patterns where you don't need
// to maintain a persistent connection.
//
// Parameters:
//   - ctx: Context for connection timeout and cancellation
//   - request: Reader containing the data to send. Data is read until EOF.
//   - fct: Callback function to handle responses. Receives the client as io.Reader.
//     Can be nil if no response is expected (fire-and-forget).
//
// The response callback receives the client itself, allowing you to read
// response data using client.Read().
//
// Behavior:
//   - Automatically connects if not connected
//   - Reads all data from request until io.EOF
//   - Calls response callback (if provided)
//   - Automatically closes socket via defer
//   - Triggers all appropriate state callbacks
//   - Error callback triggered for any errors
//
// Returns:
//   - nil: Operation completed successfully
//   - ErrInstance: If client is nil
//   - Connection or I/O errors if operation fails
//
// UNIX socket benefits for Once():
//   - Fast connection establishment (local only)
//   - No TCP handshake overhead
//   - Ideal for quick commands to daemons
//
// Example:
//
//	request := bytes.NewBufferString("GET /status")
//	err := client.Once(ctx, request, func(reader io.Reader) {
//	    response, _ := io.ReadAll(reader)
//	    fmt.Printf("Response: %s\n", response)
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Socket automatically closed
//
// See github.com/nabbar/golib/socket.Response for callback signature details.
func (o *cli) Once(ctx context.Context, request io.Reader, fct libsck.Response) error {
	if o == nil {
		return ErrInstance
	}

	defer func() {
		if err := o.Close(); err != nil {
			o.fctError(err)
		}
	}()

	var (
		err error
		nbr int64
	)

	if err = o.Connect(ctx); err != nil {
		o.fctError(err)
		return err
	}

	for {
		nbr, err = io.Copy(o, request)

		if err != nil {
			if !errors.Is(err, io.EOF) {
				o.fctError(err)
				return err
			} else {
				break
			}
		} else if nbr < 1 {
			break
		}
	}

	if fct != nil {
		fct(o)
	}

	return nil
}
