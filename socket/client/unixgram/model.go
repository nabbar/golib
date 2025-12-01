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

package unixgram

import (
	"context"
	"errors"
	"io"
	"net"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	librun "github.com/nabbar/golib/runner"
	libsck "github.com/nabbar/golib/socket"
)

// Internal atomic map keys for storing client state.
// Using uint8 keys for efficient memory usage and fast lookups.
const (
	keyNetAddr uint8 = iota // UNIX socket file path (string)
	keyFctErr               // Error callback function (libsck.FuncError)
	keyFctInfo              // Info callback function (libsck.FuncInfo)
	keyNetConn              // Active UNIX datagram connection (net.Conn)
)

// cli is the internal implementation of ClientUnix interface.
// It uses an atomic map to store all client state in a thread-safe manner,
// matching the architecture of the TCP/UDP/UNIX clients for consistency.
//
// State management:
//   - All state is stored in a thread-safe atomic.Map[uint8]
//   - Keys can be safely deleted to represent "no value" without storing nil
//   - Multiple goroutines can safely call methods concurrently
//   - State changes trigger registered callbacks asynchronously
//
// UNIX datagram-specific notes:
//   - Connectionless like UDP
//   - Uses filesystem path instead of network address
//   - No network overhead - kernel-space only
//   - Datagrams preserve message boundaries
type cli struct {
	m libatm.Map[uint8] // Atomic map storing all client state by key
}

// SetTLS is a no-op for UNIX datagram clients as UNIX sockets don't support TLS.
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
//   - Socket creation failures
//   - Datagram send/receive errors
//   - Socket path errors (invalid path, too long)
//
// Pass nil to unregister the current error callback. Only one error callback
// can be registered at a time; calling this method replaces any existing callback.
//
// The callback is executed asynchronously to avoid blocking datagram operations.
// Ensure your callback handles errors appropriately and returns quickly.
//
// Example:
//
//	client := unixgram.New("/tmp/app.sock")
//	client.RegisterFuncError(func(errs ...error) {
//	    for _, err := range errs {
//	        log.Printf("UNIX datagram error: %v", err)
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

// RegisterFuncInfo registers a callback function for datagram operation notifications.
//
// The callback is invoked asynchronously (in a separate goroutine) for various
// datagram operations:
//   - ConnectionDial: When socket creation starts
//   - ConnectionNew: When socket is associated with remote path
//   - ConnectionRead: Before each datagram read
//   - ConnectionWrite: Before each datagram write
//   - ConnectionClose: When socket is closed
//
// The callback receives:
//   - local: Local socket address (typically empty addr for datagram sockets)
//   - remote: Remote socket path as UnixAddr
//   - state: Operation state from github.com/nabbar/golib/socket.ConnState
//
// Pass nil to unregister the current info callback. Only one info callback
// can be registered at a time; calling this method replaces any existing callback.
//
// The callback is executed asynchronously to avoid blocking datagram I/O.
// Keep callback execution time minimal to avoid impacting performance.
//
// Example:
//
//	client := unixgram.New("/tmp/app.sock")
//	client.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
//	    log.Printf("Datagram state: %v (socket: %v)", state, remote)
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
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/client/unixgram/fctError", r)
		}
	}()

	if o == nil || e == nil {
		return
	}

	if v, k := o.m.Load(keyFctErr); k && v != nil {
		if fn, ok := v.(libsck.FuncError); ok && fn != nil {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						librun.RecoveryCaller("golib/socket/client/unixgram/fctError", r, e)
					}
				}()

				fn(e)
			}()
		}
	}
}

// fctInfo invokes the registered info callback if present.
// This internal method is called for datagram operation state changes.
// The callback is executed in a separate goroutine to avoid blocking I/O operations.
// If no callback is registered, this method does nothing.
func (o *cli) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/client/unixgram/fctError", r)
		}
	}()

	if o == nil {
		return
	}

	if v, k := o.m.Load(keyFctInfo); k && v != nil {
		if fn, ok := v.(libsck.FuncInfo); ok && fn != nil {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						librun.RecoveryCaller("golib/socket/client/unixgram/fctInfo", r)
					}
				}()

				fn(local, remote, state)
			}()
		}
	}
}

