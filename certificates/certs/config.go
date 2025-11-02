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
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
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

func cleanPemByte(s []byte) []byte {
	s = bytes.TrimSpace(s)

	// remove \n\r
	s = bytes.Trim(s, "\n")
	s = bytes.Trim(s, "\r")

	// do again if \r\n
	s = bytes.Trim(s, "\n")
	s = bytes.Trim(s, "\r")

	return bytes.TrimSpace(s)
}

type Config interface {
	Cert() (*tls.Certificate, error)

	IsChain() bool
	IsPair() bool

	IsFile() bool
	GetCerts() []string
}

type ConfigPair struct {
	Key string `mapstructure:"key" json:"key" yaml:"key" toml:"key"`
	Pub string `mapstructure:"pub" json:"pub" yaml:"pub" toml:"pub"`
}

func (c *ConfigPair) Cert() (*tls.Certificate, error) {
	if c == nil {
		return nil, ErrInvalidPairCertificate
	}

	var (
		k = cleanPemByte([]byte(c.Key))
		p = cleanPemByte([]byte(c.Pub))
	)

	if len(k) < 1 || len(p) < 1 {
		return nil, ErrInvalidPairCertificate
	}

	if _, e := os.Stat(string(k)); e == nil {
		if b, e := os.ReadFile(string(k)); e == nil {
			k = cleanPemByte(b)
		}
	}

	if _, e := os.Stat(string(p)); e == nil {
		if b, e := os.ReadFile(string(p)); e == nil {
			p = cleanPemByte(b)
		}
	}

	if crt, err := tls.X509KeyPair(p, k); err != nil {
		return nil, err
	} else {
		return &crt, nil
	}
}

func (c *ConfigPair) IsChain() bool {
	return false
}

func (c *ConfigPair) IsPair() bool {
	return true
}

func (c *ConfigPair) IsFile() bool {
	if c == nil {
		return false
	}

	var (
		k = cleanPemByte([]byte(c.Key))
		p = cleanPemByte([]byte(c.Pub))
	)

	if len(k) < 1 || len(p) < 1 {
		return false
	}

	if _, e := os.Stat(string(k)); e == nil {
		return true
	}

	if _, e := os.Stat(string(p)); e == nil {
		return true
	}

	return false
}

func (c *ConfigPair) GetCerts() []string {
	return []string{c.Key, c.Pub}
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
		if b, e := c.readFile(s); e == nil {
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

func (o *ConfigChain) readFile(fs string) ([]byte, error) {
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
func (c *ConfigChain) IsChain() bool {
	return true
}

func (c *ConfigChain) IsPair() bool {
	return false
}

func (c *ConfigChain) IsFile() bool {
	if c == nil {
		return false
	}

	if _, e := os.Stat(string(*c)); e == nil {
		return true
	}

	return false
}

func (c *ConfigChain) GetCerts() []string {
	return []string{string(*c)}
}
