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
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"
)

var (
	ErrInvalidPairCertificate = errors.New("invalid pair certificate")
	ErrInvalidCertificate     = errors.New("invalid certificate")
	ErrInvalidPrivateKey      = errors.New("invalid private key")
)

func cleanPem(s string) string {
	s = strings.TrimSpace(s)

	// remove \n\r
	s = strings.Trim(s, "\n")
	s = strings.Trim(s, "\r")

	// do again if \r\n
	s = strings.Trim(s, "\n")
	s = strings.Trim(s, "\r")

	return strings.TrimSpace(s)
}

type Config interface {
	Cert() (*tls.Certificate, error)
}

type ConfigPair struct {
	Key string `mapstructure:"key" json:"key" yaml:"key" toml:"key"`
	Pub string `mapstructure:"pub" json:"pub" yaml:"pub" toml:"pub"`
}

func (c *ConfigPair) Cert() (*tls.Certificate, error) {
	c.Key = cleanPem(c.Key)
	c.Pub = cleanPem(c.Pub)

	if c == nil {
		return nil, ErrInvalidPairCertificate
	} else if len(c.Key) < 1 || len(c.Pub) < 1 {
		return nil, ErrInvalidPairCertificate
	}

	if _, e := os.Stat(c.Key); e == nil {
		if b, e := os.ReadFile(c.Key); e == nil {
			c.Key = cleanPem(string(b))
		}
	}

	if _, e := os.Stat(c.Pub); e == nil {
		if b, e := os.ReadFile(c.Pub); e == nil {
			c.Pub = cleanPem(string(b))
		}
	}

	if crt, err := tls.X509KeyPair([]byte(c.Pub), []byte(c.Key)); err != nil {
		return nil, err
	} else {
		return &crt, nil
	}
}

type ConfigChain string

func (c *ConfigChain) Cert() (*tls.Certificate, error) {
	var (
		err error
		crt tls.Certificate
	)

	if c == nil {
		return nil, ErrInvalidPairCertificate
	} else if len(*c) < 1 {
		return nil, ErrInvalidPairCertificate
	}

	s := string(*c)

	if _, e := os.Stat(s); e == nil {
		if b, e := os.ReadFile(s); e == nil {
			s = cleanPem(string(b))
		}
	}

	p := []byte(cleanPem(s))

	for {
		block, rest := pem.Decode(p)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			crt.Certificate = append(crt.Certificate, block.Bytes)
		} else {
			crt.PrivateKey, err = c.getPrivateKey(block.Bytes)
			if err != nil {
				return nil, err
			}
		}

		p = rest
	}

	if len(crt.Certificate) == 0 {
		return nil, ErrInvalidCertificate
	} else if crt.PrivateKey == nil {
		return nil, ErrInvalidCertificate
	}

	return &crt, nil
}

func (c *ConfigChain) getPrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch k := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return k, nil
		default:
			return nil, ErrInvalidPrivateKey
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, ErrInvalidPrivateKey
}
