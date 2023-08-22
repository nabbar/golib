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

package kvdriver

type FctWalk[K comparable, M any] func(key K, model M) bool

type KVDriver[K comparable, M any] interface {
	Get(key K, model *M) error
	Set(key K, model M) error
	List() ([]K, error)
	Walk(fct FctWalk[K, M]) error
}

type FuncGet[K comparable, M any] func(key K) (M, error)
type FuncSet[K comparable, M any] func(key K, model M) error
type FuncList[K comparable, M any] func() ([]K, error)
type FuncWalk[K comparable, M any] func(fct FctWalk[K, M]) error

type Driver[K comparable, M any] struct {
	KVDriver[K, M]

	FctGet  FuncGet[K, M]
	FctSet  FuncSet[K, M]
	FctList FuncList[K, M]
	FctWalk FuncWalk[K, M] // optional
}
