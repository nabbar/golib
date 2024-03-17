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

package multi

import (
	"io"
	"sync"
	"sync/atomic"
)

type mlt struct {
	i *atomic.Value
	d *atomic.Value
	c *atomic.Int64
	w sync.Map
}

func (o *mlt) AddWriter(w ...io.Writer) {
	for _, wrt := range w {
		if wrt != nil {
			o.w.Store(o.c.Add(1), wrt)
		}
	}

	var l = make([]io.Writer, 0)

	o.w.Range(func(key, value any) bool {
		if value != nil {
			if v, k := value.(io.Writer); k {
				l = append(l, v)
			}
		}
		return true
	})

	if len(l) < 1 {
		o.d.Store(io.Discard)
	} else if len(l) == 1 {
		o.d.Store(l[0])
	} else {
		o.d.Store(io.MultiWriter(l...))
	}
}

func (o *mlt) Clean() {
	o.d.Store(io.Discard)

	var keys = make([]any, 0)

	o.w.Range(func(key, value any) bool {
		keys = append(keys, key)
		return true
	})

	for _, k := range keys {
		o.w.Delete(k)
	}

	o.c.Store(0)
}

func (o *mlt) SetInput(i io.ReadCloser) {
	if o == nil {
		return
	} else if i == nil {
		i = DiscardCloser{}
	}

	o.i.Store(i)
}

func (o *mlt) Writer() io.Writer {
	return o.d.Load().(io.Writer)
}

func (o *mlt) Reader() io.ReadCloser {
	return o.i.Load().(io.ReadCloser)
}

func (o *mlt) Copy() (n int64, err error) {
	return io.Copy(o.Writer(), o.Reader())
}

func (o *mlt) Read(p []byte) (n int, err error) {
	if i := o.i.Load(); i == nil {
		return 0, ErrInstance
	} else if in, ok := i.(io.Reader); !ok {
		return 0, ErrInstance
	} else {
		return in.Read(p)
	}
}

func (o *mlt) Write(p []byte) (n int, err error) {
	if i := o.d.Load(); i == nil {
		return 0, ErrInstance
	} else if v, k := i.(io.Writer); !k {
		return 0, ErrInstance
	} else {
		return v.Write(p)
	}
}

func (o *mlt) WriteString(s string) (n int, err error) {
	if i := o.d.Load(); i == nil {
		return 0, ErrInstance
	} else if v, k := i.(io.Writer); !k {
		return 0, ErrInstance
	} else {
		return io.WriteString(v, s)
	}
}

func (o *mlt) Close() error {
	if i := o.i.Load(); i == nil {
		return ErrInstance
	} else if in, ok := i.(io.ReadCloser); !ok {
		return ErrInstance
	} else {
		return in.Close()
	}
}
