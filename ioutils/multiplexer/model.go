/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package multiplexer

import (
	"fmt"
	"io"
	"sync"
	"time"

	libcbr "github.com/fxamacker/cbor/v2"
)

type mux[T comparable] struct {
	d *sync.Map
	r io.Reader
	w io.Writer
}

func (o *mux[T]) Add(key T, fct FuncWrite) {
	if o.r == nil {
		return
	}

	o.d.Store(key, fct)
}

func (o *mux[T]) Read(p []byte) (n int, err error) {
	if o.r == nil {
		return 0, fmt.Errorf("invalid stream io reader")
	}

	var (
		d = libcbr.NewDecoder(o.r)
		c T
		m Message[T]
	)

	if err = d.Decode(&m); err != nil {
		return 0, err
	} else if m.Stream == c {
		return 0, fmt.Errorf("invalid stream key '%s'", m.Stream)
	} else if len(m.Message) < 1 {
		return 0, nil
	} else if i, l := o.d.Load(m.Stream); !l {
		return 0, fmt.Errorf("invalid read func for stream key '%s'", m.Stream)
	} else if i == nil {
		return 0, fmt.Errorf("invalid read func for stream key '%s'", m.Stream)
	} else if f, k := i.(FuncWrite); !k {
		return 0, fmt.Errorf("invalid read func for stream key '%s'", m.Stream)
	} else if f == nil {
		return 0, fmt.Errorf("invalid read func for stream key '%s'", m.Stream)
	} else {
		return f(m.Message)
	}
}

func (o *mux[T]) write(key T, p []byte) (n int, err error) {
	if o.w == nil {
		return 0, fmt.Errorf("invalid stream io writer")
	}

	var (
		d = libcbr.NewEncoder(o.w)
		c T
		m = Message[T]{
			Stream:  key,
			Message: p,
		}
	)

	defer func() {
		time.Sleep(5 * time.Millisecond)
	}()

	n = len(p)

	if m.Stream == c {
		return 0, fmt.Errorf("invalid stream key '%s'", m.Stream)
	} else if len(m.Message) < 1 {
		return 0, nil
	} else if err = d.Encode(m); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

func (o *mux[T]) Writer(key T) io.Writer {
	return &writer[T]{
		f: func(p []byte) (n int, err error) {
			return o.write(key, p)
		},
	}
}
