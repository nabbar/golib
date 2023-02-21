/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package network

import (
	"math"
	"strconv"
	"strings"
)

type NetworkMode uint8

const (
	NetworkTCP NetworkMode = iota
	NetworkTCPIPv4
	NetworkTCPIPv6
	NetworkUnixSocket
)

func NetworkModeFromString(str string) NetworkMode {
	switch strings.ToLower(str) {
	case NetworkTCPIPv4.String():
		return NetworkTCPIPv4
	case NetworkTCPIPv6.String():
		return NetworkTCPIPv6
	case NetworkUnixSocket.String():
		return NetworkUnixSocket
	}

	return NetworkTCP
}

func NetworkModeFromInt(i int64) NetworkMode {
	if i > math.MaxUint8 {
		return NetworkTCP
	}

	switch NetworkMode(i) {
	case NetworkTCPIPv4:
		return NetworkTCPIPv4
	case NetworkTCPIPv6:
		return NetworkTCPIPv6
	case NetworkUnixSocket:
		return NetworkUnixSocket
	}

	return NetworkTCP
}

func (n NetworkMode) String() string {
	switch n {
	case NetworkTCPIPv4:
		return "tcp4"
	case NetworkTCPIPv6:
		return "tcp6"
	case NetworkUnixSocket:
		return "unix"
	case NetworkTCP:
		return "tcp"
	}

	return NetworkTCP.String()
}

func (n NetworkMode) Int() int64 {
	return int64(n)
}

func (n NetworkMode) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(n.String())+2)
	b = append(b, '"')
	b = append(b, []byte(n.String())...)
	b = append(b, '"')
	return b, nil
}

func (n *NetworkMode) UnmarshalJSON(data []byte) error {
	var (
		e   error
		i   int64
		a   NetworkMode
		str string
	)

	str = string(data)

	if str == "null" {
		*n = NetworkTCP
		return nil
	}

	if strings.HasPrefix(str, "\"") || strings.HasSuffix(str, "\"") {
		if str, e = strconv.Unquote(str); e != nil {
			return e
		}
	}

	if i, e = strconv.ParseInt(str, 10, 8); e != nil {
		*n = NetworkModeFromString(str)
		return nil
	} else if a = NetworkModeFromInt(i); a != NetworkTCP {
		*n = a
		return nil
	} else {
		*n = NetworkModeFromString(str)
		return nil
	}
}
