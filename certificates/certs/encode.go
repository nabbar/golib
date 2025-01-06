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
	"crypto/tls"
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func (o *Certif) unMarshall(p []byte) error {
	if o.UnmarshalJSON(p) == nil {
		return nil
	} else if o.UnmarshalYAML(&yaml.Node{Value: string(p)}) != nil {
		return nil
	} else if o.UnmarshalTOML(p) != nil {
		return nil
	} else if o.UnmarshalCBOR(p) != nil {
		return nil
	} else if o.UnmarshalText(p) != nil {
		return nil
	}

	return ErrInvalidCertificate
}

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

func (o *Certif) UnmarshalJSON(bytes []byte) error {
	var (
		cfg ConfigPair
		chn ConfigChain
		crt *tls.Certificate
		err error
	)

	if err = json.Unmarshal(bytes, &cfg); err == nil && len(cfg.Key) > 0 && len(cfg.Pub) > 0 {
		if crt, err = cfg.Cert(); err != nil {
			return err
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &cfg
			o.c = *crt
			return nil
		}
	} else if err = json.Unmarshal(bytes, &chn); err == nil && len(chn) > 0 {
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

func (o *Certif) MarshalYAML() (interface{}, error) {
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

	return yaml.Marshal(cfg)
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
			return err
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &cfg
			o.c = *crt
			return nil
		}
	} else if err = yaml.Unmarshal(src, &chn); err == nil && len(chn) > 0 {
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
	var (
		p []byte
		s string
		k bool
	)

	if p, k = i.([]byte); !k {
		if s, k = i.(string); k {
			p = []byte(s)
		} else {
			return ErrInvalidCertificate
		}
	}

	if len(p) < 1 {
		return ErrInvalidCertificate
	}

	var (
		cfg ConfigPair
		chn ConfigChain
		crt *tls.Certificate
		err error
	)

	if err = toml.Unmarshal(p, &cfg); err == nil && len(cfg.Key) > 0 && len(cfg.Pub) > 0 {
		if crt, err = cfg.Cert(); err != nil {
			return err
		} else if crt == nil || len(crt.Certificate) == 0 {
			return ErrInvalidPairCertificate
		} else {
			o.g = &cfg
			o.c = *crt
			return nil
		}
	} else if err = toml.Unmarshal(p, &chn); err == nil && len(chn) > 0 {
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
	} else if err = cbor.Unmarshal(bytes, &chn); err == nil && len(chn) > 0 {
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
