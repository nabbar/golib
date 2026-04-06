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
	// defaultPattern is the default template for rendering errors with a code.
	// It uses fmt.Sprintf format: [Error #%d] %s (code, message).
	defaultPattern      = "[Error #%d] %s"

	// defaultPatternTrace is the default template for rendering errors with a code and a stack trace.
	// It uses fmt.Sprintf format: [Error #%d] %s (%s) (code, message, trace).
	defaultPatternTrace = "[Error #%d] %s (%s)"
)

// SetDefaultPattern defines the global pattern for formatting error strings with codes.
// The pattern must be a fmt-compatible string with exactly two %v/%d/%s placeholders for (code, message).
func SetDefaultPattern(pattern string) {
	defaultPattern = pattern
}

// GetDefaultPattern returns the currently active global pattern for error code formatting.
func GetDefaultPattern() string {
	return defaultPattern
}

// SetDefaultPatternTrace defines the global pattern for formatting error strings with codes and traces.
// The pattern must be a fmt-compatible string with exactly three placeholders for (code, message, trace).
func SetDefaultPatternTrace(patternTrace string) {
	defaultPatternTrace = patternTrace
}

// GetDefaultPatternTrace returns the currently active global pattern for error/code/trace formatting.
func GetDefaultPatternTrace() string {
	return defaultPatternTrace
}

// SetTracePathFilter allows manual customization of the package path filter used in stack traces.
// This is useful if the automatic detection of the module root fails or if you want to mask certain paths.
func SetTracePathFilter(path string) {
	filterPkg = path
}
