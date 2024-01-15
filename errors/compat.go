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

package errors

var (
	defaultPattern      = "[Error #%d] %s"
	defaultPatternTrace = "[Error #%d] %s (%s)"
)

// GetDefaultPatternTrace define the pattern to be used for string of error with code.
// The pattern is fmt pattern with 2 inputs in order : code, message.
func SetDefaultPattern(pattern string) {
	defaultPattern = pattern
}

// GetDefaultPattern return the current pattern used for string of error with code.
// The pattern is fmt pattern with 2 inputs in order : code, message.
func GetDefaultPattern() string {
	return defaultPattern
}

// SetDefaultPatternTrace define the pattern to be used for string of error with code and trace.
// The pattern is fmt pattern with 3 inputs in order : code, message, trace.
func SetDefaultPatternTrace(patternTrace string) {
	defaultPatternTrace = patternTrace
}

// GetDefaultPatternTrace return the current pattern used for string of error with code and trace.
// The pattern is fmt pattern with 3 inputs in order : code, message, trace.
func GetDefaultPatternTrace() string {
	return defaultPatternTrace
}

// SetTracePathFilter customize the filter apply to filepath on trace.
func SetTracePathFilter(path string) {
	filterPkg = path
}
