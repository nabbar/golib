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

// padTimes repeats the given string n times and returns the concatenated result.
// Internal helper for padding operations.
func padTimes(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// PadLeft pads a string on the left (right-aligns the text).
// Uses UTF-8 rune counting to correctly handle multi-byte characters.
//
// Parameters:
//   - str: The string to pad
//   - len: The desired total length in runes (not bytes)
//   - pad: The padding string (typically " " or "0")
//
// Returns:
//   - Padded string with length 'len' runes
//
// UTF-8 Support: Correctly handles emojis, CJK characters, and multi-byte Unicode.
//
// Example:
//
//	PadLeft("text", 10, " ")      // Returns "      text"
//	PadLeft("5", 5, "0")          // Returns "00005"
//	PadLeft("你好", 10, " ")       // Returns "        你好" (correctly counts 2 runes)
func PadLeft(str string, len int, pad string) string {
	return padTimes(pad, len-utf8.RuneCountInString(str)) + str
}

// PadRight pads a string on the right (left-aligns the text).
// Uses UTF-8 rune counting to correctly handle multi-byte characters.
//
// Parameters:
//   - str: The string to pad
//   - len: The desired total length in runes (not bytes)
//   - pad: The padding string (typically " ")
//
// Returns:
//   - Padded string with length 'len' runes
//
// UTF-8 Support: Correctly handles emojis, CJK characters, and multi-byte Unicode.
//
// Example:
//
//	PadRight("text", 10, " ")     // Returns "text      "
//	PadRight("Name", 20, " ")     // Returns "Name                "
//	PadRight("🌍", 5, " ")         // Returns "🌍    " (correctly counts 1 rune)
func PadRight(str string, len int, pad string) string {
	return str + padTimes(pad, len-utf8.RuneCountInString(str))
}

// PadCenter centers a string with padding on both sides.
// Uses UTF-8 rune counting to correctly handle multi-byte characters.
// If padding cannot be distributed evenly, the right side gets one extra pad character.
//
// Parameters:
//   - str: The string to center
//   - len: The desired total length in runes (not bytes)
//   - pad: The padding string (typically " ", "=", or "-")
//
// Returns:
//   - Centered string with length 'len' runes
//
// UTF-8 Support: Correctly handles emojis, CJK characters, and multi-byte Unicode.
//
// Example:
//
//	PadCenter("text", 10, " ")    // Returns "   text   "
//	PadCenter("Title", 20, "=")   // Returns "=======Title========"
//	PadCenter("你好", 10, " ")     // Returns "    你好    " (correctly counts 2 runes)
func PadCenter(str string, len int, pad string) string {
	nbr := len - utf8.RuneCountInString(str)
	lft := int(math.Floor(float64(nbr) / 2))
	rgt := nbr - lft

	return padTimes(pad, lft) + str + padTimes(pad, rgt)
}

// PrintTabf prints formatted text with hierarchical indentation.
// Each indentation level adds 2 spaces. Uses ColorPrint for output.
//
// Parameters:
//   - tablLevel: The indentation level (0 = no indent, 1 = 2 spaces, 2 = 4 spaces, etc.)
//   - format: Printf-style format string
//   - args: Arguments for format string
//
// Output: Writes directly to stdout with ColorPrint colors.
//
// Use Cases:
//   - Configuration display
//   - Hierarchical data structures
//   - Nested lists
//   - Tree output
//
// Example:
//
//	console.PrintTabf(0, "Root\n")
//	console.PrintTabf(1, "Child 1\n")
//	console.PrintTabf(2, "Grandchild\n")
//	console.PrintTabf(1, "Child 2\n")
//
// Output:
//
//	Root
//	  Child 1
//	    Grandchild
//	  Child 2
func PrintTabf(tablLevel int, format string, args ...interface{}) {
	ColorPrint.Printf(strings.Repeat("  ", tablLevel)+format, args...)
}
