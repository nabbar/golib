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

package console

import (
	"math"
	"strings"
	"unicode/utf8"
)

func padTimes(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

func PadLeft(str string, len int, pad string) string {
	return padTimes(pad, len-utf8.RuneCountInString(str)) + str
}

func PadRight(str string, len int, pad string) string {
	return str + padTimes(pad, len-utf8.RuneCountInString(str))
}

func PadCenter(str string, len int, pad string) string {
	nbr := len - utf8.RuneCountInString(str)
	lft := int(math.Floor(float64(nbr) / 2))
	rgt := nbr - lft

	return padTimes(pad, lft) + str + padTimes(pad, rgt)
}

func PrintTabf(tablLevel int, format string, args ...interface{}) {
	ColorPrint.Printf(strings.Repeat("  ", tablLevel)+format, args...)
}

// @TODO : remove function
// deprecated: replaced by PrintTabf
// nolint: goprintffuncname
func PrintTab(tablLevel int, format string, args ...interface{}) {
	PrintTabf(tablLevel, format, args...)
}
