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
	"fmt"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

func (m Model[T]) MarshalJSON() ([]byte, error) {
	return m.marshal(FormatJSON)
}

func (m *Model[T]) UnmarshalJSON(data []byte) error {
	return m.unmarshal(data, FormatJSON)
}

func (m Model[T]) MarshalYAML() (interface{}, error) {
	data, err := m.marshal(FormatYAML)

	if err != nil {
		return nil, err
	}

	return string(data), nil
}

func (m *Model[T]) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var (
		tempMap map[string]interface{}
		rawData []byte
		err     error
	)

	if err = unmarshal(&tempMap); err != nil {
		return err
	}

	if rawData, err = yaml.Marshal(tempMap); err != nil {
		return err
	}

	return m.unmarshal(rawData, FormatYAML)
}

func (m Model[T]) MarshalTOML() ([]byte, error) {
	return m.marshal(FormatTOML)
}

func (m *Model[T]) UnmarshalTOML(data interface{}) error {

	rawData, err := toml.Marshal(data.(map[string]interface{}))

	if err != nil {
		return fmt.Errorf("error marshaling to TOML bytes: %w", err)
	}

	return m.unmarshal(rawData, FormatTOML)
}
