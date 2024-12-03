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
	"crypto/tls"

	tlscrt "github.com/nabbar/golib/certificates/certs"
)

func (o *config) LenCertificatePair() int {
	return len(o.cert)
}

func (o *config) CleanCertificatePair() {
	o.cert = make([]tlscrt.Cert, 0)
}

func (o *config) GetCertificatePair() []tls.Certificate {
	var res = make([]tls.Certificate, 0)

	for _, c := range o.cert {
		res = append(res, c.TLS())
	}

	return res
}

func (o *config) AddCertificatePairString(key, crt string) error {
	if c, e := tlscrt.ParsePair(key, crt); e != nil {
		return e
	} else {
		o.cert = append(o.cert, c)
		return nil
	}
}

func (o *config) AddCertificatePairFile(keyFile, crtFile string) error {
	var (
		key = make([]byte, 0)
		pub = make([]byte, 0)
		fct = func(p []byte) error {
			if len(key) < 1 {
				copy(key, p)
			} else {
				copy(pub, p)
			}
			return nil
		}
	)

	if e := checkFile(fct, keyFile, crtFile); e != nil {
		return e
	} else if c, e := tlscrt.ParsePair(string(key), string(pub)); e != nil {
		return e
	} else {
		o.cert = append(o.cert, c)
		return nil
	}
}
