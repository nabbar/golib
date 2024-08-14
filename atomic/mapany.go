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
)

type ma[K comparable] struct {
	m sync.Map
}

func (o *ma[K]) Load(key K) (value any, ok bool) {
	return o.m.Load(key)
}

func (o *ma[K]) Store(key K, value any) {
	o.m.Store(key, value)
}

func (o *ma[K]) LoadOrStore(key K, value any) (actual any, loaded bool) {
	return o.m.LoadOrStore(key, value)
}

func (o *ma[K]) LoadAndDelete(key K) (value any, loaded bool) {
	return o.m.LoadAndDelete(key)
}

func (o *ma[K]) Delete(key K) {
	o.m.Delete(key)
}

func (o *ma[K]) Swap(key K, value any) (previous any, loaded bool) {
	return o.m.Swap(key, value)
}

func (o *ma[K]) CompareAndSwap(key K, old, new any) bool {
	return o.m.CompareAndSwap(key, old, new)
}

func (o *ma[K]) CompareAndDelete(key K, old any) (deleted bool) {
	return o.m.CompareAndDelete(key, old)
}

func (o *ma[K]) Range(f func(key K, value any) bool) {
	o.m.Range(func(key, value any) bool {
		var (
			l bool
			k K
		)

		if k, l = Cast[K](key); !l {
			o.m.Delete(key)
			return true
		}

		return f(k, value)
	})
}
