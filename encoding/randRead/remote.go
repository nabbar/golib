/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package randRead

import (
	"fmt"
	"io"
	"sync/atomic"
)

type remote struct {
	f *atomic.Value // FuncRemote
	r *atomic.Value // io.ReadCloser
}

func (o *remote) reader(p []byte) (n int, err error) {
	if i := o.r.Load(); i == nil {
		return 0, io.EOF
	} else if r, k := i.(io.ReadCloser); !k {
		return 0, io.EOF
	} else {
		return r.Read(p)
	}
}

func (o *remote) readRemote() error {
	if o.f == nil {
		return fmt.Errorf("invalid reader")
	} else if i := o.f.Load(); i == nil {
		return fmt.Errorf("invalid reader")
	} else if f, ok := i.(FuncRemote); !ok {
		return fmt.Errorf("invalid reader")
	} else if r, err := f(); err != nil {
		return err
	} else {
		l := o.r.Swap(r)

		if v, k := l.(io.Closer); k && v != nil {
			_ = v.Close()
		}

		return nil
	}
}

func (o *remote) Read(p []byte) (n int, err error) {
	n, err = o.reader(p)
	if n > 0 {
		return n, nil
	}

	if err = o.readRemote(); err != nil {
		return 0, err
	}

	n, err = o.reader(p)
	if n > 0 {
		return n, nil
	}

	return 0, err
}

func (o *remote) Close() error {
	if i := o.r.Load(); i == nil {
		return nil
	} else if r, k := i.(io.ReadCloser); !k {
		return nil
	} else {
		return r.Close()
	}
}
