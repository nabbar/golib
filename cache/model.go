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

type cache[T any] struct {
	context.Context

	w sync.RWMutex
	m sync.Map
	c chan struct{}
	e time.Duration
}

func (t *cache[T]) ticker(exp time.Duration) {
	ticker := time.NewTicker(exp)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.expire()

		case <-t.Done():
			t.Clean()
			return

		case <-t.c:
			t.Clean()
			return
		}
	}
}

func (t *cache[T]) expire() {
	t.w.RLock()
	defer t.w.RUnlock()

	exp := t.e

	t.m.Range(func(key, value any) bool {
		if v, _ := parse(value, exp); v == nil {
			t.m.Delete(key)
		}
		return true
	})
}

func (t *cache[T]) Clone(ctx context.Context, exp time.Duration) Cache[T] {
	t.w.RLock()
	defer t.w.RUnlock()

	if ctx == nil {
		ctx = t.Context
	}

	if exp < time.Microsecond {
		exp = t.e
	}

	n := &cache[T]{
		Context: ctx,
		w:       sync.RWMutex{},
		m:       sync.Map{},
		c:       make(chan struct{}),
		e:       exp,
	}

	t.m.Range(func(key any, val interface{}) bool {
		n.Store(key, val)
		return true
	})

	go n.ticker(exp)

	return n
}

func (t *cache[T]) Close() {
	t.c <- struct{}{}
}

func (t *cache[T]) Clean() {
	t.w.Lock()
	defer t.w.Unlock()

	t.m = sync.Map{}
}

func (t *cache[T]) Merge(c Cache[T]) {
	t.w.RLock()
	defer t.w.RUnlock()

	c.Walk(func(key any, val interface{}, exp time.Duration) bool {
		t.m.Store(key, &cacheItem{
			t: time.Now().Add(-exp),
			v: nil,
		})
		return true
	})
}

func (t *cache[T]) Walk(fct func(key any, val interface{}, exp time.Duration) bool) {
	t.w.RLock()
	defer t.w.RUnlock()

	exp := t.e

	t.m.Range(func(key, value any) bool {
		if v, e := parse(value, exp); v == nil {
			t.m.Delete(key)
			return true
		} else {
			return fct(key, v, e)
		}
	})
}

func (t *cache[T]) Load(key any) (val interface{}, exp time.Duration, ok bool) {
	t.w.RLock()
	defer t.w.RUnlock()

	var o any

	if o, ok = t.m.Load(key); !ok {
		return nil, 0, false
	} else if val, exp = parse(o, t.e); val == nil {
		t.m.Delete(key)
		return nil, 0, false
	} else {
		return val, exp, true
	}
}

func (t *cache[T]) Store(key any, val interface{}) time.Duration {
	i := store(val)
	e := time.Now()

	t.w.RLock()
	defer t.w.RUnlock()

	t.m.Store(key, i)
	return t.e - time.Since(e)
}

func (t *cache[T]) Delete(key any) {
	t.m.LoadAndDelete(key)
}

func (t *cache[T]) LoadOrStore(key any, val interface{}) (res interface{}, exp time.Duration, loaded bool) {
	if res, exp, loaded = t.Load(key); !loaded {
		exp = t.Store(key, val)
		return val, exp, false
	}

	return res, exp, loaded
}

func (t *cache[T]) LoadAndDelete(key any) (val interface{}, loaded bool) {
	if val, _, loaded = t.Load(key); !loaded {
		return nil, false
	}

	t.w.RLock()
	defer t.w.RUnlock()

	t.m.Delete(key)
	return val, loaded
}

func (t *cache[T]) Deadline() (deadline time.Time, ok bool) {
	return t.Context.Deadline()
}

func (t *cache[T]) Done() <-chan struct{} {
	return t.Context.Done()
}

func (t *cache[T]) Err() error {
	return t.Context.Err()
}

func (t *cache[T]) Value(key any) any {
	if v, _, ok := t.Load(key); ok {
		return v
	} else {
		return t.Context.Value(key)
	}
}
