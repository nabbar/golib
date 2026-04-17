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

package unix

import (
	"os"

	libsck "github.com/nabbar/golib/socket"
)

// maxGID defines the maximum allowed Unix group ID value (32767).
// This limit is common on older 16-bit GID systems and is used here for safety.
const maxGID = 32767

// ServerUnix is a stub interface for non-Linux, non-Darwin platforms.
//
// On Linux and Darwin, this interface extends `libsck.Server` with Unix socket-specific methods.
// On unsupported platforms, this stub allows code to compile while returning no-op or nil results.
//
// # Design Pattern: Stub Implementation
// This is used to maintain cross-platform compatibility of the larger project while
// only enabling Unix-specific features on POSIX systems.
type ServerUnix interface {
	libsck.Server

	// RegisterSocket would configure the Unix socket file path, permissions, and group.
	// On non-supported platforms, this returns nil as it is a no-op stub.
	RegisterSocket(unixFile string, perm os.FileMode, gid int32) error
}

// New returns nil on non-Linux, non-Darwin platforms.
//
// Unix domain sockets are only supported on Linux and Darwin (macOS) in this implementation.
// Calling `New` on other systems (like Windows) will result in a nil pointer, allowing
// applications to gracefully handle the absence of Unix socket support.
//
// Parameters:
//   - u: Optional UpdateConn callback (unused on unsupported platforms).
//   - h: HandlerFunc function (unused on unsupported platforms).
//
// Returns:
//   - ServerUnix: Always nil on this platform.
//
// See `socket/server/unix/interface.go` for the full implementation on supported systems.
func New(u libsck.UpdateConn, h libsck.HandlerFunc) ServerUnix {
	return nil
}
