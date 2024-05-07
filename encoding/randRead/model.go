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
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
)

type remote struct {
	f *atomic.Value
	r io.ReadCloser
}

func (o *remote) readReader(p []byte) (n int, err error) {
	if o.r == nil {
		return 0, io.EOF
	}

	if n, err = o.r.Read(p); err != nil {
		return n, err
	} else {
		return n, nil
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
		if o.r != nil {
			_ = o.r.Close()
		}

		o.r = r
		return nil
	}
}

func (o *remote) Read(p []byte) (n int, err error) {
	if n, err = o.readReader(p); err != nil && !errors.Is(err, io.EOF) {
		return n, err
	}

	if err = o.readRemote(); err != nil {
		return 0, err
	}

	return o.readReader(p)
}

func (o *remote) Close() error {
	if o.r != nil {
		_ = o.r.Close()
	}

	return nil
}

type prnd struct {
	b *bufio.Reader
	r *remote
}

func (o *prnd) Read(p []byte) (n int, err error) {
	if o.b != nil {
		return o.b.Read(p)
	}

	if o.r != nil {
		o.b = bufio.NewReader(o.r)
		return o.b.Read(p)
	} else {
		return 0, fmt.Errorf("invalid reader")
	}
}

func (o *prnd) Close() error {
	if o.b != nil {
		o.b.Reset(nil)
	}

	if o.r != nil {
		_ = o.r.Close()
	}

	return nil
}
