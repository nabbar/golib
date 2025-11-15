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
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"gopkg.in/yaml.v3"
)

func isFile(path string) bool {
	f, e := os.Stat(path)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func (o *Certif) unMarshall(p []byte) error {
	if len(p) < 1 {
		return ErrInvalidPairCertificate
	}

	p = bytes.TrimSpace(p)
	p = bytes.Trim(p, "\"")
	p = bytes.Replace(p, []byte("\\n"), []byte("\n"), -1) // nolint

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
			if b, e := o.readFile(fs); e == nil {
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

func (o *Certif) readFile(fs string) ([]byte, error) {
	r, e := os.OpenRoot(filepath.Dir(fs))

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	if e != nil {
		return nil, e
	} else {
		return r.ReadFile(filepath.Base(fs))
	}
}

func (o *Certif) MarshalText() ([]byte, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return []byte(s), nil
	}
}

func (o *Certif) UnmarshalText(b []byte) error {
	return o.unMarshall(b)
}

func (o *Certif) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

func (o *Certif) UnmarshalBinary(b []byte) error {
	return o.UnmarshalCBOR(b)
}

func (o *Certif) MarshalJSON() ([]byte, error) {
	if s, e := o.SliceChain(); e != nil {
		return nil, e
	} else {
		return json.Marshal(s)
	}
}

func (o *Certif) UnmarshalJSON(b []byte) error {
	var s = make([]string, 0)

	if err := json.Unmarshal(b, &s); err != nil {
		return o.unMarshall(b)
	} else if len(s) > 0 {
		return o.unMarshall([]byte(strings.Join(s, "\n")))
	} else {
		*o = Certif{
			c: make([]*x509.Certificate, 0),
		}
		return nil
	}
}

func (o *Certif) MarshalYAML() (interface{}, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return s, nil
	}
}

func (o *Certif) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	} else {
		return o.unMarshall([]byte(value.Value))
	}
}

func (o *Certif) MarshalTOML() ([]byte, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return []byte(strconv.Quote(s)), nil
	}
}

func (o *Certif) UnmarshalTOML(i interface{}) error {
	if p, k := i.([]byte); k {
		return o.unMarshall(p)
	}

	if s, k := i.(string); k {
		return o.unMarshall([]byte(s))
	}

	return ErrInvalidCertificate
}

func (o *Certif) MarshalCBOR() ([]byte, error) {
	if s, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return cbor.Marshal(s)
	}
}

func (o *Certif) UnmarshalCBOR(b []byte) error {
	var s string
	if e := cbor.Unmarshal(b, &s); e != nil {
		return e
	} else {
		return o.unMarshall([]byte(s))
	}
}
