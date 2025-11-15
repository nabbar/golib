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

	tlscas "github.com/nabbar/golib/certificates/ca"
)

func (o *config) GetRootCA() []tlscas.Cert {
	return append(make([]tlscas.Cert, 0), o.caRoot...)
}

func (o *config) GetRootCAPool() *x509.CertPool {
	var res = x509.NewCertPool()
	for _, ca := range o.caRoot {
		ca.AppendPool(res)
	}
	return res
}

func (o *config) AddRootCA(rootCA tlscas.Cert) bool {
	if rootCA != nil && rootCA.Len() > 0 {
		o.caRoot = append(o.caRoot, rootCA)
		return true
	}

	return false
}

func (o *config) AddRootCAString(rootCA string) bool {
	if rootCA != "" {
		if c, e := tlscas.Parse(rootCA); e == nil {
			o.caRoot = append(o.caRoot, c)
			return true
		}
	}

	return false
}

func (o *config) AddRootCAFile(pemFile string) error {
	var fct = func(p []byte) error {
		if c, e := tlscas.ParseByte(p); e != nil {
			return e
		} else {
			o.caRoot = append(o.caRoot, c)
			return nil
		}
	}

	if e := checkFile(fct, pemFile); e != nil {
		return e
	}

	return nil
}
