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
	libkvi "github.com/nabbar/golib/database/kvitem"
)

type FuncWalk[K comparable, M any] func(kv libkvi.KVItem[K, M]) bool

type KVTable[K comparable, M any] interface {
	Get(key K) (libkvi.KVItem[K, M], error)
	List() ([]libkvi.KVItem[K, M], error)
	Walk(fct FuncWalk[K, M]) error
}

func New[K comparable, M any](drv libkvd.KVDriver[K, M]) KVTable[K, M] {
	d := new(atomic.Value)
	d.Store(drv)

	return &tbl[K, M]{
		d: d,
	}
}
