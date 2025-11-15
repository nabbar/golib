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

package ca

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
)

func (o *Certif) String() string {
	s, _ := o.Chain()
	return s
}

func (o *Certif) Chain() (string, error) {
	var buf = bytes.NewBuffer(make([]byte, 0))

	for _, c := range o.c {
		if c == nil {
			continue
		} else if e := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw}); e != nil {
			return "", e
		}
	}

	return buf.String(), nil
}

func (o *Certif) SliceChain() ([]string, error) {
	var (
		res = make([]string, 0)
		buf = bytes.NewBuffer(make([]byte, 0))
	)

	for _, c := range o.c {
		if c == nil {
			continue
		}

		if e := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw}); e != nil {
			return nil, e
		} else {
			res = append(res, buf.String())
			buf.Reset()
		}
	}

	return res, nil
}

func (o *Certif) AppendPool(p *x509.CertPool) {
	for _, c := range o.c {
		if c == nil {
			continue
		}
		p.AddCert(c)
	}
}
