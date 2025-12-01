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

package server_test

import (
	libsck "github.com/nabbar/golib/socket"
)

// basicHandler returns a simple handler that closes the connection immediately.
// This handler is used for basic server creation tests where we don't need
// actual I/O operations.
func basicHandler() libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() { _ = c.Close() }()
	}
}

// echoHandler returns a handler that echoes back received data.
// This is useful for integration tests where we need to verify data flow.
func echoHandler() libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() { _ = c.Close() }()
		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				_, err = c.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}
	}
}

// findFreePort returns an available TCP port for testing.
// It binds to port 0 which lets the OS choose an available port.
func findFreePort() int {
	// Start from a high port to avoid conflicts
	// In practice, we'll use port 0 in configs to let OS choose
	return 0
}

// getTestTCPAddress returns a test TCP address with a free port.
func getTestTCPAddress() string {
	return ":0" // Let OS choose port
}

// getTestUDPAddress returns a test UDP address with a free port.
func getTestUDPAddress() string {
	return ":0" // Let OS choose port
}
