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
	"crypto/x509"
	"encoding/pem"
)

func (o *Certif) String() string {
	if o == nil {
		return ""
	}

	s, _ := o.Chain()
	return cleanPem(s)
}

func (o *Certif) Pair() (pub string, key string, err error) {
	if o == nil {
		return "", "", ErrInvalidPairCertificate
	}

	var (
		bufPub = bytes.NewBuffer(make([]byte, 0))
		bufKey = bytes.NewBuffer(make([]byte, 0))
	)

	for _, certDER := range o.c.Certificate {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certDER,
		}
		if err = pem.Encode(bufPub, block); err != nil {
			return "", "", err
		}
	}

	// Afficher la clé privée si disponible
	if o.c.PrivateKey != nil {
		var p []byte
		if p, err = x509.MarshalPKCS8PrivateKey(o.c.PrivateKey); err != nil {
			return "", "", err
		} else {
			privateKeyBlock := &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: p,
			}
			if err = pem.Encode(bufKey, privateKeyBlock); err != nil {
				return "", "", err
			}
		}
	}

	return bufPub.String(), bufKey.String(), nil
}

func (o *Certif) Chain() (string, error) {
	if pub, key, err := o.Pair(); err != nil {
		return "", err
	} else {
		return pub + key, nil
	}
}

func (o *Certif) TLS() tls.Certificate {
	if o == nil {
		return tls.Certificate{}
	}
	return o.c
}
