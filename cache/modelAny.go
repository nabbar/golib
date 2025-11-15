/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package cache

import cchitm "github.com/nabbar/golib/cache/item"

// Close cancels the cache context and removes all items from the cache.
// It implements the io.Closer interface.
func (o *cc[K, V]) Close() error {
	if o.n != nil {
		o.n()
	}

	o.Clean()
	return nil
}

// Clean removes all items from the cache, regardless of their expiration status.
// This is useful for clearing the cache completely.
func (o *cc[K, V]) Clean() {
	o.v.Range(func(key K, v cchitm.CacheItem[V]) bool {
		if val, ok := o.v.LoadAndDelete(key); ok {
			val.Clean()
		}

		return true
	})
}

// Expire removes all expired items from the cache.
// This method can be called periodically to free memory used by expired items.
func (o *cc[K, V]) Expire() {
	o.v.Range(func(key K, val cchitm.CacheItem[V]) bool {
		if !val.Check() {
			o.v.Delete(key)
		}
		return true
	})
}
