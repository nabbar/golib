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

package httpcli

import "strings"

type Network uint8

const (
	NetworkTCP Network = iota
	NetworkUDP
	NetworkUnix
)

func GetNetworkFromString(str string) Network {
	switch {
	case strings.EqualFold(NetworkUDP.Code(), str):
		return NetworkUDP
	case strings.EqualFold(NetworkUnix.Code(), str):
		return NetworkUnix
	default:
		return NetworkTCP
	}
}

func (n Network) String() string {
	switch n {
	case NetworkUDP:
		return "UDP"
	case NetworkUnix:
		return "unix"
	default:
		return "TCP"
	}
}

func (n Network) Code() string {
	return strings.ToLower(n.String())
}
