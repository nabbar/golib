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

package kvmap

import (
	libkvt "github.com/nabbar/golib/database/kvtypes"
)

type FuncNew[K comparable, M any] func() libkvt.KVDriver[K, M]
type FuncGet[K comparable, MK comparable] func(key K) (map[MK]any, error)
type FuncSet[K comparable, MK comparable] func(key K, model map[MK]any) error
type FuncDel[K comparable] func(key K) error
type FuncList[K comparable] func() ([]K, error)
type FuncSearch[K comparable] func(prefix K) ([]K, error)
type FuncWalk[K comparable, M any] func(fct libkvt.FctWalk[K, M]) error

func New[K comparable, MK comparable, M any](cmp libkvt.Compare[K], fn FuncNew[K, M], fg FuncGet[K, MK], fs FuncSet[K, MK], fd FuncDel[K], fl FuncList[K], fh FuncSearch[K], fw FuncWalk[K, M]) libkvt.KVDriver[K, M] {
	return &drv[K, MK, M]{
		cmp: cmp, // compare instance

		fctNew: fn,
		fctGet: fg,
		fctSet: fs,
		fctDel: fd,
		fctLst: fl,

		fctSch: fh, // optional
		fctWlk: fw, // optional
	}
}
