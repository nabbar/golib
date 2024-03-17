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
	"bufio"
	"bytes"
	"errors"
	"io"

	libcbr "github.com/fxamacker/cbor/v2"
	enchex "github.com/nabbar/golib/encoding/hexa"
)

type dmux struct {
	r *bufio.Reader
	d byte
	m map[rune]io.Writer
}

// Read defines a Read method for the dmux type in Go.
// It reads data from dmux, handles error conditions, and performs data manipulation before writing the data to a specified channel.
func (o *dmux) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInvalidInstance
	}

	var (
		b []byte
		d = data{}
		h = enchex.New()
	)

	b, err = o.r.ReadBytes(o.d)
	b = bytes.TrimSuffix(b, []byte{o.d})

	if err != nil {
		return 0, err
	}

	if len(b) < 1 {
		return 0, nil
	}

	err = libcbr.Unmarshal(b, &d)
	if err != nil {
		return 0, err
	}

	if len(d.D) < 1 {
		return 0, nil
	} else if p, err = h.Decode(d.D); err != nil {
		return 0, err
	}

	if len(o.m) < 1 {
		return 0, ErrInvalidChannel
	} else if w, k := o.m[d.K]; !k {
		return 0, ErrInvalidChannel
	} else {
		return w.Write(p)
	}
}

// Copy defines a method Copy for a type dmux
// that continuously reads data into a buffer p in chunks of 32KB until an error occurs,
// and then returns that error except the error is io.EOF.
func (o *dmux) Copy() error {
	var (
		e error
		p = make([]byte, 32*1024)
	)

	for e == nil {
		clear(p)
		_, e = o.Read(p)
	}

	if errors.Is(e, io.EOF) {
		e = nil
	}

	return e
}

// NewChannel defines a method NewChannel for a type dmux,
// which adds a new channel to the dmux by mapping the given key to the provided writer in the dmux's map.
func (o *dmux) NewChannel(key rune, w io.Writer) {
	if o == nil {
		return
	}

	o.m[key] = w
}
