/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package sectun

import (
	"crypto/rand"
	"io"

	libsck "github.com/nabbar/golib/socket"
)

type Algo uint8

const (
	AlgoAES256 Algo = iota + 1
	AlgoChaChaPoly
)

type Curves uint8

const (
	CurveX25519 Curves = iota + 1
	CurveMLKEM768
	CurveMLKEM1024
)

type Hash uint8

const (
	HashSha2_256 = iota + 1
	HashSha2_512
	HashSha3_256
	HashSha3_384
	HashSha3_512
)

type SecureTunnel interface {
	io.Reader
	io.Writer
	io.Closer

	Reader() io.ReadCloser
	Writer() io.WriteCloser
}

func New(rnd func() io.Reader, a Algo, c Curves, h Hash, r io.Reader, w io.Writer, readFirst bool) (SecureTunnel, error) {
	if rnd == nil {
		rnd = func() io.Reader {
			return rand.Reader
		}
	}

	if a < 1 || a > 2 {
		return nil, ErrUnsupportedAlgo
	}

	if c < 1 || c > 3 {
		return nil, ErrUnsupportedCurve
	}

	if h < 1 || h > 5 {
		return nil, ErrUnsupportedHash
	}

	m := &mdl{
		n: rnd,
		r: r,
		w: w,
		a: nil,
	}

	if err := m.handshake(a, c, h, readFirst); err != nil {
		return nil, err
	}

	return m, nil
}

func NewContext(x libsck.Context, rnd func() io.Reader, a Algo, c Curves, h Hash, readFirst bool) (libsck.Context, error) {
	if t, e := New(rnd, a, c, h, x, x, readFirst); e != nil {
		return nil, e
	} else {
		return &ctx{
			x: x,
			t: t,
		}, nil
	}
}
