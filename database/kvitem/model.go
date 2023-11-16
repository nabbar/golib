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

package kvitem

import (
	"reflect"
	"sync/atomic"

	libkvt "github.com/nabbar/golib/database/kvtypes"
)

type itm[K comparable, M any] struct {
	d libkvt.KVDriver[K, M]
	k K // key

	ml *atomic.Value // model read
	ms *atomic.Value // model write

	fl *atomic.Value
	fs *atomic.Value
	fr *atomic.Value
}

func (o *itm[K, M]) getDriver() libkvt.KVDriver[K, M] {
	if o == nil {
		return nil
	}

	return o.d
}

func (o *itm[K, M]) Set(model M) {
	var val = reflect.ValueOf(model)
	o.ms.Store(val.Interface())
}

func (o *itm[K, M]) setModelLoad(mod M) {
	var val = reflect.ValueOf(mod)
	o.ml.Store(val.Interface())
}

func (o *itm[K, M]) getModelLoad() M {
	var mod M
	if i := o.ml.Load(); i == nil {
		return mod
	} else if v, k := i.(M); !k {
		return mod
	} else {
		return v
	}
}

func (o *itm[K, M]) getModelStore() M {
	var mod M
	if i := o.ms.Load(); i == nil {
		return mod
	} else if v, k := i.(M); !k {
		return mod
	} else {
		return v
	}
}

func (o *itm[K, M]) Key() K {
	if o == nil {
		var k K
		return k
	}

	return o.k
}

func (o *itm[K, M]) Get() M {
	var (
		tmp M
		mod M
	)

	if o == nil {
		return mod
	}

	mod = o.getModelStore()

	if reflect.DeepEqual(mod, tmp) {
		mod = o.getModelLoad()
	}

	// nothing load, so return new instance
	return mod
}

func (o *itm[K, M]) Load() error {
	var (
		mod M
		drv = o.getDriver()
	)

	if drv == nil {
		return ErrorLoadFunction.Error(nil)
	}

	if e := drv.Get(o.k, &mod); e == nil {
		o.setModelLoad(mod)
	} else {
		return e
	}

	return nil
}

func (o *itm[K, M]) Store(force bool) error {
	var drv = o.getDriver()

	if drv == nil {
		return ErrorStoreFunction.Error(nil)
	}

	var (
		lod M
		str M
	)

	_ = o.Load()

	str = o.getModelStore()
	if reflect.DeepEqual(lod, str) {
		str = o.getModelLoad()
	}

	lod = o.getModelLoad()
	if !reflect.DeepEqual(lod, str) {
		return drv.Set(o.k, str)
	} else if force {
		return drv.Set(o.k, lod)
	}

	return nil
}

func (o *itm[K, M]) Remove() error {
	drv := o.getDriver()

	if drv == nil {
		return ErrorStoreFunction.Error(nil)
	}

	return drv.Del(o.k)
}

func (o *itm[K, M]) Clean() {
	var tmp M
	o.setModelLoad(tmp)
	o.Set(tmp)
}

func (o *itm[K, M]) HasChange() bool {
	r := o.getModelLoad()
	w := o.getModelStore()

	return !reflect.DeepEqual(r, w)
}
