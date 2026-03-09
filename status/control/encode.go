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

// unmarshall is an internal helper method that parses a byte slice into a `Mode`.
// It delegates the actual parsing to `ParseBytes`. This method is used by all
// the public unmarshaling methods to ensure consistent behavior.
//
// It always returns `nil` error because `ParseBytes` handles invalid inputs by
// returning the default `Ignore` mode, rather than raising an error. This design
// choice simplifies configuration loading where a default fallback is often preferred.
func (c *Mode) unmarshall(val []byte) error {
	*c = ParseBytes(val)
	return nil
}

// MarshalJSON implements the `json.Marshaler` interface.
// It serializes the `Mode` into a JSON string using its PascalCase representation
// (e.g., "Must", "Should").
//
// Example:
//   data, _ := json.Marshal(control.Must)
//   // data is []byte(`"Must"`)
func (c Mode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements the `json.Unmarshaler` interface.
// It deserializes a JSON string into a `Mode`. The parsing is case-insensitive.
// If the JSON value is not a valid string or does not match a known mode,
// it defaults to `Ignore`.
//
// Example:
//   var m control.Mode
//   json.Unmarshal([]byte(`"must"`), &m)
//   // m is control.Must
func (c *Mode) UnmarshalJSON(bytes []byte) error {
	var s string
	if err := json.Unmarshal(bytes, &s); err != nil {
		return err
	}
	return c.unmarshall([]byte(s))
}

// MarshalYAML implements the `yaml.Marshaler` interface.
// It serializes the `Mode` into a YAML string using its PascalCase representation.
func (c Mode) MarshalYAML() (interface{}, error) {
	return c.String(), nil
}

// UnmarshalYAML implements the `yaml.Unmarshaler` interface.
// It deserializes a YAML value into a `Mode`. The parsing is case-insensitive.
// It extracts the value from the `yaml.Node` and parses it.
func (c *Mode) UnmarshalYAML(value *yaml.Node) error {
	return c.unmarshall([]byte(value.Value))
}

// MarshalTOML implements the `toml.Marshaler` interface.
// It serializes the `Mode` into a TOML string. It reuses `MarshalJSON` internally
// to ensure consistency.
func (c Mode) MarshalTOML() ([]byte, error) {
	return c.MarshalJSON()
}

// UnmarshalTOML implements the `toml.Unmarshaler` interface.
// It deserializes a TOML value into a `Mode`. It supports both byte slices and
// strings as input. The parsing is case-insensitive.
//
// If the input is not a byte slice or a string, it returns an error indicating
// an invalid format.
func (c *Mode) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return c.unmarshall(p)
	}
	if p, k := i.(string); k {
		return c.unmarshall([]byte(p))
	}
	return fmt.Errorf("size: value not in valid format")
}

// MarshalText implements the `encoding.TextMarshaler` interface.
// It serializes the `Mode` into a plain text byte slice using its PascalCase
// representation. This is useful for generic text-based protocols or logging.
func (c Mode) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// UnmarshalText implements the `encoding.TextUnmarshaler` interface.
// It deserializes a plain text byte slice into a `Mode`. The parsing is
// case-insensitive.
func (c *Mode) UnmarshalText(bytes []byte) error {
	return c.unmarshall(bytes)
}

// MarshalCBOR implements the `cbor.Marshaler` interface.
// It serializes the `Mode` into a CBOR (Concise Binary Object Representation)
// string. CBOR is a binary data serialization format loosely based on JSON.
func (c Mode) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(c.String())
}

// UnmarshalCBOR implements the `cbor.Unmarshaler` interface.
// It deserializes a CBOR-encoded byte slice into a `Mode`. The parsing is
// case-insensitive.
func (c *Mode) UnmarshalCBOR(bytes []byte) error {
	var s string
	if e := cbor.Unmarshal(bytes, &s); e != nil {
		return e
	} else {
		return c.unmarshall([]byte(s))
	}
}
