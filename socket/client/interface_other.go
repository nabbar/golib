//go:build !linux
// +build !linux

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

package client

import (
	"fmt"

	libptc "github.com/nabbar/golib/network/protocol"
	libsiz "github.com/nabbar/golib/size"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/tcp"
	sckclu "github.com/nabbar/golib/socket/client/udp"
)

func New(proto libptc.NetworkProtocol, sizeBufferRead libsiz.Size, address string) (libsck.Client, error) {
	switch proto {
	case libptc.NetworkTCP, libptc.NetworkTCP4, libptc.NetworkTCP6:
		return sckclt.New(sizeBufferRead, address)
	case libptc.NetworkUDP, libptc.NetworkUDP4, libptc.NetworkUDP6:
		return sckclu.New(sizeBufferRead, address)
	}

	return nil, fmt.Errorf("invalid client protocol")
}
