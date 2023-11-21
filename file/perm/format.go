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

package perm

import (
	"fmt"
	"math"
	"os"
)

func (p Perm) FileMode() os.FileMode {
	return os.FileMode(p.Uint64())
}

func (p Perm) String() string {
	return fmt.Sprintf("%#o", p.Uint64())
}

func (p Perm) Int64() int64 {
	if uint64(p) > math.MaxInt64 {
		// overflow
		return math.MaxInt64
	}

	return int64(p)
}

func (p Perm) Int32() int32 {
	if uint64(p) > math.MaxInt32 {
		// overflow
		return math.MaxInt32
	}

	return int32(p)
}

func (p Perm) Int() int {
	if uint64(p) > math.MaxInt {
		// overflow
		return math.MaxInt
	}

	return int(p)
}

func (p Perm) Uint64() uint64 {
	return uint64(p)
}

func (p Perm) Uint32() uint32 {
	if uint64(p) > math.MaxUint32 {
		// overflow
		return math.MaxUint32
	}

	return uint32(p)
}

func (p Perm) Uint() uint {
	if uint64(p) > math.MaxUint {
		// overflow
		return math.MaxUint
	}

	return uint(p)
}
