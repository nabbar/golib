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

// Package retro provides retrocompatibility support for struct serialization across different versions.
// It allows structs to be marshaled and unmarshaled with version-specific field inclusion/exclusion
// based on semantic versioning constraints defined in struct tags.
//
// The package supports multiple serialization formats (JSON, YAML, TOML) and enables backward
// compatibility by controlling which fields are included based on the struct's version field.
//
// Example usage:
//
//	type MyStruct struct {
//	    Version string `json:"version"`
//	    Name    string `json:"name" retro:">=v1.0.0"`
//	    OldField string `json:"old_field" retro:"<v2.0.0"`
//	}
//
//	model := retro.Model[MyStruct]{
//	    Struct: MyStruct{Version: "v1.5.0", Name: "test"},
//	}
//	data, _ := model.MarshalJSON()
//
// See also:
//   - encoding/json for JSON marshaling
//   - gopkg.in/yaml.v3 for YAML marshaling
//   - github.com/pelletier/go-toml for TOML marshaling
package retro

// Format represents a serialization format supported by the retro package.
type Format string

const (
	// FormatJSON represents JSON serialization format.
	// Uses encoding/json for marshaling and unmarshaling.
	FormatJSON Format = "json"

	// FormatYAML represents YAML serialization format.
	// Uses gopkg.in/yaml.v3 for marshaling and unmarshaling.
	FormatYAML Format = "yaml"

	// FormatTOML represents TOML serialization format.
	// Uses github.com/pelletier/go-toml for marshaling and unmarshaling.
	FormatTOML Format = "toml"
)

// SupportedFormats is the list of all serialization formats supported by this package.
var SupportedFormats = []Format{FormatJSON, FormatYAML, FormatTOML}

// Valid checks if the format is one of the supported serialization formats.
// Returns true if the format is FormatJSON, FormatYAML, or FormatTOML.
//
// Example:
//
//	format := retro.FormatJSON
//	if format.Valid() {
//	    // Format is supported
//	}
func (f Format) Valid() bool {
	for _, validFormat := range SupportedFormats {
		if f == validFormat {
			return true
		}
	}
	return false
}
