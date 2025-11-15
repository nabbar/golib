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

// String returns a string representation of the NetworkProtocol.
// If the NetworkProtocol is found, the corresponding string value is returned.
// Otherwise, the return string is empty
func (n NetworkProtocol) String() string {
	if s, k := ptcLkp[n]; k {
		return s
	}
	return ""
}

// Code returns a string representation of the NetworkProtocol as a code.
// The returned string is equal to the value returned by String().
// This function is alias of String to be consistently with other custom const type.
func (n NetworkProtocol) Code() string {
	return n.String()
}

// Int returns the NetworkProtocol as an int value.
// If the NetworkProtocol is found, the corresponding int value is returned.
// Otherwise, the return int value is zero.
func (n NetworkProtocol) Int() int {
	if _, k := ptcLkp[n]; k {
		return int(n)
	}
	return 0
}

// Int64 returns the NetworkProtocol as an int64 value.
// If the NetworkProtocol is found, the corresponding int64 value is returned.
// Otherwise, the return int64 value is zero.
func (n NetworkProtocol) Int64() int64 {
	if _, k := ptcLkp[n]; k {
		return int64(n)
	}
	return 0
}

// Uint returns the NetworkProtocol as an uint value.
// If the NetworkProtocol is found, the corresponding uint value is returned.
// Otherwise, the return uint value is zero.
func (n NetworkProtocol) Uint() uint {
	if _, k := ptcLkp[n]; k {
		return uint(n)
	}
	return 0
}

// Uint64 returns the NetworkProtocol as a uint64 value.
// If the NetworkProtocol is found, the corresponding uint64 value is returned.
// Otherwise, the return uint64 value is zero.
func (n NetworkProtocol) Uint64() uint64 {
	if _, k := ptcLkp[n]; k {
		return uint64(n)
	}
	return 0
}
