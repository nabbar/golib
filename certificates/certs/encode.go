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

package certs

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"strconv"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func (o *Certif) MarshalText() (text []byte, err error) {
	return []byte(o.String()), err
}

func (o *Certif) UnmarshalText(text []byte) error {
	var (
		chn = ConfigChain(text)
		crt *tls.Certificate
		err error
	)

	if crt, err = chn.Cert(); err != nil {
		return err
	} else if crt == nil || len(crt.Certificate) == 0 {
		return ErrInvalidPairCertificate
	} else {
		o.g = &chn
		o.c = *crt
		return nil
	}
}

func (o *Certif) MarshalBinary() (data []byte, err error) {
	return o.MarshalCBOR()
}

func (o *Certif) UnmarshalBinary(data []byte) error {
	return o.UnmarshalCBOR(data)
}

func (o *Certif) MarshalJSON() ([]byte, error) {
	var cfg any

	if o == nil || o.g == nil {
		return []byte(""), nil
	} else if p := o.g.GetCerts(); len(p) == 1 {
		cfg = ConfigChain(o.g.GetCerts()[0])
	} else if len(p) == 2 {
		cfg = ConfigPair{
			Key: p[0],
			Pub: p[1],
		}
	} else {
		cfg = o.g
	}

	return json.Marshal(cfg)
}

func (o *Certif) UnmarshalJSON(p []byte) error {
	var (
		cfg ConfigPair
		chn ConfigChain
		crt *tls.Certificate
		err error
	)

	if err = json.Unmarshal(p, &cfg); err == nil && len(cfg.Key) > 0 && len(cfg.Pub) > 0 {
		if crt, err = cfg.Cert(); err != nil {
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &cfg
			o.c = *crt
			return nil
		}
	}

	if err = json.Unmarshal(p, &chn); err == nil && len(chn) > 0 {
		if crt, err = chn.Cert(); err != nil {
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &chn
			o.c = *crt
			return nil
		}
	}

	p = bytes.TrimSpace(p)
	p = bytes.Trim(p, "\"")
	p = bytes.Replace(p, []byte("\\n"), []byte("\n"), -1) // nolint

	if c, e := Parse(string(p)); e == nil {
		*o = c.Model()
		return nil
	}

	return ErrInvalidCertificate
}

func (o *Certif) MarshalYAML() (interface{}, error) {
	if o == nil || o.g == nil {
		return []byte(""), nil
	} else if p, e := o.Chain(); e != nil {
		return nil, e
	} else {
		return "\"" + strconv.Quote(p) + "\"", nil
	}
}

func (o *Certif) UnmarshalYAML(value *yaml.Node) error {
	var (
		src = []byte(value.Value)
		cfg ConfigPair
		chn ConfigChain
		crt *tls.Certificate
		err error
	)

	if err = yaml.Unmarshal(src, &cfg); err == nil && len(cfg.Key) > 0 && len(cfg.Pub) > 0 {
		if crt, err = cfg.Cert(); err != nil {
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &cfg
			o.c = *crt
			return nil
		}
	}

	if err = yaml.Unmarshal(src, &chn); err == nil && len(chn) > 0 {
		if crt, err = chn.Cert(); err != nil {
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &chn
			o.c = *crt
			return nil
		}
	}

	src = bytes.TrimSpace(src)
	src = bytes.Trim(src, "\"")
	src = bytes.Replace(src, []byte("\\n"), []byte("\n"), -1) // nolint

	if c, e := Parse(string(src)); e == nil {
		*o = c.Model()
		return nil
	}

	return ErrInvalidCertificate
}

func (o *Certif) MarshalTOML() ([]byte, error) {
	var cfg any

	if o == nil || o.g == nil {
		return []byte(""), nil
	} else if p := o.g.GetCerts(); len(p) == 1 {
		cfg = ConfigChain(o.g.GetCerts()[0])
	} else if len(p) == 2 {
		cfg = ConfigPair{
			Key: p[0],
			Pub: p[1],
		}
	} else {
		cfg = o.g
	}

	return toml.Marshal(cfg)
}

func (o *Certif) UnmarshalTOML(i interface{}) error {
	if s, t := i.(map[string]interface{}); t {
		m := make(map[string]string)
		for n, v := range s {
			if u, l := v.(string); l && len(u) > 0 {
				m[n] = u
			} else if w, l := v.([]byte); l && len(w) > 0 {
				m[n] = string(w)
			}
		}
		i = m
	}

	if m, k := i.(map[string]string); k && len(m) == 2 {
		cfg := ConfigPair{
			Key: m["key"],
			Pub: m["pub"],
		}
		if c, e := cfg.Cert(); e == nil {
			*o = Certif{
				g: &cfg,
				c: *c,
			}
			return nil
		}
	}

	if p, k := i.(string); k && len(p) > 0 {
		i = []byte(p)
	}

	if p, k := i.([]byte); k && len(p) > 0 {
		p = bytes.TrimSpace(p)
		p = bytes.Trim(p, "\"")
		p = bytes.Replace(p, []byte("\\n"), []byte("\n"), -1) // nolint
		if c, e := Parse(string(p)); e == nil {
			*o = c.Model()
			return nil
		}
		if c, e := Parse(string(p)); e == nil {
			*o = c.Model()
			return nil
		}
	}

	return ErrInvalidCertificate
}

func (o *Certif) MarshalCBOR() ([]byte, error) {
	var cfg any

	if o == nil || o.g == nil {
		return []byte(""), nil
	} else if p := o.g.GetCerts(); len(p) == 1 {
		cfg = ConfigChain(o.g.GetCerts()[0])
	} else if len(p) == 2 {
		cfg = ConfigPair{
			Key: p[0],
			Pub: p[1],
		}
	} else {
		cfg = o.g
	}

	return cbor.Marshal(cfg)
}

func (o *Certif) UnmarshalCBOR(bytes []byte) error {
	var (
		cfg ConfigPair
		chn ConfigChain
		crt *tls.Certificate
		err error
	)

	if err = cbor.Unmarshal(bytes, &cfg); err == nil && len(cfg.Key) > 0 && len(cfg.Pub) > 0 {
		if crt, err = cfg.Cert(); err != nil {
			return err
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &cfg
			o.c = *crt
			return nil
		}
	}

	if err = cbor.Unmarshal(bytes, &chn); err == nil && len(chn) > 0 {
		if crt, err = chn.Cert(); err != nil {
			return err
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &chn
			o.c = *crt
			return nil
		}
	}

	return ErrInvalidCertificate
}
