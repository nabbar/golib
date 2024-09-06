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
	"net"

	libtls "github.com/nabbar/golib/certificates"
)

// DefaultBufferSize is the default buffer size
const DefaultBufferSize = 32 * 1024

// EOL is the end of line, default delimiter of the socket
const EOL byte = '\n'

// ConnState is used to process state connection
type ConnState uint8

const (
	ConnectionDial ConnState = iota
	ConnectionNew
	ConnectionRead
	ConnectionCloseRead
	ConnectionHandler
	ConnectionWrite
	ConnectionCloseWrite
	ConnectionClose
)

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

// FuncError is used to process errors during running process
type FuncError func(e ...error)

// FuncInfoSrv is used to process information about the listening server
type FuncInfoSrv func(msg string)

// FuncInfo is used to process state connection information during running server.
type FuncInfo func(local, remote net.Addr, state ConnState)

// Handler is used to process request
type Handler func(request Reader, response Writer)

// Response is used to process response
type Response func(r io.Reader)

type Server interface {
	io.Closer

	// RegisterFuncError registers a FuncError used to process errors during running server
	// f FuncError
	RegisterFuncError(f FuncError)

	// RegisterFuncInfo registers the given FuncInfo used to process state connection information during running server.
	// f FuncInfo - the FuncInfo to be registered.
	RegisterFuncInfo(f FuncInfo)

	// RegisterFuncInfoServer registers the given FuncInfoSrv used to process information about the listening server.
	// f FuncInfoSrv parameter.
	RegisterFuncInfoServer(f FuncInfoSrv)

	// SetTLS defines if the server should use TLS and if so, the configuration associated.
	// Parameters:
	//   - enable bool defines if the server should use TLS
	//   - config libtls.TLSConfig defines the configuration
	// Returns: error
	SetTLS(enable bool, config libtls.TLSConfig) error

	// Listen is a function that listens for incoming connections.
	// It takes a context.Context as a parameter and returns an error.
	Listen(ctx context.Context) error

	// Shutdown is a function that stops the server.
	// ctx context.Context
	// error
	Shutdown(ctx context.Context) error

	// IsRunning returns a boolean value indicating whether the process is currently running.
	// Returns a boolean value.
	IsRunning() bool

	// IsGone returns a boolean value indicating the process is in shutting down and has been gone.
	// Returns a boolean.
	IsGone() bool

	// Done returns a read-only channel who's value is set when the shutdown process is reached.
	// Returns <-chan struct{}.
	Done() <-chan struct{}

	// OpenConnections returns the number of open connections.
	// Returns an int64.
	OpenConnections() int64
}

type Client interface {
	io.ReadWriteCloser

	SetTLS(enable bool, config libtls.TLSConfig, serverName string) error

	// RegisterFuncError registers a FuncError used to process errors during running process
	// f FuncError
	RegisterFuncError(f FuncError)

	// RegisterFuncInfo registers the given FuncInfo used to process state connection information during running process.
	// f FuncInfo - the FuncInfo to be registered.
	RegisterFuncInfo(f FuncInfo)

	// Connect is used to establish a connection with the server.
	// ctx context.Context
	// error
	Connect(ctx context.Context) error

	// IsConnected returns a boolean value indicating whether the connection is currently established.
	// bool
	IsConnected() bool

	// Once is used to send a request to the server and wait for a response.
	// ctx context.Context, request io.Reader, fct Response.
	// error.
	Once(ctx context.Context, request io.Reader, fct Response) error
}
