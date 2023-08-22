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
	"sync/atomic"

	libkvd "github.com/nabbar/golib/database/kvdriver"
	libkvs "github.com/nabbar/golib/database/kvitem"
)

type tbl[K comparable, M any] struct {
	d *atomic.Value
}

func (o *tbl[K, M]) getDriver() libkvd.KVDriver[K, M] {
	if o == nil {
		return nil
	}

	i := o.d.Load()
	if i == nil {
		return nil
	} else if d, k := i.(libkvd.KVDriver[K, M]); !k {
		return nil
	} else {
		return d
	}
}

func (o *tbl[K, M]) Get(key K) (libkvs.KVItem[K, M], error) {
	var kvs = libkvs.New[K, M](key)

	if drv := o.getDriver(); drv == nil {
		return nil, ErrorBadDriver.Error(nil)
	} else {
		kvs.RegisterFctLoad(drv.Get)
		kvs.RegisterFctStore(drv.Set)
	}

	return kvs, kvs.Load()
}

func (o *tbl[K, M]) Walk(fct FuncWalk[K, M]) error {
	if drv := o.getDriver(); drv == nil {
		return ErrorBadDriver.Error(nil)
	} else {
		return drv.Walk(func(key K, model M) bool {
			var kvs = libkvs.New[K, M](key)

			kvs.RegisterFctStore(drv.Set)
			kvs.RegisterFctLoad(func(k K, m *M) error {
				*m = model
				return nil
			})
			_ = kvs.Load()
			kvs.RegisterFctLoad(drv.Get)

			return fct(kvs)
		})
	}
}

func (o *tbl[K, M]) List() ([]libkvs.KVItem[K, M], error) {
	var res = make([]libkvs.KVItem[K, M], 0)

	if drv := o.getDriver(); drv == nil {
		return nil, ErrorBadDriver.Error(nil)
	} else if l, e := drv.List(); e != nil {
		return nil, e
	} else {
		for _, k := range l {
			var kvs = libkvs.New[K, M](k)

			kvs.RegisterFctLoad(drv.Get)
			kvs.RegisterFctStore(drv.Set)

			res = append(res, kvs)
		}

		return res, nil
	}
}
