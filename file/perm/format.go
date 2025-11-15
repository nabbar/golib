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

// FileMode returns the file mode represented by the Perm as an os.FileMode.
//
// It panics if the Perm is not a valid file permission.
//
// Example:
// p := Perm(420)
// fmt.Println(p.FileMode()) // Output: -rw-r--
func (p Perm) FileMode() os.FileMode {
	return os.FileMode(p.Uint32())
}

// String returns a string representation of the Perm as an octal number.
//
// The returned string is in the format of "%#o", where "#" is the octal
// representation of the Perm value.
//
// Example:
// p := Perm(420)
// fmt.Println(p.String()) // Output: "0644"
func (p Perm) String() string {
	return fmt.Sprintf("%#o", p.Uint64())
}

// Int64 returns the Perm value as an int64.
//
// If the Perm value exceeds the maximum value of an int64, the function returns
// math.MaxInt64.
//
// Example:
// p := Perm(420)
// fmt.Println(p.Int64()) // Output: 420
func (p Perm) Int64() int64 {
	if uint64(p) > math.MaxInt64 {
		// overflow
		return math.MaxInt64
	}

	return int64(p)
}

// Int32 returns the Perm value as an int32.
//
// If the Perm value exceeds the maximum value of an int32, the function returns
// math.MaxInt32.
//
// Example:
// p := Perm(420)
// fmt.Println(p.Int32()) // Output: 420
func (p Perm) Int32() int32 {
	if i := uint64(p); i > uint64(math.MaxInt32) {
		// overflow
		return math.MaxInt32
	} else {
		return int32(i)
	}
}

// Int returns the Perm value as an int.
//
// If the Perm value exceeds the maximum value of an int, the function returns
// math.MaxInt.
//
// Example:
// p := Perm(420)
// fmt.Println(p.Int()) // Output: 420
func (p Perm) Int() int {
	if uint64(p) > math.MaxInt {
		// overflow
		return math.MaxInt
	}

	return int(p)
}

// Uint64 returns the Perm value as a uint64.
//
// Example:
// p := Perm(420)
// fmt.Println(p.Uint64()) // Output: 420
func (p Perm) Uint64() uint64 {
	return uint64(p)
}

// Uint32 returns the Perm value as a uint32.
//
// If the Perm value exceeds the maximum value of a uint32, the function returns
// math.MaxUint32.
//
// Example:
// p := Perm(420)
// fmt.Println(p.Uint32()) // Output: 420
func (p Perm) Uint32() uint32 {
	if uint64(p) > math.MaxUint32 {
		// overflow
		return math.MaxUint32
	}

	return uint32(p)
}

// Uint returns the Perm value as a uint.
//
// If the Perm value exceeds the maximum value of a uint, the function returns
// math.MaxUint.
//
// Example:
// p := Perm(420)
// fmt.Println(p.Uint()) // Output: 420
func (p Perm) Uint() uint {
	if uint64(p) > math.MaxUint {
		// overflow
		return math.MaxUint
	}

	return uint(p)
}
