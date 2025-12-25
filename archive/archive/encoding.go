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

package archive

import (
	"bytes"
	"encoding/json"
	"strings"
)

// MarshalText implements encoding.TextMarshaler interface.
// It converts the Algorithm to its string representation as a byte slice.
// This is used for text-based serialization formats (YAML, TOML, etc.).
//
// Returns:
//   - []byte: the algorithm name as bytes ("tar", "zip", or "none")
//   - error: always nil (this method never fails)
func (a Algorithm) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
// It parses a text representation and sets the algorithm accordingly.
// The parsing is case-insensitive and trims whitespace and quotes.
//
// Supported values:
//   - "tar" or "TAR" → Tar
//   - "zip" or "ZIP" → Zip
//   - any other value → None
//
// Parameters:
//   - b: the text bytes to parse
//
// Returns:
//   - error: always nil (this method never fails, defaults to None for invalid input)
func (a *Algorithm) UnmarshalText(b []byte) error {
	*a = None

	s := strings.Trim(string(b), "\"")
	s = strings.Trim(s, "'")
	s = strings.TrimSpace(s)

	switch {
	case strings.EqualFold(s, Tar.String()):
		*a = Tar
	case strings.EqualFold(s, Zip.String()):
		*a = Zip
	default:
		*a = None
	}

	return nil
}

// MarshalJSON implements json.Marshaler interface.
// It converts the Algorithm to JSON representation.
// None is marshaled as null, other algorithms as their string names in quotes.
//
// Examples:
//   - Tar → "tar"
//   - Zip → "zip"
//   - None → null
//
// Returns:
//   - []byte: the JSON representation
//   - error: always nil (this method never fails)
func (a Algorithm) MarshalJSON() ([]byte, error) {
	if a.IsNone() {
		return []byte("null"), nil
	}
	return append(append([]byte{'"'}, []byte(a.String())...), '"'), nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
// It parses JSON data and sets the algorithm accordingly.
// Accepts both string values and null.
//
// Supported JSON values:
//   - "tar" → Tar
//   - "zip" → Zip
//   - null → None
//   - invalid string → None
//
// Parameters:
//   - b: the JSON bytes to parse
//
// Returns:
//   - error: returns error only if the JSON structure is invalid
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
