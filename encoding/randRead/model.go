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
	"bytes"
	"fmt"
	"sync/atomic"
)

type prnd struct {
	b *atomic.Value // *bufio.Reader
	r *remote
}

func (o *prnd) buf() (*bufio.Reader, error) {
	if i := o.b.Load(); i != nil {
		if b, k := i.(*bufio.Reader); k {
			return b, nil
		}
	}

	if o.r == nil {
		return nil, fmt.Errorf("invalid reader")
	}

	b := bufio.NewReader(o.r)
	o.b.Store(b)
	return b, nil
}

func (o *prnd) reset() {
	l := o.b.Swap(bufio.NewReader(bytes.NewReader(make([]byte, 0))))

	if b, k := l.(*bufio.Reader); k {
		b.Reset(bytes.NewReader(make([]byte, 0)))
	}
}

func (o *prnd) Read(p []byte) (n int, err error) {
	if b, e := o.buf(); e != nil {
		return 0, e
	} else {
		return b.Read(p)
	}
}

func (o *prnd) Close() error {
	o.reset()

	if o.r == nil {
		return fmt.Errorf("invalid reader")
	}

	return o.r.Close()
}
