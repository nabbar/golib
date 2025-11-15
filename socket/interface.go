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

// Package socket provides a unified interface for TCP, UDP, and Unix socket communication.
//
// This package offers both client and server implementations for various socket types:
//   - TCP sockets (client/tcp, server/tcp) for reliable, connection-oriented communication
//   - UDP sockets (client/udp, server/udp) for connectionless datagram communication
//   - Unix domain sockets (client/unix, server/unix) for inter-process communication
//   - Unix datagram sockets (client/unixgram, server/unixgram) for connectionless IPC
//
// The package supports TLS encryption for TCP connections using the github.com/nabbar/golib/certificates package.
//
// Configuration is managed through the socket/config package, which provides builders for both
// client and server configurations.
//
// Example usage for a TCP server:
//
//	cfg := config.NewServer().
//	    Network(config.NetworkTCP).
//	    Address(":8080").
//	    Handler(myHandler)
//
//	server, err := cfg.Build(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := server.Listen(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// Example usage for a TCP client:
//
//	cfg := config.NewClient().
//	    Network(config.NetworkTCP).
//	    Address("localhost:8080")
//
//	client, err := cfg.Build(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := client.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
// For more details on configuration, see github.com/nabbar/golib/socket/config.
// For TLS configuration, see github.com/nabbar/golib/certificates.
package socket

import (
	"context"
	"io"
	"net"

	libtls "github.com/nabbar/golib/certificates"
)

// DefaultBufferSize defines the default buffer size for socket I/O operations (32KB).
// This value is used for read/write operations when no custom buffer size is specified.
const DefaultBufferSize = 32 * 1024

// EOL defines the end-of-line delimiter (newline character) used as the default
// message delimiter for socket communication.
const EOL byte = '\n'

// errFilterClosed contains the error message pattern for closed network connections.
// This error is filtered out by ErrorFilter to avoid propagating expected closure errors.
var (
	errFilterClosed = "use of closed network connection"
)

// ConnState represents the current state of a network connection.
// It is used to track and report connection lifecycle events through the FuncInfo callback.
type ConnState uint8

const (
	// ConnectionDial indicates the client is attempting to dial/connect to a server.
	ConnectionDial ConnState = iota
	// ConnectionNew indicates a new connection has been established.
	ConnectionNew
	// ConnectionRead indicates data is being read from the connection.
	ConnectionRead
	// ConnectionCloseRead indicates the read side of the connection is being closed.
	ConnectionCloseRead
	// ConnectionHandler indicates the request handler is being executed.
	ConnectionHandler
	// ConnectionWrite indicates data is being written to the connection.
	ConnectionWrite
	// ConnectionCloseWrite indicates the write side of the connection is being closed.
	ConnectionCloseWrite
	// ConnectionClose indicates the entire connection is being closed.
	ConnectionClose
)

// String returns a human-readable string representation of the connection state.
func (c ConnState) String() string {
	switch c {
	case ConnectionDial:
		return "Dial Connection"
	case ConnectionNew:
		return "New Connection"
	case ConnectionRead:
		return "Read Incoming Stream"
	case ConnectionCloseRead:
		return "Close Incoming Stream"
	case ConnectionHandler:
		return "Run Handler"
	case ConnectionWrite:
		return "Write Outgoing Steam"
	case ConnectionCloseWrite:
		return "Close Outgoing Stream"
	case ConnectionClose:
		return "Close Connection"
	}

	return "unknown connection state"
}

// FuncError is a callback function type for handling errors that occur during
// socket operations. It receives one or more errors as parameters.
type FuncError func(e ...error)

// FuncInfoSrv is a callback function type for receiving informational messages
// about the server's listening state and lifecycle events.
type FuncInfoSrv func(msg string)

// FuncInfo is a callback function type for tracking connection state changes.
// It receives the local address, remote address, and the current connection state.
// This is useful for logging, monitoring, and debugging connection lifecycles.
type FuncInfo func(local, remote net.Addr, state ConnState)

// Handler is a server-side callback function type for processing incoming requests.
// It receives a Reader for reading the request and a Writer for sending the response.
// The handler is responsible for reading the complete request and writing the complete response.
type Handler func(request Reader, response Writer)

// UpdateConn is a callback function type for modifying a net.Conn before it is used.
// This allows for custom connection configuration such as setting timeouts, buffer sizes,
// or other socket options.
type UpdateConn func(co net.Conn)

// Response is a client-side callback function type for processing responses received
// from the server. It receives an io.Reader to read the response data.
type Response func(r io.Reader)

