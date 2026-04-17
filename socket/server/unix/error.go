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

// Package unix provides error definitions for the Unix domain socket server.
//
// These errors are used to signal issues with configuration, lifecycle management,
// and platform-specific constraints.
package unix

import "fmt"

var (
	// ErrInvalidUnixFile is returned when the provided Unix socket path is empty or malformed.
	// This error occurs during the `RegisterSocket` or `Listen` phase if the path is not a
	// valid filesystem path or does not satisfy the requirements for a Unix socket file.
	//
	// # Usage Case:
	//   - An empty string is passed as the `unixFile` parameter in `RegisterSocket`.
	//   - The path points to a directory instead of a file.
	ErrInvalidUnixFile = fmt.Errorf("invalid unix file for socket listening")

	// ErrInvalidGroup is returned when the specified Group ID (GID) exceeds the maximum allowed value (32767).
	// Unix group IDs must be within the range supported by the operating system to ensure
	// proper file permission management via `os.Chown`.
	//
	// # Usage Case:
	//   - A GID greater than 32767 is passed to `RegisterSocket` on a system with standard 16-bit GID limits.
	ErrInvalidGroup = fmt.Errorf("invalid unix group for socket group permission")

	// ErrInvalidHandler is returned when attempting to start a server without a valid handler function.
	// A `HandlerFunc` is required by the `New()` constructor and must be provided to process
	// incoming client connections.
	//
	// # Usage Case:
	//   - The `hdl` parameter in `New()` is nil.
	//   - The server's `hdl` field has been zeroed out before calling `Listen`.
	ErrInvalidHandler = fmt.Errorf("invalid handler")

	// ErrShutdownTimeout is returned when the server shutdown process exceeds the timeout
	// specified in the context passed to the `Shutdown()` method.
	//
	// # Usage Case:
	//   - Active connections are not closing within the 25-second (or custom) timeout period during a graceful shutdown.
	//   - The server is under heavy load and cannot finish draining connections before the deadline.
	ErrShutdownTimeout = fmt.Errorf("timeout on stopping socket")

	// ErrInvalidInstance is returned when an operation is performed on a nil server instance.
	// This check is used in methods like `Shutdown` to prevent potential nil-pointer panics.
	//
	// # Usage Case:
	//   - Calling `srv.Shutdown(ctx)` when `srv` is nil.
	ErrInvalidInstance = fmt.Errorf("invalid socket instance")
)
