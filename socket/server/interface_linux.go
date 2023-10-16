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
)

func New(handler libsck.Handler, proto libptc.NetworkProtocol, sizeBufferRead int32, address string, perm os.FileMode) (libsck.Server, error) {
	switch proto {
	case libptc.NetworkUnix:
		if strings.EqualFold(runtime.GOOS, "linux") {
			s := scksrx.New(handler, sizeBufferRead)
			s.RegisterSocket(address, perm)
			return s, nil
		}
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		s := scksrt.New(handler, sizeBufferRead)
		e := s.RegisterServer(address)
		return s, e
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		s := scksru.New(handler, sizeBufferRead)
		e := s.RegisterServer(address)
		return s, e
	}

	return nil, fmt.Errorf("invalid server protocol")
}