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

package unixgram

import "fmt"

// This file contains predefined errors returned by the unixgram server implementation.
// These errors are specifically chosen to handle common failure modes in Unix domain
// datagram socket management, such as permission issues, invalid paths, and lifecycle timeouts.

var (
	// ErrInvalidUnixFile is returned when the specified filesystem path for the socket
	// is empty or contains invalid characters. The socket path is essential as it
	// serves as the endpoint for IPC communication.
	ErrInvalidUnixFile = fmt.Errorf("invalid unix file for socket listening")

	// ErrInvalidGroup is returned when the specified GID (Group ID) for the socket file
	// is invalid or exceeds the system's maximum (MaxGID, typically 32767).
	// Setting a group allows multiple users/services within the same group to communicate.
	ErrInvalidGroup = fmt.Errorf("invalid unix group for socket group permission")

	// ErrInvalidHandler is returned when attempting to start a server without a valid
	// HandlerFunc. The handler is the primary entry point for processing incoming datagrams.
	ErrInvalidHandler = fmt.Errorf("invalid handler")

	// ErrShutdownTimeout is returned when the server's graceful shutdown procedure
	// (StopListen) fails to complete within the time allocated by the provided context.
	// This usually happens if the handler function blocks indefinitely without responding
	// to the context's cancellation signal.
	//
	// Use Case:
	// Preventing a hanging service from blocking a system-wide shutdown or a container restart.
	ErrShutdownTimeout = fmt.Errorf("timeout on stopping socket")

	// ErrInvalidInstance is returned when a method is called on a nil server instance.
	// This is a defensive error to prevent panics during runtime.
	ErrInvalidInstance = fmt.Errorf("invalid socket instance")
)

/*
Error Handling Dataflow:

	[Operation Attempted] -> (Validation Fails) -> [Return ErrXXX]
	                                     |
	                                     v
	[Internal Callback] <--------- (Error Occurs)
	       |
	       v
	[User's FuncError Callback] (RegisterFuncError)
*/
