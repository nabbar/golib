/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package curves

import (
	"crypto/tls"
	"strings"
)

func (v Curves) String() string {
	switch v {
	case X25519:
		return "X25519"
	case P256:
		return "P256"
	case P384:
		return "P384"
	case P521:
		return "P521"
	default:
		return ""
	}
}

func (v Curves) Code() string {
	return strings.ToLower(v.String())
}

func (v Curves) CurveID() tls.CurveID {
	switch v {
	case X25519:
		return tls.X25519
	case P256:
		return tls.CurveP256
	case P384:
		return tls.CurveP384
	case P521:
		return tls.CurveP521
	default:
		return 0
	}
}

func (v Curves) TLS() tls.CurveID {
	return v.CurveID()
}

func (v Curves) Uint16() uint16 {
	return uint16(v.CurveID())
}

func (v Curves) Uint() uint {
	return uint(v.CurveID())
}

func (v Curves) Uint32() uint32 {
	return uint32(v.CurveID())
}

func (v Curves) Uint64() uint64 {
	return uint64(v.CurveID())
}

func (v Curves) Int() int {
	return int(v.CurveID())
}

func (v Curves) Int32() int32 {
	return int32(v.CurveID())
}

func (v Curves) Int64() int64 {
	return int64(v.CurveID())
}
