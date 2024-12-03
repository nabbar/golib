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
	"crypto/x509"

	tlsaut "github.com/nabbar/golib/certificates/auth"
	tlscas "github.com/nabbar/golib/certificates/ca"
)

func (o *config) SetClientAuth(a tlsaut.ClientAuth) {
	o.clientAuth = a
}

func (o *config) GetClientCA() []tlscas.Cert {
	var res = make([]tlscas.Cert, 0)

	for _, c := range o.clientCA {
		res = append(res, c)
	}

	return res
}

func (o *config) GetClientCAPool() *x509.CertPool {
	var res = x509.NewCertPool()

	for _, ca := range o.clientCA {
		ca.AppendPool(res)
	}

	return res
}

func (o *config) AddClientCAString(ca string) bool {
	if ca != "" {
		if c, e := tlscas.Parse(ca); e == nil {
			o.clientCA = append(o.clientCA, c)
			return true
		}
	}

	return false
}

func (o *config) AddClientCAFile(pemFile string) error {
	var fct = func(p []byte) error {
		if c, e := tlscas.ParseByte(p); e != nil {
			return e
		} else {
			o.clientCA = append(o.clientCA, c)
			return nil
		}
	}

	if e := checkFile(fct, pemFile); e != nil {
		return e
	} else {
		return nil
	}
}
