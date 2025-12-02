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

// Package tcp provides a TCP server implementation with support for TLS,
// connection management, and various callback hooks for monitoring and error handling.
//
// This package implements the github.com/nabbar/golib/socket.Server interface
// and provides a robust TCP server with features including:
//   - TLS/SSL support with certificate management
//   - Graceful shutdown with connection draining
//   - Connection lifecycle callbacks (new, read, write, close)
//   - Error and informational logging callbacks
//   - Atomic connection counting and state management
//   - Context-aware operations
//
// See github.com/nabbar/golib/socket for the Server interface definition.
package tcp

import "fmt"

var (
	// ErrInvalidAddress is returned when the server address is empty or malformed.
	// The address must be in the format "host:port" or ":port" for all interfaces.
	//
	// Example of valid addresses:
	//   - "localhost:8080" - Listen on localhost port 8080
	//   - ":8080" - Listen on all interfaces port 8080
	//   - "0.0.0.0:8080" - Explicitly listen on all IPv4 interfaces
	ErrInvalidAddress = fmt.Errorf("invalid listen address")

	// ErrInvalidHandler is returned when attempting to start a server without a valid handler function.
	// A handler must be provided via the New() constructor and must not be nil.
	//
	// Example of valid usage:
	//   handler := func(r socket.Reader, w socket.Writer) { /* ... */ }
	//   srv, err := tcp.New(nil, handler, config)
	ErrInvalidHandler = fmt.Errorf("invalid handler")

	// ErrShutdownTimeout is returned when the server shutdown exceeds the context timeout.
	// This typically happens when StopListen() takes longer than expected to complete.
	// The server will attempt to close all active connections before returning this error.
	//
	// To handle this error, you may want to implement a fallback strategy or log the event.
	ErrShutdownTimeout = fmt.Errorf("timeout on stopping socket")

	// ErrInvalidInstance is returned when operating on a nil server instance.
	// This typically occurs if the server was not properly initialized or has been set to nil.
	// Always check for this error when working with server instances that might be nil.
	ErrInvalidInstance = fmt.Errorf("invalid socket instance")
)
