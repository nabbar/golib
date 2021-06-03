/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package logger

import "github.com/sirupsen/logrus"

type Fields map[string]interface{}

func NewFields() Fields {
	return make(Fields)
}

func (f Fields) clone() map[string]interface{} {
	res := make(map[string]interface{}, 0)

	if len(f) > 0 {
		for k, v := range f {
			res[k] = v
		}
	}

	return res
}

func (f Fields) Add(key string, val interface{}) Fields {
	res := f.clone()
	res[key] = val

	return res
}

func (f Fields) Map(fct func(key string, val interface{}) interface{}) Fields {
	res := f.clone()

	for k, v := range res {
		if v = fct(k, v); v != nil {
			res[k] = v
		}
	}

	return res
}

func (f Fields) Merge(other Fields) Fields {
	if len(other) < 1 {
		return f
	} else if len(f) < 1 {
		return other
	}

	res := f.clone()

	other.Map(func(key string, val interface{}) interface{} {
		res[key] = val
		return nil
	})

	return res
}

func (f Fields) Clean(keys ...string) Fields {
	res := make(map[string]interface{}, 0)

	if len(keys) > 0 {
		f.Map(func(key string, val interface{}) interface{} {
			for _, kk := range keys {
				if kk == key {
					return nil
				}
			}

			res[key] = val
			return nil
		})
	}

	return res
}

func (f Fields) Logrus() logrus.Fields {
	return f.clone()
}
