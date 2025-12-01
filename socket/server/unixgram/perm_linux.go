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

// getListen creates and configures the Unix datagram socket listener.
//
// The function:
//  1. Sets umask to apply configured permissions
//  2. Creates the Unix datagram socket with net.ListenUnixgram() (SOCK_DGRAM)
//  3. Verifies and corrects file permissions with os.Chmod() if needed
//  4. Changes file group ownership with os.Chown() if needed
//  5. Invokes the server info callback with startup message
//
// The umask is temporarily modified to ensure correct permissions are applied
// during socket creation, then restored to the original value.
//
// Parameters:
//   - uxf: Unix socket file path
//   - adr: Resolved UnixAddr for the socket
//
// Returns:
//   - *net.UnixConn: The active Unix datagram connection (connectionless)
//   - error: Any error during socket creation or configuration
//
// Unlike connection-oriented Unix sockets, this creates a SOCK_DGRAM socket
// which operates in connectionless datagram mode.
//
// This is an internal helper called by Listen().
//
// See syscall.Umask, os.Chmod, os.Chown for permission management.
// See net.ListenUnixgram for datagram socket creation.
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
