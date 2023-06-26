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
	default:
		return NetworkEmpty
	}
}

func ParseBytes(p []byte) NetworkProtocol {
	return Parse(string(p))
}

func SizeFromInt64(val int64) NetworkProtocol {
	if val > int64(math.MaxUint8) {
		return NetworkProtocol(math.MaxUint8)
	}

	return NetworkProtocol(uint8(val))
}
