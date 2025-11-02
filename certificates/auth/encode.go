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

package auth

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

func (a *ClientAuth) unmarshall(val []byte) error {
	*a = ParseBytes(val)
	return nil
}

func (a ClientAuth) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *ClientAuth) UnmarshalJSON(bytes []byte) error {
	return a.unmarshall(bytes)
}

func (a ClientAuth) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

func (a *ClientAuth) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	return a.unmarshall([]byte(value.Value))
}

func (a ClientAuth) MarshalTOML() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *ClientAuth) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return a.unmarshall(p)
	}
	if p, k := i.(string); k {
		return a.unmarshall([]byte(p))
	}
	return fmt.Errorf("tls client Auth: value not in valid format")
}

func (a ClientAuth) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *ClientAuth) UnmarshalText(bytes []byte) error {
	return a.unmarshall(bytes)
}

func (a ClientAuth) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(a.String())
}

func (a *ClientAuth) UnmarshalCBOR(p []byte) error {
	var s string
	if e := cbor.Unmarshal(p, &s); e != nil {
		return e
	} else {
		return a.unmarshall([]byte(s))
	}
}
