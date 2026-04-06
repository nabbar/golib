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

package unixgram

import (
	"net"
	"os"
	"syscall"

	libptc "github.com/nabbar/golib/network/protocol"
)

// getListen creates and configures the Unix domain datagram socket for Linux.
//
// # Initialization Flow for Linux
//
// 1. Resolve Address: The filesystem path is resolved as a Unixgram address.
// 2. Bind Socket: Calls net.ListenUnixgram, creating a SOCK_DGRAM endpoint.
// 3. Security (Post-creation):
//    - os.Chmod: Adjusts the socket file permissions (e.g., 0600) to ensure
//      only the owner (and potentially a group) can write datagrams to the socket.
//    - os.Chown: Changes the file's group ownership to the configured GID.
//
// # Parameters:
//   - uxf: The filesystem path where the socket will be bound.
//
// # Returns:
//   - *net.UnixConn: The active listener connection.
//   - Error if any operation (bind, chmod, chown) fails.
//
// # Performance Note:
// By calling net.ListenUnixgram first and then applying permissions, we ensure
// that the socket file is correctly created by the kernel before modification.
func (o *srv) getListen(uxf string) (*net.UnixConn, error) {
	var (
		err error
		prm = o.getSocketPerm()
		grp = o.getSocketGroup()
		lis *net.UnixConn
		loc *net.UnixAddr
	)

	if loc, err = net.ResolveUnixAddr(libptc.NetworkUnixGram.Code(), uxf); err != nil {
		return nil, err
	} else if lis, err = net.ListenUnixgram(libptc.NetworkUnixGram.Code(), loc); err != nil {
		return nil, err
	} else if lis == nil {
		return nil, os.ErrNotExist
	} else if _, err = os.Stat(uxf); err != nil {
		_ = lis.Close()
		return nil, err
	} else if err = os.Chmod(uxf, prm.FileMode()); err != nil {
		_ = lis.Close()
		return nil, err
	} else if err = os.Chown(uxf, syscall.Getuid(), grp); err != nil {
		_ = lis.Close()
		return nil, err
	}

	o.fctInfoSrv("starting listening socket '%s %s'", libptc.NetworkUnixGram.String(), uxf)
	return lis, nil
}
