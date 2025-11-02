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
 */

package console

import (
	"io"
)

// BuffPrintf writes formatted, colored text to an io.Writer.
// This is useful for:
//   - Writing colored output to buffers for testing
//   - Writing colored text to files
//   - Capturing colored output for processing
//   - Non-terminal output that supports ANSI codes
//
// Parameters:
//   - buff: The io.Writer to write to (returns error if nil)
//   - format: Printf-style format string
//   - args: Arguments for format string
//
// Returns:
//   - int: Number of bytes written
//   - error: ErrorColorBufUndefined if buff is nil, or write error
//
// Example:
//
//	var buf bytes.Buffer
//	n, err := console.ColorPrint.BuffPrintf(&buf, "Hello %s", "World")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Wrote %d bytes: %s\n", n, buf.String())
func (c ColorType) BuffPrintf(buff io.Writer, format string, args ...interface{}) (int, error) {
	if buff == nil {
		return 0, ErrorColorBufUndefined.Error(nil)
	}

	return GetColor(c).Fprintf(buff, format, args...)
}
