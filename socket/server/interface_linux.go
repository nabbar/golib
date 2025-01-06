//go:build linux
// +build linux

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

package server

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	scksrt "github.com/nabbar/golib/socket/server/tcp"
	scksru "github.com/nabbar/golib/socket/server/udp"
	scksrx "github.com/nabbar/golib/socket/server/unix"
	sckgrm "github.com/nabbar/golib/socket/server/unixgram"
)

// New creates a new server based on the provided network protocol.
//
// Parameters:
// - upd: a Update Connection function or nil
// - handler: the handler for the server
// - delim: the delimiter to use to separate messages
// - proto: the network protocol to use
// - sizeBufferRead: the size of the buffer for reading
// - address: the address to bind the server to
// - perm: the file mode permissions for the socket, not applicable for non unix
// - gid: the group ID for the socket permissions, not applicable for non unix
// Return type(s):
// - libsck.Server: the created server
// - error: an error if any occurred during server creation
func New(upd libsck.UpdateConn, handler libsck.Handler, proto libptc.NetworkProtocol, address string, perm os.FileMode, gid int32) (libsck.Server, error) {
	switch proto {
	case libptc.NetworkUnix:
		if strings.EqualFold(runtime.GOOS, "linux") {
			s := scksrx.New(upd, handler)
			e := s.RegisterSocket(address, perm, gid)
			return s, e
		}
	case libptc.NetworkUnixGram:
		if strings.EqualFold(runtime.GOOS, "linux") {
			s := sckgrm.New(upd, handler)
			e := s.RegisterSocket(address, perm, gid)
			return s, e
		}
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		s := scksrt.New(upd, handler)
		e := s.RegisterServer(address)
		return s, e
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		s := scksru.New(upd, handler)
		e := s.RegisterServer(address)
		return s, e
	}

	return nil, fmt.Errorf("invalid server protocol")
}
