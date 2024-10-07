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

	"gopkg.in/yaml.v3"
)

// Model => The generic struct that handles retro functionality
type Model[T any] struct {
	Fields T
}

func (m Model[T]) marshal(format Format) ([]byte, error) {
	if !format.Valid() {
		return nil, errors.New("unsupported format")
	}

	var (
		modelMap = make(map[string]interface{})
		key      string
	)

	val := reflect.Indirect(reflect.ValueOf(&m.Fields))

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
	default:
		return nil, errors.New("unsupported format")
	}
}

func (m *Model[T]) unmarshal(data []byte, format Format) error {
	if !format.Valid() {
		return errors.New("unsupported format")
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
		err = json.Unmarshal(data, &tempMap)
	case FormatYAML:
		err = yaml.Unmarshal(data, &tempMap)
	default:
		return errors.New("unsupported format")
	}

	if err != nil {
		return err
	}

	version, _ = tempMap["version"].(string)

	if version == "" {
		version = "default"
	}

	val := reflect.Indirect(reflect.ValueOf(&m.Fields))

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

			switch format {

			case FormatJSON:
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

			case FormatYAML:
				if unmarshaler, ok := field.Interface().(yaml.Unmarshaler); ok {

					var node yaml.Node

					if rawMessage, err = yaml.Marshal(rawField); err != nil {
						return err
					}

					if err = yaml.Unmarshal(rawMessage, &node); err != nil {
						return err
					}

					if err = unmarshaler.UnmarshalYAML(&node); err != nil {
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
			}
		}
	}
	return nil
}
