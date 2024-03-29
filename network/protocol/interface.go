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

func Parse(str string) NetworkProtocol {
	switch {
	case strings.EqualFold(NetworkTCP.Code(), str):
		return NetworkTCP
	case strings.EqualFold(NetworkTCP4.Code(), str):
		return NetworkTCP4
	case strings.EqualFold(NetworkTCP6.Code(), str):
		return NetworkTCP6
	case strings.EqualFold(NetworkUDP.Code(), str):
		return NetworkUDP
	case strings.EqualFold(NetworkUnix.Code(), str):
		return NetworkUnix
	case strings.EqualFold(NetworkUnixGram.Code(), str):
		return NetworkUnixGram
	default:
		return NetworkEmpty
	}
}

func ParseBytes(p []byte) NetworkProtocol {
	return Parse(string(p))
}

func ParseInt64(val int64) NetworkProtocol {
	var v NetworkProtocol

	if val > int64(math.MaxUint8) {
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
	case NetworkUnix:
		return NetworkUnix
	case NetworkUnixGram:
		return NetworkUnixGram
	default:
		return NetworkEmpty
	}
}
