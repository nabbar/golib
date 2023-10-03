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

package crypt

import (
	"crypto/cipher"
	"encoding/hex"
)

type crt struct {
	a cipher.AEAD
	n []byte
}

func (o *crt) Encode(p []byte) []byte {
	if len(p) < 1 {
		return make([]byte, 0)
	}
	return o.a.Seal(nil, o.n, p, nil)
}

func (o *crt) EncodeHex(p []byte) []byte {
	if len(p) < 1 {
		return make([]byte, 0)
	}
	return []byte(hex.EncodeToString(o.Encode(p)))
}

func (o *crt) Decode(p []byte) ([]byte, error) {
	if len(p) < 1 {
		return make([]byte, 0), nil
	}
	return o.a.Open(nil, o.n, p, nil)
}

func (o *crt) DecodeHex(p []byte) ([]byte, error) {
	if len(p) < 1 {
		return make([]byte, 0), nil
	} else if dec, err := hex.DecodeString(string(p)); err != nil {
		return nil, err
	} else {
		return o.Decode(dec)
	}
}
