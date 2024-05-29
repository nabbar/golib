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
	"io"
	"sync/atomic"
)

type FuncRemote func() (io.ReadCloser, error)

/*
New return an interface for a random reader from remote source.
New returns io.ReadCloser, will buffer data from remote source.
The input function should return an io.ReadCloser and an error.
The http request or any other implementation could be used.
*/
func New(fct FuncRemote) io.ReadCloser {
	if fct == nil {
		return nil
	}

	f := new(atomic.Value)
	f.Store(fct)

	r := &remote{
		r: new(atomic.Value),
		f: f,
	}

	return &prnd{
		b: new(atomic.Value),
		r: r,
	}
}
