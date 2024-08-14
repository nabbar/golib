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
	"sync"
	"sync/atomic"
)

type Value[T any] interface {
	SetDefaultLoad(def T)
	SetDefaultStore(def T)

	Load() (val T)
	Store(val T)
	Swap(new T) (old T)
	CompareAndSwap(old, new T) (swapped bool)
}

type Map[K comparable] interface {
	Load(key K) (value any, ok bool)
	Store(key K, value any)

	LoadOrStore(key K, value any) (actual any, loaded bool)
	LoadAndDelete(key K) (value any, loaded bool)

	Delete(key K)
	Swap(key K, value any) (previous any, loaded bool)

	CompareAndSwap(key K, old, new any) bool
	CompareAndDelete(key K, old any) (deleted bool)

	Range(f func(key K, value any) bool)
}

type MapTyped[K comparable, V any] interface {
	Load(key K) (value V, ok bool)
	Store(key K, value V)

	LoadOrStore(key K, value V) (actual V, loaded bool)
	LoadAndDelete(key K) (value V, loaded bool)

	Delete(key K)
	Swap(key K, value V) (previous V, loaded bool)

	CompareAndSwap(key K, old, new V) bool
	CompareAndDelete(key K, old V) (deleted bool)

	Range(f func(key K, value V) bool)
}

func NewValue[T any]() Value[T] {
	var (
		tmp1 T
		tmp2 T
	)

	return NewValueDefault[T](tmp1, tmp2)
}

func NewValueDefault[T any](load, store T) Value[T] {
	o := &val[T]{
		av: new(atomic.Value),
		dl: new(atomic.Value),
		ds: new(atomic.Value),
	}

	o.SetDefaultLoad(load)
	o.SetDefaultStore(store)

	return o
}

func NewMapAny[K comparable]() Map[K] {
	return &ma[K]{
		m: sync.Map{},
	}
}

func NewMapTyped[K comparable, V any]() MapTyped[K, V] {
	return &mt[K, V]{
		m: NewMapAny[K](),
	}
}
