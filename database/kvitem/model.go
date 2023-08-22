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
)

type itm[K comparable, M any] struct {
	k K // key

	ml *atomic.Value // model read
	ms *atomic.Value // model write

	fl *atomic.Value
	fs *atomic.Value
}

func (o *itm[K, M]) RegisterFctLoad(fct FuncLoad[K, M]) {
	o.fl.Store(fct)
}

func (o *itm[K, M]) getFctLoad() FuncLoad[K, M] {
	if o == nil {
		return nil
	}

	i := o.fs.Load()
	if i == nil {
		return nil
	} else if f, k := i.(FuncLoad[K, M]); !k {
		return nil
	} else {
		return f
	}
}

func (o *itm[K, M]) RegisterFctStore(fct FuncStore[K, M]) {
	o.fs.Store(fct)
}

func (o *itm[K, M]) getFctStore() FuncStore[K, M] {
	if o == nil {
		return nil
	}

	i := o.fs.Load()
	if i == nil {
		return nil
	} else if f, k := i.(FuncStore[K, M]); !k {
		return nil
	} else {
		return f
	}
}

func (o *itm[K, M]) Set(model M) {
	if o == nil {
		return
	}

	m := o.ml.Load()

	// model not loaded, so store new model
	if m == nil {
		o.ms.Store(model)
		// model loaded and new model given not same, so store new model
	} else if !reflect.DeepEqual(m.(M), model) {
		o.ms.Store(model)
		// model loaded and given model are same, so don't store new model
	} else {
		o.ms.Store(nil)
	}
}

func (o *itm[K, M]) Get() M {
	if o == nil {
		return *(new(M))
	}

	// update exist so latest fresh value
	m := o.ms.Load()
	if m != nil {
		if v, k := m.(M); k {
			return v
		}
	}

	// load model exist so return last model load
	m = o.ml.Load()
	if m != nil {
		if v, k := m.(M); k {
			return v
		}
	}

	// nothing load, so return new instance
	return *(new(M))
}

func (o *itm[K, M]) Load() error {
	var fct FuncLoad[K, M]

	if fct = o.getFctLoad(); fct == nil {
		return ErrorLoadFunction.Error(nil)
	}

	m := *(new(M))
	e := fct(o.k, &m)

	if e == nil {
		o.ml.Store(m)
	}

	return e
}

func (o *itm[K, M]) Store(force bool) error {
	var fct FuncStore[K, M]

	if fct = o.getFctStore(); fct == nil {
		return ErrorStoreFunction.Error(nil)
	}

	m := o.ms.Load()
	if m != nil {
		return fct(o.k, m.(M))
	} else if !force {
		return nil
	}

	// no update, but force store, so use load model
	m = o.ml.Load()
	if m != nil {
		return fct(o.k, m.(M))
	}

	// no update and no load, but force store, so use new instance of model
	m = *(new(M))
	return fct(o.k, m.(M))
}

func (o *itm[K, M]) Clean() {
	o.ml.Store(nil)
	o.ms.Store(nil)
}

func (o *itm[K, M]) HasChange() bool {
	r := o.ml.Load()
	w := o.ms.Load()

	if r == nil && w == nil {
		// not loaded and not store, so no change
		return false
	} else if r == nil {
		// not loaded but store is set, so has been updated
		return true
	} else if w == nil {
		// loaded and not store, so no change
		return false
	}

	mr, kr := r.(M)
	mw, kw := w.(M)

	if !kr && !kw {
		// no valid model, so no change
		return false
	} else if !kr {
		// not valid model for load, but valid for store, so has been updated
		return true
	} else if !kw {
		// valid model for load, but not valid for store, so like no change
		return false
	}

	return !reflect.DeepEqual(mr, mw)
}
