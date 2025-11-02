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
const maxGID = 32767

// ServerUnix is a stub interface for non-Linux platforms.
// On Linux, this interface extends libsck.Server with Unix socket-specific methods.
type ServerUnix interface {
	libsck.Server
	// RegisterSocket would configure the Unix socket file path, permissions, and group.
	// On non-Linux platforms, this is a no-op stub.
	RegisterSocket(unixFile string, perm os.FileMode, gid int32) error
}

// New returns nil on non-Linux platforms.
// Unix domain sockets are only supported on Linux.
// On Linux systems, this creates a functional Unix socket server.
func New(h libsck.Handler) ServerUnix {
	return nil
}
