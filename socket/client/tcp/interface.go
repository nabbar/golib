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

package tcp

import (
	"net"
	"sync/atomic"

	libptc "github.com/nabbar/golib/network/protocol"
	libsiz "github.com/nabbar/golib/size"
	libsck "github.com/nabbar/golib/socket"
)

type ClientTCP interface {
	libsck.Client
}

func New(buffSizeRead libsiz.Size, address string) (ClientTCP, error) {
	var (
		a = new(atomic.Value)
		s = new(atomic.Int32)
	)

	if len(address) < 1 {
		return nil, ErrAddress
	} else if _, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), address); err != nil {
		return nil, err
	}

	a.Store(address)
	s.Store(buffSizeRead.Int32())

	return &cli{
		a: a,
		s: s,
		e: new(atomic.Value),
		i: new(atomic.Value),
		c: new(atomic.Value),
	}, nil
}
