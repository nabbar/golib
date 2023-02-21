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

import (
	"context"
	"sync"
	"time"
)

type FuncCache[T any] func() Cache[T]

type Cache[T any] interface {
	Clone(ctx context.Context, exp time.Duration) Cache[T]
	Close()
	Clean()

	Merge(c Cache[T])
	Walk(fct func(key any, val interface{}, exp time.Duration) bool)

	Load(key any) (val interface{}, exp time.Duration, ok bool)
	Store(key any, val interface{}) time.Duration
	Delete(key any)

	LoadOrStore(key any, val interface{}) (res interface{}, exp time.Duration, loaded bool)
	LoadAndDelete(key any) (val interface{}, loaded bool)
}

func New[T any](ctx context.Context, exp time.Duration) Cache[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	if exp < time.Microsecond {
		return nil
	}

	n := &cache[T]{
		Context: ctx,
		w:       sync.RWMutex{},
		m:       sync.Map{},
		c:       make(chan struct{}),
		e:       exp,
	}

	go n.ticker(exp)

	return n
}
