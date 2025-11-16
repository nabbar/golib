/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package tlsmode

// String returns the string representation of the TLSMode.
//
// Returns:
//   - TLSNone → "" (empty string)
//   - TLSStartTLS → "starttls"
//   - TLSStrictTLS → "tls"
//
// This is the canonical string format used for configuration files,
// DSN strings, and serialization. The returned string can be parsed
// back using the Parse function.
//
// Example:
//
//	mode := TLSStartTLS
//	fmt.Println(mode.String()) // Output: starttls
func (t TLSMode) String() string {
	switch t {
	case TLSStrictTLS:
		return "tls"
	case TLSStartTLS:
		return "starttls"
	default:
		return ""
	}
}

// Uint returns the TLSMode as a uint8.
//
// Returns:
//   - TLSNone → 0
//   - TLSStartTLS → 1
//   - TLSStrictTLS → 2
func (t TLSMode) Uint() uint8 {
	return uint8(t)
}

// Uint32 returns the TLSMode as a uint32.
//
// Returns:
//   - TLSNone → 0
//   - TLSStartTLS → 1
//   - TLSStrictTLS → 2
func (t TLSMode) Uint32() uint32 {
	return uint32(t)
}

// Uint64 returns the TLSMode as a uint64.
//
// Returns:
//   - TLSNone → 0
//   - TLSStartTLS → 1
//   - TLSStrictTLS → 2
//
// This value can be parsed back using ParseUint64.
func (t TLSMode) Uint64() uint64 {
	return uint64(t)
}

// Int returns the TLSMode as an int.
//
// Returns:
//   - TLSNone → 0
//   - TLSStartTLS → 1
//   - TLSStrictTLS → 2
func (t TLSMode) Int() int {
	return int(t)
}

// Int32 returns the TLSMode as an int32.
//
// Returns:
//   - TLSNone → 0
//   - TLSStartTLS → 1
//   - TLSStrictTLS → 2
func (t TLSMode) Int32() int32 {
	return int32(t)
}

// Int64 returns the TLSMode as an int64.
//
// Returns:
//   - TLSNone → 0
//   - TLSStartTLS → 1
//   - TLSStrictTLS → 2
//
// This value can be parsed back using ParseInt64.
func (t TLSMode) Int64() int64 {
	return int64(t)
}

// Float32 returns the TLSMode as a float32.
//
// Returns:
//   - TLSNone → 0.0
//   - TLSStartTLS → 1.0
//   - TLSStrictTLS → 2.0
func (t TLSMode) Float32() float32 {
	return float32(t)
}

// Float64 returns the TLSMode as a float64.
//
// Returns:
//   - TLSNone → 0.0
//   - TLSStartTLS → 1.0
//   - TLSStrictTLS → 2.0
//
// This value can be parsed back using ParseFloat64.
func (t TLSMode) Float64() float64 {
	return float64(t)
}
