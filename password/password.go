/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package password

import (
	"crypto/rand"
	"math/big"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz,;:!?./*%^$&\"'(-_)=+~#{[|`\\^@]}ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	// If we have 93 chars, that means 7 bits code a letter index.
	// So 63 random bits can designate 63/7 = 9 different letter indices.
	// Let's use all those 10.
	letterIdxBits  = 7                    // 6 bits to represent a letter index
	letterIdxMask  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax   = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	loopRandMaxLen = 10
)

func randIdx() int {
	size := int64(len(letterBytes))

	for n := 0; n < 100; n++ {

		if i, e := rand.Int(rand.Reader, big.NewInt(size+1)); e != nil {
			return 0
		} else {
			j := i.Int64()

			if j > 0 && j < size {
				return int(j)
			}
		}
	}

	return 0
}

func Generate(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i := n - 1; i >= 0; {
		b[i] = letterBytes[randIdx()]
		i--
	}

	return string(b)
}
