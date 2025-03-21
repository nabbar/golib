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
	"crypto/tls"
	"reflect"

	libmap "github.com/mitchellh/mapstructure"
)

type Certif struct {
	g Config
	c tls.Certificate
}

func (o *Certif) Cert() Cert {
	return o
}

func (o *Certif) Model() Certif {
	if o == nil {
		return Certif{}
	}
	return *o
}

func (o *Certif) IsChain() bool {
	if o == nil {
		return false
	}
	return o.g.IsChain()
}

func (o *Certif) IsPair() bool {
	if o == nil {
		return false
	}
	return o.g.IsPair()
}

func (o *Certif) IsFile() bool {
	if o == nil {
		return false
	}
	return o.g.IsFile()
}

func (o *Certif) GetCerts() []string {
	if o == nil {
		return make([]string, 0)
	}
	return o.g.GetCerts()
}

func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = &Certif{c: tls.Certificate{}}
			t string
			k bool
		)

		// Check if the data type matches the expected one
		if from.Kind() != reflect.String {
			return data, nil
		} else if t, k = data.(string); !k {
			return data, nil
		}

		// Check if the target type matches the expected one
		if to != reflect.TypeOf(z) {
			return data, nil
		}

		// Format/decode/parse the data and return the new value
		if e := z.unMarshall([]byte(t)); e != nil {
			return nil, e
		} else {
			return z, nil
		}
	}
}
