//go:build !linux && !darwin

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

// This file provides stub implementations for non-Linux, non-Darwin platforms.
// Unix domain datagram sockets are specific to Linux and Darwin (macOS), so this file
// prevents build errors on other operating systems by providing minimal interface
// definitions that return nil.
//
// For actual Unix datagram socket server functionality, use Linux or Darwin systems
// where the full implementation is available (see interface.go, model.go, listener.go).

package unixgram

import (
	"os"

	libsck "github.com/nabbar/golib/socket"
)

// maxGID defines the maximum allowed Unix group ID value (32767).
const maxGID = 32767

// ServerUnixGram is a stub interface for non-Linux, non-Darwin platforms.
// On Linux and Darwin, this interface extends libsck.Server with Unix datagram socket-specific methods.
type ServerUnixGram interface {
	libsck.Server
	// RegisterSocket would configure the Unix socket file path, permissions, and group.
	// On non-Linux, non-Darwin platforms, this is a no-op stub.
	RegisterSocket(unixFile string, perm os.FileMode, gid int32) error
}

// New returns nil on non-Linux, non-Darwin platforms.
// Unix domain datagram sockets are only supported on Linux and Darwin (macOS).
// On supported systems, this creates a functional Unix datagram socket server.
//
// Parameters:
//   - u: Optional UpdateConn callback (unused on unsupported platforms)
//   - h: Handler function (unused on unsupported platforms)
//
// See github.com/nabbar/golib/socket/server/unixgram.New for the full implementation.
func New(u libsck.UpdateConn, h libsck.Handler) ServerUnixGram {
	return nil
}
