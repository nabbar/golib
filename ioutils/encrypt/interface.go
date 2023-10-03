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

package encrypt

import (
	"io"

	libcrp "github.com/nabbar/golib/crypt"
)

type Encrypt interface {
	io.Writer
}

type Decrypt interface {
	io.Reader
}

func NewEncrypt(w io.Writer, hex bool, key [32]byte, nonce [12]byte) (Encrypt, error) {
	if crp, err := libcrp.New(key, nonce); err != nil {
		return nil, err
	} else {
		return &enc{
			c: crp,
			h: hex,
			w: w,
		}, nil
	}
}

func NewDecrypt(r io.Reader, hex bool, key [32]byte, nonce [12]byte) (Decrypt, error) {
	if crp, err := libcrp.New(key, nonce); err != nil {
		return nil, err
	} else {
		return &dec{
			c: crp,
			h: hex,
			r: r,
		}, nil
	}
}