// dial creates the UNIX datagram socket and associates it with the remote path.
// This internal method uses net.Dialer to create a UNIX datagram connection.
//
// Important: For UNIX datagrams, "dial" doesn't establish a connection in the
// traditional sense (it's connectionless like UDP). It creates a socket and
// associates it with the remote path, allowing subsequent Write() calls to send
// datagrams to that path without specifying the destination each time.
//
// Returns:
//   - net.Conn: The UNIX datagram socket (actually *net.UnixConn)
//   - error: ErrInstance if client is nil, ErrAddress if path is invalid,
//     or a network error if socket creation fails
func (o *cli) dial(ctx context.Context) (net.Conn, error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/client/unixgram/dial", r)
		}
	}()

	if o == nil {
		return nil, ErrInstance
	}

	if v, k := o.m.Load(keyNetAddr); !k || v == nil {
		return nil, ErrAddress
	} else if adr, ok := v.(string); !ok {
		return nil, ErrAddress
	} else {
		d := net.Dialer{}
		return d.DialContext(ctx, libptc.NetworkUnixGram.Code(), adr)
	}
}

// IsConnected checks if the client has an associated UNIX datagram socket.
//
// This method checks the local socket state by verifying that a socket
// object exists in the client's state. For UNIX datagrams, this doesn't mean
// there's an actual connection (datagram sockets are connectionless), but rather
// that a socket has been created and associated with a remote path.
//
// Important notes:
//   - Returns true if a socket was created and not yet closed
//   - Returns false if never connected or after Close() was called
//   - UNIX datagram is connectionless, so "connected" means "socket associated"
//   - Doesn't verify if the remote socket exists or is reachable
//   - Doesn't verify if datagrams can actually be sent/received
//
// Thread-safe and can be called from multiple goroutines concurrently.
//
// Example:
//
//	if client.IsConnected() {
//	    // Socket is ready for datagram operations
//	    _, err := client.Write([]byte("event data"))
//	    if err != nil {
//	        log.Printf("Send failed: %v", err)
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
	}

	return true
}

// Connect creates a UNIX datagram socket and associates it with the remote path.
//
// For UNIX datagrams, this method doesn't establish a connection in the traditional
// sense (datagram sockets are connectionless). Instead, it:
//  1. Creates a UNIX datagram socket
//  2. Associates it with the remote path specified in New()
//  3. Allows subsequent Write() calls without specifying destination
//
// The context parameter controls timeouts and cancellation:
//   - Use context.WithTimeout() to set a socket creation timeout
//   - Use context.WithCancel() to allow cancelling the operation
//   - The context is only used during socket creation
//
// Operation states:
//   - Triggers ConnectionDial callback when socket creation starts
//   - Triggers ConnectionNew callback when socket is ready
//   - Triggers error callback if operation fails
//
// If a socket already exists, it is replaced. The old socket is closed
// automatically.
//
// Returns:
//   - nil: Socket created and associated successfully
//   - ErrInstance: If client is nil
//   - network error: If socket creation fails
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
// Thread-safe: Can be called concurrently, but only one socket is active at a time.
func (o *cli) Connect(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/client/unixgram/connect", r)
		}
	}()

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

// Read reads one complete datagram from the UNIX socket into the provided buffer.
//
// This method implements io.Reader interface and reads data from the underlying
// UNIX datagram socket. Each Read() call receives exactly one complete datagram.
//
// Parameters:
//   - p: Buffer to read datagram into. The buffer should be large enough to
//     hold the complete datagram. System-dependent max size (typically 16KB-64KB).
//
// Returns:
//   - n: Number of bytes in the received datagram (0 to len(p))
//   - err: nil on success, or an error:
//   - ErrInstance if client is nil
//   - ErrConnection if not connected (call Connect() first)
//   - network error for other failures
//
// Important UNIX datagram behavior:
//   - Blocks until a datagram arrives or an error occurs
//   - If the datagram is larger than len(p), the excess bytes are discarded
//   - Each call receives one complete datagram (message boundaries preserved)
//   - No guarantee datagrams arrive in order or at all (unreliable)
//   - Thread-safe: Multiple goroutines should NOT call Read concurrently on the
//     same client (underlying socket is not safe for concurrent reads)
//
// Example:
//
//	buf := make([]byte, 8192) // Typical datagram buffer
//	n, err := client.Read(buf)
//	if err != nil {
//	    log.Printf("Read error: %v", err)
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

