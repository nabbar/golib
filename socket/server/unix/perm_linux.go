//go:build linux

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
	"net"
	"os"
	"syscall"

	libptc "github.com/nabbar/golib/network/protocol"
)

// getListen creates and configures the Unix socket listener on Linux.
//
// This internal helper is responsible for the actual `net.Listen` call and the
// subsequent filesystem-level configuration (permissions and ownership).
//
// # Linux Specific Logic Hierarchy:
//  1. Listen: Calls `net.Listen` using the "unix" network protocol.
//  2. Validation: Verifies that the socket file was physically created.
//  3. Permissions (`chmod`): Applies the configured POSIX permissions.
//  4. Ownership (`chown`): Assigns the socket file to the current UID and the configured GID.
//
// # Lifecycle:
//   - If any step fails, the listener is closed, and the error is returned.
//   - On success, it reports a startup message via the server info callback.
//
// Parameters:
//   - uxf: The filesystem path for the socket.
//
// Returns:
//   - net.Listener: The active listener.
//   - error: Any filesystem or network error encountered.
func (o *srv) getListen(uxf string) (net.Listener, error) {
	var (
		err error
		prm = o.getSocketPerm()
		grp = o.getSocketGroup()
		lis net.Listener
	)

	// Step 1: Create the listener.
	lis, err = net.Listen(libptc.NetworkUnix.Code(), uxf)

	if err != nil {
		return nil, err
	} else if lis == nil {
		return nil, os.ErrNotExist
	}

	// Step 2: Verify the file's existence in the filesystem.
	if _, err = os.Stat(uxf); err != nil {
		_ = lis.Close()
		return nil, err
	}

	// Step 3: Set POSIX permissions (e.g., 0600).
	if err = os.Chmod(uxf, prm.FileMode()); err != nil {
		_ = lis.Close()
		return nil, err
	}

	// Step 4: Set file ownership (current UID, target GID).
	if err = os.Chown(uxf, syscall.Getuid(), grp); err != nil {
		_ = lis.Close()
		return nil, err
	}

	o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUnix.String(), uxf)
	return lis, nil
}
