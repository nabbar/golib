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

package status

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements json.Marshaler interface.
// It encodes the Status as a JSON string.
//
// Example:
//
//	s := status.OK
//	b, err := s.MarshalJSON()
//	// b contains: "OK"
func (o Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// UnmarshalJSON implements json.Unmarshaler interface.
// It decodes a JSON string or number to a Status value.
// Unrecognized values default to KO.
//
// Example:
//
//	var s status.Status
//	err := json.Unmarshal([]byte(`"OK"`), &s)
//	// s is now status.OK
func (o *Status) UnmarshalJSON(p []byte) error {
	return o.unmarshall(p)
}

// MarshalYAML implements yaml.Marshaler interface.
// It encodes the Status as a YAML string.
//
// Example:
//
//	s := status.Warn
//	y, err := s.MarshalYAML()
//	// y is "Warn"
func (o Status) MarshalYAML() (interface{}, error) {
	return o.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface.
// It decodes a YAML node value to a Status.
//
// Example:
//
//	var s status.Status
//	node := &yaml.Node{Value: "OK"}
//	err := s.UnmarshalYAML(node)
//	// s is now status.OK
func (o *Status) UnmarshalYAML(value *yaml.Node) error {
	return o.unmarshall([]byte(value.Value))
}

// MarshalTOML implements toml.Marshaler interface.
// It encodes the Status as a TOML string.
func (o Status) MarshalTOML() ([]byte, error) {
	return json.Marshal(o.String())
}

// UnmarshalTOML implements toml.Unmarshaler interface.
// It decodes a TOML value (string or byte slice) to a Status.
// Returns an error if the value is not a string or byte slice.
func (o *Status) UnmarshalTOML(i interface{}) error {
	if b, k := i.([]byte); k {
		return o.unmarshall(b)
	}

	if b, k := i.(string); k {
		return o.unmarshall([]byte(b))
	}

	return fmt.Errorf("duration: value not in valid format")
}

// MarshalText implements encoding.TextMarshaler interface.
// It encodes the Status as text.
//
// Example:
//
//	s := status.OK
//	b, err := s.MarshalText()
//	// b contains: []byte("OK")
func (o Status) MarshalText() ([]byte, error) {
	return []byte(o.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
// It decodes text to a Status value.
//
// Example:
//
//	var s status.Status
//	err := s.UnmarshalText([]byte("Warn"))
//	// s is now status.Warn
func (o *Status) UnmarshalText(p []byte) error {
	return o.unmarshall(p)
}

// MarshalCBOR implements cbor.Marshaler interface.
// It encodes the Status as CBOR.
//
// Example:
//
//	s := status.OK
//	b, err := s.MarshalCBOR()
//	// b contains CBOR-encoded "OK"
func (o Status) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.String())
}

// UnmarshalCBOR implements cbor.Unmarshaler interface.
// It decodes CBOR data to a Status value.
//
// Example:
//
//	var s status.Status
//	data, _ := cbor.Marshal("OK")
//	err := s.UnmarshalCBOR(data)
//	// s is now status.OK
func (o *Status) UnmarshalCBOR(p []byte) error {
	var s string
	if e := cbor.Unmarshal(p, &s); e != nil {
		return e
	} else {
		return o.unmarshall([]byte(s))
	}
}

// unmarshall is an internal helper function that parses byte data to Status.
// It always succeeds by defaulting to KO for unrecognized values.
func (o *Status) unmarshall(val []byte) error {
	tmp := ParseByte(val)
	*o = tmp
	return nil
}
