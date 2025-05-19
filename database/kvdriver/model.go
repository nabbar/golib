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

package kvdriver

import (
	libkvt "github.com/nabbar/golib/database/kvtypes"
)

func (o *drv[K, M]) New() libkvt.KVDriver[K, M] {
	if o.fctNew != nil {
		return o.fctNew()
	}

	return &drv[K, M]{
		cmp:    o.cmp,
		fctNew: o.fctNew,
		fctGet: o.fctGet,
		fctSet: o.fctSet,
		fctDel: o.fctDel,
		fctLst: o.fctLst,
		fctSch: o.fctSch,
		fctWlk: o.fctWlk,
	}
}

func (o *drv[K, M]) Get(key K, model *M) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.fctGet == nil {
		return ErrorGetFunction.Error(nil)
	} else {
		m, e := o.fctGet(key)
		*model = m
		return e
	}
}

func (o *drv[K, M]) Set(key K, model M) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.fctSet == nil {
		return ErrorSetFunction.Error(nil)
	} else {
		return o.fctSet(key, model)
	}
}

func (o *drv[K, M]) Del(key K) error {
	if o == nil {
		return ErrorBadInstance.Error(nil)
	} else if o.fctDel == nil {
		return ErrorSetFunction.Error(nil)
	} else {
		return o.fctDel(key)
	}
}

func (o *drv[K, M]) List() ([]K, error) {
	if o == nil {
		return nil, ErrorBadInstance.Error(nil)
	} else if o.fctLst == nil {
		return nil, ErrorListFunction.Error(nil)
	} else {
		return o.fctLst()
	}
}

func (o *drv[K, M]) Search(pattern K) ([]K, error) {
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

func (o *drv[K, M]) Walk(fct libkvt.FctWalk[K, M]) error {
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

func (o *drv[K, M]) fakeSrch(pattern K) ([]K, error) {
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

func (o *drv[K, M]) fakeWalk(fct libkvt.FctWalk[K, M]) error {
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
