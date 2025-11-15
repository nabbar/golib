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
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
)

// Internal atomic map keys for storing client state.
// Using uint8 keys for efficient memory usage and fast lookups.
const (
	keyNetAddr uint8 = iota // Network address (string)
	keyTLSCfg               // TLS configuration (*tls.Config)
	keyFctErr               // Error callback function (libsck.FuncError)
	keyFctInfo              // Info callback function (libsck.FuncInfo)
	keyNetConn              // Active network connection (net.Conn)
)

// cli is the internal implementation of ClientTCP interface.
// It uses an atomic map to store all client state in a thread-safe manner,
// avoiding the need for explicit locking and preventing nil pointer panics
// that occurred with the previous atomic.Value implementation.
//
// State management:
//   - All state is stored in a thread-safe atomic.Map[uint8]
//   - Keys can be safely deleted to represent "no value" without storing nil
//   - Multiple goroutines can safely call methods concurrently
//   - Connection state changes trigger registered callbacks asynchronously
type cli struct {
	m libatm.Map[uint8] // Atomic map storing all client state by key
}

// SetTLS configures TLS encryption for the client connection.
//
// This method must be called before Connect() to enable TLS. If called on
// an already connected client, the TLS settings will apply to the next connection.
//
// Parameters:
//   - enable: Set to true to enable TLS, false to disable
//   - config: TLS configuration from github.com/nabbar/golib/certificates package.
//     Required when enable is true, ignored when false.
//   - serverName: The expected server hostname for certificate verification.
//     Used for SNI (Server Name Indication) and certificate validation.
//
// When enable is false, any existing TLS configuration is removed and the
// client will use plain TCP connections.
//
// Returns an error if:
//   - config is nil when enable is true
//   - the config cannot generate a valid *tls.Config
//
// Example:
//
//	tlsConfig := certificates.New()
//	err := tlsConfig.AddRootCA(caCert)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	client, _ := tcp.New("secure.example.com:443")
//	err = client.SetTLS(true, tlsConfig, "secure.example.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// See github.com/nabbar/golib/certificates for TLS configuration details.
func (o *cli) SetTLS(enable bool, config libtls.TLSConfig, serverName string) error {
	if !enable {
		// #nosec
		o.m.Delete(keyTLSCfg)
		return nil
	}

	if config == nil {
		return fmt.Errorf("invalid tls config")
	} else if t := config.TlsConfig(serverName); t == nil {
		return fmt.Errorf("invalid tls config")
	} else {
		o.m.Store(keyTLSCfg, t)
		return nil
	}
}

// RegisterFuncError registers a callback function for error notifications.
//
// The callback is invoked asynchronously (in a separate goroutine) whenever
// an error occurs during client operations, including:
//   - Connection failures
//   - Read/Write errors
//   - TLS handshake errors
//   - Network timeouts
//
// The callback receives variadic error parameters, typically containing a single
// error but may contain multiple errors in some scenarios.
//
// Pass nil to unregister the current error callback. Only one error callback
// can be registered at a time; calling this method replaces any existing callback.
//
// The callback is executed asynchronously to avoid blocking the main operation.
// Ensure your callback handles errors appropriately and returns quickly.
//
// Example:
//
//	client, _ := tcp.New("localhost:8080")
//	client.RegisterFuncError(func(errs ...error) {
//	    for _, err := range errs {
//	        log.Printf("TCP client error: %v", err)
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

// RegisterFuncInfo registers a callback function for connection state notifications.
//
// The callback is invoked asynchronously (in a separate goroutine) for various
// connection state changes:
//   - ConnectionDial: When dialing starts
//   - ConnectionNew: When connection is established
//   - ConnectionRead: Before each read operation
//   - ConnectionWrite: Before each write operation
//   - ConnectionClose: When connection is closed
//
// The callback receives:
//   - local: Local network address (may be nil during dial)
//   - remote: Remote network address (may be nil during dial)
//   - state: Connection state from github.com/nabbar/golib/socket.ConnState
//
// Pass nil to unregister the current info callback. Only one info callback
// can be registered at a time; calling this method replaces any existing callback.
//
// The callback is executed asynchronously to avoid blocking I/O operations.
// Keep callback execution time minimal to avoid impacting performance.
//
// Example:
//
//	client, _ := tcp.New("localhost:8080")
//	client.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
//	    log.Printf("Connection state: %v (local: %v, remote: %v)", state, local, remote)
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
// This internal method is called for connection state changes.
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

