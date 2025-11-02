/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package retro

import (
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface for Model.
// It serializes the wrapped struct to JSON format with version-aware field filtering.
//
// If Standard is false, only fields matching the version constraints in "retro" tags
// are included in the output. The "Version" field in the struct determines which
// fields are included.
//
// Returns:
//   - []byte: JSON-encoded data
//   - error: Error if marshaling fails
//
// Example:
//
//	model := Model[MyStruct]{Struct: MyStruct{Version: "v1.0.0"}}
//	data, err := json.Marshal(model) // Uses MarshalJSON
//
// See also: encoding/json.Marshaler
func (m Model[T]) MarshalJSON() ([]byte, error) {
	return m.marshal(FormatJSON)
}

// UnmarshalJSON implements the json.Unmarshaler interface for Model.
// It deserializes JSON data into the wrapped struct with version-aware field filtering.
//
// If Standard is false, only fields matching the version constraints in "retro" tags
// are populated. The version is extracted from the JSON data.
//
// Parameters:
//   - data: JSON-encoded data to unmarshal
//
// Returns:
//   - error: Error if unmarshaling fails or data is invalid
//
// Example:
//
//	var model Model[MyStruct]
//	err := json.Unmarshal(data, &model) // Uses UnmarshalJSON
//
// See also: encoding/json.Unmarshaler
func (m *Model[T]) UnmarshalJSON(data []byte) error {
	return m.unmarshal(data, FormatJSON)
}

// MarshalYAML implements the yaml.Marshaler interface for Model.
// It serializes the wrapped struct to YAML format with version-aware field filtering.
//
// The method returns an interface{} that can be marshaled by the YAML encoder.
// If Standard is false, only fields matching version constraints are included.
//
// Returns:
//   - interface{}: YAML-compatible data structure
//   - error: Error if marshaling fails
//
// Example:
//
//	model := Model[MyStruct]{Struct: MyStruct{Version: "v1.0.0"}}
//	data, err := yaml.Marshal(model) // Uses MarshalYAML
//
// See also: gopkg.in/yaml.v3.Marshaler
func (m Model[T]) MarshalYAML() (interface{}, error) {
	var result interface{}

	data, err := m.marshal(FormatYAML)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Model.
// It deserializes YAML data into the wrapped struct with version-aware field filtering.
//
// The method receives an unmarshal function from the YAML decoder and uses it to
// extract the data before applying version filtering.
//
// Parameters:
//   - unmarshal: Function provided by YAML decoder to unmarshal data
//
// Returns:
//   - error: Error if unmarshaling fails or data is invalid
//
// Example:
//
//	var model Model[MyStruct]
//	err := yaml.Unmarshal(data, &model) // Uses UnmarshalYAML
//
// See also: gopkg.in/yaml.v3.Unmarshaler
func (m *Model[T]) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var (
		tmp     interface{}
		err     error
		rawData []byte
	)

	if err = unmarshal(&tmp); err != nil {
		return err
	}

	if rawData, err = yaml.Marshal(tmp); err != nil {
		return err
	}

	return m.unmarshal(rawData, FormatYAML)
}

// MarshalTOML implements the toml.Marshaler interface for Model.
// It serializes the wrapped struct to TOML format with version-aware field filtering.
//
// If Standard is false, only fields matching version constraints in "retro" tags
// are included in the output.
//
// Returns:
//   - []byte: TOML-encoded data
//   - error: Error if marshaling fails
//
// Example:
//
//	model := Model[MyStruct]{Struct: MyStruct{Version: "v1.0.0"}}
//	data, err := toml.Marshal(model) // Uses MarshalTOML
//
// See also: github.com/pelletier/go-toml.Marshaler
func (m Model[T]) MarshalTOML() ([]byte, error) {
	return m.marshal(FormatTOML)
}

// UnmarshalTOML implements the toml.Unmarshaler interface for Model.
// It deserializes TOML data into the wrapped struct with version-aware field filtering.
//
// The method expects data to be a map[string]interface{} as provided by the TOML decoder.
// If Standard is false, only fields matching version constraints are populated.
//
// Parameters:
//   - data: TOML data as map[string]interface{}
//
// Returns:
//   - error: Error if unmarshaling fails, data is invalid, or type assertion fails
//
// Example:
//
//	var model Model[MyStruct]
//	err := toml.Unmarshal(data, &model) // Uses UnmarshalTOML
//
// See also: github.com/pelletier/go-toml.Unmarshaler
func (m *Model[T]) UnmarshalTOML(data interface{}) (err error) {

	var rawData []byte

	if rawData, err = toml.Marshal(data.(map[string]interface{})); err != nil {
		return err
	}

	return m.unmarshal(rawData, FormatTOML)
}
