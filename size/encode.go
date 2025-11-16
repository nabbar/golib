/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package size

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface.
//
// The Size value is marshaled as a human-readable string (e.g., "1.50 MB")
// rather than a raw byte count. This makes JSON output more readable and
// easier to work with in configuration files.
//
// Example:
//
//	size := size.ParseUint64(1572864) // 1.5 MB
//	json, _ := json.Marshal(size)
//	fmt.Println(string(json)) // Output: "\"1.50 MB\""
func (s Size) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// The method accepts both string representations (e.g., "10MB") and numeric
// values (e.g., 10485760). String values are parsed using the Parse function.
//
// Example:
//
//	var size size.Size
//	json.Unmarshal([]byte(`"10MB"`), &size)
//	fmt.Println(size.MegaBytes()) // Output: 10
func (s *Size) UnmarshalJSON(p []byte) error {
	return s.unmarshall(p)
}

// MarshalYAML implements the yaml.Marshaler interface.
//
// The Size value is marshaled as a human-readable string (e.g., "1.50 MB")
// for better readability in YAML configuration files.
//
// Example:
//
//	size := size.ParseUint64(1572864) // 1.5 MB
//	yaml, _ := yaml.Marshal(size)
//	fmt.Println(string(yaml)) // Output: "1.50 MB\n"
func (s Size) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// The method accepts string representations (e.g., "10MB") from YAML files.
// String values are parsed using the Parse function.
//
// Example:
//
//	var size size.Size
//	yaml.Unmarshal([]byte("10MB"), &size)
//	fmt.Println(size.MegaBytes()) // Output: 10
func (s *Size) UnmarshalYAML(value *yaml.Node) error {
	return s.unmarshall([]byte(value.Value))
}

// MarshalTOML implements the TOML marshaler interface.
//
// The Size value is marshaled as a human-readable string (e.g., "1.50 MB")
// for better readability in TOML configuration files.
//
// Example:
//
//	size := size.ParseUint64(1572864) // 1.5 MB
//	// In TOML: max_size = "1.50 MB"
func (s Size) MarshalTOML() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalTOML implements the TOML unmarshaler interface.
//
// The method accepts both []byte and string representations (e.g., "10MB")
// from TOML files. Values are parsed using the Parse function.
//
// Example:
//
//	var size size.Size
//	// From TOML: max_size = "10MB"
//	toml.Unmarshal(data, &config) // size will be 10MB
func (s *Size) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return s.unmarshall(p)
	}
	if p, k := i.(string); k {
		return s.unmarshall([]byte(p))
	}
	return fmt.Errorf("size: value not in valid format")
}

// MarshalText implements the encoding.TextMarshaler interface.
//
// The Size value is marshaled as a human-readable string (e.g., "1.50 MB").
// This method is used by various encoding packages that work with text.
//
// Example:
//
//	size := size.ParseUint64(1572864) // 1.5 MB
//	text, _ := size.MarshalText()
//	fmt.Println(string(text)) // Output: "1.50 MB"
func (s Size) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// The method accepts text representations (e.g., "10MB") and parses them
// using the Parse function.
//
// Example:
//
//	var size size.Size
//	size.UnmarshalText([]byte("10MB"))
//	fmt.Println(size.MegaBytes()) // Output: 10
func (s *Size) UnmarshalText(p []byte) error {
	return s.unmarshall(p)
}

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library.
//
// The Size value is marshaled as a human-readable string (e.g., "1.50 MB")
// in CBOR format. This is useful for compact binary serialization while
// maintaining human readability when decoded.
//
// See also: github.com/fxamacker/cbor
func (s Size) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(s.String())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// The method decodes CBOR data containing a string representation of a size
// (e.g., "10MB") and parses it into a Size value.
//
// See also: github.com/fxamacker/cbor
func (s *Size) UnmarshalCBOR(p []byte) error {
	var b string
	if e := cbor.Unmarshal(p, &b); e != nil {
		return e
	} else {
		return s.unmarshall([]byte(b))
	}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// This method uses CBOR encoding internally to provide compact binary
// serialization of Size values.
//
// Example:
//
//	size := size.ParseUint64(1048576) // 1 MB
//	data, _ := size.MarshalBinary()
//	// data can be stored or transmitted in binary format
func (s Size) MarshalBinary() ([]byte, error) {
	return s.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// This method uses CBOR decoding internally to deserialize binary data
// into a Size value.
//
// Example:
//
//	var size size.Size
//	size.UnmarshalBinary(data)
//	fmt.Println(size.MegaBytes()) // Output: 1
func (s *Size) UnmarshalBinary(p []byte) error {
	return s.UnmarshalCBOR(p)
}
