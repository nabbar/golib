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

func (s Size) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Size) UnmarshalJSON(p []byte) error {
	return s.unmarshall(p)
}

func (s Size) MarshalYAML() (interface{}, error) {
	return s.String(), nil
}

func (s *Size) UnmarshalYAML(value *yaml.Node) error {
	return s.unmarshall([]byte(value.Value))
}

func (s Size) MarshalTOML() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Size) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return s.unmarshall(p)
	}
	if p, k := i.(string); k {
		return s.unmarshall([]byte(p))
	}
	return fmt.Errorf("size: value not in valid format")
}

func (s Size) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Size) UnmarshalText(p []byte) error {
	return s.unmarshall(p)
}

func (s Size) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(s.String())
}

func (s *Size) UnmarshalCBOR(p []byte) error {
	var b string
	if e := cbor.Unmarshal(p, &b); e != nil {
		return e
	} else {
		return s.unmarshall([]byte(b))
	}
}

func (s Size) MarshalBinary() ([]byte, error) {
	return s.MarshalCBOR()
}

func (s *Size) UnmarshalBinary(p []byte) error {
	return s.UnmarshalCBOR(p)
}
