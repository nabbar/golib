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

package mux

import (
	"io"
	"sync"

	libcbr "github.com/fxamacker/cbor/v2"
	enchex "github.com/nabbar/golib/encoding/hexa"
)

type mux struct {
	w  io.Writer
	d  byte
	mu sync.Mutex
}

// NewChannel This Function defines a method NewChannel for a type mux in Go. It returns an io.Writer interface.
// The method creates a function that takes a byte slice, marshals it into a data struct, appends a field o.d to the byte slice, then writes the resulting byte slice to an io.Writer o.w.
// If successful, it returns the number of bytes written.
func (o *mux) NewChannel(key rune) io.Writer {
	if o == nil {
		return &wrt{
			f: func(p []byte) (n int, e error) {
				return 0, ErrInvalidInstance
			},
		}
	}

	var crp = enchex.New()

	return &wrt{
		f: func(p []byte) (n int, e error) {
			n = len(p)

			var d = data{
				K: key,
				D: crp.Encode(p),
			}

			clear(p)

			p, e = libcbr.Marshal(&d)
			p = append(p, o.d)

			if e != nil {
				return 0, e
			}

			o.mu.Lock()
			_, e = o.w.Write(p)
			o.mu.Unlock()

			if e != nil {
				return 0, e
			}

			return n, nil
		},
	}
}
