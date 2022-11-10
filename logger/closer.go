/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package logger

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

type _Closer interface {
	Add(clo io.Closer)
	Get() []io.Closer
	Len() int
	Clean()
	Clone() _Closer
	Close() error
}

func _NewCloser() _Closer {
	return &closer{
		m: sync.Mutex{},
		c: make([]io.Closer, 0),
	}
}

type closer struct {
	m sync.Mutex
	c []io.Closer
}

func (c *closer) Add(clo io.Closer) {
	lst := append(c.Get(), clo)

	c.m.Lock()
	defer c.m.Unlock()

	c.c = lst
}

func (c *closer) Get() []io.Closer {
	res := make([]io.Closer, 0)

	c.m.Lock()
	defer c.m.Unlock()

	if len(c.c) > 0 {
		for _, i := range c.c {
			if i == nil {
				continue
			}
			res = append(res, i)
		}
	}

	return res
}

func (c *closer) Len() int {
	c.m.Lock()
	defer c.m.Unlock()
	return len(c.c)
}

func (c *closer) Clean() {
	c.m.Lock()
	defer c.m.Unlock()

	c.c = make([]io.Closer, 0)
}

func (c *closer) Clone() _Closer {
	o := _NewCloser()
	for _, i := range c.Get() {
		o.Add(i)
	}
	return o
}

func (c *closer) Close() error {
	var e = make([]string, 0)

	for _, i := range c.Get() {
		if err := i.Close(); err != nil {
			e = append(e, err.Error())
		}
	}

	return fmt.Errorf("%s", strings.Join(e, ", "))
}
