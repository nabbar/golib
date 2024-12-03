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

package ca

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func isFile(path string) bool {
	f, e := os.Stat(path)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func (o *mod) unMarshall(p []byte) error {
	if len(p) < 1 {
		return ErrInvalidPairCertificate
	}

	p = bytes.TrimSpace(p)

	// remove \n\r
	p = bytes.Trim(p, "\n")
	p = bytes.Trim(p, "\r")

	// do again if \r\n
	p = bytes.Trim(p, "\n")
	p = bytes.Trim(p, "\r")

	p = bytes.TrimSpace(p)
	v := make([]*x509.Certificate, 0)

	for {
		block, rest := pem.Decode(p)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			if c, e := x509.ParseCertificate(block.Bytes); e == nil {
				v = append(v, c)
			}
		} else if fs := string(block.Bytes); isFile(fs) {
			if b, e := os.ReadFile(fs); e == nil {
				if crt, err := x509.ParseCertificate(b); err == nil {
					v = append(v, crt)
				}
			}
		}

		p = rest
	}

	o.c = v
	return nil
}

func (o *mod) MarshalText() (text []byte, err error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return []byte(s), nil
	}
}

func (o *mod) UnmarshalText(text []byte) error {
	return o.unMarshall(text)
}

func (o *mod) MarshalBinary() (data []byte, err error) {
	return o.MarshalCBOR()
}

func (o *mod) UnmarshalBinary(data []byte) error {
	return o.UnmarshalCBOR(data)
}

func (o *mod) MarshalJSON() ([]byte, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return json.Marshal(s)
	}
}

func (o *mod) UnmarshalJSON(bytes []byte) error {
	return o.unMarshall(bytes)
}

func (o *mod) MarshalYAML() (interface{}, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return yaml.Marshal(s)
	}
}

func (o *mod) UnmarshalYAML(value *yaml.Node) error {
	return o.unMarshall([]byte(value.Value))
}

func (o *mod) MarshalTOML() ([]byte, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return toml.Marshal(s)
	}
}

func (o *mod) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return o.unMarshall(p)
	}

	if s, k := i.(string); k {
		return o.unMarshall([]byte(s))
	}

	return ErrInvalidCertificate
}

func (o *mod) MarshalCBOR() ([]byte, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return cbor.Marshal(s)
	}
}

func (o *mod) UnmarshalCBOR(bytes []byte) error {
	return o.unMarshall(bytes)
}
