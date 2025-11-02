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

// Package udp provides a UDP server implementation with connectionless datagram support.
//
// This package implements the github.com/nabbar/golib/socket.Server interface
// for UDP protocol, providing a stateless datagram server with features including:
//   - Connectionless UDP datagram handling
//   - Single handler for all incoming datagrams
//   - Callback hooks for errors and informational messages
//   - Graceful shutdown support
//   - Atomic state management
//   - Context-aware operations
//
// Unlike TCP servers, UDP servers operate in a stateless mode where each datagram
// is processed independently without maintaining persistent connections.
//
// See github.com/nabbar/golib/socket for the Server interface definition.
package udp

import "fmt"

var (
	// ErrInvalidAddress is returned when the server address is empty or malformed.
	// The address must be in the format "host:port" or ":port" for all interfaces.
	ErrInvalidAddress = fmt.Errorf("invalid listen address")

	// ErrContextClosed is returned when an operation is cancelled due to context cancellation.
	ErrContextClosed = fmt.Errorf("context closed")

	// ErrServerClosed is returned when attempting to perform operations on a closed server.
	ErrServerClosed = fmt.Errorf("server closed")

	// ErrInvalidHandler is returned when attempting to start a server without a valid handler function.
	// A handler must be provided via the New() constructor.
	ErrInvalidHandler = fmt.Errorf("invalid handler")

	// ErrShutdownTimeout is returned when the server shutdown exceeds the context timeout.
	// This typically happens when StopListen() takes longer than expected.
	ErrShutdownTimeout = fmt.Errorf("timeout on stopping socket")

	// ErrGoneTimeout is returned when connection draining exceeds the context timeout.
	// Note: For UDP servers, this is rarely used as there are no persistent connections.
	ErrGoneTimeout = fmt.Errorf("timeout on closing connections")

	// ErrInvalidInstance is returned when operating on a nil server instance.
	ErrInvalidInstance = fmt.Errorf("invalid socket instance")
)
