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

package protocol

import (
	"math"
	"strconv"
	"strings"
)

type NetworkProtocol uint8

const (
	NetworkEmpty NetworkProtocol = iota
	NetworkUnix
	NetworkTCP
	NetworkTCP4
	NetworkTCP6
	NetworkUDP
	NetworkUDP4
	NetworkUDP6
	NetworkIP
	NetworkIP4
	NetworkIP6
	NetworkUnixGram
)

var ptcLkp = map[NetworkProtocol]string{
	NetworkTCP:      "tcp",
	NetworkTCP4:     "tcp4",
	NetworkTCP6:     "tcp6",
	NetworkUDP:      "udp",
	NetworkUDP4:     "udp4",
	NetworkUDP6:     "udp6",
	NetworkIP:       "ip",
	NetworkIP4:      "ip4",
	NetworkIP6:      "ip6",
	NetworkUnix:     "unix",
	NetworkUnixGram: "unixgram",
}

func Parse(str string) NetworkProtocol {
	str = strings.ToLower(str)
	str = strings.TrimSpace(str)
	str = strings.Trim(str, "\"")

	if s, e := strconv.Unquote(str); e == nil {
		str = s
	}

	for k, v := range ptcLkp {
		if v == str {
			return k
		}
	}

	return NetworkEmpty
}

func ParseBytes(p []byte) NetworkProtocol {
	return Parse(string(p))
}

func ParseUint64(val uint64) NetworkProtocol {
	var v NetworkProtocol

	if val > uint64(math.MaxUint8) {
		v = NetworkProtocol(math.MaxUint8)
	} else {
		v = NetworkProtocol(val)
	}

	switch v {
	case NetworkTCP:
		return NetworkTCP
	case NetworkTCP4:
		return NetworkTCP4
	case NetworkTCP6:
		return NetworkTCP6
	case NetworkUDP:
		return NetworkUDP
	case NetworkUDP4:
		return NetworkUDP4
	case NetworkUDP6:
		return NetworkUDP6
	case NetworkIP:
		return NetworkIP
	case NetworkIP4:
		return NetworkIP4
	case NetworkIP6:
		return NetworkIP6
	case NetworkUnix:
		return NetworkUnix
	case NetworkUnixGram:
		return NetworkUnixGram
	default:
		return NetworkEmpty
	}
}

func ParseInt64(val int64) NetworkProtocol {
	if val < 0 {
		return NetworkEmpty // Explicitly reject negative values
	} else {
		return ParseUint64(uint64(val))
	}
}
