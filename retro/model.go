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
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Model is a generic wrapper that provides version-aware serialization for any struct type T.
// It supports both standard serialization (using default marshaling) and retrocompatibility mode
// where fields are included/excluded based on version constraints in struct tags.
//
// Fields:
//   - Struct: The wrapped struct instance of type T
//   - Standard: If true, uses standard marshaling without version filtering.
//     If false, applies version-based field filtering using "retro" tags.
//
// The wrapped struct must have a "Version" field (string type) to enable version-based filtering.
// Version constraints in "retro" tags follow semantic versioning with operators:
//   - ">=v1.0.0" - Include field for versions >= v1.0.0
//   - "<v2.0.0" - Include field for versions < v2.0.0
//   - "v1.0.0-v2.0.0" - Include field for versions in range [v1.0.0, v2.0.0]
//
// Example:
//
//	type Config struct {
//	    Version  string `json:"version"`
//	    NewField string `json:"new_field" retro:">=v2.0.0"`
//	    OldField string `json:"old_field" retro:"<v2.0.0"`
//	}
//
//	model := Model[Config]{
//	    Struct: Config{Version: "v1.5.0", OldField: "value"},
//	    Standard: false,
//	}
//	data, _ := model.MarshalJSON() // Only includes OldField, not NewField
type Model[T any] struct {
	Struct   T
	Standard bool
}

// marshal serializes the wrapped struct to the specified format with version-aware field filtering.
// If Standard is true, uses default marshaling. Otherwise, applies version constraints from "retro" tags.
//
// The method:
//  1. Validates the format is supported
//  2. If Standard mode, delegates to standard marshaling
//  3. Otherwise, builds a map with only version-compatible fields
//  4. Respects "omitempty" tags to exclude empty values
//  5. Marshals the filtered map to the target format
//
// Parameters:
//   - format: Target serialization format (JSON, YAML, or TOML)
//
// Returns:
//   - []byte: Serialized data
//   - error: Error if format is unsupported or marshaling fails
//
// The wrapped struct must have a "Version" field. If missing or empty, "default" is used.
func (m Model[T]) marshal(format Format) ([]byte, error) {

	if !format.Valid() {
		return nil, errors.New("unsupported format")
	}

	if m.Standard {
		switch format {
		case FormatJSON:
			return json.Marshal(m.Struct)
		case FormatYAML:
			return yaml.Marshal(m.Struct)
		case FormatTOML:
			return toml.Marshal(m.Struct)
		default:
			return nil, errors.New("unsupported format")
		}
	}

	var (
		modelMap = make(map[string]interface{})
		key      string
	)

	val := reflect.Indirect(reflect.ValueOf(&m.Struct))

	version := val.FieldByName("Version").String()

	if version == "" {
		version = "default"
	}

	for i := 0; i < val.NumField(); i++ {

		typeField := val.Type().Field(i)

		formatTag := typeField.Tag.Get(string(format))

		retroTag := typeField.Tag.Get("retro")

		supported := isVersionSupported(version, retroTag)

		if !supported {
			continue
		}

		if formatTag != "" {
			tagParts := strings.Split(formatTag, ",")
			key = tagParts[0]
		} else {
			key = val.Type().Field(i).Name
		}

		fieldValue := val.Field(i)

		// Check for "omitempty"
		if len(strings.Split(formatTag, ",")) > 1 &&
			strings.Split(formatTag, ",")[1] == "omitempty" {

			if isEmptyValue(fieldValue) {
				continue
			}
		}
		// Marshal the field value and add to the map
		modelMap[key] = fieldValue.Interface()
	}

	switch format {
	case FormatJSON:
		return json.Marshal(modelMap)
	case FormatYAML:
		return yaml.Marshal(modelMap)
	case FormatTOML:
		return toml.Marshal(modelMap)
	default:
		return nil, errors.New("unsupported format")
	}
}

