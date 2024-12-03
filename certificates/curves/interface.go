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
	"math"
	"strings"
)

type Curves uint16

const (
	Unknown Curves = iota
	X25519         = Curves(tls.X25519)
	P256           = Curves(tls.CurveP256)
	P384           = Curves(tls.CurveP384)
	P521           = Curves(tls.CurveP521)
)

func List() []Curves {
	return []Curves{
		X25519,
		P256,
		P384,
		P521,
	}
}

func ListString() []string {
	var res = make([]string, 0)
	for _, c := range List() {
		res = append(res, c.String())
	}
	return res
}

func Parse(s string) Curves {
	s = strings.ToLower(s)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "'", "", -1)
	s = strings.Replace(s, "x", "", -1)
	s = strings.Replace(s, "X", "", -1)
	s = strings.Replace(s, "p", "", -1)
	s = strings.Replace(s, "P", "", -1)
	s = strings.Replace(s, ".", "", -1)
	s = strings.Replace(s, "-", "", -1)
	s = strings.Replace(s, "_", "", -1)
	s = strings.Replace(s, " ", "", -1)
	s = strings.TrimSpace(s)

	switch {
	case strings.EqualFold(s, "25519"):
		return X25519
	case strings.EqualFold(s, "256"):
		return P256
	case strings.EqualFold(s, "384"):
		return P384
	case strings.EqualFold(s, "521"):
		return P521
	default:
		return Unknown
	}
}

func ParseInt(d int) Curves {
	if d > math.MaxUint16 {
		d = math.MaxUint16
	} else if d < 1 {
		d = 0
	}

	switch tls.CurveID(d) {
	case tls.X25519:
		return X25519
	case tls.CurveP256:
		return P256
	case tls.CurveP384:
		return P384
	case tls.CurveP521:
		return P521
	default:
		return Unknown
	}
}

func Check(curves uint16) bool {
	if c := ParseInt(int(curves)); c == Unknown {
		return false
	}
	return true
}

func parseBytes(p []byte) Curves {
	return Parse(string(p))
}
