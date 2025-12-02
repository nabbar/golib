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

// Package unix provides a Unix domain socket server implementation with session support.
//
// This package implements the github.com/nabbar/golib/socket.Server interface
// for Unix domain sockets (AF_UNIX), providing a connection-oriented server with features including:
//   - Unix domain socket file creation and management
//   - File permissions and group ownership control
//   - Persistent connections with session handling
//   - Connection lifecycle management (accept, read, write, close)
//   - Callback hooks for errors and connection events
//   - Graceful shutdown with connection draining
//   - Atomic state management
//   - Context-aware operations
//
// Unix domain sockets provide inter-process communication (IPC) on the same host
// with lower overhead than TCP sockets. They appear as special files in the filesystem.
//
// Platform support: Linux and Darwin (macOS). See ignore.go for other platforms.
//
// See github.com/nabbar/golib/socket for the Server interface definition.
package unix

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