// unmarshal deserializes data into the wrapped struct with version-aware field filtering.
// If Standard is true, uses default unmarshaling. Otherwise, only populates fields that match
// version constraints in "retro" tags.
//
// The method:
//  1. Validates the format is supported
//  2. If Standard mode, delegates to standard unmarshaling
//  3. Otherwise, unmarshals to a temporary map
//  4. Extracts version from the map (defaults to "default" if missing)
//  5. For each struct field, checks version compatibility
//  6. Populates only compatible fields, respecting custom unmarshalers
//
// Parameters:
//   - data: Serialized data to unmarshal
//   - format: Source serialization format (JSON, YAML, or TOML)
//
// Returns:
//   - error: Error if format is unsupported, data is invalid, or unmarshaling fails
//
// The method supports custom unmarshalers (json.Unmarshaler, yaml.Unmarshaler, toml.Unmarshaler)
// for individual fields.
func (m *Model[T]) unmarshal(data []byte, format Format) error {

	if !format.Valid() {
		return errors.New("unsupported format")
	}

	if m.Standard {
		switch format {
		case FormatJSON:
			return json.Unmarshal(data, &m.Struct)
		case FormatYAML:
			return yaml.Unmarshal(data, &m.Struct)
		case FormatTOML:
			return toml.Unmarshal(data, &m.Struct)
		default:
			return errors.New("unsupported format")
		}
	}

	var (
		tempMap    map[string]interface{}
		version    string
		exists     bool
		rawField   interface{}
		rawMessage []byte
		err        error
	)

	// Unmarshal based on the provided format
	switch format {
	case FormatJSON:
		if err = json.Unmarshal(data, &tempMap); err != nil {
			return err
		}
	case FormatYAML:
		if err = yaml.Unmarshal(data, &tempMap); err != nil {
			return err
		}
	case FormatTOML:
		if err = toml.Unmarshal(data, &tempMap); err != nil {
			return err
		}
	default:
		return errors.New("unsupported format")
	}

	version, _ = tempMap["version"].(string)

	if version == "" {
		version = "default"
	}

	val := reflect.Indirect(reflect.ValueOf(&m.Struct))

	for i := 0; i < val.NumField(); i++ {

		typeField := val.Type().Field(i)

		formatTag := typeField.Tag.Get(string(format))

		fieldName := strings.Split(formatTag, ",")[0]

		retroTag := typeField.Tag.Get("retro")

		if !isVersionSupported(version, retroTag) {
			continue
		}

		if rawField, exists = tempMap[fieldName]; exists {

			field := val.Field(i)

			if field.CanAddr() {
				field = field.Addr()
			}

			if format == FormatJSON { // nolint

				if unmarshaler, ok := field.Interface().(json.Unmarshaler); ok {

					if rawMessage, err = json.Marshal(rawField); err != nil {
						return err
					}

					if err = unmarshaler.UnmarshalJSON(rawMessage); err != nil {
						return err
					}

				} else {

					if rawMessage, err = json.Marshal(rawField); err != nil {
						return err
					}

					if err = json.Unmarshal(rawMessage, field.Interface()); err != nil {
						return err
					}
				}

			} else if format == FormatYAML { // nolint

				if unmarshaler, ok := field.Interface().(yaml.Unmarshaler); ok {

					var node1 yaml.Node

					if rawMessage, err = yaml.Marshal(rawField); err != nil {
						return err
					}

					if err = yaml.Unmarshal(rawMessage, &node1); err != nil {
						return err
					}

					if err = unmarshaler.UnmarshalYAML(&node1); err != nil {
						return err
					}

				} else {

					if rawMessage, err = yaml.Marshal(rawField); err != nil {
						return err
					}

					if err = yaml.Unmarshal(rawMessage, field.Interface()); err != nil {
						return err
					}
				}

			} else if format == FormatTOML { // nolint

				if unmarshaler, ok := field.Interface().(toml.Unmarshaler); ok {

					if rawMessage, err = toml.Marshal(map[string]interface{}{
						fieldName: rawField,
					}); err != nil {
						return err
					}

					if err = unmarshaler.UnmarshalTOML(rawMessage); err != nil {
						return err
					}

				} else {

					if rawMessage, err = toml.Marshal(map[string]interface{}{
						fieldName: rawField,
					}); err != nil {
						return err
					}

					if err = toml.Unmarshal(rawMessage, &m.Struct); err != nil {
						return err
					}
				}
			}

		}

	}

	return nil
}
