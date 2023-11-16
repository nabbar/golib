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
	"fmt"

	"gopkg.in/yaml.v3"
)

func (p Perm) MarshalJSON() ([]byte, error) {
	t := p.String()
	b := make([]byte, 0, len(t)+2)
	b = append(b, '"')
	b = append(b, []byte(t)...)
	b = append(b, '"')
	return b, nil
}

func (p *Perm) UnmarshalJSON(bytes []byte) error {
	return p.unmarshall(bytes)
}

func (p Perm) MarshalYAML() (interface{}, error) {
	return p.MarshalJSON()
}

func (p *Perm) UnmarshalYAML(value *yaml.Node) error {
	return p.unmarshall([]byte(value.Value))
}

func (p Perm) MarshalTOML() ([]byte, error) {
	return p.MarshalJSON()
}

func (p *Perm) UnmarshalTOML(i interface{}) error {
	if b, k := i.([]byte); k {
		return p.unmarshall(b)
	}

	if b, k := i.(string); k {
		return p.parseString(b)
	}

	return fmt.Errorf("size: value not in valid format")
}

func (p Perm) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Perm) UnmarshalText(bytes []byte) error {
	return p.unmarshall(bytes)
}

func (p Perm) MarshalCBOR() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Perm) UnmarshalCBOR(bytes []byte) error {
	return p.unmarshall(bytes)
}
