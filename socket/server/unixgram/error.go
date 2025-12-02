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

// Package unixgram provides a Unix domain datagram socket server implementation.
//
// This package implements the github.com/nabbar/golib/socket.Server interface
// for Unix domain sockets in datagram mode (SOCK_DGRAM), providing a connectionless
// server with features including:
//   - Unix domain socket file creation and management
//   - File permissions and group ownership control
//   - Datagram handling without persistent connections
//   - Single handler for all incoming datagrams
//   - Callback hooks for errors and datagram events
//   - Graceful shutdown support
//   - Atomic state management
//   - Context-aware operations
//
// Unix datagram sockets provide connectionless inter-process communication (IPC)
// on the same host. Like UDP, they operate in datagram mode but use filesystem
// paths instead of IP addresses and ports. They appear as special files.
//
// Key differences from unix package (connection-oriented):
//   - No persistent connections (like UDP vs TCP)
//   - Single handler processes all datagrams
//   - OpenConnections() returns 1 when running, 0 when stopped
//   - No per-client state maintained
//
// Platform support: Linux and Darwin (macOS). See ignore.go for other platforms.
//
// See github.com/nabbar/golib/socket for the Server interface definition.
// See github.com/nabbar/golib/socket/server/unix for connection-oriented Unix sockets.
package unixgram

import "fmt"

var (
	ErrInvalidUnixFile = fmt.Errorf("invalid unix file for socket listening")

	// ErrInvalidGroup is returned when the specified GID exceeds the maximum allowed value (32767).
	// Unix group IDs must be within the valid range for the operating system.
	ErrInvalidGroup = fmt.Errorf("invalid unix group for socket group permission")

	// ErrInvalidHandler is returned when attempting to start a server without a valid handler function.
	// A handler must be provided via the New() constructor.
	ErrInvalidHandler = fmt.Errorf("invalid handler")

	// ErrShutdownTimeout is returned when the server shutdown exceeds the context timeout.
	// This typically happens when StopListen() takes longer than expected.
	ErrShutdownTimeout = fmt.Errorf("timeout on stopping socket")

	// ErrInvalidInstance is returned when operating on a nil server instance.
	ErrInvalidInstance = fmt.Errorf("invalid socket instance")
)
