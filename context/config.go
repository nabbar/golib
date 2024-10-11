/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context

import (
	"context"
	"slices"
	"sync"
	"time"
)

type FuncContext func() context.Context
type FuncContextConfig[T comparable] func() Config[T]
type FuncWalk[T comparable] func(key T, val interface{}) bool

type MapManage[T comparable] interface {
	Clean()
	Load(key T) (val interface{}, ok bool)
	Store(key T, cfg interface{})
	Delete(key T)
}

type Context interface {
	GetContext() context.Context
}

type Config[T comparable] interface {
	context.Context
	MapManage[T]
	Context

	SetContext(ctx FuncContext)
	Clone(ctx context.Context) Config[T]
	Merge(cfg Config[T]) bool
	Walk(fct FuncWalk[T]) bool
	WalkLimit(fct FuncWalk[T], validKeys ...T) bool

	LoadOrStore(key T, cfg interface{}) (val interface{}, loaded bool)
	LoadAndDelete(key T) (val interface{}, loaded bool)
}

func NewConfig[T comparable](ctx FuncContext) Config[T] {
	if ctx == nil {
		ctx = context.Background
	}

	return &configContext[T]{
		Context: ctx(),
		n:       sync.RWMutex{},
		m:       sync.Map{},
		x:       ctx,
	}
}

type configContext[T comparable] struct {
	context.Context
	n sync.RWMutex
	m sync.Map
	x FuncContext
}

func (c *configContext[T]) SetContext(ctx FuncContext) {
	c.n.Lock()
	defer c.n.Unlock()

	if ctx == nil {
		ctx = context.Background
	}

	c.Context = ctx()
	c.x = ctx
}

func (c *configContext[T]) Delete(key T) {
	if c.Err() != nil {
		c.Clean()
		return
	}

	c.n.RLock()
	defer c.n.RUnlock()

	c.m.Delete(key)
}

func (c *configContext[T]) Load(key T) (val interface{}, ok bool) {
	c.n.RLock()
	defer c.n.RUnlock()

	return c.m.Load(key)
}

func (c *configContext[T]) Store(key T, cfg interface{}) {
	if c.Err() != nil {
		c.Clean()
		return
	}

	c.n.RLock()
	defer c.n.RUnlock()

	c.m.Store(key, cfg)
}

func (c *configContext[T]) LoadOrStore(key T, cfg interface{}) (val interface{}, loaded bool) {
	if c.Err() != nil {
		c.Clean()
		return nil, false
	}

	c.n.RLock()
	defer c.n.RUnlock()

	return c.m.LoadOrStore(key, cfg)
}

func (c *configContext[T]) LoadAndDelete(key T) (val interface{}, loaded bool) {
	c.n.RLock()
	defer c.n.RUnlock()

	return c.m.LoadAndDelete(key)
}

func (c *configContext[T]) Walk(fct FuncWalk[T]) bool {
	return c.WalkLimit(fct)
}

func (c *configContext[T]) WalkLimit(fct FuncWalk[T], validKeys ...T) bool {
	c.n.RLock()
	defer c.n.RUnlock()

	c.m.Range(func(key, value any) bool {
		if i, k := key.(T); !k {
			return true
		} else if len(validKeys) < 1 {
			return fct(i, value)
		} else if slices.Contains(validKeys, i) {
			return fct(i, value)
		}
		return true
	})

	return true
}

func (c *configContext[T]) Merge(cfg Config[T]) bool {
	if c.Err() != nil {
		c.Clean()
		return false
	} else if cfg == nil {
		return false
	}

	c.n.RLock()
	defer c.n.RUnlock()

	cfg.Walk(func(key T, val interface{}) bool {
		c.m.Store(key, val)
		return true
	})

	return true
}

func (c *configContext[T]) Clean() {
	c.n.Lock()
	defer c.n.Unlock()

	c.m = sync.Map{}
}

func (c *configContext[T]) Clone(ctx context.Context) Config[T] {
	if c.Err() != nil {
		c.Clean()
		return nil
	}

	c.n.RLock()
	defer c.n.RUnlock()

	if ctx == nil {
		ctx = c.x()
	}

	if ctx == nil {
		ctx = c.Context
	}

	n := &configContext[T]{
		Context: ctx,
		n:       sync.RWMutex{},
		m:       sync.Map{},
		x:       c.x,
	}

	c.m.Range(func(key any, val interface{}) bool {
		if i, k := key.(T); k {
			n.Store(i, val)
		}
		return true
	})

	return n
}

func (c *configContext[T]) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *configContext[T]) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *configContext[T]) Err() error {
	return c.Context.Err()
}

func (c *configContext[T]) Value(key any) any {
	if i, k := key.(T); !k {
		return c.Context.Value(key)
	} else if v, ok := c.Load(i); ok {
		return v
	} else {
		return c.Context.Value(key)
	}
}

func (c *configContext[T]) GetContext() context.Context {
	c.n.RLock()
	defer c.n.RUnlock()

	if c.x == nil {
		return context.Background()
	} else if x := c.x(); x == nil {
		return context.Background()
	} else {
		return x
	}
}