// Write sends one complete datagram to the remote UNIX socket path.
//
// This method implements io.Writer interface and sends data to the remote
// UNIX socket. Each Write() call sends exactly one complete datagram.
//
// Parameters:
//   - p: Buffer containing data to send. All len(p) bytes will be sent as one datagram.
//
// Returns:
//   - n: Number of bytes sent (0 to len(p))
//   - err: nil if datagram sent successfully, or an error:
//   - ErrInstance if client is nil
//   - ErrConnection if not connected (call Connect() first)
//   - network error if send fails
//
// Important UNIX datagram behavior:
//   - Each Write() sends one complete datagram
//   - System-dependent maximum size (typically 16KB-64KB)
//   - No guarantee of delivery (unreliable protocol, like UDP)
//   - No automatic retransmission
//   - Faster than stream sockets for small messages
//   - Thread-safe: Multiple goroutines should NOT call Write concurrently on the
//     same client (underlying socket is not safe for concurrent writes)
//
// Recommended datagram sizes:
//   - < 8KB: Generally safe and efficient
//   - < 16KB: Should work on most systems
//   - > 16KB: May fail or be truncated depending on system limits
//
// Example:
//
//	data := []byte("Event: user logout")
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

// Close closes the UNIX datagram socket and releases associated resources.
//
// This method closes the underlying socket and removes it from the client's
// state. After calling Close(), IsConnected() will return false and any
// subsequent Read() or Write() calls will return ErrConnection.
//
// Behavior:
//   - Triggers ConnectionClose callback before closing
//   - Closes the underlying socket
//   - Removes socket from client state atomically
//   - Safe to call multiple times (subsequent calls return ErrConnection)
//   - Thread-safe: Can be called concurrently with other operations
//   - Does NOT remove the socket file (server's responsibility)
//
// Returns:
//   - nil: Socket closed successfully
//   - ErrInstance: If client is nil
//   - ErrConnection: If no socket exists (already closed or never connected)
//   - network error: If underlying Close() fails (rare)
//
// Best practice: Always defer Close() after successful Connect() to ensure
// proper cleanup even if errors occur.
//
// Example:
//
//	client := unixgram.New("/tmp/app.sock")
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

// Once performs a one-shot datagram send operation.
//
// This method is a convenience function that:
//  1. Creates and associates a UNIX datagram socket
//  2. Sends all data from the request reader as datagrams
//  3. Invokes the response callback to handle received datagrams (if provided)
//  4. Closes the socket automatically
//
// This is useful for simple request/response patterns where you don't need
// to maintain a persistent socket association.
//
// Parameters:
//   - ctx: Context for socket creation timeout and cancellation
//   - request: Reader containing the data to send. Data is read until EOF.
//     Note: For UNIX datagrams, consider the data size - system limits apply.
//   - fct: Callback function to handle responses. Receives the client as io.Reader.
//     Can be nil if no response is expected (fire-and-forget).
//
// The response callback receives the client itself, allowing you to read
// response datagrams using client.Read().
//
// Behavior:
//   - Automatically creates socket if not connected
//   - Reads all data from request until io.EOF
//   - Calls response callback (if provided)
//   - Automatically closes socket via defer
//   - Triggers all appropriate state callbacks
//   - Error callback triggered for any errors
//
// Returns:
//   - nil: Operation completed successfully
//   - ErrInstance: If client is nil
//   - Socket creation or I/O errors if operation fails
//
// Important UNIX datagram considerations:
//   - No guarantee the datagram(s) will be received (unreliable)
//   - Response callback may timeout waiting for datagrams that never arrive
//   - Consider using a timeout context for response reading
//
// Example:
//
//	request := bytes.NewBufferString("STATUS")
//	err := client.Once(ctx, request, func(reader io.Reader) {
//	    buf := make([]byte, 8192)
//	    n, _ := reader.Read(buf)
//	    fmt.Printf("Response: %s\n", buf[:n])
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Socket automatically closed
//
// See github.com/nabbar/golib/socket.Response for callback signature details.
func (o *cli) Once(ctx context.Context, request io.Reader, fct libsck.Response) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/socket/client/unixgram/once", r)
		}
	}()

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

	if request != nil {
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
	}

	if fct != nil {
		fct(o)
	}

	return nil
}
