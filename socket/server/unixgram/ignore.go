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

// Package unixgram provides stubs for non-Linux, non-Darwin platforms.
//
// # Platform Limitations
//
// Unix domain sockets in datagram mode (SOCK_DGRAM) are a feature of POSIX-compliant
// operating systems (primarily Linux and BSD-based systems like macOS/Darwin).
//
// While newer versions of Windows (Windows 10+, Windows Server 2019+) support Unix domain
// sockets (AF_UNIX) for stream mode (SOCK_STREAM), datagram mode is not natively
// supported in the same way.
//
// # Design Alternatives for Other Platforms
//
// If you are developing for a platform where this package is disabled:
//  1. UDP: Use the `udp` package for local datagram-based IPC via the loopback address (127.0.0.1).
//  2. Named Pipes: Use Windows-specific named pipes (not currently part of this package).
//  3. Shared Memory: For high-performance local IPC.
//
// This file ensures that projects including the `unixgram` package will still compile
// on unsupported systems, returning `nil` for server instances.

package unixgram

import (
	"os"

	libsck "github.com/nabbar/golib/socket"
)

// maxGID defines the maximum allowed Unix group ID value (32767).
// This is a stub for cross-platform compatibility.
const maxGID = 32767

// ServerUnixGram is a stub interface for non-Linux, non-Darwin platforms.
// It matches the interface on supported platforms but does not provide
// any functional implementation.
type ServerUnixGram interface {
	libsck.Server

	// RegisterSocket would configure the Unix socket file path, permissions, and group.
	// On non-supported platforms, this returns nil and performs no action.
	RegisterSocket(unixFile string, perm os.FileMode, gid int32) error
}

// New returns nil on non-Linux, non-Darwin platforms.
//
// Since Unix domain datagram sockets are not available, this function
// serves as a placeholder to allow the rest of the application to compile
// without platform-specific guards around the constructor call.
//
// Important: Developers should check for nil results when using this constructor
// in a cross-platform codebase.
func New(u libsck.UpdateConn, h libsck.HandlerFunc) ServerUnixGram {
	return nil
}
