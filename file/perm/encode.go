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

package perm

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON returns the JSON encoding of p.
//
// The JSON encoding is a simple string encoding of the Perm value,
// surrounded by double quotes.
//
// The returned bytes are valid JSON.
//
// The function does not return an error.
func (p Perm) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON parses a JSON encoding of a file permission into a Perm.
// The JSON encoding is expected to be a simple string encoding of the Perm value,
// surrounded by double quotes.
//
// The function returns an error if the JSON encoding is not a valid file permission.
//
// The function does not return an error if the JSON encoding is valid.
func (p *Perm) UnmarshalJSON(b []byte) error {
	return p.unmarshall(b)
}

// MarshalYAML returns the YAML encoding of p.
//
// The YAML encoding is a simple string encoding of the Perm value.
//
// The returned value is valid YAML.
//
// The function does not return an error.
func (p Perm) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

// UnmarshalYAML parses a YAML encoding of a file permission into a Perm.
// The YAML encoding is expected to be a simple string encoding of the Perm value.
//
// The function returns an error if the YAML encoding is not a valid file permission.
//
// The function does not return an error if the YAML encoding is valid.
func (p *Perm) UnmarshalYAML(value *yaml.Node) error {
	return p.unmarshall([]byte(value.Value))
}

// MarshalTOML returns the TOML encoding of p.
//
// The TOML encoding is equivalent to the JSON encoding returned by MarshalJSON.
//
// The returned bytes are valid TOML.
//
// The function does not return an error.
func (p Perm) MarshalTOML() ([]byte, error) {
	return p.MarshalJSON()
}

// UnmarshalTOML parses a TOML encoding of a file permission into a Perm.
//
// The TOML encoding is expected to be either a byte slice or a string.
// If the TOML encoding is a byte slice, it is expected to be a valid TOML
// encoding of a string. If the TOML encoding is a string, it is expected
// to be a valid string representation of a file permission.
//
// The function returns an error if the TOML encoding is not a valid file
// permission. The function does not return an error if the TOML encoding is
// valid.
func (p *Perm) UnmarshalTOML(i interface{}) error {
	if b, k := i.([]byte); k {
		return p.unmarshall(b)
	}

	if b, k := i.(string); k {
		return p.parseString(b)
	}

	return fmt.Errorf("file perm: value not in valid format")
}

// MarshalText returns the text encoding of p.
//
// The text encoding is a simple string encoding of the Perm value.
//
// The returned bytes are valid text.
//
// The function does not return an error.
func (p Perm) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText parses a text encoding of a file permission into a Perm.
//
// The text encoding is expected to be a simple string encoding of the Perm value.
//
// The function returns an error if the text encoding is not a valid file permission.
//
// The function does not return an error if the text encoding is valid.
func (p *Perm) UnmarshalText(b []byte) error {
	return p.unmarshall(b)
}

// MarshalCBOR returns the CBOR encoding of p.
//
// The CBOR encoding is a byte slice that represents the Perm value as a string.
//
// The returned bytes are valid CBOR.
//
// The function does not return an error.
//
// The function returns the CBOR encoding of p.String().
func (p Perm) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(p.String())
}

// UnmarshalCBOR parses a CBOR encoding of a file permission into a Perm.
//
// The CBOR encoding is expected to be a byte slice that represents the Perm value as a string.
//
// The function returns an error if the CBOR encoding is not a valid file permission.
//
// The function does not return an error if the CBOR encoding is valid.
func (p *Perm) UnmarshalCBOR(b []byte) error {
	var s string
	if e := cbor.Unmarshal(b, &s); e != nil {
		return e
	} else {
		return p.unmarshall([]byte(s))
	}
}
