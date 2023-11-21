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

package config

import (
	"os"

	libdur "github.com/nabbar/golib/duration"
	libptc "github.com/nabbar/golib/network/protocol"
	libsiz "github.com/nabbar/golib/size"
	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server"
)

type ServerConfig struct {
	Network      libptc.NetworkProtocol ``
	Address      string
	PermFile     os.FileMode
	BuffSizeRead libsiz.Size
	TimeoutRead  libdur.Duration
	TimeoutWrite libdur.Duration
}

func (o ServerConfig) New(handler libsck.Handler) (libsck.Server, error) {
	s, e := scksrv.New(handler, o.Network, o.BuffSizeRead, o.Address, o.PermFile)

	if e != nil {
		s.SetReadTimeout(o.TimeoutRead.Time())
		s.SetWriteTimeout(o.TimeoutWrite.Time())
	}

	return s, e
}
