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

package fields

import (
	"context"
	"encoding/json"

	libctx "github.com/nabbar/golib/context"
	"github.com/sirupsen/logrus"
)

type fldModel struct {
	libctx.Config[string]
}

func (o *fldModel) Add(key string, val interface{}) Fields {
	if o == nil {
		return nil
	} else if o.Config == nil {
		return nil
	}

	o.Store(key, val)

	return o
}

func (o *fldModel) Logrus() logrus.Fields {
	var res = make(logrus.Fields, 0)

	if o == nil {
		return res
	} else if o.Config == nil {
		return res
	}

	o.Walk(func(key string, val interface{}) bool {
		res[key] = val
		return true
	})
	return res
}

func (o *fldModel) Map(fct func(key string, val interface{}) interface{}) Fields {
	if o == nil {
		return nil
	} else if o.Config == nil {
		return nil
	}

	o.Walk(func(key string, val interface{}) bool {
		o.Store(key, fct(key, val))
		return true
	})

	return o
}

func (o *fldModel) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Logrus())
}

func (o *fldModel) UnmarshalJSON(bytes []byte) error {
	var l = make(logrus.Fields)

	if e := json.Unmarshal(bytes, &l); e != nil {
		return e
	} else if len(l) > 0 {
		for k, v := range l {
			o.Store(k, v)
		}
	}

	return nil
}

func (o *fldModel) FieldsClone(ctx context.Context) Fields {
	if o == nil {
		return nil
	} else if o.Config == nil {
		return nil
	} else {
		return &fldModel{
			o.Config.Clone(ctx),
		}
	}
}
