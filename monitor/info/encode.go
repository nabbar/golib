/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package info

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Encode defines the interface for an encoding model that can be
// represented as a string or a byte slice.
type Encode interface {
	// String returns a single-line string representation of the info data.
	// It typically formats as "Name (key1: val1, key2: val2)".
	String() string
	// Bytes returns the byte slice representation of the info data,
	// which is equivalent to the result of String().
	Bytes() []byte
}

// encMod is the internal struct used for encoding the Info object.
// It holds the name and info data for marshaling.
type encMod struct {
	Name string
	Info map[string]interface{}
}

// String implements the Encode interface, providing a compact, single-line
// representation of the info data.
func (e *encMod) String() string {
	if i := e.stringInfo(); len(i) > 0 {
		return e.stringClean(fmt.Sprintf("%s (%s)", e.Name, i))
	} else {
		return e.stringClean(e.Name)
	}
}

// Bytes implements the Encode interface. It returns the same single-line
// representation as String(), but as a byte slice.
func (e *encMod) Bytes() []byte {
	return []byte(e.String())
}

// stringInfo concatenates all key-value pairs from the Info map into a
// single comma-separated string.
func (e *encMod) stringInfo() string {
	var buf = bytes.NewBuffer(make([]byte, 0))
	for n, i := range e.Info {
		buf.WriteString(fmt.Sprintf("%s: %v,", n, i)) // nolint
	}
	return strings.Trim(strings.TrimSpace(buf.String()), ",")
}

// stringClean removes newline and carriage return characters from a string,
// ensuring the output is a single line.
func (e *encMod) stringClean(str string) string {
	str = strings.Replace(str, "\n", " ", -1) // nolint
	str = strings.Replace(str, "\r", "", -1)  // nolint
	return str
}

// getEncodeModel creates and returns an encoding model (encMod) populated
// with the current name and info data from the Info object.
func (o *inf) getEncodeModel() Encode {
	return &encMod{
		Name: o.Name(),
		Info: o.Info(),
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
// It provides a compact, human-readable, single-line text representation
// of the Info object.
func (o *inf) MarshalText() (text []byte, err error) {
	return o.getEncodeModel().Bytes(), nil
}

// MarshalJSON implements the json.Marshaler interface.
// It provides a structured JSON representation of the Info object,
// including its name and info data.
func (o *inf) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.getEncodeModel())
}
