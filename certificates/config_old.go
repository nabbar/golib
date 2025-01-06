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
	tlsaut "github.com/nabbar/golib/certificates/auth"
	tlscas "github.com/nabbar/golib/certificates/ca"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"
)

type CertifOld struct {
	Key string `mapstructure:"key" json:"key" yaml:"key" toml:"key"`
	Pem string `mapstructure:"pem" json:"pem" yaml:"pem" toml:"pem"`
}

type ConfigOld struct {
	CurveList            []string    `mapstructure:"curveList" json:"curveList" yaml:"curveList" toml:"curveList"`
	CipherList           []string    `mapstructure:"cipherList" json:"cipherList" yaml:"cipherList" toml:"cipherList"`
	RootCAString         []string    `mapstructure:"rootCA" json:"rootCA" yaml:"rootCA" toml:"rootCA"`
	RootCAFile           []string    `mapstructure:"rootCAFiles" json:"rootCAFiles" yaml:"rootCAFiles" toml:"rootCAFiles"`
	ClientCAString       []string    `mapstructure:"clientCA" json:"clientCA" yaml:"clientCA" toml:"clientCA"`
	ClientCAFiles        []string    `mapstructure:"clientCAFiles" json:"clientCAFiles" yaml:"clientCAFiles" toml:"clientCAFiles"`
	CertPairString       []CertifOld `mapstructure:"certPair" json:"certPair" yaml:"certPair" toml:"certPair"`
	CertPairFile         []CertifOld `mapstructure:"certPairFiles" json:"certPairFiles" yaml:"certPairFiles" toml:"certPairFiles"`
	VersionMin           string      `mapstructure:"versionMin" json:"versionMin" yaml:"versionMin" toml:"versionMin"`
	VersionMax           string      `mapstructure:"versionMax" json:"versionMax" yaml:"versionMax" toml:"versionMax"`
	AuthClient           string      `mapstructure:"authClient" json:"authClient" yaml:"authClient" toml:"authClient"`
	InheritDefault       bool        `mapstructure:"inheritDefault" json:"inheritDefault" yaml:"inheritDefault" toml:"inheritDefault"`
	DynamicSizingDisable bool        `mapstructure:"dynamicSizingDisable" json:"dynamicSizingDisable" yaml:"dynamicSizingDisable" toml:"dynamicSizingDisable"`
	SessionTicketDisable bool        `mapstructure:"sessionTicketDisable" json:"sessionTicketDisable" yaml:"sessionTicketDisable" toml:"sessionTicketDisable"`
}

func (c *ConfigOld) ToConfig() Config {
	var car tlscas.Cert
	for _, v := range c.RootCAString {
		if car == nil {
			if i, e := tlscas.Parse(v); e == nil {
				car = i
			}
		} else {
			_ = car.AppendString(v)
		}
	}

	for _, v := range c.RootCAFile {
		if car == nil {
			if i, e := tlscas.Parse(v); e == nil {
				car = i
			}
		} else {
			_ = car.AppendString(v)
		}
	}

	var cac tlscas.Cert
	for _, v := range c.ClientCAFiles {
		if cac == nil {
			if i, e := tlscas.Parse(v); e == nil {
				cac = i
			}
		} else {
			_ = cac.AppendString(v)
		}
	}

	for _, v := range c.ClientCAString {
		if cac == nil {
			if i, e := tlscas.Parse(v); e == nil {
				cac = i
			}
		} else {
			_ = cac.AppendString(v)
		}
	}

	var crt = make([]tlscrt.Certif, 0)
	for _, v := range c.CertPairFile {
		if i, e := tlscrt.ParsePair(v.Key, v.Pem); e == nil {
			crt = append(crt, i.Model())
		}
	}

	for _, v := range c.CertPairString {
		if i, e := tlscrt.ParsePair(v.Key, v.Pem); e == nil {
			crt = append(crt, i.Model())
		}
	}

	cip := make([]tlscpr.Cipher, 0)
	for _, v := range c.CipherList {
		if i := tlscpr.Parse(v); i.Check() {
			cip = append(cip, i)
		}
	}

	crv := make([]tlscrv.Curves, 0)
	for _, v := range c.CurveList {
		if i := tlscrv.Parse(v); i.Check() {
			crv = append(crv, i)
		}
	}

	return Config{
		CurveList:            crv,
		CipherList:           cip,
		RootCA:               append(make([]tlscas.Cert, 0), car),
		ClientCA:             append(make([]tlscas.Cert, 0), cac),
		Certs:                crt,
		VersionMin:           tlsvrs.Parse(c.VersionMin),
		VersionMax:           tlsvrs.Parse(c.VersionMax),
		AuthClient:           tlsaut.Parse(c.AuthClient),
		InheritDefault:       c.InheritDefault,
		DynamicSizingDisable: c.DynamicSizingDisable,
		SessionTicketDisable: c.SessionTicketDisable,
	}
}