// dial establishes the actual network connection to the server.
// This internal method handles both plain TCP and TLS connections based on
// the current TLS configuration. It uses context for cancellation and timeout.
//
// The dialer is configured with a 5-minute keep-alive to maintain long-lived connections.
// For TLS connections, it uses tls.Dialer with the configured TLS settings.
//
// Returns:
//   - net.Conn: The established connection (may be *net.TCPConn or *tls.Conn)
//   - error: ErrInstance if client is nil, ErrAddress if address is invalid,
//     or a network error if connection fails
func (o *cli) dial(ctx context.Context) (net.Conn, error) {
	if o == nil {
		return nil, ErrInstance
	}

	d := &net.Dialer{
		KeepAlive: 5 * time.Minute,
	}

	if v, k := o.m.Load(keyNetAddr); !k || v == nil {
		return nil, ErrAddress
	} else if adr, ok := v.(string); !ok {
		return nil, ErrAddress
	} else if i, k := o.m.Load(keyTLSCfg); k && i != nil {
		if t, k := i.(*tls.Config); k {
			u := &tls.Dialer{
				NetDialer: d,
				Config:    t,
			}
			return u.DialContext(ctx, libptc.NetworkTCP.Code(), adr)
		} else {
			return d.DialContext(ctx, libptc.NetworkTCP.Code(), adr)
		}
	} else {
		return d.DialContext(ctx, libptc.NetworkTCP.Code(), adr)
	}
}

// IsConnected checks if the client has an active connection.
//
// This method checks the local connection state by verifying that a connection
// object exists in the client's state. It does NOT perform any network I/O to
// verify if the connection is still alive.
//
// Important notes:
//   - Returns true if a connection was established and not yet closed locally
//   - Returns false if never connected or after Close() was called
//   - Does NOT detect if the remote server has closed the connection
//   - Does NOT detect network failures until actual I/O is attempted
//
// To detect if a connection is truly alive, you must attempt a Read() or Write()
// operation, which will return an error if the connection is broken.
//
// Thread-safe and can be called from multiple goroutines concurrently.
//
// Example:
//
//	if client.IsConnected() {
//	    // Attempt I/O to verify connection is truly alive
//	    _, err := client.Write([]byte("ping"))
//	    if err != nil {
//	        log.Printf("Connection is broken: %v", err)
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

// Connect establishes a connection to the server.
//
// This method dials the server address specified in New() and stores the
// resulting connection. If TLS was configured via SetTLS(), a TLS handshake
// is performed as part of the connection process.
//
// The context parameter controls timeouts and cancellation:
//   - Use context.WithTimeout() to set a connection timeout
//   - Use context.WithCancel() to allow cancelling the connection attempt
//   - The context is only used during connection establishment, not for the lifetime
//
// Connection states:
//   - Triggers ConnectionDial callback when dialing starts
//   - Triggers ConnectionNew callback when connection succeeds
//   - Triggers error callback if connection fails
//
// If a connection already exists, it is replaced. The old connection is closed
// after verifying it's still valid (by attempting a zero-byte write).
//
// Returns:
//   - nil: Connection established successfully
//   - ErrInstance: If client is nil
//   - network error: If connection fails (includes TLS handshake errors)
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	o.fctInfo(&net.TCPAddr{}, &net.TCPAddr{}, libsck.ConnectionDial)
	if con, err = o.dial(ctx); err != nil {
		o.fctError(err)
		return err
	}

	o.fctInfo(con.LocalAddr(), con.RemoteAddr(), libsck.ConnectionNew)
	if i, k := o.m.Swap(keyNetConn, con); !k || i == nil {
		return nil
	} else if c, k := i.(net.Conn); !k || c == nil {
		return nil
	} else if _, e := c.Write(nil); e != nil {
		return nil
	} else {
		_ = c.Close()
		return nil
	}
}

