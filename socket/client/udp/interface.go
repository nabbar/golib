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

package udp

import (
	"net/url"
	"strconv"
	"sync/atomic"

	libsck "github.com/nabbar/golib/socket"
)

type ClientUDP interface {
	libsck.Client
}

func New(address string) (ClientUDP, error) {
	var (
		a = new(atomic.Value)
		u = &url.URL{
			Host: address,
		}
	)

	if len(u.Hostname()) < 1 {
		return nil, ErrHostName
	} else if len(u.Port()) < 1 {
		return nil, ErrHostPort
	} else if i, e := strconv.Atoi(u.Port()); e != nil {
		return nil, e
	} else if i < 1 || i > 65534 {
		return nil, ErrHostPort
	} else {
		a.Store(u)
	}

	return &cltu{
		a:  a,
		e:  new(atomic.Value),
		i:  new(atomic.Value),
		tr: new(atomic.Value),
		tw: new(atomic.Value),
	}, nil
}
