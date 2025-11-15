/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package tlsversion

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

func (v *Version) marshallByte() ([]byte, error) {
	return []byte("\"" + v.String() + "\""), nil
}

func (v *Version) unmarshall(val []byte) error {
	*v = ParseBytes(val)
	return nil
}

func (v Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *Version) UnmarshalJSON(bytes []byte) error {
	return v.unmarshall(bytes)
}

func (v Version) MarshalYAML() (interface{}, error) {
	return v.String(), nil
}

func (v *Version) UnmarshalYAML(value *yaml.Node) error {
	return v.unmarshall([]byte(value.Value))
}

func (v Version) MarshalTOML() ([]byte, error) {
	return v.marshallByte()
}

func (v *Version) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return v.unmarshall(p)
	}
	if p, k := i.(string); k {
		return v.unmarshall([]byte(p))
	}
	return fmt.Errorf("tls version: value not in valid format")
}

func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Version) UnmarshalText(bytes []byte) error {
	return v.unmarshall(bytes)
}

func (v Version) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(v.String())
}

func (v *Version) UnmarshalCBOR(bytes []byte) error {
	var t string
	if err := cbor.Unmarshal(bytes, &t); err != nil {
		return err
	} else {
		*v = Parse(t)
		return nil
	}
}
