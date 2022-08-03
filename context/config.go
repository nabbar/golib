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
	"sync"
	"sync/atomic"
)

type Config interface {
	context.Context

	Merge(cfg Config) bool

	Store(key string, cfg interface{})
	Load(key string) interface{}
}

func NewConfig(ctx context.Context) Config {
	return &configContext{
		Context: ctx,
		cfg:     make(map[string]*atomic.Value, 0),
	}
}

type configContext struct {
	context.Context
	m   sync.Mutex
	cfg map[string]*atomic.Value
}

func (c configContext) Load(key string) interface{} {
	c.m.Lock()
	defer c.m.Unlock()

	var (
		v  *atomic.Value
		i  interface{}
		ok bool
	)

	if c.cfg == nil {
		c.cfg = make(map[string]*atomic.Value, 0)
	} else if v, ok = c.cfg[key]; !ok || v == nil {
		return nil
	} else if i = v.Load(); i != nil {
		return i
	}

	return nil
}

func (c configContext) Store(key string, cfg interface{}) {
	c.m.Lock()
	defer c.m.Unlock()

	var ok bool

	if c.cfg == nil {
		c.cfg = make(map[string]*atomic.Value, 0)
	}

	if _, ok = c.cfg[key]; !ok {
		c.cfg[key] = new(atomic.Value)
	}

	c.cfg[key].Store(cfg)
}

func (c *configContext) Merge(cfg Config) bool {
	var (
		x  *configContext
		ok bool
	)

	if x, ok = cfg.(*configContext); !ok {
		return false
	}

	x.m.Lock()
	defer x.m.Unlock()

	for k, v := range x.cfg {
		if k == "" || v == nil {
			continue
		}

		if i := v.Load(); i == nil {
			continue
		} else {
			c.Store(k, i)
		}
	}

	return true
}
