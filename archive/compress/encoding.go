/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package compress

import (
	"bytes"
	"encoding/json"
	"strings"
)

// MarshalText implements encoding.TextMarshaler.
// It returns the lowercase string representation of the algorithm.
// This is used for text-based serialization formats like YAML or TOML.
func (a Algorithm) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It parses the algorithm from a text representation.
// The parsing is case-insensitive and trims whitespace, quotes, and apostrophes.
// Unknown or invalid values result in None being set.
func (a *Algorithm) UnmarshalText(b []byte) error {
	*a = None

	s := strings.TrimSpace(string(b))
	s = strings.Trim(s, "\"")
	s = strings.Trim(s, "'")
	s = strings.TrimSpace(s)

	switch {
	case strings.EqualFold(s, Gzip.String()):
		*a = Gzip
	case strings.EqualFold(s, Bzip2.String()):
		*a = Bzip2
	case strings.EqualFold(s, LZ4.String()):
		*a = LZ4
	case strings.EqualFold(s, XZ.String()):
		*a = XZ
	default:
		*a = None
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
// It returns the lowercase string representation of the algorithm as a JSON string.
// The None algorithm is marshaled as JSON null for semantic correctness.
//
// Examples:
//   - Gzip   → "gzip"
//   - None   → null
//   - Bzip2  → "bzip2"
func (a Algorithm) MarshalJSON() ([]byte, error) {
	if a.IsNone() {
		return []byte("null"), nil
	}
	return append(append([]byte{'"'}, []byte(a.String())...), '"'), nil
}

// UnmarshalJSON implements json.Unmarshaler.
// It parses the algorithm from a JSON string value.
// JSON null is interpreted as None. The parsing delegates to UnmarshalText
// for the actual string-to-algorithm conversion.
//
// Examples:
//   - "gzip"  → Gzip
//   - null    → None
//   - "lz4"   → LZ4
func (a *Algorithm) UnmarshalJSON(b []byte) error {
	var s string

	if n := []byte("null"); bytes.Equal(b, n) {
		*a = None
		return nil
	} else if err := json.Unmarshal(b, &s); err != nil {
		return err
	} else {
		return a.UnmarshalText([]byte(s))
	}
}
