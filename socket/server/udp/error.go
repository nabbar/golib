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

// Package udp provides a high-performance, stateless UDP server implementation.
package udp

import "fmt"

var (
	// ErrInvalidAddress is returned when the provided listen address fails validation.
	//
	// # Validation Logic
	//
	// The address is parsed using net.ResolveUDPAddr. Common valid formats:
	//   - ":port" (all interfaces, IPv4 and IPv6)
	//   - "0.0.0.0:port" (all IPv4 interfaces)
	//   - "127.0.0.1:port" (loopback only)
	//   - "[::1]:port" (IPv6 loopback)
	//
	// # Use Case
	//
	// Typically returned during RegisterServer() or at the start of Listen().
	ErrInvalidAddress = fmt.Errorf("invalid listen address")

	// ErrInvalidHandler is returned when trying to start the server with a nil handler.
	//
	// # Mandatory Requirement
	//
	// Since the server only spawns a single handler for all UDP traffic, a non-nil
	// HandlerFunc is required by the New() constructor.
	ErrInvalidHandler = fmt.Errorf("invalid handler")

	// ErrShutdownTimeout is returned when the graceful shutdown period expires.
	//
	// # Shutdown Logic
	//
	// When Shutdown(ctx) is called, the server:
	//   1. Closes the broadcast channel (gnc).
	//   2. Closes the listener socket.
	//   3. Monitors the IsRunning flag for cleanup completion.
	//
	// If the provided context 'ctx' times out before the cleanup is finished,
	// this error is returned.
	ErrShutdownTimeout = fmt.Errorf("timeout on stopping socket")

	// ErrInvalidInstance is returned when operating on a nil *srv pointer.
	//
	// # Defensive Coding
	//
	// This error prevents panics in scenarios where methods are called on an
	// uninitialized or nil server instance.
	ErrInvalidInstance = fmt.Errorf("invalid socket instance")
)
