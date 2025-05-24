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

	libkvt "github.com/nabbar/golib/database/kvtypes"
)

type drv[K comparable, MK comparable, M any] struct {
	cmp libkvt.Compare[K]

	fctNew FuncNew[K, M]
	fctGet FuncGet[K, MK]
	fctSet FuncSet[K, MK]
	fctDel FuncDel[K]
	fctLst FuncList[K]

	fctSch FuncSearch[K]  // optional
	fctWlk FuncWalk[K, M] // optional
}

func (o *drv[K, MK, M]) serialize(model *M, modelMap *map[MK]any) error {
	if p, e := json.Marshal(model); e != nil {
		return e
	} else {
		return json.Unmarshal(p, modelMap)
	}
}

func (o *drv[K, MK, M]) unSerialize(modelMap *map[MK]any, model *M) error {
	if p, e := json.Marshal(modelMap); e != nil {
		return e
	} else {
		return json.Unmarshal(p, model)
	}
}

func (o *drv[K, MK, M]) New() libkvt.KVDriver[K, M] {
	return &drv[K, MK, M]{
		fctGet: o.fctGet,
		fctSet: o.fctSet,
		fctDel: o.fctDel,
		fctLst: o.fctLst,
		fctSch: o.fctSch,
		fctWlk: o.fctWlk,
	}
}

func (o *drv[K, MK, M]) Get(key K, model *M) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.fctGet == nil {
		return ErrorGetFunction.Error(nil)
	} else if m, e := o.fctGet(key); e != nil {
		return e
	} else {
		return o.unSerialize(&m, model)
	}
}

func (o *drv[K, MK, M]) Set(key K, model M) error {
	var m = make(map[MK]any)

	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.fctSet == nil {
		return ErrorSetFunction.Error(nil)
	} else if e := o.serialize(&model, &m); e != nil {
		return e
	} else {
		return o.fctSet(key, m)
	}
}

func (o *drv[K, MK, M]) Del(key K) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.fctDel == nil {
		return ErrorDelFunction.Error(nil)
	} else {
		return o.fctDel(key)
	}
}

func (o *drv[K, MK, M]) List() ([]K, error) {
	if o == nil {
		return nil, ErrorBadInstance.Error(nil)
	} else if o.fctLst == nil {
		return nil, ErrorListFunction.Error(nil)
	} else {
		return o.fctLst()
	}
}

func (o *drv[K, MK, M]) Search(pattern K) ([]K, error) {
	if o == nil {
		return nil, ErrorBadInstance.Error(nil)
	} else if o.cmp == nil {
		return nil, ErrorBadInstance.Error(nil)
	} else if o.cmp.IsEmpty(pattern) {
		return o.List()
	} else if o.fctSch != nil {
		return o.fctSch(pattern)
	} else {
		return o.fakeSrch(pattern)
	}
}

func (o *drv[K, MK, M]) Walk(fct libkvt.FctWalk[K, M]) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if fct == nil {
		return ErrorFunctionParams.Error(nil)
	} else if o.fctWlk == nil {
		return o.fakeWalk(fct)
	} else {
		return o.fctWlk(fct)
	}
}

func (o *drv[K, MK, M]) fakeSrch(pattern K) ([]K, error) {
	if l, e := o.List(); e != nil {
		return nil, e
	} else {
		var res = make([]K, 0)
		for _, k := range l {
			if o.cmp.IsContains(k, pattern) {
				res = append(res, k)
			}
		}
		return res, nil
	}
}

func (o *drv[K, MK, M]) fakeWalk(fct libkvt.FctWalk[K, M]) error {
	if l, e := o.List(); e != nil {
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
