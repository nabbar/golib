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
	"crypto/x509"
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

type mod struct {
	c []*x509.Certificate
}

func (m *mod) Len() int {
	return len(m.c)
}

func (m *mod) AppendBytes(p []byte) error {
	c := &mod{
		c: make([]*x509.Certificate, 0),
	}

	if e := c.unMarshall(p); e != nil {
		return e
	}

	for _, i := range c.c {
		m.c = append(m.c, i)
	}

	return nil
}

func (m *mod) AppendString(str string) error {
	return m.AppendBytes([]byte(str))
}

func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = &mod{
				c: make([]*x509.Certificate, 0),
			}
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
		if to.Kind() != reflect.Interface {
			return data, nil
		}

		// Format/decode/parse the data and return the new value
		if e := z.unMarshall([]byte(t)); e != nil {
			return data, nil
		} else {
			return z, nil
		}
	}
}
