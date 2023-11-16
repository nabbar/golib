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
type FuncList[K comparable, MK comparable] func() ([]K, error)

type drv[K comparable, MK comparable, M any] struct {
	FctNew  FuncNew[K, M]
	FctGet  FuncGet[K, MK]
	FctSet  FuncSet[K, MK]
	FctDel  FuncDel[K]
	FctList FuncList[K, MK]
}

func New[K comparable, MK comparable, M any](fn FuncNew[K, M], fg FuncGet[K, MK], fs FuncSet[K, MK], fd FuncDel[K], fl FuncList[K, MK]) libkvt.KVDriver[K, M] {
	return &drv[K, MK, M]{
		FctNew:  fn,
		FctGet:  fg,
		FctSet:  fs,
		FctDel:  fd,
		FctList: fl,
	}
}
