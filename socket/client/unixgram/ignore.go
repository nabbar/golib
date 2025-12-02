//go:build !linux && !darwin

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

// This file provides stub implementations for non-Linux platforms where
// UNIX domain datagram sockets are not available. It prevents compilation errors
// while making it clear that UNIX datagram sockets are platform-specific.

package unixgram

import libsck "github.com/nabbar/golib/socket"

// ClientUnix represents a UNIX domain datagram socket client interface.
// On non-Linux platforms, this is a stub interface that cannot be instantiated.
//
// UNIX domain datagram sockets require platform-specific support and are only
// available on Linux and Darwin (macOS) systems. Attempting to use this on other
// platforms will result in a nil client.
//
// See github.com/nabbar/golib/socket.Client for the base interface definition.
type ClientUnix interface {
	libsck.Client
}

// New returns nil on non-Linux platforms as UNIX datagram sockets are not supported.
//
// UNIX domain datagram sockets are a platform-specific feature available only on
// Linux and Darwin systems. This stub implementation allows code to compile on other
// platforms but will always return nil, indicating that UNIX datagram sockets are
// unavailable.
//
// Parameters:
//   - unixfile: Socket file path (ignored on this platform)
//
// Returns:
//   - nil: Always, as UNIX datagram sockets are not available on this platform
//
// For cross-platform code, check for nil before using:
//
//	client := unixgram.New("/tmp/app.sock")
//	if client == nil {
//	    log.Println("UNIX datagram sockets not available on this platform")
//	    // Use UDP or TCP instead
//	}
func New(unixfile string) ClientUnix {
	return nil
}
