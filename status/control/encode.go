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

package control

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

// unmarshall is an internal helper that parses bytes into a Mode.
// It always returns nil error for compatibility with unmarshaling interfaces.
func (c *Mode) unmarshall(val []byte) error {
	*c = ParseBytes(val)
	return nil
}

// MarshalJSON implements json.Marshaler interface.
//
// The Mode is marshaled as a JSON string using its String() representation.
//
// Example:
//
//	mode := control.Must
//	data, _ := json.Marshal(mode)
//	fmt.Println(string(data)) // Output: "Must"
func (c Mode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements json.Unmarshaler interface.
//
// The Mode is unmarshaled from a JSON string. The parsing is case-insensitive.
// If the JSON value is not a valid Mode string, it defaults to Ignore.
//
// Example:
//
//	var mode control.Mode
//	json.Unmarshal([]byte(`"must"`), &mode)
//	fmt.Println(mode) // Output: Must
func (c *Mode) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	return c.unmarshall([]byte(s))
}

// MarshalYAML implements yaml.Marshaler interface.
//
// The Mode is marshaled as a YAML string using its String() representation.
//
// Example:
//
//	mode := control.Must
//	data, _ := yaml.Marshal(mode)
//	fmt.Println(string(data)) // Output: Must
func (c Mode) MarshalYAML() (interface{}, error) {
	return c.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface.
//
// The Mode is unmarshaled from a YAML string. The parsing is case-insensitive.
// If the YAML value is not a valid Mode string, it defaults to Ignore.
//
// Example:
//
//	var mode control.Mode
//	yaml.Unmarshal([]byte("must"), &mode)
//	fmt.Println(mode) // Output: Must
func (c *Mode) UnmarshalYAML(value *yaml.Node) error {
	return c.unmarshall([]byte(value.Value))
}

// MarshalTOML implements toml.Marshaler interface.
//
// The Mode is marshaled as a TOML string using its JSON representation.
//
// Example:
//
//	mode := control.Must
//	data, _ := mode.MarshalTOML()
//	fmt.Println(string(data)) // Output: "Must"
func (c Mode) MarshalTOML() ([]byte, error) {
	return c.MarshalJSON()
}

// UnmarshalTOML implements toml.Unmarshaler interface.
//
// The Mode is unmarshaled from a TOML value which can be either a byte slice
// or a string. The parsing is case-insensitive.
// If the value is not a valid Mode string or type, it returns an error.
//
// Example:
//
//	var mode control.Mode
//	mode.UnmarshalTOML("must")
//	fmt.Println(mode) // Output: Must
func (c *Mode) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return c.unmarshall(p)
	}
	if p, k := i.(string); k {
		return c.unmarshall([]byte(p))
	}
	return fmt.Errorf("size: value not in valid format")
}

// MarshalText implements encoding.TextMarshaler interface.
//
// The Mode is marshaled as plain text using its String() representation.
// This is useful for text-based protocols and formats.
//
// Example:
//
//	mode := control.Must
//	data, _ := mode.MarshalText()
//	fmt.Println(string(data)) // Output: Must
func (c Mode) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
//
// The Mode is unmarshaled from plain text. The parsing is case-insensitive.
// If the text is not a valid Mode string, it defaults to Ignore.
//
// Example:
//
//	var mode control.Mode
//	mode.UnmarshalText([]byte("must"))
//	fmt.Println(mode) // Output: Must
func (c *Mode) UnmarshalText(bytes []byte) error {
	return c.unmarshall(bytes)
}

// MarshalCBOR implements cbor.Marshaler interface.
//
// The Mode is marshaled as a CBOR string using its String() representation.
// CBOR (Concise Binary Object Representation) is a binary data serialization
// format based on JSON.
//
// Example:
//
//	mode := control.Must
//	data, _ := mode.MarshalCBOR()
//	// data contains CBOR-encoded "Must"
//
// See also: github.com/fxamacker/cbor for CBOR encoding/decoding.
func (c Mode) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(c.String())
}

// UnmarshalCBOR implements cbor.Unmarshaler interface.
//
// The Mode is unmarshaled from CBOR-encoded data. The parsing is case-insensitive.
// If the CBOR value is not a valid Mode string, it defaults to Ignore.
//
// Example:
//
//	var mode control.Mode
//	cborData, _ := cbor.Marshal("must")
//	mode.UnmarshalCBOR(cborData)
//	fmt.Println(mode) // Output: Must
//
// See also: github.com/fxamacker/cbor for CBOR encoding/decoding.
func (c *Mode) UnmarshalCBOR(bytes []byte) error {
	var s string
	if e := cbor.Unmarshal(bytes, &s); e != nil {
		return e
	} else {
		return c.unmarshall([]byte(s))
	}
}
