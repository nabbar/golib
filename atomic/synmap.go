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

type mt[K comparable, V any] struct {
	m Map[K]
}

func (o *mt[K, V]) castBool(in any, chk bool) (value V, ok bool) {
	if v, k := Cast[V](in); k {
		return v, chk
	}

	return value, false
}

func (o *mt[K, V]) Load(key K) (value V, ok bool) {
	return o.castBool(o.m.Load(key))
}

func (o *mt[K, V]) Store(key K, value V) {
	o.m.Store(key, value)
}

func (o *mt[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	return o.castBool(o.m.LoadOrStore(key, value))
}

func (o *mt[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	return o.castBool(o.m.LoadAndDelete(key))
}

func (o *mt[K, V]) Delete(key K) {
	o.m.Delete(key)
}

func (o *mt[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	return o.castBool(o.m.Swap(key, value))
}

func (o *mt[K, V]) CompareAndSwap(key K, old, new V) bool {
	return o.m.CompareAndSwap(key, old, new)
}

func (o *mt[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return o.m.CompareAndDelete(key, old)
}

func (o *mt[K, V]) Range(f func(key K, value V) bool) {
	o.m.Range(func(key K, value any) bool {
		var (
			l bool
			v V
		)

		if v, l = Cast[V](value); !l {
			o.m.Delete(key)
			return true
		}

		return f(key, v)
	})
}
