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

	libval "github.com/go-playground/validator/v10"
	tlsaut "github.com/nabbar/golib/certificates/auth"
	tlscas "github.com/nabbar/golib/certificates/ca"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"
	liberr "github.com/nabbar/golib/errors"
)

type Config struct {
	CurveList            []tlscrv.Curves   `mapstructure:"curveList" json:"curveList" yaml:"curveList" toml:"curveList"`
	CipherList           []tlscpr.Cipher   `mapstructure:"cipherList" json:"cipherList" yaml:"cipherList" toml:"cipherList"`
	RootCA               []tlscas.Cert     `mapstructure:"rootCA" json:"rootCA" yaml:"rootCA" toml:"rootCA"`
	ClientCA             []tlscas.Cert     `mapstructure:"clientCA" json:"clientCA" yaml:"clientCA" toml:"clientCA"`
	Certs                []tlscrt.Certif   `mapstructure:"certs" json:"certs" yaml:"certs" toml:"certs"`
	VersionMin           tlsvrs.Version    `mapstructure:"versionMin" json:"versionMin" yaml:"versionMin" toml:"versionMin"`
	VersionMax           tlsvrs.Version    `mapstructure:"versionMax" json:"versionMax" yaml:"versionMax" toml:"versionMax"`
	AuthClient           tlsaut.ClientAuth `mapstructure:"authClient" json:"authClient" yaml:"authClient" toml:"authClient"`
	InheritDefault       bool              `mapstructure:"inheritDefault" json:"inheritDefault" yaml:"inheritDefault" toml:"inheritDefault"`
	DynamicSizingDisable bool              `mapstructure:"dynamicSizingDisable" json:"dynamicSizingDisable" yaml:"dynamicSizingDisable" toml:"dynamicSizingDisable"`
	SessionTicketDisable bool              `mapstructure:"sessionTicketDisable" json:"sessionTicketDisable" yaml:"sessionTicketDisable" toml:"sessionTicketDisable"`
}

func (c *Config) Validate() liberr.Error {
	err := ErrorValidatorError.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.StructNamespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c *Config) New() TLSConfig {
	if c.InheritDefault {
		return c.NewFrom(Default)
	} else {
		return c.NewFrom(nil)
	}
}

// nolint #gocognit
func (c *Config) NewFrom(cfg TLSConfig) TLSConfig {
	var t *Config

	if cfg != nil {
		t = cfg.Config()
	}

	if t == nil {
		t = &Config{}
	}

	if c.VersionMin != tlsvrs.VersionUnknown {
		t.VersionMin = c.VersionMin
	}

	if c.VersionMax != tlsvrs.VersionUnknown {
		t.VersionMax = c.VersionMax
	}

	if c.DynamicSizingDisable {
		t.DynamicSizingDisable = true
	}

	if c.SessionTicketDisable {
		t.SessionTicketDisable = true
	}

	if c.AuthClient != tlsaut.NoClientCert {
		t.AuthClient = c.AuthClient
	}

	if len(c.CipherList) > 0 {
		for _, a := range c.CipherList {
			if tlscpr.Check(a.Uint16()) {
				t.CipherList = append(t.CipherList, a)
			}
		}
	}

	if len(c.CurveList) > 0 {
		for _, a := range c.CurveList {
			if tlscrv.Check(a.Uint16()) {
				t.CurveList = append(t.CurveList, a)
			}
		}
	}

	if len(c.RootCA) > 0 {
		for _, s := range c.RootCA {
			t.RootCA = append(t.RootCA, s)
		}
	}

	if len(c.ClientCA) > 0 {
		for _, s := range c.ClientCA {
			t.ClientCA = append(t.ClientCA, s)
		}
	}

	if len(c.Certs) > 0 {
		for _, s := range c.Certs {
			t.Certs = append(t.Certs, s)
		}
	}

	res := &config{
		rand:                  nil,
		cert:                  make([]tlscrt.Cert, 0),
		cipherList:            make([]tlscpr.Cipher, 0),
		curveList:             make([]tlscrv.Curves, 0),
		caRoot:                make([]tlscas.Cert, 0),
		clientAuth:            t.AuthClient,
		clientCA:              make([]tlscas.Cert, 0),
		tlsMinVersion:         t.VersionMin,
		tlsMaxVersion:         t.VersionMax,
		dynSizingDisabled:     t.DynamicSizingDisable,
		ticketSessionDisabled: t.SessionTicketDisable,
	}

	if len(t.Certs) > 0 {
		for _, s := range t.Certs {
			res.cert = append(res.cert, s.Cert())
		}
	}

	if len(t.CipherList) > 0 {
		for _, s := range t.CipherList {
			res.cipherList = append(res.cipherList, s)
		}
	}

	if len(t.CurveList) > 0 {
		for _, s := range t.CurveList {
			res.curveList = append(res.curveList, s)
		}
	}

	if len(t.RootCA) > 0 {
		for _, s := range t.RootCA {
			res.caRoot = append(res.caRoot, s)
		}
	}

	if len(t.ClientCA) > 0 {
		for _, s := range t.ClientCA {
			res.clientCA = append(res.clientCA, s)
		}
	}

	return res
}
