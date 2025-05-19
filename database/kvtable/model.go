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

package kvtable

import (
	libkvs "github.com/nabbar/golib/database/kvitem"
	libkvt "github.com/nabbar/golib/database/kvtypes"
)

type tbl[K comparable, M any] struct {
	d libkvt.KVDriver[K, M]
}

func (o *tbl[K, M]) Get(key K) (libkvt.KVItem[K, M], error) {
	if drv := o.getDriver(); drv == nil {
		return nil, ErrorBadDriver.Error(nil)
	} else {
		var kvi = libkvs.New[K, M](drv.New(), key)
		e := kvi.Load()
		return kvi, e
	}
}

func (o *tbl[K, M]) Del(key K) error {
	if drv := o.getDriver(); drv == nil {
		return ErrorBadDriver.Error(nil)
	} else {
		return drv.Del(key)
	}
}

func (o *tbl[K, M]) List() ([]libkvt.KVItem[K, M], error) {
	var res = make([]libkvt.KVItem[K, M], 0)

	if drv := o.getDriver(); drv == nil {
		return nil, ErrorBadDriver.Error(nil)
	} else if l, e := drv.List(); e != nil {
		return nil, e
	} else {
		for _, k := range l {
			res = append(res, libkvs.New[K, M](drv.New(), k))
		}

		return res, nil
	}
}

func (o *tbl[K, M]) Search(pattern K) ([]libkvt.KVItem[K, M], error) {
	var res = make([]libkvt.KVItem[K, M], 0)

	if drv := o.getDriver(); drv == nil {
		return nil, ErrorBadDriver.Error(nil)
	} else if l, e := drv.Search(pattern); e != nil {
		return nil, e
	} else {
		for _, k := range l {
			res = append(res, libkvs.New[K, M](drv.New(), k))
		}

		return res, nil
	}
}

func (o *tbl[K, M]) Walk(fct libkvt.FuncWalk[K, M]) error {
	if drv := o.getDriver(); drv == nil {
		return ErrorBadDriver.Error(nil)
	} else {
		return drv.Walk(func(key K, model M) bool {
			kvi := libkvs.New[K, M](drv.New(), key)
			kvi.Set(model)
			return fct(kvi)
		})
	}
}

func (o *tbl[K, M]) getDriver() libkvt.KVDriver[K, M] {
	if o == nil {
		return nil
	}

	if o.d == nil {
		return nil
	} else {
		return o.d
	}
}