// Server defines the interface for socket server implementations.
// It provides methods for configuring, starting, and managing a socket server
// that can handle multiple concurrent connections.
//
// Server implementations are available for TCP, UDP, Unix domain sockets,
// and Unix datagram sockets in their respective subpackages.
//
// See github.com/nabbar/golib/socket/server/tcp for TCP server implementation.
// See github.com/nabbar/golib/socket/server/udp for UDP server implementation.
// See github.com/nabbar/golib/socket/server/unix for Unix socket server implementation.
// See github.com/nabbar/golib/socket/server/unixgram for Unix datagram server implementation.
type Server interface {
	io.Closer

	// RegisterFuncError registers a callback function to handle errors that occur
	// during server operation. Multiple error handlers can be registered.
	RegisterFuncError(f FuncError)

	// RegisterFuncInfo registers a callback function to track connection state changes.
	// This is useful for monitoring, logging, and debugging connection lifecycles.
	RegisterFuncInfo(f FuncInfo)

	// RegisterFuncInfoServer registers a callback function to receive informational
	// messages about the server's state and operations.
	RegisterFuncInfoServer(f FuncInfoSrv)

	// SetTLS enables or disables TLS encryption for the server.
	// This is only supported for TCP servers. For other socket types, this method
	// may return an error or be a no-op.
	//
	// Parameters:
	//   - enable: true to enable TLS, false to disable it
	//   - config: TLS configuration (see github.com/nabbar/golib/certificates)
	//
	// Returns an error if TLS cannot be enabled or the configuration is invalid.
	SetTLS(enable bool, config libtls.TLSConfig) error

	// Listen starts the server and begins accepting incoming connections.
	// This method blocks until the context is canceled or an error occurs.
	// The context can be used to control the server's lifetime.
	//
	// Returns an error if the server cannot start or encounters a fatal error.
	Listen(ctx context.Context) error

	// Shutdown gracefully stops the server and closes all open connections.
	// The context controls how long to wait for connections to close cleanly.
	//
	// Returns an error if the shutdown process fails.
	Shutdown(ctx context.Context) error

	// IsRunning returns true if the server is currently running and accepting connections.
	IsRunning() bool

	// IsGone returns true if the server has completed its shutdown process.
	// Once IsGone returns true, the server cannot be restarted.
	IsGone() bool

	// Done returns a channel that is closed when the server has finished shutting down.
	// This can be used to wait for the server to completely stop.
	Done() <-chan struct{}

	// OpenConnections returns the current number of active connections being handled
	// by the server. This is useful for monitoring and graceful shutdown.
	OpenConnections() int64
}

// Client defines the interface for socket client implementations.
// It provides methods for configuring, connecting to, and communicating with a socket server.
//
// Client implementations are available for TCP, UDP, Unix domain sockets,
// and Unix datagram sockets in their respective subpackages.
//
// See github.com/nabbar/golib/socket/client/tcp for TCP client implementation.
// See github.com/nabbar/golib/socket/client/udp for UDP client implementation.
// See github.com/nabbar/golib/socket/client/unix for Unix socket client implementation.
// See github.com/nabbar/golib/socket/client/unixgram for Unix datagram client implementation.
type Client interface {
	io.ReadWriteCloser

	// SetTLS enables or disables TLS encryption for the client connection.
	// This is only supported for TCP clients.
	//
	// Parameters:
	//   - enable: true to enable TLS, false to disable it
	//   - config: TLS configuration (see github.com/nabbar/golib/certificates)
	//   - serverName: the server name for TLS certificate verification
	//
	// Returns an error if TLS cannot be enabled or the configuration is invalid.
	SetTLS(enable bool, config libtls.TLSConfig, serverName string) error

	// RegisterFuncError registers a callback function to handle errors that occur
	// during client operations.
	RegisterFuncError(f FuncError)

	// RegisterFuncInfo registers a callback function to track connection state changes
	// during client operations.
	RegisterFuncInfo(f FuncInfo)

	// Connect establishes a connection to the server.
	// The context can be used to set a timeout or cancel the connection attempt.
	//
	// Returns an error if the connection cannot be established.
	Connect(ctx context.Context) error

	// IsConnected returns true if the client is currently connected to the server.
	IsConnected() bool

	// Once sends a single request to the server and processes the response.
	// This is a convenience method for one-shot request/response patterns.
	// The connection is automatically established if not already connected.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeout control
	//   - request: reader containing the request data to send
	//   - fct: callback function to process the response
	//
	// Returns an error if the request cannot be sent or the response cannot be read.
	Once(ctx context.Context, request io.Reader, fct Response) error
}

// ErrorFilter filters out expected network errors that don't require handling.
// It returns nil for errors related to closed connections, which are normal during
// shutdown or connection cleanup. Other errors are returned unchanged.
//
// This function is useful for error handling in connection management code where
// "use of closed network connection" errors are expected and can be safely ignored.
func ErrorFilter(err error) error {
	if err == nil {
		return nil
	}

	if err.Error() == errFilterClosed {
		return nil
	}

	return err
}
