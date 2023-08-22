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

import "sync/atomic"

type FuncLoad[K comparable, M any] func(key K, model *M) error
type FuncStore[K comparable, M any] func(key K, model M) error

type KVItem[K comparable, M any] interface {
	Set(model M)
	Get() M

	Load() error
	Store(force bool) error
	Clean()

	HasChange() bool

	RegisterFctLoad(fct FuncLoad[K, M])
	RegisterFctStore(fct FuncStore[K, M])
}

func New[K comparable, M any](key K) KVItem[K, M] {
	var (
		ml = new(atomic.Value)
		mw = new(atomic.Value)
	)

	ml.Store(nil)
	mw.Store(nil)

	return &itm[K, M]{
		k:  key,
		ml: ml,
		ms: mw,
		fl: nil,
		fs: nil,
	}

}
