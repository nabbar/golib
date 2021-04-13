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
		cfg:     new(atomic.Value),
	}
}

type configContext struct {
	context.Context
	cfg *atomic.Value
}

func (c configContext) getMap() map[string]*atomic.Value {
	var (
		v  interface{}
		s  map[string]*atomic.Value
		ok bool
	)

	if c.cfg == nil {
		c.cfg = new(atomic.Value)
	} else if v = c.cfg.Load(); v == nil {
		s = make(map[string]*atomic.Value)

	} else if s, ok = v.(map[string]*atomic.Value); !ok {
		s = make(map[string]*atomic.Value)
	}

	return s
}

func (c *configContext) Store(key string, cfg interface{}) {
	s := c.getMap()

	if _, ok := s[key]; !ok {
		s[key] = &atomic.Value{}
	}

	s[key].Store(cfg)
	c.cfg.Store(s)
}

func (c *configContext) Load(key string) interface{} {
	s := c.getMap()

	if _, ok := s[key]; !ok {
		return nil
	} else {
		return s[key].Load()
	}
}

func (c *configContext) Merge(cfg Config) bool {
	var (
		x  *configContext
		ok bool
	)

	if x, ok = cfg.(*configContext); !ok {
		return false
	}

	s := c.getMap()

	for k, v := range x.getMap() {
		if k == "" || v == nil {
			continue
		}

		if i := v.Load(); i == nil {
			continue
		} else {
			s[k] = &atomic.Value{}
			s[k].Store(i)
		}
	}

	c.cfg.Store(s)

	return true
}

//Deprecated: use Store.
func (c *configContext) ObjectStore(key string, obj interface{}) {
	c.Store(key, obj)
}

//Deprecated: use Load.
func (c *configContext) ObjectLoad(key string) interface{} {
	return c.Load(key)
}
