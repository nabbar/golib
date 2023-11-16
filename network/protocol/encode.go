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

package protocol

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	dblQuote = "\""
	smpQuote = "'"
)

func (s *NetworkProtocol) unmarshall(val []byte) error {
	val = bytes.Trim(bytes.Trim(val, smpQuote), dblQuote)
	*s = ParseBytes(val)
	return nil
}

func (s NetworkProtocol) MarshalJSON() ([]byte, error) {
	t := s.String()
	b := make([]byte, 0, len(t)+2)
	b = append(b, '"')
	b = append(b, []byte(t)...)
	b = append(b, '"')
	return b, nil
}

func (s *NetworkProtocol) UnmarshalJSON(bytes []byte) error {
	return s.unmarshall(bytes)
}

func (s NetworkProtocol) MarshalYAML() (interface{}, error) {
	return []byte(s.String()), nil
}

func (s *NetworkProtocol) UnmarshalYAML(value *yaml.Node) error {
	return s.unmarshall([]byte(value.Value))
}

func (s NetworkProtocol) MarshalTOML() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *NetworkProtocol) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return s.unmarshall(p)
	}
	if p, k := i.(string); k {
		return s.unmarshall([]byte(p))
	}
	return fmt.Errorf("size: value not in valid format")
}

func (s NetworkProtocol) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *NetworkProtocol) UnmarshalText(bytes []byte) error {
	return s.unmarshall(bytes)
}

func (s NetworkProtocol) MarshalCBOR() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *NetworkProtocol) UnmarshalCBOR(bytes []byte) error {
	return s.unmarshall(bytes)
}
