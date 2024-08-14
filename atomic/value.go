/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package atomic

import (
	"sync/atomic"
)

type val[T any] struct {
	av *atomic.Value // atomic value of T
	dl *atomic.Value // default value for load
	ds *atomic.Value // default value for store
}

func (o *val[T]) SetDefaultLoad(def T) {
	o.dl.Store(newDefault[T](def))
}

func (o *val[T]) SetDefaultStore(def T) {
	o.ds.Store(newDefault[T](def))
}

func (o *val[T]) getDefault(i any) T {
	if v, k := Cast[defaultValue[T]](i); !k {
		var tmp T
		return tmp
	} else {
		return v.GetDefault()
	}
}

func (o *val[T]) getDefaultLoad() T {
	return o.getDefault(o.dl.Load())
}

func (o *val[T]) getDefaultStore() T {
	return o.getDefault(o.ds.Load())
}

func (o *val[T]) Load() (val T) {
	if v, k := Cast[T](o.av.Load()); !k {
		return o.getDefaultLoad()
	} else {
		return v
	}
}

func (o *val[T]) Store(val T) {
	if IsEmpty[T](val) {
		o.av.Store(o.getDefaultStore())
	} else {
		o.av.Store(val)
	}
}

func (o *val[T]) Swap(new T) (old T) {
	if IsEmpty[T](new) {
		new = o.getDefaultStore()
	}

	if v, k := Cast[T](o.av.Swap(new)); !k {
		return o.getDefaultLoad()
	} else {
		return v
	}
}

func (o *val[T]) CompareAndSwap(old, new T) (swapped bool) {
	if IsEmpty[T](old) {
		old = o.getDefaultStore()
	}

	if IsEmpty[T](new) {
		new = o.getDefaultStore()
	}

	return o.av.CompareAndSwap(old, new)
}