// Read reads data from the connection into the provided buffer.
//
// This method implements io.Reader interface and reads data from the underlying
// TCP connection. The method blocks until data is available, an error occurs,
// or the connection is closed.
//
// Parameters:
//   - p: Buffer to read data into. The read will fill up to len(p) bytes.
//
// Returns:
//   - n: Number of bytes actually read (0 to len(p))
//   - err: nil on success, or an error:
//   - ErrInstance if client is nil
//   - ErrConnection if not connected (call Connect() first)
//   - io.EOF if connection closed cleanly by remote
//   - network error for other failures
//
// Behavior:
//   - Blocks until data is available or error occurs
//   - Triggers ConnectionRead callback before reading
//   - Triggers error callback if an error occurs
//   - Thread-safe: Multiple goroutines SHOULD NOT call Read concurrently on the
//     same client (net.Conn is not safe for concurrent reads)
//
// Important: TCP Read() does not support timeouts directly. Use SetReadDeadline()
// on the underlying connection or use context cancellation at the application level.
//
// Example:
//
//	buf := make([]byte, 4096)
//	n, err := client.Read(buf)
//	if err != nil {
//	    if err == io.EOF {
//	        log.Println("Connection closed by server")
//	    } else {
//	        log.Printf("Read error: %v", err)
//	    }
//	    return
//	}
//	log.Printf("Read %d bytes: %s", n, buf[:n])
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

// Write writes data to the connection from the provided buffer.
//
// This method implements io.Writer interface and writes data to the underlying
// TCP connection. The method blocks until all data is written or an error occurs.
//
// Parameters:
//   - p: Buffer containing data to write. All len(p) bytes will be written.
//
// Returns:
//   - n: Number of bytes actually written (0 to len(p))
//   - err: nil if all data written successfully, or an error:
//   - ErrInstance if client is nil
//   - ErrConnection if not connected (call Connect() first)
//   - network error if write fails (e.g., broken pipe, connection reset)
//
// Behavior:
//   - Blocks until all data is written or error occurs
//   - Triggers ConnectionWrite callback before writing
//   - Triggers error callback if an error occurs
//   - Thread-safe: Multiple goroutines SHOULD NOT call Write concurrently on the
//     same client (net.Conn is not safe for concurrent writes)
//
// Short writes (n < len(p) with no error) are not expected with TCP connections.
// If they occur, it indicates an unusual network condition.
//
// Example:
//
//	data := []byte("Hello, server!\n")
//	n, err := client.Write(data)
//	if err != nil {
//	    log.Printf("Write error: %v", err)
//	    return
//	}
//	if n != len(data) {
//	    log.Printf("Short write: %d of %d bytes", n, len(data))
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

// Close closes the connection and releases associated resources.
//
// This method closes the underlying TCP connection and removes it from the
// client's state. After calling Close(), IsConnected() will return false and
// any subsequent Read() or Write() calls will return ErrConnection.
//
// Behavior:
//   - Triggers ConnectionClose callback before closing
//   - Closes the underlying net.Conn
//   - Removes connection from client state atomically
//   - Safe to call multiple times (subsequent calls return ErrConnection)
//   - Thread-safe: Can be called concurrently with other operations
//
// Returns:
//   - nil: Connection closed successfully
//   - ErrInstance: If client is nil
//   - ErrConnection: If no connection exists (already closed or never connected)
//   - network error: If underlying Close() fails (rare)
//
// Best practice: Always defer Close() after successful Connect() to ensure
// proper cleanup even if errors occur.
//
// Example:
//
//	client, err := tcp.New("localhost:8080")
//	if err != nil {
//	    log.Fatal(err)
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
//  1. Connects to the server
//  2. Writes all data from the request reader
//  3. Invokes the response callback to handle the server's response
//  4. Closes the connection automatically
//
// This is useful for simple request/response protocols where you don't need
// to maintain a persistent connection.
//
// Parameters:
//   - ctx: Context for connection timeout and cancellation
//   - request: Reader containing the request data to send (e.g., bytes.Buffer)
//   - fct: Callback function to handle the response. Receives the client as io.Reader.
//     Can be nil if no response is expected.
//
// The response callback receives the client itself, allowing you to read the
// response using client.Read() or io.Copy(), etc.
//
// Behavior:
//   - Automatically connects if not connected
//   - Reads all data from request until io.EOF
//   - Calls response callback (if provided)
//   - Automatically closes connection via defer
//   - Triggers all appropriate state callbacks
//   - Error callback triggered for any errors
//
// Returns:
//   - nil: Request/response completed successfully
//   - ErrInstance: If client is nil
//   - Connection or I/O errors if operation fails
//
// Example:
//
//	request := bytes.NewBufferString("GET / HTTP/1.0\r\n\r\n")
//	err := client.Once(ctx, request, func(reader io.Reader) {
//	    response, _ := io.ReadAll(reader)
//	    fmt.Printf("Response: %s\n", response)
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Connection automatically closed
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
