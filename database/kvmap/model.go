/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package kvmap

import (
	"encoding/json"

	libkvd "github.com/nabbar/golib/database/kvdriver"
)

func (o *Driver[K, MK, M]) serialize(model *M, modelMap *map[MK]any) error {
	if p, e := json.Marshal(model); e != nil {
		return e
	} else {
		return json.Unmarshal(p, modelMap)
	}
}

func (o *Driver[K, MK, M]) unSerialize(modelMap *map[MK]any, model *M) error {
	if p, e := json.Marshal(modelMap); e != nil {
		return e
	} else {
		return json.Unmarshal(p, model)
	}
}

func (o *Driver[K, MK, M]) Get(key K, model *M) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.FctGet == nil {
		return ErrorGetFunction.Error(nil)
	} else if m, e := o.FctGet(key); e != nil {
		return e
	} else {
		return o.unSerialize(&m, model)
	}
}

func (o *Driver[K, MK, M]) Set(key K, model M) error {
	var m = make(map[MK]any)

	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.FctSet == nil {
		return ErrorSetFunction.Error(nil)
	} else if e := o.serialize(&model, &m); e != nil {
		return e
	} else {
		return o.FctSet(key, m)
	}
}

func (o *Driver[K, MK, M]) List() ([]K, error) {
	if o == nil {
		return nil, ErrorBadInstance.Error(nil)
	} else if o.FctList == nil {
		return nil, ErrorListFunction.Error(nil)
	} else {
		return o.FctList()
	}
}

func (o *Driver[K, MK, M]) Walk(fct libkvd.FctWalk[K, M]) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if fct == nil {
		return ErrorFunctionParams.Error(nil)
	} else if l, e := o.List(); e != nil {
		return e
	} else {
		for _, k := range l {
			var m = *(new(M))

			if er := o.Get(k, &m); er != nil {
				return er
			}

			if !fct(k, m) {
				return nil
			}
		}
	}

	return nil
}
