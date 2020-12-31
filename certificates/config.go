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

package certificates

import (
	"fmt"

	valid "github.com/go-playground/validator/v10"
	liberr "github.com/nabbar/golib/errors"
)

type Certif struct {
	Key string `mapstructure:"key" json:"key" yaml:"key" toml:"key"`
	Pem string `mapstructure:"pem" json:"pem" yaml:"pem" toml:"pem"`
}

type Config struct {
	InheritDefault       bool     `mapstructure:"inheritDefault" json:"inheritDefault" yaml:"inheritDefault" toml:"inheritDefault"`
	VersionMin           string   `mapstructure:"versionMin" json:"versionMin" yaml:"versionMin" toml:"versionMin"`
	VersionMax           string   `mapstructure:"versionMax" json:"versionMax" yaml:"versionMax" toml:"versionMax"`
	DynmaicSizingDisable bool     `mapstructure:"dynamicSizingDisable" json:"dynamicSizingDisable" yaml:"dynamicSizingDisable" toml:"dynamicSizingDisable"`
	SessionTicketDisable bool     `mapstructure:"sessionTicketDisable" json:"sessionTicketDisable" yaml:"sessionTicketDisable" toml:"sessionTicketDisable"`
	AuthClient           string   `mapstructure:"authClient" json:"authClient" yaml:"authClient" toml:"authClient"`
	CurveList            []string `mapstructure:"curveList" json:"curveList" yaml:"curveList" toml:"curveList"`
	CipherList           []string `mapstructure:"cipherList" json:"cipherList" yaml:"cipherList" toml:"cipherList"`
	RootCAString         []string `mapstructure:"rootCA" json:"rootCA" yaml:"rootCA" toml:"rootCA"`
	RootCAFile           []string `mapstructure:"rootCAFiles" json:"rootCAFiles" yaml:"rootCAFiles" toml:"rootCAFiles"`
	ClientCAString       []string `mapstructure:"clientCA" json:"clientCA" yaml:"clientCA" toml:"clientCA"`
	ClientCAFiles        []string `mapstructure:"clientCAFiles" json:"clientCAFiles" yaml:"clientCAFiles" toml:"clientCAFiles"`
	CertPairString       []Certif `mapstructure:"certPair" json:"certPair" yaml:"certPair" toml:"certPair"`
	CertPairFile         []Certif `mapstructure:"certPairFiles" json:"certPairFiles" yaml:"certPairFiles" toml:"certPairFiles"`
}

func (c *Config) Validate() liberr.Error {
	var e = ErrorValidatorError.Error(nil)

	if err := valid.New().Struct(c); err != nil {
		if er, ok := err.(*valid.InvalidValidationError); ok {
			e.AddParent(er)
		}

		for _, err := range err.(valid.ValidationErrors) {
			e.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", err.StructNamespace(), err.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (c *Config) New() (TLSConfig, liberr.Error) {
	if c.InheritDefault {
		return c.NewFrom(Default)
	} else {
		return c.NewFrom(nil)
	}
}

func (c *Config) NewFrom(cfg TLSConfig) (TLSConfig, liberr.Error) {
	var t *config

	if cfg != nil {
		t = asStruct(cfg.Clone())
	}

	if t == nil {
		t = asStruct(New())
		t.caRoot = SystemRootCA()
	}

	if c.VersionMin != "" {
		t.tlsMinVersion = StringToTlsVersion(c.VersionMin)
	}

	if c.VersionMax != "" {
		t.tlsMaxVersion = StringToTlsVersion(c.VersionMax)
	}

	if c.DynmaicSizingDisable {
		t.dynSizingDisabled = true
	}

	if c.SessionTicketDisable {
		t.ticketSessionDisabled = true
	}

	if c.AuthClient != "" {
		t.clientAuth = StringToClientAuth(c.AuthClient)
	}

	if len(c.CipherList) > 0 {
		for _, a := range c.CipherList {
			if len(a) < 1 {
				continue
			}
			t.cipherList = append(t.cipherList, StringToCipherKey(a))
		}
	}

	if len(c.CurveList) > 0 {
		for _, a := range c.CurveList {
			if len(a) < 1 {
				continue
			}
			t.curveList = append(t.curveList, StringToCurveID(a))
		}
	}

	if len(c.RootCAString) > 0 {
		for _, s := range c.RootCAString {
			if len(s) < 1 {
				continue
			}
			t.AddRootCAString(s)
		}
	}

	if len(c.RootCAFile) > 0 {
		for _, f := range c.RootCAFile {
			if len(f) < 1 {
				continue
			}
			if e := t.AddRootCAFile(f); e != nil {
				return nil, e
			}
		}
	}

	if len(c.ClientCAString) > 0 {
		for _, s := range c.ClientCAString {
			if len(s) < 1 {
				continue
			}
			t.AddClientCAString(s)
		}
	}

	if len(c.ClientCAFiles) > 0 {
		for _, f := range c.ClientCAFiles {
			if len(f) < 1 {
				continue
			}
			if e := t.AddClientCAFile(f); e != nil {
				return nil, e
			}
		}
	}

	if len(c.CertPairString) > 0 {
		for _, s := range c.CertPairString {
			if len(s.Key) < 1 || len(s.Pem) < 1 {
				continue
			}
			if e := t.AddCertificatePairString(s.Key, s.Pem); e != nil {
				return nil, e
			}
		}
	}

	if len(c.CertPairFile) > 0 {
		for _, f := range c.CertPairFile {
			if len(f.Key) < 1 || len(f.Pem) < 1 {
				continue
			}
			if e := t.AddCertificatePairFile(f.Key, f.Pem); e != nil {
				return nil, e
			}
		}
	}

	return t, nil
}
