/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context

import (
	"context"

	libatm "github.com/nabbar/golib/atomic"
)

type ccx[T comparable] struct {
	m libatm.Map[T]
	x context.Context
}

func (c *ccx[T]) Clone(ctx context.Context) Config[T] {
	if c.Err() != nil {
		c.Clean()
		return nil
	} else if ctx == nil {
		ctx = c.GetContext()
	}

	n := &ccx[T]{
		m: libatm.NewMapAny[T](),
		x: ctx,
	}

	c.m.Range(func(key T, value any) bool {
		n.Store(key, value)
		return true
	})

	return n
}

func (c *ccx[T]) Merge(cfg Config[T]) bool {
	if c.Err() != nil {
		c.Clean()
		return false
	} else if cfg == nil {
		return false
	}

	cfg.Walk(func(k T, v interface{}) bool {
		c.m.Store(k, v)
		return true
	})

	return true
}
