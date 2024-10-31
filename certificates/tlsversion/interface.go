/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package tlsversion

import (
	"crypto/tls"
	"strings"
)

type Version int

const (
	VersionUnknown Version = iota
	VersionTLS10           = Version(tls.VersionTLS10)
	VersionTLS11           = Version(tls.VersionTLS11)
	VersionTLS12           = Version(tls.VersionTLS12)
	VersionTLS13           = Version(tls.VersionTLS13)
)

func List() []Version {
	return []Version{
		VersionTLS13,
		VersionTLS12,
		VersionTLS11,
		VersionTLS10,
	}
}

func ListHigh() []Version {
	return []Version{
		VersionTLS13,
		VersionTLS12,
	}
}

func Parse(s string) Version {
	s = strings.ToLower(s)
	s = strings.Replace(s, "tls", "", -1)
	s = strings.Replace(s, "ssl", "", -1)
	s = strings.Replace(s, ".", "", -1)
	s = strings.Replace(s, "-", "", -1)
	s = strings.Replace(s, "_", "", -1)
	s = strings.Replace(s, " ", "", -1)
	s = strings.TrimSpace(s)

	switch {
	case strings.EqualFold(s, "1"):
		return VersionTLS10
	case strings.EqualFold(s, "10"):
		return VersionTLS10
	case strings.EqualFold(s, "11"):
		return VersionTLS11
	case strings.EqualFold(s, "12"):
		return VersionTLS12
	case strings.EqualFold(s, "13"):
		return VersionTLS13
	default:
		return VersionUnknown
	}
}

func ParseInt(d int) Version {
	switch d {
	case tls.VersionTLS10:
		return VersionTLS10
	case tls.VersionTLS11:
		return VersionTLS11
	case tls.VersionTLS12:
		return VersionTLS12
	case tls.VersionTLS13:
		return VersionTLS13
	default:
		return VersionUnknown
	}
}

func parseBytes(p []byte) Version {
	return Parse(string(p))
}
