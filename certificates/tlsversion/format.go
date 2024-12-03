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

package tlsversion

import (
	"crypto/tls"
	"strings"
)

func (v Version) String() string {
	switch v {
	case VersionTLS10:
		return "TLS 1.0"
	case VersionTLS11:
		return "TLS 1.1"
	case VersionTLS12:
		return "TLS 1.2"
	case VersionTLS13:
		return "TLS 1.3"
	default:
		return ""
	}
}

func (v Version) Code() string {
	s := strings.ToLower(v.String())
	s = strings.Replace(s, " ", "_", -1)
	return s
}

func (v Version) TLS() uint16 {
	return v.Uint16()
}

func (v Version) Uint16() uint16 {
	switch v {
	case VersionTLS10:
		return tls.VersionTLS10
	case VersionTLS11:
		return tls.VersionTLS11
	case VersionTLS12:
		return tls.VersionTLS12
	case VersionTLS13:
		return tls.VersionTLS13
	default:
		return 0
	}
}

func (v Version) Uint() uint {
	return uint(v.Uint16())
}

func (v Version) Uint32() uint32 {
	return uint32(v.Uint16())
}

func (v Version) Uint64() uint64 {
	return uint64(v.Uint16())
}

func (v Version) Int() int {
	return int(v.Uint16())
}

func (v Version) Int32() int32 {
	return int32(v.Uint16())
}

func (v Version) Int64() int64 {
	return int64(v.Uint16())
}
