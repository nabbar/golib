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

// Package tcp provides a robust and performance-oriented TCP server implementation.
// It integrates with nabbar/golib/socket for common interfaces and configuration.
//
// Features include:
//   - TLS support (v1.2, v1.3) with certificate management.
//   - Graceful shutdown with connection draining.
//   - High-performance memory pooling (sync.Pool).
//   - Centralized idle connection scanning.
//   - TCP_NODELAY and Keep-Alive tuning.
package tcp

import "fmt"

var (
	// ErrInvalidAddress is returned when the provided server address is empty,
	// malformed, or cannot be resolved using net.ResolveTCPAddr.
	//
	// Valid address examples:
	//   - "127.0.0.1:8080" (IPv4 localhost)
	//   - "[::1]:8080"     (IPv6 localhost)
	//   - ":8080"          (All interfaces)
	ErrInvalidAddress = fmt.Errorf("invalid listen address")

	// ErrInvalidHandler is returned when the required HandlerFunc is not provided 
	// during server initialization via New().
	//
	// The handler must be a function matching the libsck.HandlerFunc signature 
	// and is responsible for processing each client connection.
	ErrInvalidHandler = fmt.Errorf("invalid handler")

	// ErrShutdownTimeout is returned when the graceful shutdown process exceeds 
	// the provided context's deadline.
	//
	// This error occurs during the draining phase, when active connections 
	// fail to close or finish their task within the allocated time.
	ErrShutdownTimeout = fmt.Errorf("timeout on stopping socket")

	// ErrInvalidInstance is returned when a method is called on a nil server instance
	// or an instance that has not been properly initialized via New().
	ErrInvalidInstance = fmt.Errorf("invalid socket instance")
)
