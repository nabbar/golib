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

package mapCloser

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
)

type closer struct {
	c *atomic.Bool
	f func() // Context Func Cancel
	i *atomic.Uint64
	x libctx.Config[uint64]
}

func (o *closer) idx() uint64 {
	return o.i.Load()
}

func (o *closer) idxInc() uint64 {
	o.i.Add(1)
	return o.idx()
}

func (o *closer) Add(clo ...io.Closer) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	} else if o.x.Err() != nil {
		return
	}

	for _, c := range clo {
		o.x.Store(o.idxInc(), c)
	}
}

func (o *closer) Get() []io.Closer {
	var res = make([]io.Closer, 0)

	if o == nil {
		return res
	} else if o.x == nil {
		return res
	} else if o.x.Err() != nil {
		return res
	}

	o.x.Walk(func(key uint64, val interface{}) bool {
		if val == nil {
			return true
		}
		if v, k := val.(io.Closer); !k {
			return true
		} else {
			res = append(res, v)
			return true
		}
	})
	return res
}

func (o *closer) Len() int {
	i := o.idx()

	if i > math.MaxInt {
		// overflow
		return math.MaxInt
	} else {
		return int(i)
	}
}

func (o *closer) Len64() uint64 {
	return o.idx()
}

func (o *closer) Clean() {
	if o == nil {
		return
	} else if o.x == nil {
		return
	} else if o.x.Err() != nil {
		return
	}

	o.i.Store(0)
	o.x.Clean()
}

func (o *closer) Clone() Closer {
	if o == nil {
		return nil
	} else if o.x == nil {
		return nil
	} else if o.x.Err() != nil {
		return nil
	}

	i := new(atomic.Uint64)
	i.Store(o.idx())

	c := new(atomic.Bool)
	c.Store(o.c.Load())

	return &closer{
		c: c,
		f: o.f,
		i: i,
		x: o.x.Clone(nil),
	}
}

func (o *closer) Close() error {
	var e = make([]string, 0)

	if o == nil {
		return fmt.Errorf("not initialized")
	}

	o.c.Store(true)

	if o.f != nil {
		defer o.f()
	}

	if o.x == nil {
		return fmt.Errorf("not initialized")
	} else if o.x.Err() != nil {
		return o.x.Err()
	}

	o.x.Walk(func(key uint64, val interface{}) bool {
		if c, k := val.(io.Closer); !k {
			return true
		} else if err := c.Close(); err != nil {
			e = append(e, err.Error())
		}
		return true
	})

	if len(e) > 0 {
		return fmt.Errorf("%s", strings.Join(e, ", "))
	}

	return nil
}
